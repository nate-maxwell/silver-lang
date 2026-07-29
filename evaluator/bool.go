package evaluator

import "silver/object"

// Canonical boolean singletons are the boolean portion of constant pooling.
// Unlike the unbounded evaluator-local pools, there are only two possible
// values, so these can be safely shared by every evaluator.
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// isTruthy implements Silver truthiness. Only null and False are falsey.
func isTruthy(value object.Object) bool {
	switch value {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// evalBangOperatorExpression implements logical negation using canonical
// singleton values.
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// nativeBoolToBooleanObject returns canonical boolean instances because
// equality for these values relies on object identity.
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}
