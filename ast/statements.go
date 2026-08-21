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
Variable assignment statements
---------------------------------------------------------------------------------------------------------- */

// AssignmentStatement replaces the value of an existing variable binding.
type AssignmentStatement struct {
	Token token.Token // the = token
	Name  *Identifier
	Value Expression
}

func (as *AssignmentStatement) statementNode()       {}
func (as *AssignmentStatement) TokenLiteral() string { return as.Token.Literal }

func (as *AssignmentStatement) Position() token.Position {
	if as.Name != nil {
		return as.Name.Position()
	}
	return as.Token.Position
}

func (as *AssignmentStatement) String() string {
	var out bytes.Buffer
	if as.Name != nil {
		out.WriteString(as.Name.String())
	}
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Member assignment statements
---------------------------------------------------------------------------------------------------------- */

// MemberAssignmentStatement replaces one field on an existing struct value.
type MemberAssignmentStatement struct {
	Token  token.Token // the = token
	Target *MemberExpression
	Value  Expression
}

// statementNode marks MemberAssignmentStatement as a Statement.
func (mas *MemberAssignmentStatement) statementNode() {}

// TokenLiteral returns the assignment operator.
func (mas *MemberAssignmentStatement) TokenLiteral() string { return mas.Token.Literal }

// Position points at the assignment target.
func (mas *MemberAssignmentStatement) Position() token.Position {
	if mas.Target != nil {
		return mas.Target.Position()
	}
	return mas.Token.Position
}

// String renders the target and replacement value.
func (mas *MemberAssignmentStatement) String() string {
	var out bytes.Buffer
	if mas.Target != nil {
		out.WriteString(mas.Target.String())
	}
	out.WriteString(" = ")
	if mas.Value != nil {
		out.WriteString(mas.Value.String())
	}
	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Index assignment statements
---------------------------------------------------------------------------------------------------------- */

// IndexAssignmentStatement assigns through a map or an indexable struct.
type IndexAssignmentStatement struct {
	Token  token.Token // the = token
	Target *IndexExpression
	Value  Expression
}

// statementNode marks IndexAssignmentStatement as a Statement.
func (ias *IndexAssignmentStatement) statementNode() {}

// TokenLiteral returns the assignment operator.
func (ias *IndexAssignmentStatement) TokenLiteral() string { return ias.Token.Literal }

// Position points at the indexed assignment target.
func (ias *IndexAssignmentStatement) Position() token.Position {
	if ias.Target != nil {
		return ias.Target.Position()
	}
	return ias.Token.Position
}

// String renders the target and replacement value.
func (ias *IndexAssignmentStatement) String() string {
	var out bytes.Buffer
	if ias.Target != nil {
		out.WriteString(ias.Target.String())
	}
	out.WriteString(" = ")
	if ias.Value != nil {
		out.WriteString(ias.Value.String())
	}
	return out.String()
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

/* ----------------------------------------------------------------------------------------------------------
Assert statements
---------------------------------------------------------------------------------------------------------- */

// AssertStatement requires Condition to be truthy. Message, when present, is
// evaluated only after the condition fails and becomes the AssertionError
// detail.
type AssertStatement struct {
	Token     token.Token // the 'assert' token
	Condition Expression
	Message   Expression
}

// statementNode marks AssertStatement as a Statement.
func (as *AssertStatement) statementNode() {}

// TokenLiteral returns the assert keyword.
func (as *AssertStatement) TokenLiteral() string { return as.Token.Literal }

// Position returns the assert keyword's position.
func (as *AssertStatement) Position() token.Position { return as.Token.Position }

// String renders the condition and optional failure message without a
// statement terminator.
func (as *AssertStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.TokenLiteral())
	if as.Condition != nil {
		out.WriteString(" ")
		out.WriteString(as.Condition.String())
	}
	if as.Message != nil {
		out.WriteString(", ")
		out.WriteString(as.Message.String())
	}
	return out.String()
}

/* ----------------------------------------------------------------------------------------------------------
Defer statements
---------------------------------------------------------------------------------------------------------- */

// DeferStatement schedules Call for the end of the surrounding function,
// module, or script. The evaluator captures the callable and arguments when
// this statement executes.
type DeferStatement struct {
	Token token.Token // the 'defer' token
	Call  *CallExpression
}

// statementNode marks DeferStatement as a Statement.
func (ds *DeferStatement) statementNode() {}

// TokenLiteral returns the defer keyword.
func (ds *DeferStatement) TokenLiteral() string { return ds.Token.Literal }

// Position returns the defer keyword's position.
func (ds *DeferStatement) Position() token.Position { return ds.Token.Position }

// String renders the deferred call without a statement terminator.
func (ds *DeferStatement) String() string {
	if ds.Call == nil {
		return ds.TokenLiteral()
	}
	return ds.TokenLiteral() + " " + ds.Call.String()
}

/* ----------------------------------------------------------------------------------------------------------
Export statements
---------------------------------------------------------------------------------------------------------- */

// ExportStatement selects the top-level bindings exposed by an imported
// module. A source file without one exports every top-level binding.
type ExportStatement struct {
	Token token.Token // the 'export' token
	Names []*Identifier
}

func (es *ExportStatement) statementNode()           {}
func (es *ExportStatement) TokenLiteral() string     { return es.Token.Literal }
func (es *ExportStatement) Position() token.Position { return es.Token.Position }

// String renders the exported names as a brace-delimited list.
func (es *ExportStatement) String() string {
	var out bytes.Buffer
	out.WriteString(es.TokenLiteral())
	out.WriteString(" {")
	for index, name := range es.Names {
		if index > 0 {
			out.WriteString(", ")
		}
		out.WriteString(name.String())
	}
	out.WriteString("}")
	return out.String()
}
