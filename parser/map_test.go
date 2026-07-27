package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestParsingHashLiteralsWithExpressions(t *testing.T) {
	p := New(lexer.New(`{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	hash := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.HashLiteral)
	if len(hash.Pairs) != 3 {
		t.Fatalf("hash has %d pairs, want 3", len(hash.Pairs))
	}
	tests := map[string]func(ast.Expression){
		"one":   func(expression ast.Expression) { testInfixExpression(t, expression, 0, "+", 1) },
		"two":   func(expression ast.Expression) { testInfixExpression(t, expression, 10, "-", 8) },
		"three": func(expression ast.Expression) { testInfixExpression(t, expression, 15, "/", 5) },
	}
	for key, value := range hash.Pairs {
		tests[key.(*ast.StringLiteral).String()](value)
	}
}

func TestParsingEmptyHashLiteral(t *testing.T) {
	p := New(lexer.New("{}"))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	hash := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.HashLiteral)
	if len(hash.Pairs) != 0 {
		t.Fatalf("hash has %d pairs, want 0", len(hash.Pairs))
	}
}

func TestParsingHashLiteralsStringKeys(t *testing.T) {
	p := New(lexer.New(`{"one": 1, "two": 2, "three": 3}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	hash := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.HashLiteral)
	expected := map[string]int64{"one": 1, "two": 2, "three": 3}
	for key, value := range hash.Pairs {
		testIntegerLiteral(t, value, expected[key.(*ast.StringLiteral).String()])
	}
}
