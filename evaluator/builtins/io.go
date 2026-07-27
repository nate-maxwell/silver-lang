package builtins

import (
	"fmt"
	"io"
	"silver/object"
)

// ioDefinitions builds I/O functions around the evaluator's configured
// writer. Capturing the writer keeps the package independent of os.Stdout and
// lets callers redirect program output.
func ioDefinitions(out io.Writer, null *object.Null) []definition {
	return []definition{
		{name: "print", fn: builtinPrint(out, null)},
	}
}

// builtinPrint creates a print function bound to out. Each argument is written
// on its own line, and the function returns null because it exists for its side
// effect rather than for a value.
func builtinPrint(out io.Writer, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		for _, argument := range args {
			fmt.Fprintln(out, argument.Inspect())
		}
		return null
	}
}
