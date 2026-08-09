package parser

import "silver/token"

// precedences assigns binding power to infix token types. Tokens absent from
// the table use LOWEST, which terminates the current Pratt-parser expression.
var precedences = map[token.TokenType]int{
	token.OR:       OR,
	token.AND:      AND,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.INT_DIV:  PRODUCT,
	token.ASTERISK: PRODUCT,
	token.POWER:    POWER,
	token.LPAREN:   CALL,
	token.LBRACE:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      MEMBER,
}

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
