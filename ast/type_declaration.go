package ast

import (
	"silver/token"
	"strings"
)

// TypeExpression describes the right-hand side of a type declaration.
// Type syntax is separate from runtime value expressions.
type TypeExpression interface {
	Node
	typeExpressionNode()
}

// TypeStatement binds a nominal definition or reusable type contract.
type TypeStatement struct {
	Token token.Token
	Name  *Identifier
	Value TypeExpression
}

func (ts *TypeStatement) statementNode()           {}
func (ts *TypeStatement) TokenLiteral() string     { return ts.Token.Literal }
func (ts *TypeStatement) Position() token.Position { return ts.Token.Position }
func (ts *TypeStatement) String() string {
	return "type " + ts.Name.String() + " = " + ts.Value.String()
}

// StructTypeLiteral describes the fields of a nominal type declaration.
// StructLiteral, in contrast, constructs an instance of an existing type.
type StructTypeLiteral struct {
	Token  token.Token
	Fields []*Identifier
}

func (st *StructTypeLiteral) typeExpressionNode()      {}
func (st *StructTypeLiteral) TokenLiteral() string     { return st.Token.Literal }
func (st *StructTypeLiteral) Position() token.Position { return st.Token.Position }
func (st *StructTypeLiteral) String() string {
	fields := make([]string, 0, len(st.Fields))
	for _, field := range st.Fields {
		fields = append(fields, field.DeclarationString())
	}
	return "struct { " + strings.Join(fields, ", ") + " }"
}

// EnumTypeLiteral describes the singleton members of a nominal declaration.
type EnumTypeLiteral struct {
	Token   token.Token
	Members []*Identifier
}

func (et *EnumTypeLiteral) typeExpressionNode()      {}
func (et *EnumTypeLiteral) TokenLiteral() string     { return et.Token.Literal }
func (et *EnumTypeLiteral) Position() token.Position { return et.Token.Position }
func (et *EnumTypeLiteral) String() string {
	members := make([]string, 0, len(et.Members))
	for _, member := range et.Members {
		members = append(members, member.String())
	}
	return "enum { " + strings.Join(members, ", ") + " }"
}
