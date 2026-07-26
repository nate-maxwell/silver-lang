package evaluator

import "silver/object"

// coreBuiltinDefinitions returns general-purpose primitives that are available
// as global identifiers in every evaluated Silver program.
func coreBuiltinDefinitions() []builtinDefinition {
	return []builtinDefinition{
		{name: "len", fn: builtinLen},
	}
}

// builtinLen returns the number of elements in an array or the number of bytes
// in a string. String length is byte-based for now, so a multi-byte UTF-8
// character may contribute more than one to the result.
func builtinLen(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}

	switch arg := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(arg.Elements))}
	case *object.String:
		return &object.Integer{Value: int64(len(arg.Value))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}
