package parser

import (
	"silver/ast"
	"silver/lexer"
	"silver/token"
)

// Pratt binding powers increase from weak logical operators to tight calls,
// indexing, and member access.
const (
	_ int = iota
	LOWEST
	OR          // ||
	AND         // &&
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	POWER       // ** (binds more tightly than a leading prefix)
	CALL        // myFunction(X)
	INDEX       // array[index]
	MEMBER      // module.member
)

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
	l                *lexer.Lexer // token source
	errors           []string     // accumulated diagnostics; parsing attempts to continue
	stopAtBlockBrace bool         // top-level { terminates an unparenthesized if condition

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
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.TEMPLATE_START, p.parseTemplateStringLiteral)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)
	p.registerPrefix(token.LBRACE, p.parseMapLiteral)
	p.registerPrefix(token.IMPORT, p.parseImportExpression)
	p.registerPrefix(token.TASK, p.parseTaskExpression)
	p.registerPrefix(token.COLLECT, p.parseCollectExpression)
	p.registerPrefix(token.TRY, p.parseTryExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.INT_DIV, p.parseInfixExpression)
	p.registerInfix(token.MODULO, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.POWER, p.parsePowerExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACE, p.parseStructLiteral)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)
	p.registerInfix(token.DOT, p.parseMemberExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// registerPrefix installs a parser for expressions beginning with tokenType.
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix installs a parser for expressions extending a left operand.
func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
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
	if len(p.errors) == 0 {
		p.validateTaskUsage(program)
	}

	return program
}

// lineBreakBeforePeek reports whether ignored source whitespace crossed a
// physical line. Newlines terminate expressions without becoming AST tokens.
func (p *Parser) lineBreakBeforePeek() bool {
	current := p.curToken.Position
	peek := p.peekToken.Position
	return current.IsValid() && peek.IsValid() && current.Source == peek.Source && peek.Line > current.Line
}
