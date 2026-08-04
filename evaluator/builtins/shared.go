package builtins

import (
	"math"
	"math/big"
	"silver/object"
)

// builtinLen returns the number of array elements, map pairs, or string bytes.
// String length remains byte-based rather than counting Unicode code points.
func builtinLen(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}

	switch argument := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(argument.Elements))}
	case *object.Hash:
		return &object.Integer{Value: int64(argument.Len())}
	case *object.String:
		return &object.Integer{Value: int64(len(argument.Value))}
	default:
		return newError(object.RuntimeErrorKindType, "argument to `len` not supported, got %s", args[0].Type())
	}
}

// builtinContains dispatches collection membership for the globally available
// contains function and for the array and map receiver methods.
func builtinContains(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		switch collection := args[0].(type) {
		case *object.Array:
			for _, element := range collection.Elements {
				if objectsEqual(element, args[1]) {
					return trueValue
				}
			}
			return falseValue
		case *object.Hash:
			key, ok := args[1].(object.Hashable)
			if !ok {
				return newError(object.RuntimeErrorKindType, "unusable as hash key: %s", args[1].Type())
			}
			if _, ok := collection.Get(key.HashKey()); ok {
				return trueValue
			}
			return falseValue
		default:
			return newError(object.RuntimeErrorKindType, "argument to `contains` must be ARRAY or HASH, got %s", args[0].Type())
		}
	}
}

func objectsEqual(left, right object.Object) bool {
	if isNumber(left) && isNumber(right) {
		if numberIsNaN(left) || numberIsNaN(right) {
			return false
		}
		return compareNumbers(left, right) == 0
	}
	if left.Type() != right.Type() {
		return false
	}
	switch left := left.(type) {
	case *object.String:
		return left.Value == right.(*object.String).Value
	case *object.Boolean:
		return left.Value == right.(*object.Boolean).Value
	default:
		return left == right
	}
}

func isNumber(value object.Object) bool {
	return value.Type() == object.INTEGER_OBJ || value.Type() == object.FLOAT_OBJ
}

func numberIsNaN(value object.Object) bool {
	float, ok := value.(*object.Float)
	return ok && math.IsNaN(float.Value)
}

func compareNumbers(left, right object.Object) int {
	leftFloat, leftIsFloat := left.(*object.Float)
	rightFloat, rightIsFloat := right.(*object.Float)
	if leftIsFloat && (math.IsNaN(leftFloat.Value) || math.IsInf(leftFloat.Value, 0)) ||
		rightIsFloat && (math.IsNaN(rightFloat.Value) || math.IsInf(rightFloat.Value, 0)) {
		leftValue := numberAsFloat(left)
		rightValue := numberAsFloat(right)
		switch {
		case math.IsNaN(leftValue):
			if math.IsNaN(rightValue) {
				return 0
			}
			return 1
		case math.IsNaN(rightValue):
			return -1
		case leftValue < rightValue:
			return -1
		case leftValue > rightValue:
			return 1
		default:
			return 0
		}
	}
	return numberAsRat(left).Cmp(numberAsRat(right))
}

func numberAsFloat(value object.Object) float64 {
	if integer, ok := value.(*object.Integer); ok {
		return float64(integer.Value)
	}
	return value.(*object.Float).Value
}

func numberAsRat(value object.Object) *big.Rat {
	if integer, ok := value.(*object.Integer); ok {
		return new(big.Rat).SetInt64(integer.Value)
	}
	return new(big.Rat).SetFloat64(value.(*object.Float).Value)
}
