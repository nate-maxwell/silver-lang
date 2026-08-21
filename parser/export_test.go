package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestExportStatement(t *testing.T) {
	p := New(lexer.New(`export {
    symbol_a,
    symbol_b,
}`))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program has %d statements, want 1", len(program.Statements))
	}
	statement, ok := program.Statements[0].(*ast.ExportStatement)
	if !ok {
		t.Fatalf("statement is %T, want *ast.ExportStatement", program.Statements[0])
	}
	if len(statement.Names) != 2 || statement.Names[0].Value != "symbol_a" || statement.Names[1].Value != "symbol_b" {
		t.Fatalf("export names are %#v, want symbol_a and symbol_b", statement.Names)
	}
	if got, want := statement.String(), "export {symbol_a, symbol_b}"; got != want {
		t.Fatalf("statement string is %q, want %q", got, want)
	}
}

func TestEmptyExportStatement(t *testing.T) {
	p := New(lexer.New("export {}"))
	program := p.ParseProgram()
	checkParserErrors(t, p)

	statement := program.Statements[0].(*ast.ExportStatement)
	if len(statement.Names) != 0 {
		t.Fatalf("empty export has %d names", len(statement.Names))
	}
}

func TestExportStatementValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
	}{
		{"duplicate name", "export { value, value }", "duplicate exported symbol"},
		{"duplicate declaration", "export { first }\nexport { second }", "only one export declaration"},
		{"nested declaration", "if True { export { value } }", "only valid at the top level"},
		{"missing body", "export value", "expected next token to be {"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := New(lexer.New(test.input))
			p.ParseProgram()
			if len(p.Errors()) == 0 || !strings.Contains(strings.Join(p.Errors(), "\n"), test.message) {
				t.Fatalf("errors are %v, want one containing %q", p.Errors(), test.message)
			}
		})
	}
}
