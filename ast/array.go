package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// ArrayLiteral contains expressions evaluated from left to right into an array.
type ArrayLiteral struct {
	Token    token.Token // the '[' token
	Elements []Expression
}

// expressionNode marks ArrayLiteral as an Expression.
func (al *ArrayLiteral) expressionNode() {}

// TokenLiteral returns the opening bracket.
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }

// Position returns the opening bracket's position.
func (al *ArrayLiteral) Position() token.Position { return al.Token.Position }

// String renders the array's comma-separated element expressions.
func (al *ArrayLiteral) String() string {
	var out bytes.Buffer

	elements := []string{}
	for _, element := range al.Elements {
		elements = append(elements, element.String())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
