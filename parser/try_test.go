package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestTryCatchParsing(t *testing.T) {
	p := New(lexer.New(`try {
read()
} catch FileNotFound err {
err.message
} catch io.PermissionDenied err {
err.message
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	expression, ok := statement.Expression.(*ast.TryExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.TryExpression", statement.Expression)
	}
	if len(expression.Catches) != 2 {
		t.Fatalf("got %d catches, want 2", len(expression.Catches))
	}
	first := expression.Catches[0]
	if first.ErrorType.String() != "FileNotFound" || first.Binding.Value != "err" {
		t.Fatalf("first catch parsed as %#v", first)
	}
	if got := expression.Catches[1].ErrorType.String(); got != "io.PermissionDenied" {
		t.Fatalf("second catch type is %q", got)
	}
}

func TestTryRequiresCatch(t *testing.T) {
	p := New(lexer.New(`try { 42 }`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted try without catch")
	}
}

func TestCollectInsideTryParticipatesInTaskUsageValidation(t *testing.T) {
	p := New(lexer.New(`
let work = fn() int { 1 }
let handle = task work
try {
	collect handle
} catch IOError err {
	False
}
collect handle
`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser did not detect task collected again after try")
	}
}
