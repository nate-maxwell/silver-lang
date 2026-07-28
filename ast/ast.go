package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// Node is the common interface implemented by every syntax-tree node.
type Node interface {
	TokenLiteral() string
	Position() token.Position
	String() string
}

// Statement is a node that can appear in a program or block statement list.
type Statement interface {
	Node
	statementNode()
}

/* ----------------------------------------------------------------------------------------------------------
Program
---------------------------------------------------------------------------------------------------------- */

// Program is the root node for a parsed Silver source unit.
type Program struct {
	Statements []Statement
}

// Position returns the position of the first statement, or an invalid position
// for an empty program.
func (p *Program) Position() token.Position {
	if len(p.Statements) == 0 {
		return token.Position{}
	}
	return p.Statements[0].Position()
}

// TokenLiteral returns the first statement's token text.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	} else {
		return ""
	}
}

// String renders all statements in source order.
func (p *Program) String() string {
	var out bytes.Buffer

	for _, s := range p.Statements {
		out.WriteString(s.String())
	}

	return out.String()
}

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

	for _, s := range bs.Statements {
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
Identifiers
---------------------------------------------------------------------------------------------------------- */

// Identifier names a variable, parameter, builtin, or module binding.
type Identifier struct {
	Token token.Token // the token.IDENT token
	Value string
	Type  *TypeAnnotation // optional declaration annotation
}

// expressionNode marks Identifier as an Expression.
func (i *Identifier) expressionNode() {}

// TokenLiteral returns the identifier's source spelling.
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// Position returns the identifier's position.
func (i *Identifier) Position() token.Position {
	return i.Token.Position
}

// String returns the identifier's resolved source name.
func (i *Identifier) String() string { return i.Value }

// DeclarationString renders an identifier together with its optional type.
// Expression identifiers deliberately use String so annotations never leak
// into ordinary name references.
func (i *Identifier) DeclarationString() string {
	if i.Type == nil {
		return i.Value
	}
	return i.Value + ": " + i.Type.String()
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

// String renders the binding as a let statement.
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.DeclarationString())
	out.WriteString(" = ")

	if ls.Value != nil {
		out.WriteString(ls.Value.String())
	}

	out.WriteString(";")

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

// String renders the return statement.
func (rs *ReturnStatement) String() string {
	var out bytes.Buffer

	out.WriteString(rs.TokenLiteral() + " ")

	if rs.ReturnValue != nil {
		out.WriteString(rs.ReturnValue.String())
	}

	out.WriteString(";")

	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Function literals + calls
---------------------------------------------------------------------------------------------------------- */

// FunctionLiteral declares parameters and a body for a Silver closure.
type FunctionLiteral struct {
	Token      token.Token // The 'fn' token
	Parameters []*Identifier
	ReturnType *TypeAnnotation
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
		out.WriteString(": ")
		out.WriteString(fl.ReturnType.String())
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

// IndexExpression looks up Index in Left, which may be an array or hash.
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
