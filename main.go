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

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, in io.Reader, out, errOut io.Writer) int {
	switch len(args) {
	case 0:
		currentUser, err := user.Current()
		if err != nil {
			fmt.Fprintf(errOut, "could not determine current user: %s\n", err)
			return 1
		}
		fmt.Fprintf(out, "Hello %s! This is the Monkey programming language!\n", currentUser.Username)
		fmt.Fprintln(out, "Feel free to type in commands")
		repl.Start(in, out)
		return 0
	case 1:
		engine := evaluator.NewWithOutput(out)
		result := engine.EvalFile(args[0], object.NewEnvironment())
		if result != nil && result.Type() == object.ERROR_OBJ {
			fmt.Fprintln(errOut, result.Inspect())
			return 1
		}
		return 0
	default:
		fmt.Fprintln(errOut, "usage: monkey [file]")
		return 2
	}
}
