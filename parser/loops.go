package parser

import (
	"silver/ast"
	"silver/token"
)

// parseForStatement parses either `for item in array { ... }` or
// `for key, value in map { ... }`. Collection shape is checked at runtime.
func (p *Parser) parseForStatement() *ast.ForStatement {
	statement := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}
	statement.Key = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		statement.Value = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	}

	if !p.expectPeek(token.IN) {
		return nil
	}
	if p.peekTokenIs(token.LBRACE) {
		p.addError(p.peekToken.Position, "expected iterable before for body")
		return nil
	}

	p.nextToken()
	previous := p.stopAtBlockBrace
	p.stopAtBlockBrace = true
	statement.Iterable = p.parseExpression(LOWEST)
	p.stopAtBlockBrace = previous

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	statement.Body = p.parseBlockStatement()
	p.consumeStatementEnd()
	return statement
}

// parseWhileStatement parses `while condition { ... }`.
func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	statement := &ast.WhileStatement{Token: p.curToken}

	if p.peekTokenIs(token.LBRACE) {
		p.addError(p.peekToken.Position, "expected condition before while body")
		return nil
	}
	p.nextToken()
	previous := p.stopAtBlockBrace
	p.stopAtBlockBrace = true
	statement.Condition = p.parseExpression(LOWEST)
	p.stopAtBlockBrace = previous

	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	statement.Body = p.parseBlockStatement()
	p.consumeStatementEnd()
	return statement
}
