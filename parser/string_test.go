package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestStringLiteralExpression(t *testing.T) {
	p := New(lexer.New(`"hello world";`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.StringLiteral)
	if literal.Value != "hello world" {
		t.Fatalf("literal value is %q, want hello world", literal.Value)
	}
}
