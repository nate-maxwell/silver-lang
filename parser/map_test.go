package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
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

func TestMapLiteralRejectsDuplicateKeys(t *testing.T) {
	for _, input := range []string{
		`let t = {1: 10, 1: 20}`,
		`let t = {"key": 10, "key": 20}`,
		`let t = {"\t": 10, "\x09": 20}`,
		`let t = {True: 10, True: 20}`,
		`let t = {False: 10, False: 20}`,
		`let t = {1.5: 10, 1.50: 20}`,
		`let t = {1: 10, 1.0: 20}`,
		`let t = {1.0: 10, 1: 20}`,
		`let t = {-1: 10, -1.0: 20}`,
		`let t = {-1.5: 10, -1.50: 20}`,
		`let t = {0: 10, -0.0: 20}`,
		`let t = {(1): 10, 1: 20}`,
		`let t = {"nested": {1: 10, 1: 20}}`,
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			p.ParseProgram()
			if errors := p.Errors(); len(errors) != 1 || !strings.Contains(errors[0], "duplicate map key") {
				t.Fatalf("parser errors are %v, want one duplicate-map-key error", errors)
			}
		})
	}
}

func TestMapLiteralDuplicateKeyPosition(t *testing.T) {
	p := New(lexer.NewWithSource("let t = {\n  1: 10,\n  1: 20,\n}\nlet next = 30", "duplicate.slv"))
	program := p.ParseProgram()
	if errors := p.Errors(); len(errors) != 1 || errors[0] != "duplicate.slv:3:3: duplicate map key 1" {
		t.Fatalf("unexpected parser errors: %v", errors)
	}
	if len(program.Statements) != 2 {
		t.Fatalf("parser returned %d statements, want 2", len(program.Statements))
	}
}

func TestMapLiteralAllowsDistinctAndDynamicKeys(t *testing.T) {
	for _, input := range []string{
		`let t = {1: 10, "1": 20, True: 30, False: 40, 0: 50, "True": 60}`,
		`let t = {1: 10, -1: 20, 1.5: 30}`,
		`let t = {9007199254740993: 10, 9007199254740992.0: 20}`,
		`let t = {"first": {1: 10}, "second": {1: 20}}`,
		`let t = {key: 10, key: 20}`,
		`let t = {next(): 10, next(): 20}`,
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			p.ParseProgram()
			checkParserErrors(t, p)
		})
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
