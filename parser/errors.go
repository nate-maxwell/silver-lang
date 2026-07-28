package parser

import (
	"fmt"
	"silver/token"
)

// Errors returns the parser diagnostics accumulated so far.
func (p *Parser) Errors() []string {
	return p.errors
}

// peekError records an unexpected-lookahead diagnostic.
func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.addError(p.peekToken.Position, msg)
}

// addError prefixes a diagnostic with source coordinates when available.
func (p *Parser) addError(position token.Position, message string) {
	if position.IsValid() {
		message = fmt.Sprintf("%s:%d:%d: %s", position.Source, position.Line, position.Column, message)
	}
	p.errors = append(p.errors, message)
}

// noPrefixParseFnError records that the current token cannot begin an
// expression.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(p.curToken.Position, msg)
}
