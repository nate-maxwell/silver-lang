package builtins

import "silver/object"

// coreDefinitions returns general-purpose primitives available as global
// identifiers in every evaluated Silver program.
func coreDefinitions() []definition {
	return []definition{
		{name: "len", fn: builtinLen},
	}
}

// builtinLen returns the number of elements in an array or the number of bytes
// in a string. String length is byte-based, so a multi-byte UTF-8 character may
// contribute more than one to the result.
func builtinLen(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}

	switch argument := args[0].(type) {
	case *object.Array:
		return &object.Integer{Value: int64(len(argument.Elements))}
	case *object.String:
		return &object.Integer{Value: int64(len(argument.Value))}
	default:
		return newError("argument to `len` not supported, got %s", args[0].Type())
	}
}
