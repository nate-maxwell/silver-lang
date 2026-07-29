package evaluator

import (
	"os"
	"path/filepath"
	"silver/ast"
	"silver/astcache"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"strings"
)

// EvalFile parses and evaluates path in env. It also sets env's source
// directory so relative imports resolve beside the entry file.
func (e *Evaluator) EvalFile(path string, env *object.Environment) object.Object {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return newError("could not resolve file %q: %s", path, err)
	}
	absolutePath = filepath.Clean(absolutePath)

	program, parseError := parseFile(absolutePath)
	if parseError != nil {
		return parseError
	}

	env.SetSourceDir(filepath.Dir(absolutePath))
	return e.Eval(program, env)
}

// importModule resolves, loads, and evaluates a module in an isolated top-level
// environment. Successful modules are cached by canonical absolute path.
func (e *Evaluator) importModule(path string, env *object.Environment) object.Object {
	absolutePath, err := resolveImportPath(path, env.SourceDir())
	if err != nil {
		return newError("could not resolve import %q: %s", path, err)
	}

	if module, ok := e.modules[absolutePath]; ok {
		return module
	}
	if e.loading[absolutePath] {
		return newError("circular import detected while loading %q", absolutePath)
	}

	e.loading[absolutePath] = true
	defer delete(e.loading, absolutePath)

	program, parseError := parseFile(absolutePath)
	if parseError != nil {
		return parseError
	}

	moduleEnv := object.NewEnvironment()
	moduleEnv.SetSourceDir(filepath.Dir(absolutePath))
	e.pushContext("<module>")
	defer e.popContext()
	result := e.Eval(program, moduleEnv)
	if isError(result) {
		return result
	}

	module := &object.Module{Path: absolutePath, Exports: moduleEnv.Bindings()}
	e.modules[absolutePath] = module
	return module
}

// resolveImportPath resolves path relative to sourceDir, falling back to the
// process working directory for in-memory evaluation.
func resolveImportPath(path, sourceDir string) (string, error) {
	if sourceDir == "" {
		var err error
		sourceDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(sourceDir, path)
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolutePath), nil
}

// parseFile reads a source file and parses it with its absolute path attached
// to every token for diagnostics and tracebacks.
func parseFile(path string) (*ast.Program, *object.Error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, newError("could not read %q: %s", path, err)
	}
	if program, ok := astcache.Load(path, input); ok {
		return program, nil
	}

	p := parser.New(lexer.NewWithSource(string(input), path))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return nil, newError("could not parse %q:\n%s", path, strings.Join(p.Errors(), "\n"))
	}
	// A cache is an optimization only. Read-only directories and other cache
	// write failures must not prevent valid source from running.
	_ = astcache.Store(path, input, program)
	return program, nil
}
