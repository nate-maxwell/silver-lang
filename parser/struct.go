package parser

import (
	"fmt"
	"silver/ast"
	"silver/token"
)

// parseStructStatement parses struct Name { field, ... }. Field names must be
// unique identifiers; a legacy trailing semicolon remains accepted.
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
		if !p.expectPeek(token.IDENT) {
			return nil
		}

		field := p.parseDeclarationIdentifier()
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
