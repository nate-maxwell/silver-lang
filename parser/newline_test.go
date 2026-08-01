package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestNewlinesSeparateStatements(t *testing.T) {
	p := New(lexer.New("let age: int = 36\nlet name: str = \"Ada\"\nage"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 3 {
		t.Fatalf("program has %d statements, want 3", len(program.Statements))
	}
	if got, want := program.String(), "let age: int = 36\nlet name: str = Ada\nage"; got != want {
		t.Fatalf("program string is %q, want %q", got, want)
	}
}

func TestSameLineStatementsRequireSeparator(t *testing.T) {
	p := New(lexer.New(`let first = 1 let second = 2`))
	p.ParseProgram()

	if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], "expected newline between statements") {
		t.Fatalf("parser errors are %v, want a statement-boundary error", p.Errors())
	}
}

func TestSemicolonsAreRejected(t *testing.T) {
	p := New(lexer.New("let first = 1;\nlet second = 2"))
	p.ParseProgram()

	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted a semicolon")
	}
}

func TestNewlineEndsCallExpression(t *testing.T) {
	p := New(lexer.New("callable\n(42)"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program has %d statements, want 2", len(program.Statements))
	}
	first := program.Statements[0].(*ast.ExpressionStatement)
	if _, ok := first.Expression.(*ast.Identifier); !ok {
		t.Fatalf("first expression is %T, want *ast.Identifier", first.Expression)
	}
}

func TestEnumMembersMayBeNewlineSeparated(t *testing.T) {
	p := New(lexer.New("enum Direction {\nNorth\nSouth\n}"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	declaration := program.Statements[0].(*ast.EnumStatement)
	if len(declaration.Members) != 2 {
		t.Fatalf("enum has %d members, want 2", len(declaration.Members))
	}
}
