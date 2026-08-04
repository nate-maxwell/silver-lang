// Package builtins contains the native functions made available to Silver
// programs by the evaluator.
package builtins

import (
	"fmt"
	"io"
	"silver/ast"
	"silver/object"
)

// definition is the declarative form of a builtin. Domain files return slices
// of definitions, and the registry turns them into callable Silver objects.
type definition struct {
	name      string
	fn        object.BuiltinFunction
	signature *ast.TypeAnnotation
}

// Registry contains the native functions available to one evaluator. Each
// evaluator owns a registry so dependencies such as print's writer remain
// configurable and isolated between execution sessions.
type Registry struct {
	values  map[string]*object.Builtin
	methods map[object.ObjectType]map[string]*object.Builtin
}

// New constructs the standard builtin registry from the evaluator's canonical
// null and boolean singletons.
func New(out io.Writer, null *object.Null, trueValue, falseValue *object.Boolean) *Registry {
	if out == nil {
		out = io.Discard
	}
	if null == nil {
		panic("builtins: canonical null value must not be nil")
	}
	if trueValue == nil || falseValue == nil {
		panic("builtins: canonical boolean values must not be nil")
	}

	arrays := arrayDefinitions(null, trueValue, falseValue)
	maps := mapDefinitions(null, trueValue, falseValue)
	registry := newRegistry(
		coreDefinitions(),
		arrays,
		mapGlobalDefinitions(maps),
		ioDefinitions(out, null),
	)
	registry.addMethods(object.ARRAY_OBJ, arrays)
	registry.addMethods(object.MAP_OBJ, maps)
	return registry
}

// newRegistry materializes definitions as object.Builtin values. A duplicate
// name is a programming error, so construction fails instead of silently
// replacing the earlier definition.
func newRegistry(groups ...[]definition) *Registry {
	registry := &Registry{
		values:  make(map[string]*object.Builtin),
		methods: make(map[object.ObjectType]map[string]*object.Builtin),
	}
	for _, group := range groups {
		for _, builtin := range group {
			if _, exists := registry.values[builtin.name]; exists {
				panic(fmt.Sprintf("builtin %q registered more than once", builtin.name))
			}
			registry.values[builtin.name] = &object.Builtin{Fn: builtin.fn, Signature: builtin.signature}
		}
	}
	return registry
}

// addMethods exposes a definition group as receiver-style methods for one
// runtime type. The underlying functions still receive the receiver as their
// first argument, which lets the same implementation support global calls.
func (r *Registry) addMethods(objectType object.ObjectType, definitions []definition) {
	methods := make(map[string]*object.Builtin, len(definitions))
	for _, definition := range definitions {
		methods[definition.name] = &object.Builtin{Fn: definition.fn, Signature: definition.signature}
	}
	r.methods[objectType] = methods
}

// Lookup returns the builtin registered under the identifier used by Silver
// source code.
func (r *Registry) Lookup(name string) (*object.Builtin, bool) {
	builtin, ok := r.values[name]
	return builtin, ok
}

// LookupMethod returns a builtin method registered for a runtime type.
func (r *Registry) LookupMethod(objectType object.ObjectType, name string) (*object.Builtin, bool) {
	methods, ok := r.methods[objectType]
	if !ok {
		return nil, false
	}
	builtin, ok := methods[name]
	return builtin, ok
}

// requireArgumentCount converts an arity mismatch into a Silver error object.
// Builtins call it before indexing args so invalid programs cannot cause a Go
// index-out-of-range panic.
func requireArgumentCount(args []object.Object, want int) *object.Error {
	if len(args) == want {
		return nil
	}
	return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", len(args), want)
}

// requireArray performs the shared runtime type check for array builtins.
func requireArray(name string, value object.Object) (*object.Array, *object.Error) {
	array, ok := value.(*object.Array)
	if !ok {
		return nil, newError(object.RuntimeErrorKindType, "argument to `%s` must be ARRAY, got %s", name, value.Type())
	}
	return array, nil
}

// newError creates a runtime error that the evaluator will annotate with the
// Silver call site's traceback information.
func newError(kind object.RuntimeErrorKind, format string, args ...interface{}) *object.Error {
	return object.NewError(kind, fmt.Sprintf(format, args...))
}
