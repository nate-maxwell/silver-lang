package builtins

import "silver/object"

// arrayDefinitions groups the native array operations. Keeping this list next
// to their implementations makes their Silver names easy to audit.
func arrayDefinitions(null *object.Null) []definition {
	return []definition{
		{name: "first", fn: builtinFirst(null)},
		{name: "last", fn: builtinLast(null)},
		{name: "rest", fn: builtinRest(null)},
		{name: "push", fn: builtinPush},
	}
}

// builtinFirst creates a builtin that returns the first array element, or null
// for an empty array.
func builtinFirst(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("first", args[0])
		if err != nil {
			return err
		}
		if len(array.Elements) == 0 {
			return null
		}
		return array.Elements[0]
	}
}

// builtinLast creates a builtin that returns the last array element, or null
// for an empty array.
func builtinLast(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("last", args[0])
		if err != nil {
			return err
		}
		if len(array.Elements) == 0 {
			return null
		}
		return array.Elements[len(array.Elements)-1]
	}
}

// builtinRest creates a builtin that returns a new array containing every
// element except the first. Empty input produces null.
func builtinRest(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("rest", args[0])
		if err != nil {
			return err
		}
		if len(array.Elements) == 0 {
			return null
		}

		elements := make([]object.Object, len(array.Elements)-1)
		copy(elements, array.Elements[1:])
		return &object.Array{Elements: elements}
	}
}

// builtinPush returns a new array with the supplied value appended. The input
// array is not mutated.
func builtinPush(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	array, err := requireArray("push", args[0])
	if err != nil {
		return err
	}

	elements := make([]object.Object, len(array.Elements)+1)
	copy(elements, array.Elements)
	elements[len(array.Elements)] = args[1]
	return &object.Array{Elements: elements}
}
