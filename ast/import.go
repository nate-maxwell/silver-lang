package ast

import "silver/token"

// ImportExpression loads a module from the string literal in Path.
type ImportExpression struct {
	Token token.Token // The 'import' token
	Path  *StringLiteral
}

// expressionNode marks ImportExpression as an Expression.
func (ie *ImportExpression) expressionNode() {}

// TokenLiteral returns the import keyword.
func (ie *ImportExpression) TokenLiteral() string { return ie.Token.Literal }

// Position returns the import keyword's position.
func (ie *ImportExpression) Position() token.Position {
	return ie.Token.Position
}

// String renders the import call and quoted path.
func (ie *ImportExpression) String() string {
	return ie.TokenLiteral() + "(\"" + ie.Path.Value + "\")"
}
