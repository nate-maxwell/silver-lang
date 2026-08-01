package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestArrayForStatement(t *testing.T) {
	p := New(lexer.New(`for element in values {
print(element)
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ForStatement", program.Statements[0])
	}
	if statement.Key.Value != "element" || statement.Value != nil {
		t.Fatalf("bindings are (%v, %v), want element and no second binding", statement.Key, statement.Value)
	}
	if got := statement.Iterable.String(); got != "values" {
		t.Fatalf("iterable is %q, want values", got)
	}
	if len(statement.Body.Statements) != 1 {
		t.Fatalf("body has %d statements, want 1", len(statement.Body.Statements))
	}
}

func TestMapForStatementUsesUserDefinedNames(t *testing.T) {
	p := New(lexer.New(`for customKey, customValue in entries {
customValue
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ForStatement)
	if statement.Key.Value != "customKey" || statement.Value.Value != "customValue" {
		t.Fatalf("bindings are %q and %q", statement.Key.Value, statement.Value.Value)
	}
}

func TestWhileStatement(t *testing.T) {
	p := New(lexer.New(`while count < 3 {
let count = count + 1
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.WhileStatement", program.Statements[0])
	}
	if !testInfixExpression(t, statement.Condition, "count", "<", 3) {
		t.FailNow()
	}
}

func TestLoopsRequireHeaderExpressions(t *testing.T) {
	for _, input := range []string{"for item in {}", "while {}"} {
		p := New(lexer.New(input))
		p.ParseProgram()
		if len(p.Errors()) == 0 {
			t.Fatalf("parser accepted %q", input)
		}
		if !strings.Contains(p.Errors()[0], "expected") {
			t.Fatalf("unexpected parser errors for %q: %v", input, p.Errors())
		}
	}
}
