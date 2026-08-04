package evaluator

import (
	"fmt"
	"silver/object"
)

// newError creates the built-in Error struct inside the same Error wrapper
// used for declared struct errors. Eval attaches its source origin centrally.
func newError(format string, a ...interface{}) *object.Error {
	return object.NewError(fmt.Sprintf(format, a...))
}

// isError identifies the single failure object used for evaluation unwinding.
func isError(obj object.Object) bool {
	return obj != nil && obj.Type() == object.ERROR_OBJ
}
