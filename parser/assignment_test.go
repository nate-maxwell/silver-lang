package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestMemberAssignmentStatement(t *testing.T) {
	p := New(lexer.New(`actor.location = move(actor.location)`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.MemberAssignmentStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.MemberAssignmentStatement", program.Statements[0])
	}
	if got, want := statement.Target.String(), "actor.location"; got != want {
		t.Fatalf("target is %q, want %q", got, want)
	}
	if got, want := statement.Value.String(), "move(actor.location)"; got != want {
		t.Fatalf("value is %q, want %q", got, want)
	}
	if got, want := statement.String(), "actor.location = move(actor.location)"; got != want {
		t.Fatalf("statement is %q, want %q", got, want)
	}
}

func TestVariableAssignmentStatement(t *testing.T) {
	p := New(lexer.New(`value = 42`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement, ok := program.Statements[0].(*ast.AssignmentStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.AssignmentStatement", program.Statements[0])
	}
	if statement.Name.Value != "value" || statement.Value.String() != "42" {
		t.Fatalf("unexpected assignment: %s", statement.String())
	}
}

func TestAssignmentRejectsNonAssignableExpression(t *testing.T) {
	p := New(lexer.New(`call() = 42`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted a call as an assignment target")
	}
}
