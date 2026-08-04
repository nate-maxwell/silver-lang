package evaluator

import "silver/object"

// evalStringInfixExpression implements value-based string operators. Comparing
// contents explicitly keeps equality independent of constant pooling.
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	leftValue := left.(*object.String).Value
	rightValue := right.(*object.String).Value

	switch operator {
	case "+":
		return &object.String{Value: leftValue + rightValue}
	case "==":
		return nativeBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return nativeBoolToBooleanObject(leftValue != rightValue)
	default:
		return newError(object.RuntimeErrorKindType, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}
