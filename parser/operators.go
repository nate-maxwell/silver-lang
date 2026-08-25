package parser

import (
	"fmt"
	"silver/token"
)

// InfixRegistry holds the symbolic operators known to an interpreter or
// standalone parser session. Evaluators share one registry across files,
// imports, tasks, and REPL submissions.
type InfixRegistry struct {
	precedences map[string]int
	definitions map[string]token.Position
}

var definedLanguageInfixOperators = map[string]bool{
	"=": true, ".": true, ":": true, "...": true,
	"||": true, "&&": true, "==": true, "!=": true,
	"<": true, ">": true, "<=": true, ">=": true,
	"+": true, "-": true, "*": true, "/": true,
	"//": true, "%": true, "**": true,
}

func NewInfixRegistry() *InfixRegistry {
	return &InfixRegistry{
		precedences: make(map[string]int),
		definitions: make(map[string]token.Position),
	}
}

// HasUserOperators reports whether this session's grammar has been extended.
func (r *InfixRegistry) HasUserOperators() bool {
	return len(r.definitions) != 0
}

func (r *InfixRegistry) define(symbol string, power int, position token.Position) string {
	if definedLanguageInfixOperators[symbol] {
		return fmt.Sprintf("operator %q is already defined by the language", symbol)
	}
	if previous, exists := r.definitions[symbol]; exists {
		message := fmt.Sprintf("operator %q is already defined", symbol)
		if previous.IsValid() {
			message += fmt.Sprintf(" at %s:%d:%d", previous.Source, previous.Line, previous.Column)
		}
		return message
	}
	r.precedences[symbol] = power
	r.definitions[symbol] = position
	return ""
}

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
	token.MODULO:   PRODUCT,
	token.ASTERISK: PRODUCT,
	token.POWER:    POWER,
	token.LPAREN:   CALL,
	token.LBRACE:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      MEMBER,
}

// peekPrecedence returns the binding power of the lookahead token.
// It checks the InfixRegistry first, then the precedences map.
func (p *Parser) peekPrecedence() int {
	if precedence, ok := p.operators.precedences[p.peekToken.Literal]; ok {
		return precedence
	}
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence returns the binding power of the current token.
func (p *Parser) curPrecedence() int {
	if precedence, ok := p.operators.precedences[p.curToken.Literal]; ok {
		return precedence
	}
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
