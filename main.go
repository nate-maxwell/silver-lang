package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"silver/evaluator"
	"silver/object"
	"silver/repl"
)

// main delegates process setup to run so CLI behavior can be tested with
// in-memory streams.
func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// run executes either the interactive REPL or one source file. It returns a
// process-style status code instead of exiting directly: 0 for success, 1 for
// evaluation failure, and 2 for invalid command-line usage.
func run(args []string, in io.Reader, out, errOut io.Writer) int {
	switch len(args) {
	case 0:
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintf(errOut, "could not determine current user: %s\n", err)
			return 1
		}
		fmt.Fprintf(out, "Hello %s! This is the Silver programming language!\n", currentUser.Username)
		fmt.Fprintln(out, "Feel free to type in commands")
		repl.Start(in, out)
		return 0
	case 1:
		engine := evaluator.NewWithWriters(out, errOut)
		result := engine.EvalFile(args[0], object.NewEnvironment())
		if _, failed := result.(*object.Error); failed {
			fmt.Fprintln(errOut, result.Inspect())
			return 1
		}
		return 0
	default:
		fmt.Fprintln(errOut, "usage: silver [file]")
		return 2
	}
}
