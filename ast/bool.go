package ast

import "silver/token"

// Boolean is a True or False literal.
type Boolean struct {
	Token token.Token
	Value bool
}

// expressionNode marks Boolean as an Expression.
func (b *Boolean) expressionNode() {}

// TokenLiteral returns the boolean literal's source spelling.
func (b *Boolean) TokenLiteral() string { return b.Token.Literal }

// Position returns the literal's position.
func (b *Boolean) Position() token.Position { return b.Token.Position }

// String returns the boolean literal's source spelling.
func (b *Boolean) String() string { return b.Token.Literal }
