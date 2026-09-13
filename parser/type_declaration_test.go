package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestTypeDeclarations(t *testing.T) {
	for _, definition := range []string{
		"struct { x: float, y: float }", "enum { North, East, South, West }",
		"struct {  }", "enum {  }", "array[str]", "call(Point) Point",
		"int", "models.Point", "array[array[call() int]]", "call()",
		"call(values: int...) array[str] | Failure", "struct { self: Point, item:: Item, +: call(Point) Point }",
	} {
		t.Run(definition, func(t *testing.T) {
			input := "type Point = " + definition
			p := New(lexer.NewWithSource(input+"\nlet next = 42", "types.slv"))
			program := p.ParseProgram()
			checkParserErrors(t, p)
			if len(program.Statements) != 2 {
				t.Fatalf("got %d statements, want 2", len(program.Statements))
			}
			statement, ok := program.Statements[0].(*ast.TypeStatement)
			if !ok {
				t.Fatalf("got %T, want type declaration", program.Statements[0])
			}
			if statement.Name.Value != "Point" || statement.TokenLiteral() != "type" || statement.String() != input {
				t.Fatalf("unexpected declaration: %#v, %q", statement, statement.String())
			}
			if position := statement.Position(); position.Source != "types.slv" || position.Line != 1 || position.Column != 1 {
				t.Fatalf("unexpected declaration position: %+v", position)
			}
			switch {
			case strings.HasPrefix(definition, "struct"):
				if _, ok := statement.Value.(*ast.StructTypeLiteral); !ok {
					t.Fatalf("got %T, want struct definition", statement.Value)
				}
			case strings.HasPrefix(definition, "enum"):
				if _, ok := statement.Value.(*ast.EnumTypeLiteral); !ok {
					t.Fatalf("got %T, want enum definition", statement.Value)
				}
			default:
				if _, ok := statement.Value.(*ast.TypeAnnotation); !ok {
					t.Fatalf("got %T, want type contract", statement.Value)
				}
			}
		})
	}
}

func TestMalformedTypeDeclarations(t *testing.T) {
	for _, input := range []string{
		"type Point", "type Point struct {}", "type Point: int = int",
		"type Point =", "type Point = 42", "type Point = fn() {}", "type Point = <int>",
		"type Point = struct Named {}", "type Point = enum Named {}",
		"type Point = struct { x: int, x: str }", "type Point = enum { One, One }",
		"type Point = struct { x: }", "type Point = struct { x: int",
		"type Point = enum { One", "type Point = enum { 1 }",
		"type Point = array[]", "type Point = array[int, str]", "type Point = call(int... , str)",
		"type Point = call() Missing extra", "type Point = int + str",
		"type Point = int let value = 1", "type Point = enum {} let value = 1",
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatal("invalid type declaration was accepted")
			}
		})
	}
}

func TestMultilineTypeDeclarations(t *testing.T) {
	p := New(lexer.New(`type Point = struct {
    x: float
    y: float
}
type Direction = enum {
    North
    South,
}
type Transform = call(
    Point,
    array[int]
) Point
let heading = Direction.North
`))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 4 {
		t.Fatalf("got %d statements, want 4", len(program.Statements))
	}
}
