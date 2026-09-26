package parser

import (
	"silver/ast"
	"silver/token"
)

// parseTryExpression parses a value-producing try body followed by one or
// more typed catch clauses. Each clause explicitly names the local variable
// that receives the caught struct: catch ErrorType binding { ... }.
func (p *Parser) parseTryExpression() ast.Expression {
	expression := &ast.TryExpression{Token: p.curToken}
	if !p.expectPeek(token.LBRACE) {
		return nil
	}
	expression.Body = p.parseBlockStatement()

	for p.peekTokenIs(token.CATCH) {
		p.nextToken()
		clause := &ast.CatchClause{Token: p.curToken}
		clause.ErrorType = p.parseTypeAnnotation()
		if clause.ErrorType == nil {
			return nil
		}
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		clause.Binding = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal, Type: clause.ErrorType}
		if !p.expectPeek(token.LBRACE) {
			return nil
		}
		clause.Body = p.parseBlockStatement()
		expression.Catches = append(expression.Catches, clause)
	}

	if len(expression.Catches) == 0 {
		p.addError(expression.Position(), "try requires at least one catch clause")
	}
	return expression
}

// parseExpression is the Pratt parser core. It parses one prefix expression,
// then repeatedly consumes tighter-binding infix expressions.
// The caller's precedence is the stopping threshold: equal binding power is
// left for the caller, making operators at the same power left-associative.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	// A brace can start a struct initializer or an if/switch body. The caller
	// marks the latter context explicitly; nested delimiters temporarily lift
	// that restriction in parseNestedExpression. Physical newlines also stop
	// extension here even though the lexer discards whitespace.
	for !(p.stopAtBlockBrace && p.peekTokenIs(token.LBRACE)) &&
		!p.lineBreakBeforePeek() && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if p.operators.Has(p.peekToken.Literal) {
			infix = p.parseInfixExpression
		}
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

// parseExpressionStatement parses an expression, binding assignment, member
// assignment, or map-index assignment and validates its statement boundary.
func (p *Parser) parseExpressionStatement() ast.Statement {
	firstToken := p.curToken
	expression := p.parseExpression(LOWEST)

	// Assignment is a statement form, so it is recognized after parsing the
	// entire target expression rather than as another Pratt infix operator.
	if p.peekTokenIs(token.ASSIGN) {
		p.nextToken()
		assignmentToken := p.curToken
		p.nextToken()
		value := p.parseExpression(LOWEST)
		p.consumeStatementEnd()

		switch target := expression.(type) {
		case *ast.Identifier:
			return &ast.AssignmentStatement{Token: assignmentToken, Name: target, Value: value}
		case *ast.MemberExpression:
			return &ast.MemberAssignmentStatement{Token: assignmentToken, Target: target, Value: value}
		case *ast.IndexExpression:
			return &ast.IndexAssignmentStatement{Token: assignmentToken, Target: target, Value: value}
		default:
			p.addError(firstToken.Position, "invalid assignment target; expected identifier, struct member, or map index")
			return &ast.ExpressionStatement{Token: firstToken, Expression: expression}
		}
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

// parseGroupedExpression parses a parenthesized expression without adding a
// distinct grouping node to the AST.
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseNestedExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parseIfExpression parses a condition, consequence, and optional else branch.
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}
	previous := p.stopAtBlockBrace
	defer func() { p.stopAtBlockBrace = previous }()

	if p.peekTokenIs(token.LBRACE) {
		p.addError(p.peekToken.Position, "expected condition before if body")
		return nil
	}
	p.nextToken()
	p.stopAtBlockBrace = true
	expression.Condition = p.parseExpression(LOWEST)
	p.stopAtBlockBrace = false

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if p.peekTokenIs(token.IF) {
			p.nextToken()
			ifToken := p.curToken
			alternative := p.parseIfExpression()
			if alternative == nil {
				return nil
			}

			// Keep alternatives represented as blocks by desugaring `else if`
			// into an alternative block containing another if expression.
			expression.Alternative = &ast.BlockStatement{
				Token: ifToken,
				Statements: []ast.Statement{
					&ast.ExpressionStatement{Token: ifToken, Expression: alternative},
				},
			}
			return expression
		}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}

	return expression
}

// parseSwitchExpression parses ordered case clauses and an optional final
// default clause. Clause bodies end at the next clause or closing brace.
func (p *Parser) parseSwitchExpression() ast.Expression {
	expression := &ast.SwitchExpression{Token: p.curToken}
	previous := p.stopAtBlockBrace
	defer func() { p.stopAtBlockBrace = previous }()

	if p.peekTokenIs(token.LBRACE) {
		p.addError(p.peekToken.Position, "expected value before switch body")
		return nil
	}
	p.nextToken()
	p.stopAtBlockBrace = true
	expression.Value = p.parseExpression(LOWEST)
	p.stopAtBlockBrace = false
	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()
	seenDefault := false
	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		switch p.curToken.Type {
		case token.CASE:
			if seenDefault {
				p.addError(p.curToken.Position, "case clauses must appear before default")
			}
			switchCase := &ast.SwitchCase{Token: p.curToken}
			if p.peekTokenIs(token.COLON) || p.peekTokenIs(token.RBRACE) || p.peekTokenIs(token.EOF) {
				p.addError(p.curToken.Position, "case requires a value expression")
				return expression
			}
			p.nextToken()
			switchCase.Value = p.parseExpression(LOWEST)
			if !p.expectPeek(token.COLON) {
				return expression
			}
			switchCase.Body = p.parseSwitchClauseBody()
			expression.Cases = append(expression.Cases, switchCase)

		case token.DEFAULT:
			if seenDefault {
				p.addError(p.curToken.Position, "switch may contain only one default clause")
			}
			seenDefault = true
			if !p.expectPeek(token.COLON) {
				return expression
			}
			expression.Default = p.parseSwitchClauseBody()

		default:
			p.addError(p.curToken.Position, "expected case, default, or } in switch")
			return expression
		}
	}

	if p.curTokenIs(token.EOF) {
		p.addError(expression.Position(), "unterminated switch body")
	}
	return expression
}

// parseSwitchClauseBody parses statements after a clause's colon without
// treating case and default as ordinary statement-leading expressions.
func (p *Parser) parseSwitchClauseBody() *ast.BlockStatement {
	body := &ast.BlockStatement{Token: p.curToken, Statements: []ast.Statement{}}
	p.nextToken()
	for !p.curTokenIs(token.CASE) && !p.curTokenIs(token.DEFAULT) &&
		!p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		body.Statements = append(body.Statements, p.parseStatement())
		p.nextToken()
	}
	return body
}

// parseImportExpression accepts one path expression inside import(...).
func (p *Parser) parseImportExpression() ast.Expression {
	expression := &ast.ImportExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	arguments := p.parseExpressionList(token.RPAREN)
	if len(arguments) != 1 {
		p.addError(expression.Position(), "import expects exactly one path expression")
		return expression
	}
	expression.Path = arguments[0]
	return expression
}

// parseMemberExpression parses .name access on left.
func (p *Parser) parseMemberExpression(left ast.Expression) ast.Expression {
	expression := &ast.MemberExpression{Token: p.curToken, Object: left}

	// The standard library exposes core.type as a module member.
	if p.peekTokenIs(token.TYPE) {
		p.nextToken()
	} else if !p.expectPeek(token.IDENT) {
		return nil
	}

	expression.Member = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	return expression
}

// parseIndexExpression parses a bracketed index applied to left.
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseNestedExpression(LOWEST)

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
	list = append(list, p.parseNestedExpression(LOWEST))

	for !p.peekTokenIs(end) {
		if !p.expectPeek(token.COMMA) {
			return nil
		}
		p.nextToken()
		list = append(list, p.parseNestedExpression(LOWEST))
	}

	p.nextToken()
	return list
}

// parseNestedExpression permits braces inside an explicitly delimited
// expression while an enclosing unparenthesized if condition is being parsed.
func (p *Parser) parseNestedExpression(precedence int) ast.Expression {
	previous := p.stopAtBlockBrace
	p.stopAtBlockBrace = false
	expression := p.parseExpression(precedence)
	p.stopAtBlockBrace = previous
	return expression
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
