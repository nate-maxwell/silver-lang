package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestOperatorDeclarationAndUse(t *testing.T) {
	p := New(lexer.New(`operator |> = fn(left, right: call) | PossibleError {
    return right(left)
}
first |> second |> third`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	declaration, ok := program.Statements[0].(*ast.OperatorStatement)
	if !ok {
		t.Fatalf("first statement is %T, want *ast.OperatorStatement", program.Statements[0])
	}
	if declaration.Symbol != "|>" {
		t.Fatalf("operator is %q, want |>", declaration.Symbol)
	}
	if len(declaration.Function.Parameters) != 2 {
		t.Fatalf("operator has %d parameters, want 2", len(declaration.Function.Parameters))
	}
	if len(declaration.Function.ErrorTypes) != 1 || declaration.Function.ErrorTypes[0].String() != "PossibleError" {
		t.Fatalf("operator errors are %#v, want PossibleError", declaration.Function.ErrorTypes)
	}
	if got, want := program.Statements[1].String(), "((first |> second) |> third)"; got != want {
		t.Fatalf("operator expression is %q, want %q", got, want)
	}
}

func TestOperatorUsesLowestInfixPrecedence(t *testing.T) {
	p := New(lexer.New(`operator @ = fn(left, right) { return left }
a + b @ c < d`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if got, want := program.Statements[1].String(), "((a + b) @ (c < d))"; got != want {
		t.Fatalf("precedence expression is %q, want %q", got, want)
	}
}

func TestExistingDelimiterCanBeDeclaredAsOperator(t *testing.T) {
	p := New(lexer.New(`operator :: = fn(left, right) { return left }
left :: right`))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if got, want := program.Statements[1].String(), "(left :: right)"; got != want {
		t.Fatalf("expression is %q, want %q", got, want)
	}
}

func TestOperatorFunctionRequiresTwoParameters(t *testing.T) {
	p := New(lexer.New(`operator ;; = fn(value) { value }`))
	p.ParseProgram()
	if got := strings.Join(p.Errors(), "\n"); !strings.Contains(got, "exactly two non-variadic parameters") {
		t.Fatalf("errors are %q, want operator arity diagnostic", got)
	}
}

func TestOperatorMustBeDeclaredBeforeUse(t *testing.T) {
	p := New(lexer.New(`left |> right
operator |> = fn(left, right) { return left }`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted an operator before its declaration")
	}
}

func TestOperatorCannotBeRedefined(t *testing.T) {
	p := New(lexer.New(`operator @ = fn(left, right) { return left }
operator @ = fn(left, right) { return right }`))
	p.ParseProgram()
	if got := strings.Join(p.Errors(), "\n"); !strings.Contains(got, `operator "@" is already defined`) {
		t.Fatalf("errors are %q, want duplicate-operator diagnostic", got)
	}
}

func TestLanguageOperatorCannotBeRedefined(t *testing.T) {
	p := New(lexer.New(`operator + = fn(left, right) { return left }`))
	p.ParseProgram()
	if got := strings.Join(p.Errors(), "\n"); !strings.Contains(got, `operator "+" is already defined by the language`) {
		t.Fatalf("errors are %q, want language-operator diagnostic", got)
	}
}

func TestOperatorBindingPowerCannotBeSpecified(t *testing.T) {
	p := New(lexer.New(`operator @ 5 = fn(left, right) { return left }`))
	p.ParseProgram()
	if len(p.Errors()) == 0 {
		t.Fatal("parser accepted a user-specified operator precedence")
	}
}

func TestOperatorSpellingCanEndWithEquals(t *testing.T) {
	p := New(lexer.New(`operator ??= = fn(left, right) { return left }
left ??= right`))
	program := p.ParseProgram()
	checkParserErrors(t, p)
	if got, want := program.Statements[1].String(), "(left ??= right)"; got != want {
		t.Fatalf("expression is %q, want %q", got, want)
	}
}
