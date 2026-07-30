package evaluator

import (
	"math"
	"silver/object"
)

// evalIntegerInfixExpression implements integer arithmetic and comparisons.
// Division by zero becomes a Silver error rather than a Go panic.
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftValue := left.(*object.Integer).Value
	rightValue := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftValue + rightValue}
	case "-":
		return &object.Integer{Value: leftValue - rightValue}
	case "*":
		return &object.Integer{Value: leftValue * rightValue}
	case "**":
		if rightValue < 0 {
			if leftValue == 0 {
				return newError("division by zero")
			}
			return &object.Float{Value: math.Pow(float64(leftValue), float64(rightValue))}
		}
		return &object.Integer{Value: integerPower(leftValue, rightValue)}
	case "/":
		if rightValue == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftValue / rightValue}
	case "<":
		return nativeBoolToBooleanObject(leftValue < rightValue)
	case ">":
		return nativeBoolToBooleanObject(leftValue > rightValue)
	case "==":
		return nativeBoolToBooleanObject(leftValue == rightValue)
	case "!=":
		return nativeBoolToBooleanObject(leftValue != rightValue)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// integerPower uses exponentiation by squaring. Like Silver's other integer
// arithmetic, overflow follows Go's signed-integer wraparound behavior.
func integerPower(base, exponent int64) int64 {
	result := int64(1)
	for exponent > 0 {
		if exponent&1 == 1 {
			result *= base
		}
		exponent >>= 1
		if exponent > 0 {
			base *= base
		}
	}
	return result
}
