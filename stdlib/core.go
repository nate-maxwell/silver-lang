package stdlib

import "silver/object"

// coreDefinitions contains the functions exported by import("core").
func coreDefinitions() []definition {
	return []definition{
		{name: "len", fn: builtinLen},
		{name: "range", fn: builtinRange},
		{name: "type", fn: builtinType},
	}
}

// builtinLen returns the number of array elements, map pairs, or string bytes.
// String length remains byte-based rather than counting Unicode code points.
func builtinLen(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}

	switch argument := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(argument.Elements))}
	case *object.Map:
		return &object.Integer{Value: int64(argument.Len())}
	case *object.String:
		return &object.Integer{Value: int64(len(argument.Value))}
	default:
		return newError(object.RuntimeErrorKindType, "argument to `len` not supported, got %s", args[0].Type())
	}
}

// builtinRange returns consecutive integers from start (inclusive) to end
// (exclusive). An end at or below start produces an empty array.
func builtinRange(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	start, ok := args[0].(*object.Integer)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument 1 to `range` must be INTEGER, got %s", args[0].Type())
	}
	end, ok := args[1].(*object.Integer)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument 2 to `range` must be INTEGER, got %s", args[1].Type())
	}
	if end.Value <= start.Value {
		return &object.Array{Elements: []object.Object{}}
	}

	length := uint64(end.Value) - uint64(start.Value)
	const maximumRangeLength = 1_000_000
	if length > maximumRangeLength {
		return newError(object.RuntimeErrorKindValue, "range contains too many elements: %d (maximum %d)", length, maximumRangeLength)
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
