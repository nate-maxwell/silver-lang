package repl

import (
	"bufio"
	"io"
	"silver/evaluator"
	"silver/lexer"
	"silver/object"
	"silver/parser"
)

const prompt = ">> "

// Start runs a persistent read-evaluate-print loop. It reuses one environment
// and evaluator so bindings and imported modules survive between input lines.
func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := object.NewEnvironment()
	engine := evaluator.NewWithOutput(out)

	for {
		io.WriteString(out, prompt)
		scanned := scanner.Scan()
		if !scanned {
			return
		}

		line := scanner.Text()
		l := lexer.NewWithSource(line, "<repl>")
		p := parser.New(l)

		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}

		evaluated := engine.Eval(program, env)
		if evaluated != nil {
			io.WriteString(out, evaluated.Inspect())
			io.WriteString(out, "\n")
		}
	}
}

// printParserErrors writes parser diagnostics in the REPL's indented format.
func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		io.WriteString(out, "\t"+msg+"\n")
	}
}
