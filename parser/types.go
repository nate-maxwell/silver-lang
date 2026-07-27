package parser

import (
	"fmt"
	"silver/ast"
	"silver/token"
	"strconv"
)

// parseIntegerLiteral converts the current token into a signed 64-bit value.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		message := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.addError(p.curToken.Position, message)
		return nil
	}

	literal.Value = value
	return literal
}

// parseFloatLiteral converts the current decimal token to an IEEE-754
// double-precision value.
func (p *Parser) parseFloatLiteral() ast.Expression {
	literal := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		message := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.addError(p.curToken.Position, message)
		return nil
	}

	literal.Value = value
	return literal
}

// parseBoolean maps the True and False token types to a boolean AST value.
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// parseStringLiteral builds a literal from the lexer's unquoted token value.
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

// parseArrayLiteral parses a bracket-delimited expression list.
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

// parseHashLiteral parses comma-separated key:value expression pairs.
func (p *Parser) parseHashLiteral() ast.Expression {
	hash := &ast.HashLiteral{Token: p.curToken}
	hash.Pairs = make(map[ast.Expression]ast.Expression)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		hash.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return hash
}

// parseEnumStatement parses enum Name { Member, ... } and an optional trailing
// semicolon. Member names must be unique identifiers.
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

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}
	return statement
}

// parseEnumMembers parses a possibly empty comma-separated member list. A
// trailing comma before the closing brace is accepted.
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
		if !p.expectPeek(token.COMMA) {
			return nil
		}
		if p.peekTokenIs(token.RBRACE) {
			p.nextToken()
			return members
		}
	}
}
