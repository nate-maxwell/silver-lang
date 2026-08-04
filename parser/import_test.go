package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestImportAndMemberExpression(t *testing.T) {
	p := New(lexer.New("let math = import(\"./math.slv\")\nmath.add(1, 2)"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	if len(program.Statements) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(program.Statements))
	}

	letStatement, ok := program.Statements[0].(*ast.LetStatement)
	if !ok {
		t.Fatalf("first statement is %T, want *ast.LetStatement", program.Statements[0])
	}
	importExpression, ok := letStatement.Value.(*ast.ImportExpression)
	if !ok {
		t.Fatalf("let value is %T, want *ast.ImportExpression", letStatement.Value)
	}
	if importExpression.Path.Value != "./math.slv" {
		t.Fatalf("import path is %q", importExpression.Path.Value)
	}

	expressionStatement := program.Statements[1].(*ast.ExpressionStatement)
	call, ok := expressionStatement.Expression.(*ast.CallExpression)
	if !ok {
		t.Fatalf("expression is %T, want *ast.CallExpression", expressionStatement.Expression)
	}
	member, ok := call.Function.(*ast.MemberExpression)
	if !ok {
		t.Fatalf("call target is %T, want *ast.MemberExpression", call.Function)
	}
	if member.Object.String() != "math" || member.Member.Value != "add" {
		t.Fatalf("unexpected member expression: %s", member.String())
	}
}
