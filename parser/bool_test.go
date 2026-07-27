package parser

import (
	"fmt"
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestBooleanExpression(t *testing.T) {
	for _, tt := range []struct {
		input string
		want  bool
	}{{"True;", true}, {"False;", false}} {
		p := New(lexer.New(tt.input))
		program := p.ParseProgram()
		checkParserErrors(t, p)
		boolean := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.Boolean)
		if boolean.Value != tt.want {
			t.Fatalf("boolean value is %t, want %t", boolean.Value, tt.want)
		}
	}
}

func testBooleanLiteral(t *testing.T, exp ast.Expression, value bool) bool {
	b, ok := exp.(*ast.Boolean)
	if !ok {
		t.Errorf("exp not *ast.Boolean. got=%T", exp)
		return false
	}

	if b.Value != value {
		t.Errorf("b.Value not %t. got=%t", value, b.Value)
		return false
	}

	if strings.ToLower(b.TokenLiteral()) != fmt.Sprintf("%t", value) {
		t.Errorf("b.TokenLiteral not %t. got=%s", value, b.TokenLiteral())
		return false
	}

	return true
}
