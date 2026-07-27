package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestParsingArrayLiterals(t *testing.T) {
	p := New(lexer.New("[1, 2 * 2, 3 + 3]"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	array := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.ArrayLiteral)
	if len(array.Elements) != 3 {
		t.Fatalf("array has %d elements, want 3", len(array.Elements))
	}
	testIntegerLiteral(t, array.Elements[0], 1)
	testInfixExpression(t, array.Elements[1], 2, "*", 2)
	testInfixExpression(t, array.Elements[2], 3, "+", 3)
}
