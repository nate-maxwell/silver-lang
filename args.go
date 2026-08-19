package main

import (
	"fmt"
	"io"
	"os/user"
	"silver/astgen"
	"silver/evaluator"
	"silver/formatter"
	"silver/internal/version"
	"silver/object"
	"silver/repl"
)

const usage = `usage:
  silver [file]
  silver astgen <path>
  silver frmt <file>
  silver version`

type command struct {
	name string
	run  func(args []string, in io.Reader, out, errOut io.Writer) int
}

var commands = []command{
	{name: "astgen", run: astgen.RunASTGen},
	{name: "frmt", run: runFormat},
	{name: "version", run: runVersion},
}

// run parses command-line arguments and returns a process-style status code:
// 0 for success, 1 for an execution failure, and 2 for invalid usage.
func run(args []string, in io.Reader, out, errOut io.Writer) int {
	if len(args) == 0 {
		return runREPL(in, out, errOut)
	}

	for _, candidate := range commands {
		if args[0] == candidate.name {
			return candidate.run(args[1:], in, out, errOut)
		}
	}

	if len(args) == 1 {
		return runFile(args[0], in, out, errOut)
	}

	fmt.Fprintln(errOut, usage)
	return 2
}

func runREPL(in io.Reader, out, errOut io.Writer) int {
	currentUser, err := user.Current()
	if err != nil {
		fmt.Fprintf(errOut, "could not determine current user: %s\n", err)
		return 1
	}
	fmt.Fprintf(out, "Hello %s! This is the Silver programming language!\n", currentUser.Username)
	fmt.Fprintln(out, "Feel free to type in commands")
	repl.Start(in, out)
	return 0
}

func runFile(path string, in io.Reader, out, errOut io.Writer) int {
	engine := evaluator.NewWithStreams(in, out, errOut)
	result := engine.EvalFile(path, object.NewEnvironment())
	if _, failed := result.(*object.Error); failed {
		fmt.Fprintln(errOut, result.Inspect())
		return 1
	}
	return 0
}

func runFormat(args []string, _ io.Reader, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: silver frmt <file>")
		return 2
	}
	changed, err := formatter.File(args[0])
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	if changed {
		fmt.Fprintln(out, args[0])
	}
	return 0
}

func runVersion(args []string, _ io.Reader, out, errOut io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(errOut, "usage: silver version")
		return 2
	}
	fmt.Fprintf(out, "silver %s\n", version.String())
	return 0
}
