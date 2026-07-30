package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// StructStatement declares a named product type. Fields retain declaration
// order because constructor arguments are positional.
type StructStatement struct {
	Token  token.Token // the struct keyword
	Name   *Identifier
	Fields []*Identifier
}

// statementNode marks StructStatement as a Statement.
func (ss *StructStatement) statementNode() {}

// TokenLiteral returns the struct keyword.
func (ss *StructStatement) TokenLiteral() string { return ss.Token.Literal }

// Position returns the struct keyword's position.
func (ss *StructStatement) Position() token.Position { return ss.Token.Position }

// String renders the struct name and comma-separated fields.
func (ss *StructStatement) String() string {
	fields := make([]string, 0, len(ss.Fields))
	for _, field := range ss.Fields {
		fields = append(fields, field.DeclarationString())
	}
	return "struct " + ss.Name.String() + " { " + strings.Join(fields, ", ") + " }"
}

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
