package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestEnumStatement(t *testing.T) {
	p := New(lexer.New(`
enum Direction { North, East, South, West, }
let direction = Direction.North
`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program has %d statements, want 2", len(program.Statements))
	}
	statement, ok := program.Statements[0].(*ast.EnumStatement)
	if !ok {
		t.Fatalf("first statement is %T, want *ast.EnumStatement", program.Statements[0])
	}
	if statement.Name.Value != "Direction" {
		t.Fatalf("enum name is %q, want Direction", statement.Name.Value)
	}

	wantMembers := []string{"North", "East", "South", "West"}
	if len(statement.Members) != len(wantMembers) {
		t.Fatalf("enum has %d members, want %d", len(statement.Members), len(wantMembers))
	}
	for i, want := range wantMembers {
		if statement.Members[i].Value != want {
			t.Fatalf("member %d is %q, want %q", i, statement.Members[i].Value, want)
		}
	}
	if got, want := statement.String(), "enum Direction { North, East, South, West }"; got != want {
		t.Fatalf("enum string is %q, want %q", got, want)
	}
}

func TestEnumRejectsDuplicateMembers(t *testing.T) {
	p := New(lexer.NewWithSource(`enum State { Ready, Ready }`, "duplicate.slvr"))
	p.ParseProgram()

	if len(p.Errors()) != 1 {
		t.Fatalf("parser returned %d errors, want 1: %v", len(p.Errors()), p.Errors())
	}
	if message := p.Errors()[0]; !strings.Contains(message, `duplicate enum member "Ready"`) || !strings.Contains(message, "duplicate.slvr:1:21") {
		t.Fatalf("unexpected parser error: %q", message)
	}
}
