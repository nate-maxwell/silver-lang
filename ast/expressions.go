package ast

import (
	"bytes"
	"silver/token"
)

/* ----------------------------------------------------------------------------------------------------------
Prefix expressions
---------------------------------------------------------------------------------------------------------- */

// PrefixExpression applies a unary operator to its right-hand expression.
type PrefixExpression struct {
	Token    token.Token // The prefix token, e.g. ! or -
	Operator string
	Right    Expression
}

// expressionNode marks PrefixExpression as an Expression.
func (pe *PrefixExpression) expressionNode() {}

// TokenLiteral returns the prefix operator.
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }

// Position returns the operator's position.
func (pe *PrefixExpression) Position() token.Position {
	return pe.Token.Position
}

// String renders the prefix expression with explicit parentheses.
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")

	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Infix expressions
---------------------------------------------------------------------------------------------------------- */

// InfixExpression applies a binary operator to left and right expressions.
type InfixExpression struct {
	Token    token.Token // The operator token, e.g. + or *
	Left     Expression
	Operator string
	Right    Expression
}

// MemberExpression accesses a named member on a module, enum, or struct value.
type MemberExpression struct {
	Token  token.Token // The . token
	Object Expression
	Member *Identifier
}

// expressionNode marks MemberExpression as an Expression.
func (me *MemberExpression) expressionNode() {}

// TokenLiteral returns the member-access dot token.
func (me *MemberExpression) TokenLiteral() string { return me.Token.Literal }

// Position returns the beginning of the complete member expression.
func (me *MemberExpression) Position() token.Position {
	return me.Object.Position()
}

// String renders object.member.
func (me *MemberExpression) String() string {
	return me.Object.String() + "." + me.Member.String()
}

// expressionNode marks InfixExpression as an Expression.
func (ie *InfixExpression) expressionNode() {}

// TokenLiteral returns the binary operator.
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }

// Position returns the operator's position, which is the most useful location
// for binary-operation diagnostics.
func (ie *InfixExpression) Position() token.Position {
	return ie.Token.Position
}

// String renders the infix expression with explicit parentheses.
func (ie *InfixExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")

	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
If expressions
---------------------------------------------------------------------------------------------------------- */

// IfExpression contains a condition and its consequence, plus an optional
// alternative block.
type IfExpression struct {
	Token       token.Token // The 'if' token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

// expressionNode marks IfExpression as an Expression.
func (ie *IfExpression) expressionNode() {}

// TokenLiteral returns the if keyword.
func (ie *IfExpression) TokenLiteral() string { return ie.Token.Literal }

// Position returns the if keyword's position.
func (ie *IfExpression) Position() token.Position {
	return ie.Token.Position
}

// String renders the conditional and its branches.
func (ie *IfExpression) String() string {
	var out bytes.Buffer

	out.WriteString("if")
	out.WriteString(ie.Condition.String())
	out.WriteString(" ")
	out.WriteString(ie.Consequence.String())

	if ie.Alternative != nil {
		out.WriteString("else")
		out.WriteString(ie.Alternative.String())
	}

	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Index expressions
---------------------------------------------------------------------------------------------------------- */

// IndexExpression looks up Index in Left, which may be an array or map.
type IndexExpression struct {
	Token token.Token // the [ token
	Left  Expression
	Index Expression
}

// expressionNode marks IndexExpression as an Expression.
func (ie *IndexExpression) expressionNode() {}

// TokenLiteral returns the opening index bracket.
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

// Position returns the opening index bracket's position.
func (ie *IndexExpression) Position() token.Position {
	return ie.Token.Position
}

// String renders the indexed expression.
func (ie *IndexExpression) String() string {
	var out bytes.Buffer

	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	out.WriteString(ie.Index.String())
	out.WriteString("])")

	return out.String()
}
