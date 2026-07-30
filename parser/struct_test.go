package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestStructStatement(t *testing.T) {
	p := New(lexer.New(`
	struct Point {
		x: int
		y: int
	}
	let point: Point = Point{1, 2}
`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 2 {
		t.Fatalf("program has %d statements, want 2", len(program.Statements))
	}
	statement, ok := program.Statements[0].(*ast.StructStatement)
	if !ok {
		t.Fatalf("first statement is %T, want *ast.StructStatement", program.Statements[0])
	}
	if statement.Name.Value != "Point" {
		t.Fatalf("struct name is %q, want Point", statement.Name.Value)
	}

	wantFields := []string{"x", "y"}
	if len(statement.Fields) != len(wantFields) {
		t.Fatalf("struct has %d fields, want %d", len(statement.Fields), len(wantFields))
	}
	for i, want := range wantFields {
		if statement.Fields[i].Value != want {
			t.Fatalf("field %d is %q, want %q", i, statement.Fields[i].Value, want)
		}
		if statement.Fields[i].Type == nil || statement.Fields[i].Type.String() != "int" {
			t.Fatalf("field %d type is %v, want int", i, statement.Fields[i].Type)
		}
	}
	if got, want := statement.String(), "struct Point { x: int, y: int }"; got != want {
		t.Fatalf("struct string is %q, want %q", got, want)
	}
}

func TestStructRejectsDuplicateFields(t *testing.T) {
	p := New(lexer.NewWithSource(`struct Pair { x: int, x: int }`, "duplicate.slvr"))
	p.ParseProgram()

	if len(p.Errors()) != 1 {
		t.Fatalf("parser returned %d errors, want 1: %v", len(p.Errors()), p.Errors())
	}
	if message := p.Errors()[0]; !strings.Contains(message, `duplicate struct field "x"`) || !strings.Contains(message, "duplicate.slvr:1:23") {
		t.Fatalf("unexpected parser error: %q", message)
	}
}

func TestStructLiteral(t *testing.T) {
	p := New(lexer.New(`Location{0.0, 1.0, 2.0}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExpressionStatement)
	literal, ok := statement.Expression.(*ast.StructLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.StructLiteral", statement.Expression)
	}
	if !testIdentifier(t, literal.StructType, "Location") {
		return
	}
	if len(literal.Values) != 3 {
		t.Fatalf("literal has %d values, want 3", len(literal.Values))
	}
	if got, want := literal.String(), "Location{0.0, 1.0, 2.0}"; got != want {
		t.Fatalf("literal string is %q, want %q", got, want)
	}
}

func TestStructLiteralRequiresCommaSeparatedValues(t *testing.T) {
	for _, input := range []string{
		`Location{0.0 1.0, 2.0}`,
		"Location{\n0.0\n1.0\n2.0\n}",
	} {
		p := New(lexer.New(input))
		p.ParseProgram()
		if len(p.Errors()) == 0 {
			t.Fatalf("parser accepted struct values without commas: %q", input)
		}
	}
}
