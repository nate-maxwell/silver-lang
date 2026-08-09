package parser

import (
	"silver/ast"
	"silver/token"
)

// parseStatement dispatches according to the current statement-leading token.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.ENUM:
		return p.parseEnumStatement()
	case token.STRUCT:
		return p.parseStructStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.ASSERT:
		return p.parseAssertStatement()
	case token.DEFER:
		return p.parseDeferStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseAssertStatement parses Python-style `assert condition` and
// `assert condition, message` statements.
func (p *Parser) parseAssertStatement() *ast.AssertStatement {
	stmt := &ast.AssertStatement{Token: p.curToken}
	if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) || p.lineBreakBeforePeek() {
		p.addError(p.curToken.Position, "assert requires a condition")
		p.consumeStatementEnd()
		return stmt
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)
	if p.peekTokenIs(token.COMMA) && !p.lineBreakBeforePeek() {
		p.nextToken()
		if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) || p.lineBreakBeforePeek() {
			p.addError(p.curToken.Position, "assert message requires an expression")
			p.consumeStatementEnd()
			return stmt
		}
		p.nextToken()
		stmt.Message = p.parseExpression(LOWEST)
	}
	p.consumeStatementEnd()
	return stmt
}

// parseDeferStatement parses a call that will run when the surrounding
// function, module, or script exits.
func (p *Parser) parseDeferStatement() *ast.DeferStatement {
	stmt := &ast.DeferStatement{Token: p.curToken}
	if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) || p.lineBreakBeforePeek() {
		p.addError(p.curToken.Position, "defer requires a function call")
		p.consumeStatementEnd()
		return stmt
	}

	p.nextToken()
	expression := p.parseExpression(LOWEST)
	call, ok := expression.(*ast.CallExpression)
	if !ok {
		p.addError(p.curToken.Position, "defer requires a function call")
	} else {
		stmt.Call = call
	}
	p.consumeStatementEnd()
	return stmt
}

// consumeStatementEnd accepts a physical newline, a closing delimiter, or
// EOF. Adjacent statements on one line are rejected so token adjacency cannot
// silently change program structure.
func (p *Parser) consumeStatementEnd() {
	if p.peekTokenIs(token.EOF) || p.peekTokenIs(token.RBRACE) || p.lineBreakBeforePeek() {
		return
	}
	p.addError(p.peekToken.Position, "expected newline between statements")
}

// parseLetStatement parses a named binding and its value expression.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		stmt.Name.Type = p.parseTypeAnnotation()
		if stmt.Name.Type == nil {
			return nil
		}
	}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	p.consumeStatementEnd()

	return stmt
}

// parseReturnStatement parses the expression returned by a function.
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	if p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) || p.lineBreakBeforePeek() {
		p.consumeStatementEnd()
		return stmt
	}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	p.consumeStatementEnd()

	return stmt
}

// parseBlockStatement parses statements until a closing brace or EOF. The
// opening brace is expected to be the current token.
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		block.Statements = append(block.Statements, stmt)

		p.nextToken()
	}

	return block
}
