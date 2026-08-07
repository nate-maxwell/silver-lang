package ast

import (
	"silver/token"
	"strconv"
)

// ImportExpression loads a module from the string produced by Path.
type ImportExpression struct {
	Token token.Token // The 'import' token
	Path  Expression
}

// expressionNode marks ImportExpression as an Expression.
func (ie *ImportExpression) expressionNode() {}

// TokenLiteral returns the import keyword.
func (ie *ImportExpression) TokenLiteral() string { return ie.Token.Literal }

// Position returns the import keyword's position.
func (ie *ImportExpression) Position() token.Position {
	return ie.Token.Position
}

// String renders the import call and its path expression.
func (ie *ImportExpression) String() string {
	path := ie.Path.String()
	if literal, ok := ie.Path.(*StringLiteral); ok {
		path = strconv.Quote(literal.Value)
	}
	return ie.TokenLiteral() + "(" + path + ")"
}
