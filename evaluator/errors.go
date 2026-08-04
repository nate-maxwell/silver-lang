package evaluator

import (
	"fmt"
	"silver/object"
)

// newError creates a typed built-in runtime error struct. Eval attaches its
// source origin centrally while declared error structs use the same wrapper.
func newError(kind object.RuntimeErrorKind, format string, a ...interface{}) *object.Error {
	return object.NewError(kind, fmt.Sprintf(format, a...))
}

// isError identifies the single failure object used for evaluation unwinding.
func isError(obj object.Object) bool {
	_, ok := obj.(*object.Error)
	return ok
}
