package evaluator

import (
	"os"
	"path"
	"path/filepath"
	"silver/ast"
	"silver/lexer"
	"silver/object"
	"silver/packages"
	"silver/parser"
	"silver/source"
	"silver/stdlib"
	"strings"
)

const importPathEnvironment = "SILVER_PATH"

// EvalFile parses and evaluates path in env. It also sets env's source
// directory and discovers its package operator scope. Unlike an import, this
// executes in the caller's environment on every call; moduleStore caches only
// imported module values. Package parsing may still reuse a prepared AST.
func (e *Evaluator) EvalFile(path string, env *object.Environment) object.Object {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return newError(object.RuntimeErrorKindValue, "could not resolve file %q: %s", path, err)
	}
	absolutePath = filepath.Clean(absolutePath)
	if err := e.refreshPackageIndex(); err != nil {
		return newError(object.RuntimeErrorKindImport, "could not load SILVER_PATH: %s", err)
	}
	manifest, err := e.packages.DiscoverFor(absolutePath)
	if err != nil {
		return newError(object.RuntimeErrorKindImport, "could not load package for %q: %s", absolutePath, err)
	}

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

// importModule loads a package-qualified module. Files are reachable only through
// their registered manifest, and internal members only from their owning package.
func (e *Evaluator) importModule(request string, env *object.Environment) object.Object {
	packageName, _, err := packages.ParseImport(request)
	if err != nil {
		return newError(object.RuntimeErrorKindImport, "%s", err)
	}
	if packageName == "core" {
		if module, ok := e.standardLibrary.LookupModule(request); ok {
			return e.modules.load(module.ID, module.Path, func() (*object.Module, *object.Error) { return module, nil })
		}
		if module, ok := e.standardLibrary.LookupSourceFrom(request, env.PackageID()); ok {
			return e.importSourceModule(module)
		}
		return newError(object.RuntimeErrorKindImport, "standard-library module %q does not exist", request)
	}
	if err := e.refreshPackageIndex(); err != nil {
		return newError(object.RuntimeErrorKindImport, "could not load SILVER_PATH: %s", err)
	}
	absolutePath, manifest, found, err := e.packages.ResolveFrom(request, env.PackageID())
	if err != nil {
		return newError(object.RuntimeErrorKindImport, "could not resolve import %q: %s", request, err)
	}
	if !found {
		return newError(object.RuntimeErrorKindImport, "package module %q not found; register its package.yaml in SILVER_PATH and list the module in export", request)
	}
	// Validate the owning manifest before returning a cached module.
	if info, err := os.Stat(manifest.Path()); err != nil {
		return newError(object.RuntimeErrorKindImport, "cannot import %q: required package YAML manifest %q is unavailable: %s", request, manifest.Path(), err)
	} else if !info.Mode().IsRegular() {
		return newError(object.RuntimeErrorKindImport, "cannot import %q: required package YAML manifest %q is not a regular file", request, manifest.Path())
	}
	id := source.FileID(absolutePath)
	return e.modules.load(id, absolutePath, func() (*object.Module, *object.Error) {
		return e.evaluateFileModule(id, absolutePath, manifest)
	})
}

// evaluateFileModule executes an imported file in an isolated top-level scope.
// Its program drains module defers before exports are collected. The caller
// publishes the resulting Module in moduleStore only after both steps succeed.
func (e *Evaluator) evaluateFileModule(id source.ModuleID, absolutePath string, manifest *packages.Manifest) (*object.Module, *object.Error) {
	program, parseError := e.parseFile(absolutePath, manifest)
	if parseError != nil {
		return nil, parseError
	}

	moduleEnv := object.NewEnvironment()
	moduleEnv.SetSourceDir(filepath.Dir(absolutePath))
	moduleEnv.SetPackageID(manifest.ID())
	e.pushContext("<module>")
	defer e.popContext()
	result := e.Eval(program, moduleEnv)
	if failure, ok := result.(*object.Error); ok {
		return nil, failure
	}

	exports, exportError := e.moduleExports(program, moduleEnv)
	if exportError != nil {
		return nil, exportError
	}
	return &object.Module{ID: id, Path: absolutePath, Exports: exports}, nil
}

// importSourceModule evaluates one embedded Silver standard-library module.
// It deliberately uses the same isolated environment, cache, and circular
// import protection as file modules, while retaining its bundled name as
// the module identity.
func (e *Evaluator) importSourceModule(module stdlib.SourceModule) object.Object {
	return e.modules.load(source.BundledID(module.Name), module.Name, func() (*object.Module, *object.Error) {
		return e.evaluateSourceModule(module)
	})
}

// evaluateSourceModule follows the file-module lifecycle using embedded source.
// NativeBindings seed the entry environment so Silver code can wrap Go-backed
// operations; the source's export declaration determines the resulting API.
func (e *Evaluator) evaluateSourceModule(module stdlib.SourceModule) (*object.Module, *object.Error) {
	state, parseError := e.prepareSourcePackage(module)
	if parseError != nil {
		return nil, parseError
	}
	program := state.programs[source.BundledID(module.Name)]

	moduleEnv := object.NewEnvironment()
	moduleEnv.SetPackageID(module.PackageID)
	moduleEnv.SetSourceDir(path.Dir(module.SourceName))
	for name, binding := range module.NativeBindings {
		moduleEnv.Set(name, binding)
	}
	e.pushContext("<module>")
	defer e.popContext()
	result := e.Eval(program, moduleEnv)
	if failure, ok := result.(*object.Error); ok {
		return nil, failure
	}

	exports, exportError := e.moduleExports(program, moduleEnv)
	if exportError != nil {
		return nil, exportError
	}
	return &object.Module{ID: source.BundledID(module.Name), Path: module.Name, Exports: exports}, nil
}

// moduleExports returns every top-level binding unless the program contains
// an export declaration, in which case only its listed names are exposed.
// Collection happens after execution, so declarations can export bindings
// created later in the file. The map is copied, but exported objects retain
// their identities and closures retain the module's private environment.
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

// parseFile reads a source file and parses it with its absolute path attached
// to every token for diagnostics and tracebacks.
func (e *Evaluator) parseFile(path string, manifest *packages.Manifest) (*ast.Program, *object.Error) {
	if manifest != nil {
		state, parseError := e.preparePackage(manifest)
		if parseError != nil {
			return nil, parseError
		}
		if program := state.programs[source.FileID(path)]; program != nil {
			return program, nil
		}
	}
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, newError(object.RuntimeErrorKindImport, "could not read %q: %s", path, err)
	}
	registry := e.operatorScope("").registry
	return ParseSourceWithRegistry(path, input, registry)
}

// ParseSource parses and optimizes source with a diagnostic name that need not
// refer to a filesystem path.
func ParseSource(sourceName string, input []byte) (*ast.Program, *object.Error) {
	p := parser.New(lexer.NewWithSource(string(input), sourceName))
	return finishParse(sourceName, p)
}

// ParseSourceWithRegistry parses and optimizes a source with a shared package
// operator registry. Package tooling must predefine all member declarations
// before parsing the first file, just as runtime package preparation does.
func ParseSourceWithRegistry(sourceName string, input []byte, registry *parser.InfixRegistry) (*ast.Program, *object.Error) {
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
