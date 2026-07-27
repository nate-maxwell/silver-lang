package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// HashLiteral stores key/value expression pairs. Evaluation later validates
// that each key produces a hashable runtime object.
type HashLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

// expressionNode marks HashLiteral as an Expression.
func (hl *HashLiteral) expressionNode() {}

// TokenLiteral returns the opening brace.
func (hl *HashLiteral) TokenLiteral() string { return hl.Token.Literal }

// Position returns the opening brace's position.
func (hl *HashLiteral) Position() token.Position { return hl.Token.Position }

// String renders the hash's key/value expressions.
func (hl *HashLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range hl.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
