package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// StructLiteral constructs a struct value from positional field values.
type StructLiteral struct {
	Token      token.Token // the opening {
	StructType Expression
	Values     []Expression
}

// expressionNode marks StructLiteral as an Expression.
func (sl *StructLiteral) expressionNode() {}

// TokenLiteral returns the opening brace.
func (sl *StructLiteral) TokenLiteral() string { return sl.Token.Literal }

// Position points at the struct type rather than the opening brace.
func (sl *StructLiteral) Position() token.Position { return sl.StructType.Position() }

// String renders the struct type and its comma-separated positional values.
func (sl *StructLiteral) String() string {
	var out bytes.Buffer
	values := make([]string, 0, len(sl.Values))
	for _, value := range sl.Values {
		values = append(values, value.String())
	}
	out.WriteString(sl.StructType.String())
	out.WriteString("{")
	out.WriteString(strings.Join(values, ", "))
	out.WriteString("}")
	return out.String()
}
