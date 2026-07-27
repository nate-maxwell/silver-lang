package parser

import (
	"fmt"
	"silver/ast"
	"silver/lexer"
	"silver/token"
)

// Pratt binding powers increase from weak equality/comparison operators to
// tight calls, indexing, and member access.
const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
	MEMBER      // module.member
)

// precedences assigns binding power to infix token types. Tokens absent from
// the table use LOWEST, which terminates the current Pratt-parser expression.
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      MEMBER,
}

/* ----------------------------------------------------------------------------------------------------------
Main parser
---------------------------------------------------------------------------------------------------------- */

// Parser callback types used by the Pratt dispatch tables.
type (
	// prefixParseFn parses an expression that begins with the current token.
	prefixParseFn func() ast.Expression
	// infixParseFn extends an already-parsed left expression.
	infixParseFn func(ast.Expression) ast.Expression
)

// Parser implements a two-token-lookahead Pratt parser. Prefix and infix
// parse-function tables keep syntax extension localized by token type.
type Parser struct {
	l      *lexer.Lexer // token source
	errors []string     // accumulated diagnostics; parsing attempts to continue

	curToken  token.Token // token currently being parsed
	peekToken token.Token // one-token lookahead

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// New constructs a parser, registers the language grammar, and primes current
// and lookahead tokens.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT, p.parseFloatLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNCTION, p.parseFunctionLitearl)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseHashLiteral)
	p.registerPrefix(token.IMPORT, p.parseImportExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// ParseProgram parses statements until EOF and returns the AST root. Callers
// should inspect Errors before evaluating the result.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
		p.nextToken()
	}

	return program
}

// parseIdentifier converts the current identifier token into an AST node.
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

/* ----------------------------------------------------------------------------------------------------------
Function literal parsing
---------------------------------------------------------------------------------------------------------- */

// parseFunctionLitearl parses an fn expression, including its parameter list
// and brace-delimited body.
func (p *Parser) parseFunctionLitearl() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses zero or more comma-separated parameter names.
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

// parseCallExpression extends function with a parenthesized argument list.
func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

/* ----------------------------------------------------------------------------------------------------------
Statement parsing
---------------------------------------------------------------------------------------------------------- */

// parseStatement dispatches according to the current statement-leading token.
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.ENUM:
		return p.parseEnumStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseLetStatement parses a named binding and its value expression.
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parses the expression returned by a function.
func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

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

/* ----------------------------------------------------------------------------------------------------------
Expression parsing
---------------------------------------------------------------------------------------------------------- */

// parseExpressionStatement wraps a top-level expression and consumes an
// optional trailing semicolon.
func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

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

/* ----------------------------------------------------------------------------------------------------------
Tokens
---------------------------------------------------------------------------------------------------------- */

// nextToken advances both current and lookahead tokens by one position.
func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// curTokenIs reports whether the current token has type t.
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

// peekTokenIs reports whether the lookahead token has type t.
func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

// expectPeek advances when lookahead has the requested type; otherwise it
// records a positioned parser error.
func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

/* ----------------------------------------------------------------------------------------------------------
Errors
---------------------------------------------------------------------------------------------------------- */

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

/* ----------------------------------------------------------------------------------------------------------
Expression general
---------------------------------------------------------------------------------------------------------- */

// registerPrefix installs a parser for expressions beginning with tokenType.
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix installs a parser for expressions extending a left operand.
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// parseExpression is the Pratt parser core. It parses one prefix expression,
// then repeatedly consumes tighter-binding infix expressions.
func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

/* ----------------------------------------------------------------------------------------------------------
Expression error logging
---------------------------------------------------------------------------------------------------------- */

// noPrefixParseFnError records that the current token cannot begin an
// expression.
func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.addError(p.curToken.Position, msg)
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

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

/* ----------------------------------------------------------------------------------------------------------
Operator precedence
---------------------------------------------------------------------------------------------------------- */

// peekPrecedence returns the binding power of the lookahead token.
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence returns the binding power of the current token.
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

/* ----------------------------------------------------------------------------------------------------------
Primitive types
---------------------------------------------------------------------------------------------------------- */

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
