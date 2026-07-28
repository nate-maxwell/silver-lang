package evaluator

import (
	"fmt"
	"silver/object"
)

// newError formats a Silver runtime error without a source frame. Eval attaches
// the origin as the error propagates out of the AST node that created it.
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError safely identifies runtime error objects, including a nil result.
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
