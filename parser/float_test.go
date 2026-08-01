package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestFloatLiteralExpression(t *testing.T) {
	p := New(lexer.New("12.25"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := statement.Expression.(*ast.FloatLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.FloatLiteral", statement.Expression)
	}
	if literal.Value != 12.25 || literal.TokenLiteral() != "12.25" {
		t.Fatalf("unexpected float literal: %+v", literal)
	}
}

func TestNegativeFloatParsesAsPrefixExpression(t *testing.T) {
	p := New(lexer.New("-1.5"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	prefix, ok := statement.Expression.(*ast.PrefixExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.PrefixExpression", statement.Expression)
	}
	literal, ok := prefix.Right.(*ast.FloatLiteral)
	if !ok || prefix.Operator != "-" || literal.Value != 1.5 {
		t.Fatalf("unexpected negative float AST: %+v", prefix)
	}
}
