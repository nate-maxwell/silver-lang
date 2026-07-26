package evaluator

import "silver/object"

// arrayBuiltinDefinitions groups the native array operations. Keeping this
// list next to their implementations makes the public builtin names easy to
// audit while the registry remains responsible for combining all domains.
func arrayBuiltinDefinitions() []builtinDefinition {
	return []builtinDefinition{
		{name: "first", fn: builtinFirst},
		{name: "last", fn: builtinLast},
		{name: "rest", fn: builtinRest},
		{name: "push", fn: builtinPush},
	}
}

// builtinFirst returns the first array element, or null for an empty array.
func builtinFirst(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray("first", args[0])
	if err != nil {
		return err
	}
	if len(array.Elements) == 0 {
		return NULL
	}
	return array.Elements[0]
}

// builtinLast returns the last array element, or null for an empty array.
func builtinLast(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray("last", args[0])
	if err != nil {
		return err
	}
	if len(array.Elements) == 0 {
		return NULL
	}
	return array.Elements[len(array.Elements)-1]
}

// builtinRest returns a new array containing every element except the first.
// It returns null for an empty input and never mutates the original array.
func builtinRest(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray("rest", args[0])
	if err != nil {
		return err
	}
	if len(array.Elements) == 0 {
		return NULL
	}

	elements := make([]object.Object, len(array.Elements)-1)
	copy(elements, array.Elements[1:])
	return &object.Array{Elements: elements}
}

// builtinPush returns a new array with the supplied value appended. Silver
// arrays are treated immutably here, so the input array is left unchanged.
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
