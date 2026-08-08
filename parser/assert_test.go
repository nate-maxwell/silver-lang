package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestAssertStatement(t *testing.T) {
	p := New(lexer.New("assert value > 0, \"must be positive\""))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.AssertStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.AssertStatement", program.Statements[0])
	}
	if got, want := statement.Condition.String(), "(value > 0)"; got != want {
		t.Fatalf("condition is %q, want %q", got, want)
	}
	if got, want := statement.Message.String(), "must be positive"; got != want {
		t.Fatalf("message is %q, want %q", got, want)
	}
	if got, want := statement.String(), "assert (value > 0), must be positive"; got != want {
		t.Fatalf("statement is %q, want %q", got, want)
	}
}

func TestAssertStatementWithoutMessage(t *testing.T) {
	p := New(lexer.New("assert True\n42"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.AssertStatement)
	if statement.Message != nil || len(program.Statements) != 2 {
		t.Fatalf("parsed assertion is %#v in %d statements", statement, len(program.Statements))
	}
}

func TestAssertStatementDiagnostics(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "assert", want: "assert requires a condition"},
		{input: "assert True,", want: "assert message requires an expression"},
	}
	for _, test := range tests {
		p := New(lexer.New(test.input))
		p.ParseProgram()
		if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], test.want) {
			t.Fatalf("errors for %q are %q, want %q", test.input, p.Errors(), test.want)
		}
	}
}
