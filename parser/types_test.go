package parser

import (
	"silver/ast"
	"silver/lexer"
	"testing"
)

func TestTypeAliasLiterals(t *testing.T) {
	for _, contract := range []string{
		"int", "models.Token", "array[str]", "array[array[int]]",
		"array[call(str) array[models.Token]]", "call()", "call(array[int]) array[str]",
		"call(values: array[Token]...) array[Node] | ParseError", "EarlierAlias",
	} {
		t.Run(contract, func(t *testing.T) {
			p := New(lexer.New("let signature = <" + contract + ">"))
			program := p.ParseProgram()
			checkParserErrors(t, p)
			literal := program.Statements[0].(*ast.LetStatement).Value.(*ast.TypeAliasLiteral)
			if got := literal.Annotation.String(); got != contract {
				t.Fatalf("contract = %q, want %q", got, contract)
			}
			if got := literal.String(); got != "<"+contract+">" {
				t.Fatalf("literal = %q", got)
			}
		})
	}
}

func TestArrayTypeAnnotations(t *testing.T) {
	p := New(lexer.New(`fn(values: array[array[str]]) array[call(int) bool] {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	function := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
	if got := function.Parameters[0].Type.ElementType.ElementType.String(); got != "str" {
		t.Fatalf("nested element type = %q", got)
	}
	if !function.ReturnType.ElementType.IsCallSignature() {
		t.Fatal("return array element is not a callable signature")
	}
}

func TestMalformedTypeAliasesAndArrayTypes(t *testing.T) {
	for _, input := range []string{
		"let T = <>", "let T = <int", "let T = <array[]>", "let T = <array[int, str]>",
		"let T = <array[int>", "let T = <map[str]>", "let T = <call(str) array[]>",
		"let values: array[] = []", "let values: map[str] = {}", "let T = <int + str>",
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatal("invalid type syntax was accepted")
			}
		})
	}
}

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
	p := New(lexer.New(`fn(person: models.Person, active: bool) str { person.name }`))
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

func TestVariadicFunctionParameter(t *testing.T) {
	p := New(lexer.New(`fn(prefix: str, parts: str...) {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	function := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
	if len(function.Parameters) != 2 || !function.Parameters[1].Variadic {
		t.Fatalf("parameters are %#v, want a variadic final parameter", function.Parameters)
	}
	if got, want := function.String(), "fn(prefix: str, parts: str...) "; got != want {
		t.Fatalf("function is %q, want %q", got, want)
	}
}

func TestUntypedVariadicFunctionParameter(t *testing.T) {
	p := New(lexer.New(`fn(values...) {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	parameter := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral).Parameters[0]
	if !parameter.Variadic || parameter.Type != nil {
		t.Fatalf("parameter is %#v, want untyped variadic", parameter)
	}
}

func TestVariadicFunctionParameterMustBeLast(t *testing.T) {
	p := New(lexer.New(`fn(parts: str..., suffix: str) {}`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted a non-final variadic parameter")
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

func TestVariadicCallSignatureTypeAnnotation(t *testing.T) {
	p := New(lexer.New(`let callback: call(prefix: str, parts: str...) = fn(prefix: str, parts: str...) {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	signature := program.Statements[0].(*ast.LetStatement).Name.Type
	if !signature.Variadic {
		t.Fatal("call signature is not variadic")
	}
	if got, want := signature.String(), "call(prefix: str, parts: str...)"; got != want {
		t.Fatalf("signature is %q, want %q", got, want)
	}
}

func TestVariadicCallSignatureParameterMustBeLast(t *testing.T) {
	p := New(lexer.New(`let callback: call(parts: str..., suffix: str) = fn() {}`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted a non-final variadic call parameter")
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

func TestFunctionReturnTypeAlternatives(t *testing.T) {
	p := New(lexer.New(`fn(filepath: str) str | FileNotFound | io.PermissionDenied {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	function := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
	if got, want := function.ReturnType.String(), "str"; got != want {
		t.Fatalf("success type is %q, want %q", got, want)
	}
	if len(function.ErrorTypes) != 2 || function.ErrorTypes[0].String() != "FileNotFound" || function.ErrorTypes[1].String() != "io.PermissionDenied" {
		t.Fatalf("error types are %v", function.ErrorTypes)
	}
	if got, want := function.String(), "fn(filepath: str) str | FileNotFound | io.PermissionDenied "; len(got) < len(want) || got[:len(want)] != want {
		t.Fatalf("function string is %q, want prefix %q", got, want)
	}
}

func TestFunctionReturnUnionMayOmitNullSuccessType(t *testing.T) {
	p := New(lexer.New(`fn(filepath: str) | PermissionDenied {}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	function := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral)
	if function.ReturnType != nil {
		t.Fatalf("success type is %v, want implicit null", function.ReturnType)
	}
	if len(function.ErrorTypes) != 1 || function.ErrorTypes[0].String() != "PermissionDenied" {
		t.Fatalf("error types are %v", function.ErrorTypes)
	}
}

func TestCallSignatureReturnTypeAlternatives(t *testing.T) {
	p := New(lexer.New(`let opener: call(str) str | FileNotFound = fn(filepath: str) str { filepath }`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	signature := program.Statements[0].(*ast.LetStatement).Name.Type
	if got, want := signature.String(), "call(str) str | FileNotFound"; got != want {
		t.Fatalf("signature is %q, want %q", got, want)
	}
}
