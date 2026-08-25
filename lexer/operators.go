package lexer

import (
	"silver/token"
	"sort"
	"strings"
)

// RegisterOperator teaches this lexer one symbolic spelling. Registration is
// per lexer so unrelated and concurrently parsed programs retain independent
// grammars.
func (l *Lexer) RegisterOperator(operator string) {
	for _, existing := range l.operators {
		if existing == operator {
			return
		}
	}
	l.operators = append(l.operators, operator)
	sort.SliceStable(l.operators, func(i, j int) bool {
		return len(l.operators[i]) > len(l.operators[j])
	})
}

func (l *Lexer) registeredOperator() string {
	remaining := l.input[l.position:]
	for _, operator := range l.operators {
		if strings.HasPrefix(remaining, operator) {
			return operator
		}
	}
	return ""
}

// lookupRegisteredOperator preserves grammar-significant token types when a
// built-in spelling is overridden. Other declarations use one generic type.
func lookupRegisteredOperator(literal string) token.TokenType {
	switch literal {
	case "=":
		return token.ASSIGN
	case "!":
		return token.BANG
	case "|":
		return token.PIPE
	case "&&":
		return token.AND
	case "||":
		return token.OR
	case "+":
		return token.PLUS
	case "-":
		return token.MINUS
	case "*":
		return token.ASTERISK
	case "**":
		return token.POWER
	case "/":
		return token.SLASH
	case "//":
		return token.INT_DIV
	case "%":
		return token.MODULO
	case "<":
		return token.LT
	case ">":
		return token.GT
	case "<=":
		return token.LTE
	case ">=":
		return token.GTE
	case "==":
		return token.EQ
	case "!=":
		return token.NOT_EQ
	case "::":
		return token.EMBED
	default:
		return token.CUSTOM_OPERATOR
	}
}
