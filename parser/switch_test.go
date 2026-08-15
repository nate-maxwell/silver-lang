package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestSwitchExpression(t *testing.T) {
	p := New(lexer.New(`let label = switch value {
case 1:
    "one"
case candidate():
    "candidate"
default:
    "other"
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.LetStatement)
	expression, ok := statement.Value.(*ast.SwitchExpression)
	if !ok {
		t.Fatalf("binding value is %T, want *ast.SwitchExpression", statement.Value)
	}
	if !testIdentifier(t, expression.Value, "value") {
		t.FailNow()
	}
	if len(expression.Cases) != 2 {
		t.Fatalf("switch has %d cases, want 2", len(expression.Cases))
	}
	testLiteralExpression(t, expression.Cases[0].Value, 1)
	if got := expression.Cases[1].Value.String(); got != "candidate()" {
		t.Fatalf("second case is %q, want candidate()", got)
	}
	if len(expression.Cases[0].Body.Statements) != 1 || len(expression.Cases[1].Body.Statements) != 1 {
		t.Fatal("case bodies were not parsed independently")
	}
	if expression.Default == nil || len(expression.Default.Statements) != 1 {
		t.Fatal("default body was not parsed")
	}
}

func TestSwitchWithoutDefault(t *testing.T) {
	p := New(lexer.New(`switch value {
case 1:
    "one"
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	expression := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.SwitchExpression)
	if expression.Default != nil {
		t.Fatal("switch unexpectedly has a default")
	}
}

func TestSwitchRejectsCaseTypeAnnotation(t *testing.T) {
	p := New(lexer.New(`switch value {
case candidate: int:
    candidate
}`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted a type annotation on a case")
	}
}

func TestSwitchRequiresDefaultToBeLast(t *testing.T) {
	p := New(lexer.New(`switch value {
default:
    0
case 1:
    1
}`))
	p.ParseProgram()
	if len(p.Errors()) == 0 || !strings.Contains(strings.Join(p.Errors(), " "), "before default") {
		t.Fatalf("unexpected parser errors: %v", p.Errors())
	}
}
