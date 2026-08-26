package parser

import (
	"fmt"
	"silver/ast"
	"silver/token"
)

// parseStructStatement parses struct Name { field, ... }. Fields may be named
// identifiers or operator symbols known to this source's operator registry.
func (p *Parser) parseStructStatement() *ast.StructStatement {
	statement := &ast.StructStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	statement.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	statement.Fields = p.parseStructFields()

	p.consumeStatementEnd()
	return statement
}

// parseStructFields parses a possibly empty field list separated by newlines
// or commas. A trailing comma before the closing brace is accepted.
func (p *Parser) parseStructFields() []*ast.Identifier {
	var fields []*ast.Identifier
	seen := make(map[string]bool)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return fields
	}

	for {
		if !p.isStructFieldToken(p.peekToken) {
			p.peekError(token.IDENT)
			return nil
		}
		p.nextToken()

		field := p.parseDeclarationIdentifier()
		if p.peekTokenIs(token.EMBED) {
			p.nextToken()
			field.Type = p.parseTypeAnnotation()
			field.Embedded = true
		}
		if seen[field.Value] {
			p.addError(field.Position(), fmt.Sprintf("duplicate struct field %q", field.Value))
		} else {
			seen[field.Value] = true
			fields = append(fields, field)
		}

		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return fields
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		} else if !p.lineBreakBeforePeek() {
			p.peekError(token.COMMA)
			return nil
		}
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return fields
		}
	}
}

// isStructFieldToken accepts identifiers, overloadable built-in operators,
// and any spelling visible in the parser's package-local registry.
func (p *Parser) isStructFieldToken(candidate token.Token) bool {
	if p.operators.Has(candidate.Literal) {
		return true
	}
	switch candidate.Type {
	case token.IDENT, token.CUSTOM_OPERATOR,
		token.PLUS, token.MINUS, token.ASTERISK, token.POWER,
		token.SLASH, token.INT_DIV, token.MODULO,
		token.EQ, token.NOT_EQ, token.LT, token.GT, token.LTE, token.GTE:
		return true
	default:
		return false
	}
}
