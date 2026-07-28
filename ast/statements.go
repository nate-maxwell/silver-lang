package ast

import (
	"bytes"
	"silver/token"
)

/* ----------------------------------------------------------------------------------------------------------
Block Statements
---------------------------------------------------------------------------------------------------------- */

// BlockStatement is a brace-delimited sequence of statements.
type BlockStatement struct {
	Token      token.Token // the { token
	Statements []Statement
}

// statementNode marks BlockStatement as a Statement.
func (bs *BlockStatement) statementNode() {}

// TokenLiteral returns the opening brace token.
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// Position returns the opening brace's position.
func (bs *BlockStatement) Position() token.Position {
	return bs.Token.Position
}

// String renders the statements contained by the block.
func (bs *BlockStatement) String() string {
	var out bytes.Buffer

	for index, s := range bs.Statements {
		if index > 0 {
			out.WriteString("\n")
		}
		out.WriteString(s.String())
	}

	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Exprsesion statements
---------------------------------------------------------------------------------------------------------- */

// Expression is a syntax node that evaluates to a value.
type Expression interface {
	Node
	expressionNode()
}

// ExpressionStatement wraps an expression so it can appear where a statement
// is required.
type ExpressionStatement struct {
	Token      token.Token // the first token of the expression
	Expression Expression
}

// statementNode marks ExpressionStatement as a Statement.
func (es *ExpressionStatement) statementNode() {}

// TokenLiteral returns the first token in the expression.
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// Position returns the first token's position.
func (es *ExpressionStatement) Position() token.Position {
	return es.Token.Position
}

// String renders the wrapped expression, or an empty string if parsing failed.
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}

	return ""
}

/* ----------------------------------------------------------------------------------------------------------
Let statements
---------------------------------------------------------------------------------------------------------- */

// LetStatement binds the evaluated Value to Name in the current environment.
type LetStatement struct {
	Token token.Token // the token.LET token
	Name  *Identifier
	Value Expression
}

// statementNode marks LetStatement as a Statement.
func (ls *LetStatement) statementNode() {}

// TokenLiteral returns the let keyword.
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }

// Position returns the let keyword's position.
func (ls *LetStatement) Position() token.Position {
	return ls.Token.Position
}

// String renders the binding without a statement terminator.
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.DeclarationString())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Return statements
---------------------------------------------------------------------------------------------------------- */

// ReturnStatement exits the nearest function with ReturnValue.
type ReturnStatement struct {
	Token       token.Token // the 'return' token
	ReturnValue Expression
}

// statementNode marks ReturnStatement as a Statement.
func (rs *ReturnStatement) statementNode() {}

// TokenLiteral returns the return keyword.
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// Position returns the return keyword's position.
func (rs *ReturnStatement) Position() token.Position {
	return rs.Token.Position
}

// String renders the return statement without a statement terminator.
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral())

	if rs.ReturnValue != nil {
		out.WriteString(" ")
		out.WriteString(rs.ReturnValue.String())
	}

	return out.String()
}
