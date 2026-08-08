// Package stdlib contains importable modules that ship with Silver. The
// evaluator depends only on the library exposed by this package.
package stdlib

import (
	"fmt"
	"io"
	"silver/ast"
	"silver/object"
)

// definition is the declarative form of one Go-backed module export.
type definition struct {
	name      string
	fn        object.BuiltinFunction
	value     object.Object
	signature *ast.TypeAnnotation
}

// Library contains the importable standard-library modules available to one
// evaluator.
type Library struct {
	modules map[string]*object.Module
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
		"map":         mapDefinitions(null, trueValue, falseValue),
		"math":        mathDefinitions(),
		"path":        pathDefinitions(null, trueValue, falseValue),
		"string":      stringDefinitions(trueValue, falseValue),
		"system":      systemDefinitions(null),
	})
}

func newLibrary(definitions map[string][]definition) *Library {
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
	return &Library{modules: modules}
}

// LookupModule returns the standard-library module with the given bare import
// name, such as "math".
func (l *Library) LookupModule(name string) (*object.Module, bool) {
	module, ok := l.modules[name]
	return module, ok
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
