package stdlib

import (
	"math"
	"math/big"
	"silver/object"
)

// mathDefinitions contains the Go implementations exported by import("math").
func mathDefinitions() []definition {
	return []definition{
		{name: "factorial", fn: mathFactorial},
		{name: "lcd", fn: mathLCD},
		{name: "lcm", fn: mathLCM},
		{name: "gcd", fn: mathGCD},
		{name: "isqrt", fn: mathISqrt},

		{name: "ceil", fn: mathCeil},
		{name: "abs", fn: mathAbs},
		{name: "floor", fn: mathFloor},
		{name: "round", fn: mathRound},
		{name: "fmod", fn: binaryFloatFunction("fmod", math.Mod)},
		{name: "modf", fn: mathModf},
		{name: "remainer", fn: binaryFloatFunction("remainer", math.Remainder)},
		{name: "remainder", fn: binaryFloatFunction("remainder", math.Remainder)},
		{name: "truc", fn: mathTruc},
		{name: "trunc", fn: mathTrunc},

		{name: "acos", fn: unaryFloatFunction("acos", math.Acos)},
		{name: "asin", fn: unaryFloatFunction("asin", math.Asin)},
		{name: "atan", fn: unaryFloatFunction("atan", math.Atan)},
		{name: "cos", fn: unaryFloatFunction("cos", math.Cos)},
		{name: "sin", fn: unaryFloatFunction("sin", math.Sin)},
		{name: "tan", fn: unaryFloatFunction("tan", math.Tan)},

		{name: "cbrt", fn: unaryFloatFunction("cbrt", math.Cbrt)},
		{name: "exp", fn: unaryFloatFunction("exp", math.Exp)},
		{name: "exp2", fn: unaryFloatFunction("exp2", math.Exp2)},
		{name: "expm1", fn: unaryFloatFunction("expm1", math.Expm1)},
		{name: "log", fn: mathLog},
		{name: "log1p", fn: unaryFloatFunction("log1p", math.Log1p)},
		{name: "log2", fn: unaryFloatFunction("log2", math.Log2)},
		{name: "log10", fn: unaryFloatFunction("log10", math.Log10)},
		{name: "sqrt", fn: unaryFloatFunction("sqrt", math.Sqrt)},

		{name: "degrees", fn: unaryFloatFunction("degrees", func(value float64) float64 { return value * 180 / math.Pi })},
		{name: "radians", fn: unaryFloatFunction("radians", func(value float64) float64 { return value * math.Pi / 180 })},

		{name: "min", fn: mathMin},
		{name: "max", fn: mathMax},

		{name: "pi", value: &object.Float{Value: math.Pi}},
		{name: "e", value: &object.Float{Value: math.E}},
		{name: "tau", value: &object.Float{Value: 2 * math.Pi}},
		{name: "nan", value: &object.Float{Value: math.NaN()}},
	}
}

func unaryFloatFunction(name string, operation func(float64) float64) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireNumber(name, 0, args[0])
		if err != nil {
			return err
		}
		return &object.Float{Value: operation(value)}
	}
}

func binaryFloatFunction(name string, operation func(float64, float64) float64) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		left, err := requireNumber(name, 0, args[0])
		if err != nil {
			return err
		}
		right, err := requireNumber(name, 1, args[1])
		if err != nil {
			return err
		}
		return &object.Float{Value: operation(left, right)}
	}
}

func requireNumber(name string, index int, value object.Object) (float64, *object.Error) {
	if !isNumber(value) {
		return 0, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be INTEGER or FLOAT, got %s", index+1, name, value.Type())
	}
	return numberAsFloat(value), nil
}

func requireInteger(name string, value object.Object) (int64, *object.Error) {
	integer, ok := value.(*object.Integer)
	if !ok {
		return 0, newError(object.RuntimeErrorKindType, "argument to `%s` must be INTEGER, got %s", name, value.Type())
	}
	return integer.Value, nil
}

func mathFactorial(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireInteger("factorial", args[0])
	if err != nil {
		return err
	}
	if value < 0 {
		return newError(object.RuntimeErrorKindValue, "argument to `factorial` must be nonnegative")
	}

	var result int64 = 1
	for factor := int64(2); factor <= value; factor++ {
		if result > math.MaxInt64/factor {
			return newError(object.RuntimeErrorKindValue, "result of `factorial` is out of range for INTEGER")
		}
		result *= factor
	}
	return &object.Integer{Value: result}
}

func mathGCD(args ...object.Object) object.Object {
	return mathArrayIntegerOperation("gcd", args, 0, false, gcdMagnitude)
}

// LCD is the public spelling requested by Silver's math API. Its operation is
// the least common multiple of all integers in the supplied array.
func mathLCD(args ...object.Object) object.Object {
	return mathArrayIntegerOperation("lcd", args, 1, true, lcmMagnitude)
}

func mathLCM(args ...object.Object) object.Object {
	return mathArrayIntegerOperation("lcm", args, 1, true, lcmMagnitude)
}

func mathArrayIntegerOperation(name string, args []object.Object, identity uint64, zeroAbsorbing bool, operation func(uint64, uint64) (uint64, bool)) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray(name, args[0])
	if err != nil {
		return err
	}

	magnitudes := make([]uint64, len(array.Elements))
	for index, element := range array.Elements {
		integer, ok := element.(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "element %d of argument to `%s` must be INTEGER, got %s", index+1, name, element.Type())
		}
		magnitudes[index] = integerMagnitude(integer.Value)
	}
	if zeroAbsorbing {
		for _, magnitude := range magnitudes {
			if magnitude == 0 {
				return &object.Integer{Value: 0}
			}
		}
	}

	result := identity
	for _, magnitude := range magnitudes {
		var fits bool
		result, fits = operation(result, magnitude)
		if !fits {
			return newError(object.RuntimeErrorKindValue, "result of `%s` is out of range for INTEGER", name)
		}
	}
	if result > math.MaxInt64 {
		return newError(object.RuntimeErrorKindValue, "result of `%s` is out of range for INTEGER", name)
	}
	return &object.Integer{Value: int64(result)}
}

func integerMagnitude(value int64) uint64 {
	if value >= 0 {
		return uint64(value)
	}
	return uint64(-(value + 1)) + 1
}

func gcdMagnitude(left, right uint64) (uint64, bool) {
	for right != 0 {
		left, right = right, left%right
	}
	return left, true
}

func lcmMagnitude(left, right uint64) (uint64, bool) {
	if left == 0 || right == 0 {
		return 0, true
	}
	divisor, _ := gcdMagnitude(left, right)
	left /= divisor
	if left > math.MaxUint64/right {
		return 0, false
	}
	return left * right, true
}

func mathISqrt(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireInteger("isqrt", args[0])
	if err != nil {
		return err
	}
	if value < 0 {
		return newError(object.RuntimeErrorKindValue, "argument to `isqrt` must be nonnegative")
	}

	root := int64(math.Sqrt(float64(value)))
	for root > 0 && root > value/root {
		root--
	}
	for root < math.MaxInt64 && root+1 <= value/(root+1) {
		root++
	}
	return &object.Integer{Value: root}
}

func mathCeil(args ...object.Object) object.Object {
	return roundedInteger("ceil", math.Ceil, args)
}

func mathFloor(args ...object.Object) object.Object {
	return roundedInteger("floor", math.Floor, args)
}

func mathRound(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	if _, ok := args[0].(*object.Float); !ok {
		return newError(object.RuntimeErrorKindType, "argument 1 to `round` must be FLOAT, got %s", args[0].Type())
	}
	return roundedInteger("round", math.Round, args)
}

func mathTruc(args ...object.Object) object.Object {
	return roundedInteger("truc", math.Trunc, args)
}

func mathTrunc(args ...object.Object) object.Object {
	return roundedInteger("trunc", math.Trunc, args)
}

func roundedInteger(name string, operation func(float64) float64, args []object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	if integer, ok := args[0].(*object.Integer); ok {
		return &object.Integer{Value: integer.Value}
	}
	value, err := requireNumber(name, 0, args[0])
	if err != nil {
		return err
	}
	result := operation(value)
	if math.IsNaN(result) || math.IsInf(result, 0) || result < math.MinInt64 || result >= float64(math.MaxInt64) {
		return newError(object.RuntimeErrorKindValue, "result of `%s` is out of range for INTEGER", name)
	}
	return &object.Integer{Value: int64(result)}
}

func mathModf(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	value, err := requireNumber("modf", 0, args[0])
	if err != nil {
		return err
	}
	integer, fraction := math.Modf(value)
	return &object.Array{Elements: []object.Object{
		&object.Float{Value: fraction},
		&object.Float{Value: integer},
	}}
}

func mathLog(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	value, err := requireNumber("log", 0, args[0])
	if err != nil {
		return err
	}
	base, err := requireNumber("log", 1, args[1])
	if err != nil {
		return err
	}
	return &object.Float{Value: math.Log(value) / math.Log(base)}
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

func mathAbs(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	switch value := args[0].(type) {
	case *object.Integer:
		const minimumInteger = -1 << 63
		if value.Value == minimumInteger {
			return newError(object.RuntimeErrorKindValue, "absolute value out of range for INTEGER")
		}
		if value.Value < 0 {
			return &object.Integer{Value: -value.Value}
		}
		return &object.Integer{Value: value.Value}
	case *object.Float:
		return &object.Float{Value: math.Abs(value.Value)}
	default:
		return newError(object.RuntimeErrorKindType, "argument to `abs` must be INTEGER or FLOAT, got %s", args[0].Type())
	}
}

func mathMin(args ...object.Object) object.Object {
	return numericExtreme("min", true, args)
}

func mathMax(args ...object.Object) object.Object {
	return numericExtreme("max", false, args)
}

// numericExtreme returns one of the original operands so integer and
// floating-point representations remain intact.
func numericExtreme(name string, minimum bool, args []object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	for index, argument := range args {
		if !isNumber(argument) {
			return newError(object.RuntimeErrorKindType, "argument %d to `%s` must be INTEGER or FLOAT, got %s", index+1, name, argument.Type())
		}
	}
	comparison := compareNumbers(args[0], args[1])
	if minimum && comparison > 0 || !minimum && comparison < 0 {
		return args[1]
	}
	return args[0]
}
