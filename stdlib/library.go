// Package stdlib contains importable modules that ship with Silver. The
// evaluator depends only on the library exposed by this package.
package stdlib

import (
	"embed"
	"fmt"
	"io"
	"path"
	"silver/ast"
	"silver/object"
)

// silverModuleFiles contains the Silver-authored portion of the standard
// library. Embedding the directory recursively allows each module to choose
// its own internal layout. Bare import names still come only from the explicit
// source definitions in New.
//
// Regenerate the embedded .astc files after changing any Silver source or the
// AST cache format.
//
//go:generate go run ./internal/astgen
//go:embed silver
var silverModuleFiles embed.FS

// definition is the declarative form of one Go-backed module export.
type definition struct {
	name      string
	fn        object.BuiltinFunction
	value     object.Object
	signature *ast.TypeAnnotation
}

// sourceDefinition explicitly maps one public bare import name to an embedded
// Silver entry file. The source path is independent of the module's public
// name and may be nested anywhere under stdlib/silver.
type sourceDefinition struct {
	name string
	path string
}

// Library contains the importable standard-library modules available to one
// evaluator.
type Library struct {
	modules       map[string]*object.Module
	sourceModules map[string]sourceModule
}

// sourceModule is an embedded, Silver-authored standard-library module. The
// evaluator owns parsing and execution so stdlib does not depend on evaluator.
type sourceModule struct {
	source     string
	sourceName string
	cache      []byte
}

// New constructs Silver's standard library around the evaluator's configured
// output and canonical singleton values.
func New(out io.Writer, null *object.Null, trueValue, falseValue *object.Boolean) *Library {
	if out == nil {
		out = io.Discard
	}
	if null == nil {
		panic("stdlib: canonical null value must not be nil")
	}
	if trueValue == nil || falseValue == nil {
		panic("stdlib: canonical boolean values must not be nil")
	}

	return newLibrary(map[string][]definition{
		"array":       arrayDefinitions(null, trueValue, falseValue),
		"collections": collectionDefinitions(null),
		"core":        coreDefinitions(),
		"io":          ioDefinitions(out, null),
		"json":        jsonDefinitions(null, trueValue, falseValue),
		"map":         mapDefinitions(null, trueValue, falseValue),
		"math":        mathDefinitions(),
		"path":        pathDefinitions(null, trueValue, falseValue),
		"string":      stringDefinitions(trueValue, falseValue),
		"system":      systemDefinitions(null),
		"time":        timeDefinitions(null, trueValue, falseValue),
	}, []sourceDefinition{
		{name: "testing", path: "silver/testing/testing.slv"},
	})
}

func newLibrary(definitions map[string][]definition, sourceDefinitions []sourceDefinition) *Library {
	modules := make(map[string]*object.Module, len(definitions))
	for name, moduleDefinitions := range definitions {
		exports := make(map[string]object.Object, len(moduleDefinitions))
		for _, definition := range moduleDefinitions {
			if _, exists := exports[definition.name]; exists {
				panic(fmt.Sprintf("standard library module %q exports %q more than once", name, definition.name))
			}
			if definition.value != nil {
				exports[definition.name] = definition.value
				continue
			}
			exports[definition.name] = &object.Builtin{Fn: definition.fn, Signature: definition.signature}
		}
		modules[name] = &object.Module{Path: name, Exports: exports}
	}

	sourceModules := loadSourceModules(modules, sourceDefinitions)
	return &Library{modules: modules, sourceModules: sourceModules}
}

// loadSourceModules loads only explicitly registered embedded entry files.
// Embedding keeps them available in compiled Silver executables without
// making standard-library behavior depend on the process working directory.
func loadSourceModules(nativeModules map[string]*object.Module, definitions []sourceDefinition) map[string]sourceModule {
	modules := make(map[string]sourceModule, len(definitions))
	for _, definition := range definitions {
		if _, exists := nativeModules[definition.name]; exists {
			panic(fmt.Sprintf("standard library module %q is defined in both Go and Silver", definition.name))
		}
		if _, exists := modules[definition.name]; exists {
			panic(fmt.Sprintf("standard library Silver module %q is defined more than once", definition.name))
		}

		source, err := silverModuleFiles.ReadFile(definition.path)
		if err != nil {
			panic(fmt.Sprintf("stdlib: could not read embedded Silver module %q from %q: %s", definition.name, definition.path, err))
		}
		cachePath := definition.path + ".astc"
		cache, err := silverModuleFiles.ReadFile(cachePath)
		if err != nil {
			panic(fmt.Sprintf("stdlib: could not read embedded AST cache for Silver module %q from %q: %s", definition.name, cachePath, err))
		}
		modules[definition.name] = sourceModule{
			source:     string(source),
			sourceName: path.Join("stdlib", definition.path),
			cache:      cache,
		}
	}
	return modules
}

// LookupModule returns the standard-library module with the given bare import
// name, such as "math".
func (l *Library) LookupModule(name string) (*object.Module, bool) {
	module, ok := l.modules[name]
	return module, ok
}

// LookupSourceModule returns an embedded Silver implementation registered for
// the bare import name. Evaluation remains the evaluator's responsibility.
func (l *Library) LookupSourceModule(name string) (source, sourceName string, cache []byte, ok bool) {
	module, ok := l.sourceModules[name]
	if !ok {
		return "", "", nil, false
	}
	return module.source, module.sourceName, module.cache, true
}

func requireArgumentCount(args []object.Object, want int) *object.Error {
	if len(args) == want {
		return nil
	}
	return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", len(args), want)
}

func requireArray(name string, value object.Object) (*object.Array, *object.Error) {
	array, ok := value.(*object.Array)
	if !ok {
		return nil, newError(object.RuntimeErrorKindType, "argument to `%s` must be ARRAY, got %s", name, value.Type())
	}
	return array, nil
}

func newError(kind object.RuntimeErrorKind, format string, args ...interface{}) *object.Error {
	return object.NewError(kind, fmt.Sprintf(format, args...))
}
