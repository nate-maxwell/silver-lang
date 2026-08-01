package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// FunctionLiteral declares parameters and a body for a Silver closure.
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	ReturnType *TypeAnnotation   // success type; nil means null
	ErrorTypes []*TypeAnnotation // ordinary struct values returned on failure
	Body       *BlockStatement
}

// expressionNode marks FunctionLiteral as an Expression.
func (fl *FunctionLiteral) expressionNode() {}

// TokenLiteral returns the fn keyword.
func (fl *FunctionLiteral) TokenLiteral() string { return fl.Token.Literal }

// Position returns the fn keyword's position.
func (fl *FunctionLiteral) Position() token.Position {
	return fl.Token.Position
}

// String renders the function signature and body.
func (fl *FunctionLiteral) String() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range fl.Parameters {
		params = append(params, p.DeclarationString())
	}

	out.WriteString(fl.TokenLiteral())
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if fl.ReturnType != nil {
		out.WriteString(" ")
		out.WriteString(fl.ReturnType.String())
	}
	for _, errorType := range fl.ErrorTypes {
		out.WriteString(" | ")
		out.WriteString(errorType.String())
	}
	out.WriteString(" ")
	out.WriteString(fl.Body.String())

	return out.String()
}

// CallExpression invokes Function with evaluated Arguments. Token is the
// opening parenthesis, while Position deliberately points to Function.
type CallExpression struct {
	Token     token.Token // the '(' token
	Function  Expression  // Identifier or FunctionLiteral
	Arguments []Expression
}

// expressionNode marks CallExpression as an Expression.
func (ce *CallExpression) expressionNode() {}

// TokenLiteral returns the opening parenthesis.
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// Position returns the beginning of the callable expression.
func (ce *CallExpression) Position() token.Position {
	// Point at the callable rather than its opening parenthesis. This makes a
	// traceback column identify the beginning of the expression a user called.
	return ce.Function.Position()
}

// String renders the callable and comma-separated arguments.
func (ce *CallExpression) String() string {
	var out bytes.Buffer

	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}

	out.WriteString(ce.Function.String())
	out.WriteString("(")
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")

	return out.String()
}
