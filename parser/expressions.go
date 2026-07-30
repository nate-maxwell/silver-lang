package parser

import (
	"silver/ast"
	"silver/token"
)

// parseExpression is the Pratt parser core. It parses one prefix expression,
// then repeatedly consumes tighter-binding infix expressions.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && !p.lineBreakBeforePeek() && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier converts the current identifier token into an AST node.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// parseExpressionStatement parses an expression or a struct-member assignment
// and validates its newline-delimited statement boundary.
func (p *Parser) parseExpressionStatement() ast.Statement {
	firstToken := p.curToken
	expression := p.parseExpression(LOWEST)

	if p.peekTokenIs(token.ASSIGN) {
		target, ok := expression.(*ast.MemberExpression)
		p.nextToken()
		statement := &ast.MemberAssignmentStatement{Token: p.curToken, Target: target}
		if !ok {
			p.addError(firstToken.Position, "invalid assignment target; expected struct member")
		}
		p.nextToken()
		statement.Value = p.parseExpression(LOWEST)
		p.consumeStatementEnd()
		return statement
	}

	stmt := &ast.ExpressionStatement{Token: firstToken, Expression: expression}
	p.consumeStatementEnd()
	return stmt
}

// parsePrefixExpression parses the right operand at PREFIX precedence.
func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression combines left with the operator at the current token and
// a right operand parsed at that operator's precedence.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parsePowerExpression parses exponentiation one level below its own binding
// power on the right, making chains right-associative: a ** b ** c is
// interpreted as a ** (b ** c).
func (p *Parser) parsePowerExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	p.nextToken()
	expression.Right = p.parseExpression(POWER - 1)

	return expression
}

// parseGroupedExpression parses a parenthesized expression without adding a
// distinct grouping node to the AST.
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression parses a condition, consequence, and optional else block.
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseImportExpression accepts the restricted import("path") form.
func (p *Parser) parseImportExpression() ast.Expression {
	expression := &ast.ImportExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	if !p.expectPeek(token.STRING) {
		return nil
	}

	expression.Path = &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return expression
}

// parseMemberExpression parses .name access on left.
func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	expression := &ast.MemberExpression{Token: p.curToken, Object: left}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return expression
}

// parseIndexExpression parses a bracketed index applied to left.
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

// parseExpressionList parses a possibly empty, comma-separated list ending in
// the requested delimiter. It is shared by calls and arrays.
func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for !p.peekTokenIs(end) {
		if !p.expectPeek(token.COMMA) {
			return nil
		}
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	p.nextToken()
	return list
}

// parseCallExpression extends function with a parenthesized argument list.
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

// parseStructLiteral constructs a struct value using brace-delimited,
// comma-separated positional field values.
func (p *Parser) parseStructLiteral(structType ast.Expression) ast.Expression {
	literal := &ast.StructLiteral{Token: p.curToken, StructType: structType}
	literal.Values = p.parseExpressionList(token.RBRACE)
	return literal
}
