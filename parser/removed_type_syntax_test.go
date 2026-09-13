package parser

import (
	"silver/lexer"
	"testing"
)

func TestRemovedTypeSyntaxIsRejected(t *testing.T) {
	for _, input := range []string{
		"struct Point { x: float, y: float }",
		"enum Direction { North, South }",
		"let Names = <array[str]>",
		"let Transform = <call(int) int>",
		"<int>",
		"let contracts = [<int>]",
		"let make = fn() { struct Point {} }",
		"let make = fn() { enum Direction { North } }",
		"let make = fn() { let Names = <array[str]> }",
		"type Point = struct Named {}",
		"type Direction = enum Named { North }",
		"type Names = <array[str]>",
		"let Point = struct { x: float }",
		"let type = int",
		"type(42)",
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatal("removed syntax was accepted")
			}
		})
	}
}

func TestComparisonOperatorsRemainExpressions(t *testing.T) {
	p := New(lexer.New("let below = 1 < 2\nlet above = 2 > 1\nlet same = 2 >= 2"))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if len(program.Statements) != 3 {
		t.Fatalf("got %d statements, want 3", len(program.Statements))
	}
}
