package evaluator

import (
	"math"
	"math/big"
	"silver/object"
)

// isNumeric reports whether value participates in Silver numeric promotion.
func isNumeric(value object.Object) bool {
	return value.Type() == object.INTEGER_OBJ || value.Type() == object.FLOAT_OBJ
}

// evalFloatInfixExpression promotes numeric operands to floats for arithmetic.
// Comparisons use exact rational representations for finite values so large
// integers do not become spuriously equal after float conversion.
func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftValue := numericAsFloat(left)
	rightValue := numericAsFloat(right)

	switch operator {
	case "+":
		return &object.Float{Value: leftValue + rightValue}
	case "-":
		return &object.Float{Value: leftValue - rightValue}
	case "*":
		return &object.Float{Value: leftValue * rightValue}
	case "/":
		if rightValue == 0 {
			return newError("division by zero")
		}
		return &object.Float{Value: leftValue / rightValue}
	case "<", ">", "==", "!=":
		return nativeBoolToBooleanObject(compareNumeric(operator, left, right))
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// numericAsFloat performs arithmetic promotion from an integer or float object.
func numericAsFloat(value object.Object) float64 {
	if integer, ok := value.(*object.Integer); ok {
		return float64(integer.Value)
	}
	return value.(*object.Float).Value
}

// compareNumeric compares finite values as exact rationals. Non-finite floats
// retain normal IEEE-754 comparison behavior, including NaN being unequal to
// every value.
func compareNumeric(operator string, left, right object.Object) bool {
	leftFloat := numericAsFloat(left)
	rightFloat := numericAsFloat(right)
	if math.IsInf(leftFloat, 0) || math.IsInf(rightFloat, 0) || math.IsNaN(leftFloat) || math.IsNaN(rightFloat) {
		return compareFloat64(operator, leftFloat, rightFloat)
	}

	comparison := numericAsRat(left).Cmp(numericAsRat(right))
	switch operator {
	case "<":
		return comparison < 0
	case ">":
		return comparison > 0
	case "==":
		return comparison == 0
	case "!=":
		return comparison != 0
	default:
		return false
	}
}

// numericAsRat returns the exact rational value represented by a finite
// integer or IEEE-754 float.
func numericAsRat(value object.Object) *big.Rat {
	if integer, ok := value.(*object.Integer); ok {
		return new(big.Rat).SetInt64(integer.Value)
	}
	return new(big.Rat).SetFloat64(value.(*object.Float).Value)
}

// compareFloat64 applies an operator using IEEE-754 rules.
func compareFloat64(operator string, left, right float64) bool {
	switch operator {
	case "<":
		return left < right
	case ">":
		return left > right
	case "==":
		return left == right
	case "!=":
		return left != right
	default:
		return false
	}
}
