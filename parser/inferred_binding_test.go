package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestBindingDeclarationForms(t *testing.T) {
	for _, tt := range []struct {
		input    string
		inferred bool
		typed    bool
	}{
		{"let value := 32", true, false},
		{"let value = 32", false, false},
		{"let value: int = 32", false, true},
		{"let value: any = 32", false, true},
	} {
		t.Run(tt.input, func(t *testing.T) {
			p := New(lexer.New(tt.input))
			program := p.ParseProgram()
			checkParserErrors(t, p)
			statement := program.Statements[0].(*ast.LetStatement)
			if statement.Inferred != tt.inferred || (statement.Name.Type != nil) != tt.typed {
				t.Fatalf("unexpected binding mode: %#v", statement)
			}
			if got := statement.String(); got != tt.input {
				t.Fatalf("binding string = %q, want %q", got, tt.input)
			}
		})
	}
}

func TestInvalidInferredDeclarations(t *testing.T) {
	for _, input := range []string{
		"let value: int := 32", "let value :=", "let := 32",
		"value := 32", "let value : = 32", "type Value := int",
	} {
		t.Run(input, func(t *testing.T) {
			p := New(lexer.New(input))
			p.ParseProgram()
			if len(p.Errors()) == 0 {
				t.Fatal("invalid declaration was accepted")
			}
		})
	}
}

func TestInferredBindingOperatorIsReserved(t *testing.T) {
	p := New(lexer.New(`operator := = fn(left, right) { return left }`))
	p.ParseProgram()
	if got := strings.Join(p.Errors(), "\n"); !strings.Contains(got, `operator ":=" is already defined by the language`) {
		t.Fatalf("errors = %q, want reserved operator diagnostic", got)
	}
}
