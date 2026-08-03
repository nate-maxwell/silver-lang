package builtins

import (
	"math"
	"silver/object"
)

// coreDefinitions returns general-purpose primitives available as global
// identifiers in every evaluated Silver program.
func coreDefinitions() []definition {
	return []definition{
		{name: "abs", fn: builtinAbs},
		{name: "len", fn: builtinLen},
		{name: "min", fn: builtinMin},
		{name: "max", fn: builtinMax},
		{name: "range", fn: builtinRange},
		{name: "type", fn: builtinType},
	}
}

// builtinAbs returns the non-negative magnitude of an integer or float.
func builtinAbs(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	switch value := args[0].(type) {
	case *object.Integer:
		const minimumInteger = -1 << 63
		if value.Value == minimumInteger {
			return newError("absolute value out of range for INTEGER")
		}
		if value.Value < 0 {
			return &object.Integer{Value: -value.Value}
		}
		return &object.Integer{Value: value.Value}
	case *object.Float:
		return &object.Float{Value: math.Abs(value.Value)}
	default:
		return newError("argument to `abs` must be INTEGER or FLOAT, got %s", args[0].Type())
	}
}

func builtinMin(args ...object.Object) object.Object {
	return builtinNumericExtreme("min", true, args)
}

func builtinMax(args ...object.Object) object.Object {
	return builtinNumericExtreme("max", false, args)
}

// builtinNumericExtreme returns one of the original numeric operands, keeping
// its integer or floating-point representation intact.
func builtinNumericExtreme(name string, minimum bool, args []object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	for index, argument := range args {
		if !isNumber(argument) {
			return newError("argument %d to `%s` must be INTEGER or FLOAT, got %s", index+1, name, argument.Type())
		}
	}
	comparison := compareNumbers(args[0], args[1])
	if minimum && comparison > 0 || !minimum && comparison < 0 {
		return args[1]
	}
	return args[0]
}

// builtinRange returns consecutive integers from start (inclusive) to end
// (exclusive). An end at or below start produces an empty array.
func builtinRange(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	start, ok := args[0].(*object.Integer)
	if !ok {
		return newError("argument 1 to `range` must be INTEGER, got %s", args[0].Type())
	}
	end, ok := args[1].(*object.Integer)
	if !ok {
		return newError("argument 2 to `range` must be INTEGER, got %s", args[1].Type())
	}
	if end.Value <= start.Value {
		return &object.Array{Elements: []object.Object{}}
	}

	length := uint64(end.Value) - uint64(start.Value)
	const maximumRangeLength = 1_000_000
	if length > maximumRangeLength {
		return newError("range contains too many elements: %d (maximum %d)", length, maximumRangeLength)
	}
	elements := make([]object.Object, int(length))
	for index := range elements {
		elements[index] = &object.Integer{Value: start.Value + int64(index)}
	}
	return &object.Array{Elements: elements}
}

// builtinType returns a first-class primitive type or the nominal definition
// associated with a struct or enum value.
func builtinType(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	return object.TypeOf(args[0])
}
