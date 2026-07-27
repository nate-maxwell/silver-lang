package ast

import (
	"silver/token"
	"strings"
)

// EnumStatement binds a named collection of singleton enum values. Members
// retain declaration order for stable inspection and identity assignment.
type EnumStatement struct {
	Token   token.Token // the enum keyword
	Name    *Identifier
	Members []*Identifier
}

// statementNode marks EnumStatement as a Statement.
func (es *EnumStatement) statementNode() {}

// TokenLiteral returns the enum keyword.
func (es *EnumStatement) TokenLiteral() string { return es.Token.Literal }

// Position returns the enum keyword's position.
func (es *EnumStatement) Position() token.Position {
	return es.Token.Position
}

// String renders the enum name and comma-separated members.
func (es *EnumStatement) String() string {
	members := make([]string, 0, len(es.Members))
	for _, member := range es.Members {
		members = append(members, member.String())
	}
	return "enum " + es.Name.String() + " { " + strings.Join(members, ", ") + " }"
}
