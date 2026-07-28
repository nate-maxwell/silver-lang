package ast

import (
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
