package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestParsingMapLiteralsWithExpressions(t *testing.T) {
	p := New(lexer.New(`{"one": 0 + 1, "two": 10 - 8, "three": 15 / 5}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	mapLiteral := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MapLiteral)
	if len(mapLiteral.Pairs) != 3 {
		t.Fatalf("map has %d pairs, want 3", len(mapLiteral.Pairs))
	}
	tests := map[string]func(ast.Expression){
		"one":   func(expression ast.Expression) { testInfixExpression(t, expression, 0, "+", 1) },
		"two":   func(expression ast.Expression) { testInfixExpression(t, expression, 10, "-", 8) },
		"three": func(expression ast.Expression) { testInfixExpression(t, expression, 15, "/", 5) },
	}
	for key, value := range mapLiteral.Pairs {
		tests[key.(*ast.StringLiteral).String()](value)
	}
}

func TestParsingEmptyMapLiteral(t *testing.T) {
	p := New(lexer.New("{}"))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	mapLiteral := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MapLiteral)
	if len(mapLiteral.Pairs) != 0 {
		t.Fatalf("map has %d pairs, want 0", len(mapLiteral.Pairs))
	}
}

func TestParsingMapLiteralsStringKeys(t *testing.T) {
	p := New(lexer.New(`{"one": 1, "two": 2, "three": 3}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	mapLiteral := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.MapLiteral)
	expected := map[string]int64{"one": 1, "two": 2, "three": 3}
	for key, value := range mapLiteral.Pairs {
		testIntegerLiteral(t, value, expected[key.(*ast.StringLiteral).String()])
	}
}
