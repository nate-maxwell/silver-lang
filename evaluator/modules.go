package evaluator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"silver/ast"
	"silver/astcache"
	"silver/lexer"
	"silver/object"
	"silver/packages"
	"silver/parser"
	"strings"
)

const importPathEnvironment = "SILVER_PATH"

// EvalFile parses and evaluates path in env. It also sets env's source
// directory so relative imports resolve beside the entry file.
func (e *Evaluator) EvalFile(path string, env *object.Environment) object.Object {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return newError(object.RuntimeErrorKindValue, "could not resolve file %q: %s", path, err)
	}
	absolutePath = filepath.Clean(absolutePath)
	if err := e.refreshPackageIndex(); err != nil {
		return newError(object.RuntimeErrorKindImport, "could not load SILVER_PATH: %s", err)
	}
	manifest := e.packages.ManifestFor(absolutePath)

	program, parseError := e.parseFile(absolutePath, manifest)
	if parseError != nil {
		return parseError
	}

	env.SetSourceDir(filepath.Dir(absolutePath))
	if manifest != nil {
		env.SetPackageID(manifest.ID())
	}
	return e.Eval(program, env)
}

// importModule first resolves bundled standard-library names, including
// embedded Silver implementations, then loads user file modules in an isolated
// top-level environment. Successful modules are cached by standard-library
// name or canonical absolute path.
func (e *Evaluator) importModule(path string, env *object.Environment) object.Object {
	if module, ok := e.standardLibrary.LookupModule(path); ok {
		if cached, exists := e.modules[path]; exists {
			return cached
		}
		e.modules[path] = module
		return module
	}
	if source, sourceName, cache, ok := e.standardLibrary.LookupSourceModule(path); ok {
		return e.importSourceModule(path, sourceName, source, cache)
	}

	absolutePath, manifest, err := e.resolveImportPath(path, env.SourceDir())
	if err != nil {
		return newError(object.RuntimeErrorKindImport, "could not resolve import %q: %s", path, err)
	}

	if module, ok := e.modules[absolutePath]; ok {
		return module
	}
	if e.loading[absolutePath] {
		return newError(object.RuntimeErrorKindImport, "circular import detected while loading %q", absolutePath)
	}

	e.loading[absolutePath] = true
	defer delete(e.loading, absolutePath)

	program, parseError := e.parseFile(absolutePath, manifest)
	if parseError != nil {
		return parseError
	}

	moduleEnv := object.NewEnvironment()
	moduleEnv.SetSourceDir(filepath.Dir(absolutePath))
	if manifest != nil {
		moduleEnv.SetPackageID(manifest.ID())
	}
	e.pushContext("<module>")
	defer e.popContext()
	result := e.Eval(program, moduleEnv)
	if isError(result) {
		return result
	}

	exports, exportError := e.moduleExports(program, moduleEnv)
	if exportError != nil {
		return exportError
	}
	module := &object.Module{Path: absolutePath, Exports: exports}
	e.modules[absolutePath] = module
	return module
}

// importSourceModule evaluates one embedded Silver standard-library module.
// It deliberately uses the same isolated environment, cache, and circular
// import protection as file modules, while retaining its bare import name as
// the module identity.
func (e *Evaluator) importSourceModule(name, sourceName, source string, cache []byte) object.Object {
	if module, ok := e.modules[name]; ok {
		return module
	}
	if e.loading[name] {
		return newError(object.RuntimeErrorKindImport, "circular import detected while loading %q", name)
	}

	e.loading[name] = true
	defer delete(e.loading, name)

	sourceBytes := []byte(source)
	program, cached := astcache.LoadBytes(sourceName, sourceBytes, cache)
	packageID := "stdlib"
	registry := e.operatorScope(packageID).registry
	if registry.HasUserOperators() || bytes.Contains(sourceBytes, []byte("operator")) {
		cached = false
	}
	if !cached {
		var parseError *object.Error
		program, parseError = parseSourceWithRegistry(sourceName, sourceBytes, registry)
		if parseError != nil {
			return parseError
		}
	}

	moduleEnv := object.NewEnvironment()
	moduleEnv.SetPackageID(packageID)
	e.pushContext("<module>")
	defer e.popContext()
	result := e.Eval(program, moduleEnv)
	if isError(result) {
		return result
	}

	exports, exportError := e.moduleExports(program, moduleEnv)
	if exportError != nil {
		return exportError
	}
	module := &object.Module{Path: name, Exports: exports}
	e.modules[name] = module
	return module
}

// moduleExports returns every top-level binding unless the program contains
// an export declaration, in which case only its listed names are exposed.
func (e *Evaluator) moduleExports(program *ast.Program, env *object.Environment) (map[string]object.Object, *object.Error) {
	bindings := env.Bindings()
	var declaration *ast.ExportStatement
	for _, statement := range program.Statements {
		if export, ok := statement.(*ast.ExportStatement); ok {
			declaration = export
			break
		}
	}
	if declaration == nil {
		return bindings, nil
	}

	exports := make(map[string]object.Object, len(declaration.Names))
	for _, name := range declaration.Names {
		value, ok := bindings[name.Value]
		if !ok {
			failure := newError(object.RuntimeErrorKindName, "exported symbol %q is not defined", name.Value)
			failure.SetOrigin(e.traceFrame(name))
			return nil, failure
		}
		exports[name.Value] = value
	}
	return exports, nil
}

// resolveImportPath checks the importer directory first, then package
// manifests, explicit source files, and legacy directory entries in
// SILVER_PATH.
func (e *Evaluator) resolveImportPath(path, sourceDir string) (string, *packages.Manifest, error) {
	if err := e.refreshPackageIndex(); err != nil {
		return "", nil, fmt.Errorf("could not load SILVER_PATH: %w", err)
	}
	if filepath.IsAbs(path) {
		absolute := filepath.Clean(path)
		return absolute, e.packages.ManifestFor(absolute), nil
	}
	if sourceDir == "" {
		var err error
		sourceDir, err = os.Getwd()
		if err != nil {
			return "", nil, err
		}
	}
	localPath, err := filepath.Abs(filepath.Join(sourceDir, path))
	if err != nil {
		return "", nil, err
	}
	localPath = filepath.Clean(localPath)
	if importCandidateExists(localPath) {
		return localPath, e.packages.ManifestFor(localPath), nil
	}
	if exposed, manifest, ok, err := e.packages.Resolve(path); err != nil {
		return "", nil, err
	} else if ok {
		return exposed, manifest, nil
	}
	return localPath, nil, nil
}

// importCandidateExists treats errors other than non-existence as a match so
// parseFile can report the underlying permission or file-type error instead of
// silently continuing to a different module with the same name.
func importCandidateExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

// parseFile reads a source file and parses it with its absolute path attached
// to every token for diagnostics and tracebacks.
func (e *Evaluator) parseFile(path string, manifest *packages.Manifest) (*ast.Program, *object.Error) {
	if manifest != nil {
		state, parseError := e.preparePackage(manifest)
		if parseError != nil {
			return nil, parseError
		}
		if program := state.programs[packagePathKey(path)]; program != nil {
			return program, nil
		}
	}
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, newError(object.RuntimeErrorKindImport, "could not read %q: %s", path, err)
	}
	registry := e.operatorScope("").registry
	if program, ok := astcache.Load(path, input); ok && !registry.HasUserOperators() && !bytes.Contains(input, []byte("operator")) {
		return program, nil
	}
	program, parseError := parseSourceWithRegistry(path, input, registry)
	if parseError != nil {
		return nil, parseError
	}
	// A cache is an optimization only. Read-only directories and other cache
	// write failures must not prevent valid source from running.
	if !registry.HasUserOperators() {
		_ = astcache.Store(path, input, program)
	}
	return program, nil
}

// ParseSource parses and optimizes source with a diagnostic name that need not
// refer to a filesystem path. Embedded standard-library cache generation uses
// the same pipeline as ordinary file parsing through this entry point.
func ParseSource(sourceName string, input []byte) (*ast.Program, *object.Error) {
	p := parser.New(lexer.NewWithSource(string(input), sourceName))
	return finishParse(sourceName, p)
}

func parseSourceWithRegistry(sourceName string, input []byte, registry *parser.InfixRegistry) (*ast.Program, *object.Error) {
	p := parser.NewWithInfixRegistry(lexer.NewWithSource(string(input), sourceName), registry)
	return finishParse(sourceName, p)
}

func finishParse(sourceName string, p *parser.Parser) (*ast.Program, *object.Error) {
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return nil, newError(object.RuntimeErrorKindSyntax, "could not parse %q:\n%s", sourceName, strings.Join(p.Errors(), "\n"))
	}
	program = foldConstants(program)
	return program, nil
}
