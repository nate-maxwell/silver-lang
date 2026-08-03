package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// MapLiteral stores key/value expression pairs. Evaluation later validates
// that each key produces a hashable runtime object.
type MapLiteral struct {
	Token token.Token // the '{' token
	Pairs map[Expression]Expression
}

// expressionNode marks MapLiteral as an Expression.
func (ml *MapLiteral) expressionNode() {}

// TokenLiteral returns the opening brace.
func (ml *MapLiteral) TokenLiteral() string { return ml.Token.Literal }

// Position returns the opening brace's position.
func (ml *MapLiteral) Position() token.Position { return ml.Token.Position }

// String renders the map's key/value expressions.
func (ml *MapLiteral) String() string {
	var out bytes.Buffer

	pairs := []string{}
	for key, value := range ml.Pairs {
		pairs = append(pairs, key.String()+":"+value.String())
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
