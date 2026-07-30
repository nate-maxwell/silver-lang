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
	p := New(lexer.New(`fn(person: models.Person, active: bool) str { person.name; }`))
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
	if function.ReturnType == nil || function.ReturnType.String() != "str" {
		t.Fatalf("return type is %v, want str", function.ReturnType)
	}
}

func TestCallSignatureTypeAnnotation(t *testing.T) {
	p := New(lexer.New(`fn(transform: call(int, models.Person) str) call(str) bool {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	expression := program.Statements[0].(*ast.ExpressionStatement)
	function := expression.Expression.(*ast.FunctionLiteral)
	parameterType := function.Parameters[0].Type
	if got, want := parameterType.String(), "call(int, models.Person) str"; got != want {
		t.Fatalf("parameter type is %q, want %q", got, want)
	}
	if len(parameterType.ParameterTypes) != 2 {
		t.Fatalf("parameter signature has %d arguments, want 2", len(parameterType.ParameterTypes))
	}
	if got, want := function.ReturnType.String(), "call(str) bool"; got != want {
		t.Fatalf("return type is %q, want %q", got, want)
	}
}

func TestCallSignatureMayOmitNullReturnType(t *testing.T) {
	p := New(lexer.New(`let callback: call(value: int) = fn(value: int) {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.LetStatement)
	signature := statement.Name.Type
	if signature.ReturnType != nil {
		t.Fatalf("return type is %v, want implicit null", signature.ReturnType)
	}
	if len(signature.ParameterNames) != 1 || signature.ParameterNames[0] != "value" {
		t.Fatalf("parameter names are %v, want [value]", signature.ParameterNames)
	}
	if got, want := signature.String(), "call(value: int)"; got != want {
		t.Fatalf("signature is %q, want %q", got, want)
	}
}
