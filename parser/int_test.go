package parser

import (
	"fmt"
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestIntegerLiteralExpression(t *testing.T) {
	p := New(lexer.New("5"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.IntegerLiteral)
	if literal.Value != 5 || literal.TokenLiteral() != "5" {
		t.Fatalf("unexpected integer literal: %+v", literal)
	}
}

func testIntegerLiteral(t *testing.T, il ast.Expression, value int64) bool {
	integ, ok := il.(*ast.IntegerLiteral)
	if !ok {
		t.Errorf("il not *ast.IntegerLiteral. got=%T", il)
		return false
	}

	if integ.Value != value {
		t.Errorf("integ.Value not %d. got=%d", value, integ.Value)
		return false
	}

	if integ.TokenLiteral() != fmt.Sprintf("%d", value) {
		t.Errorf("integ.TokenLiteral not %d/ got=%s", value, integ.TokenLiteral())
		return false
	}

	return true
}
