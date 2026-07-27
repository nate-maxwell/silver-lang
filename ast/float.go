package ast

import "silver/token"

// FloatLiteral is a parsed IEEE-754 double-precision decimal value.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

// expressionNode marks FloatLiteral as an Expression.
func (fl *FloatLiteral) expressionNode() {}

// TokenLiteral returns the float's source spelling.
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }

// Position returns the literal's position.
func (fl *FloatLiteral) Position() token.Position { return fl.Token.Position }

// String returns the float's original source spelling.
func (fl *FloatLiteral) String() string { return fl.Token.Literal }
