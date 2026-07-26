package evaluator

import (
	"fmt"
	"io"
	"silver/object"
)

// builtinDefinition is the declarative form of a builtin. Domain files return
// slices of definitions, and the registry turns them into callable Silver
// objects in one place.
type builtinDefinition struct {
	name string
	fn   object.BuiltinFunction
}

// builtinRegistry contains the native functions available to one Evaluator.
// Keeping it on the evaluator instead of in a package-level map makes runtime
// dependencies, such as the writer used by print, configurable and testable.
type builtinRegistry map[string]*object.Builtin

// newDefaultBuiltinRegistry explicitly composes every builtin group shipped by
// the interpreter. Add a new group here when introducing another native domain.
func newDefaultBuiltinRegistry(out io.Writer) builtinRegistry {
	return newBuiltinRegistry(
		coreBuiltinDefinitions(),
		arrayBuiltinDefinitions(),
		ioBuiltinDefinitions(out),
	)
}

// newBuiltinRegistry materializes definitions as object.Builtin values. A
// duplicate name is a programming/configuration error, so construction fails
// immediately instead of silently allowing one definition to replace another.
func newBuiltinRegistry(groups ...[]builtinDefinition) builtinRegistry {
	registry := make(builtinRegistry)
	for _, group := range groups {
		for _, definition := range group {
			if _, exists := registry[definition.name]; exists {
				panic(fmt.Sprintf("builtin %q registered more than once", definition.name))
			}
			registry[definition.name] = &object.Builtin{Fn: definition.fn}
		}
	}
	return registry
}

// get looks up a builtin by the identifier Silver source uses to call it.
func (r builtinRegistry) get(name string) (*object.Builtin, bool) {
	builtin, ok := r[name]
	return builtin, ok
}

// requireArgumentCount converts an arity mismatch into a Silver error object.
// Builtins should call it before indexing args so bad programs cannot cause a
// Go index-out-of-range panic.
func requireArgumentCount(args []object.Object, want int) *object.Error {
	if len(args) == want {
		return nil
	}
	return newError("wrong number of arguments. got=%d, want=%d", len(args), want)
}

// requireArray performs the common runtime type check for array builtins. The
// builtin name is included in the error to identify which call rejected the
// value.
func requireArray(name string, value object.Object) (*object.Array, *object.Error) {
	array, ok := value.(*object.Array)
	if !ok {
		return nil, newError("argument to `%s` must be ARRAY, got %s", name, value.Type())
	}
	return array, nil
}
