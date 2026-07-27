package ast

import "silver/token"

// IntegerLiteral is a parsed signed 64-bit integer value.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

// expressionNode marks IntegerLiteral as an Expression.
func (il *IntegerLiteral) expressionNode() {}

// TokenLiteral returns the integer's source spelling.
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// Position returns the literal's position.
func (il *IntegerLiteral) Position() token.Position { return il.Token.Position }

// String returns the integer's source spelling.
func (il *IntegerLiteral) String() string { return il.Token.Literal }
