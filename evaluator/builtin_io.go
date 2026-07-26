package evaluator

import (
	"fmt"
	"io"
	"silver/object"
)

// ioBuiltinDefinitions builds I/O functions around the evaluator's configured
// writer. Capturing the writer here keeps builtins independent of os.Stdout and
// allows callers and tests to redirect program output.
func ioBuiltinDefinitions(out io.Writer) []builtinDefinition {
	return []builtinDefinition{
		{name: "print", fn: builtinPrint(out)},
	}
}

// builtinPrint creates a print function bound to the supplied writer. Each
// argument is written on its own line, and the function returns null because
// the operation exists for its side effect rather than for a value.
func builtinPrint(out io.Writer) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		for _, arg := range args {
			fmt.Fprintln(out, arg.Inspect())
		}
		return NULL
	}
}
