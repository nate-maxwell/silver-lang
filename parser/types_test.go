package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestTypedLetStatement(t *testing.T) {
	p := New(lexer.New(`let age: int = 36`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.LetStatement)
	if statement.Name.Type == nil || statement.Name.Type.String() != "int" {
		t.Fatalf("let type is %v, want int", statement.Name.Type)
	}
	if got, want := statement.String(), "let age: int = 36"; got != want {
		t.Fatalf("let string is %q, want %q", got, want)
	}
}

func TestTypedFunctionLiteral(t *testing.T) {
	p := New(lexer.New(`fn(person: models.Person, active: bool): string { person.name; }`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	expression := program.Statements[0].(*ast.ExpressionStatement)
	function := expression.Expression.(*ast.FunctionLiteral)
	if got := function.Parameters[0].Type.String(); got != "models.Person" {
		t.Fatalf("first parameter type is %q, want models.Person", got)
	}
	if got := function.Parameters[1].Type.String(); got != "bool" {
		t.Fatalf("second parameter type is %q, want bool", got)
	}
	if function.ReturnType == nil || function.ReturnType.String() != "string" {
		t.Fatalf("return type is %v, want string", function.ReturnType)
	}
}
