package main

import (
	"fmt"
	"io"
	"os/user"
	"silver/evaluator"
	"silver/internal/version"
	"silver/object"
	"silver/packages"
	"silver/repl"
)

const usage = `usage:
  silver [file]
  silver package init <package_name>
  silver version`

type command struct {
	name string
	run  func(args []string, in io.Reader, out, errOut io.Writer) int
}

var commands = []command{
	{name: "package", run: runPackage},
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

func runPackage(args []string, _ io.Reader, out, errOut io.Writer) int {
	if len(args) != 2 || args[0] != "init" {
		fmt.Fprintln(errOut, "usage: silver package init <package_name>")
		return 2
	}

	path, err := packages.Init(".", args[1])
	if err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	fmt.Fprintf(out, "created %s\n", path)
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
