package ast

import "silver/token"

// StringLiteral is a quoted string value. Value excludes the quote characters.
type StringLiteral struct {
	Token token.Token
	Value string
}

// expressionNode marks StringLiteral as an Expression.
func (sl *StringLiteral) expressionNode() {}

// TokenLiteral returns the string's unquoted token literal.
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// Position returns the opening quote's position.
func (sl *StringLiteral) Position() token.Position { return sl.Token.Position }

// String returns the string token's unquoted value.
func (sl *StringLiteral) String() string { return sl.Token.Literal }
