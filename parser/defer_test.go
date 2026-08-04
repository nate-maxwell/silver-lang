package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestDeferStatement(t *testing.T) {
	p := New(lexer.New("defer file.close()"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}
	statement, ok := program.Statements[0].(*ast.DeferStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.DeferStatement", program.Statements[0])
	}
	if got, want := statement.String(), "defer file.close()"; got != want {
		t.Fatalf("statement is %q, want %q", got, want)
	}
}

func TestDeferRequiresCall(t *testing.T) {
	p := New(lexer.New("defer file"))
	p.ParseProgram()

	if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], "defer requires a function call") {
		t.Fatalf("errors are %q, want defer call diagnostic", p.Errors())
	}
}
