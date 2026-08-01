package parser

import (
	"fmt"
	"silver/ast"
	"silver/token"
)

// parseEnumStatement parses enum Name { Member, ... }. Member names must be
// unique identifiers.
func (p *Parser) parseEnumStatement() *ast.EnumStatement {
	statement := &ast.EnumStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	statement.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	statement.Members = p.parseEnumMembers()

	p.consumeStatementEnd()
	return statement
}

// parseEnumMembers parses a possibly empty member list separated by newlines
// or commas. A trailing comma before the closing brace is accepted.
func (p *Parser) parseEnumMembers() []*ast.Identifier {
	var members []*ast.Identifier
	seen := make(map[string]bool)

	if p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		return members
	}

	for {
		if !p.expectPeek(token.IDENT) {
			return nil
		}

		member := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		if seen[member.Value] {
			p.addError(member.Position(), fmt.Sprintf("duplicate enum member %q", member.Value))
		} else {
			seen[member.Value] = true
			members = append(members, member)
		}

		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return members
		}
		if p.peekTokenIs(token.COMMA) {
			p.nextToken()
		} else if !p.lineBreakBeforePeek() {
			p.peekError(token.COMMA)
			return nil
		}
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return members
		}
	}
}
