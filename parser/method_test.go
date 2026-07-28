package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestStructMethodFunctionLiteral(t *testing.T) {
	p := New(lexer.New(`fn[Person](greeting: str): str { greeting }`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	expression := program.Statements[0].(*ast.ExpressionStatement)
	function := expression.Expression.(*ast.FunctionLiteral)
	if function.ReceiverType == nil || function.ReceiverType.String() != "Person" {
		t.Fatalf("receiver type is %v, want Person", function.ReceiverType)
	}
	if len(function.Parameters) != 1 || function.Parameters[0].Type.String() != "str" {
		t.Fatalf("parameters are %#v, want one str parameter", function.Parameters)
	}
	if function.ReturnType == nil || function.ReturnType.String() != "str" {
		t.Fatalf("return type is %v, want str", function.ReturnType)
	}
	if got, want := function.String(), "fn[Person](greeting: str): str greeting"; got != want {
		t.Fatalf("function string is %q, want %q", got, want)
	}
}

func TestQualifiedStructMethodReceiver(t *testing.T) {
	p := New(lexer.New(`fn[models.Person]() {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	function := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
	if got := function.ReceiverType.String(); got != "models.Person" {
		t.Fatalf("receiver type is %q, want models.Person", got)
	}
}
