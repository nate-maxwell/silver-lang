package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestStringLiteralExpression(t *testing.T) {
	p := New(lexer.New(`"hello world"`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.StringLiteral)
	if literal.Value != "hello world" {
		t.Fatalf("literal value is %q, want hello world", literal.Value)
	}
}

func TestStringLiteralEscapesAreDecoded(t *testing.T) {
	p := New(lexer.New(`"hello\n\t\"Silver\""`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.StringLiteral)
	if got, want := literal.Value, "hello\n\t\"Silver\""; got != want {
		t.Fatalf("literal value is %q, want %q", got, want)
	}
}

func TestInvalidStringEscapeReportsDiagnostic(t *testing.T) {
	p := New(lexer.NewWithSource(`"bad\q"`, "escape.slv"))
	p.ParseProgram()

	if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], `escape.slv:1:1: unknown escape sequence \q`) {
		t.Fatalf("errors are %q, want invalid escape diagnostic", p.Errors())
	}
}

func TestTemplateStringLiteralParsesTextAndExpressions(t *testing.T) {
	p := New(lexer.New("```hello {name + \"!\"}```"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal, ok := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.TemplateStringLiteral)
	if !ok {
		t.Fatalf("expression is %T, want *ast.TemplateStringLiteral", program.Statements[0])
	}
	if len(literal.Parts) != 2 || literal.Parts[0].Text != "hello " {
		t.Fatalf("parts are %#v, want text followed by interpolation", literal.Parts)
	}
	expression, ok := literal.Parts[1].Expression.(*ast.InfixExpression)
	if !ok || expression.Operator != "+" {
		t.Fatalf("interpolation is %T (%v), want addition", literal.Parts[1].Expression, literal.Parts[1].Expression)
	}
}

func TestTemplateStringLiteralParsesNestedBraces(t *testing.T) {
	input := "```value: {maps.get({\"answer\": 42}, \"answer\")}```"
	p := New(lexer.New(input))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	literal := program.Statements[0].(*ast.ExpressionStatement).Expression.(*ast.TemplateStringLiteral)
	if len(literal.Parts) != 2 || literal.Parts[1].Expression == nil {
		t.Fatalf("parts are %#v, want nested map expression", literal.Parts)
	}
}

func TestTemplateStringLiteralRejectsMalformedInterpolation(t *testing.T) {
	tests := []string{
		"```empty: {}```",
		"```missing close: {value```",
		"```unmatched: }```",
		"```unterminated",
	}
	for _, input := range tests {
		p := New(lexer.NewWithSource(input, "template.slv"))
		p.ParseProgram()
		if len(p.Errors()) == 0 {
			t.Errorf("parser accepted malformed template %q", input)
		}
	}
}
