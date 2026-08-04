package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestStringLiteralExpression(t *testing.T) {
	p := New(lexer.New(`"hello world"`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.StringLiteral)
	if literal.Value != "hello world" {
		t.Fatalf("literal value is %q, want hello world", literal.Value)
	}
}

func TestStringLiteralEscapesAreDecoded(t *testing.T) {
	p := New(lexer.New(`"hello\n\t\"Silver\""`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.StringLiteral)
	if got, want := literal.Value, "hello\n\t\"Silver\""; got != want {
		t.Fatalf("literal value is %q, want %q", got, want)
	}
}

func TestInvalidStringEscapeReportsDiagnostic(t *testing.T) {
	p := New(lexer.NewWithSource(`"bad\q"`, "escape.slv"))
	p.ParseProgram()

	if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], `escape.slv:1:1: unknown escape sequence \q`) {
		t.Fatalf("errors are %q, want invalid escape diagnostic", p.Errors())
	}
}
