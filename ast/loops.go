package ast

import (
	"bytes"
	"silver/token"
)

// ForStatement iterates an array element or a map key/value pair. Value is
// nil for the single-binding array form.
type ForStatement struct {
	Token    token.Token // the 'for' token
	Key      *Identifier
	Value    *Identifier
	Iterable Expression
	Body     *BlockStatement
}

func (fs *ForStatement) statementNode() {}

func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }

func (fs *ForStatement) Position() token.Position { return fs.Token.Position }

func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for ")
	out.WriteString(fs.Key.String())
	if fs.Value != nil {
		out.WriteString(", ")
		out.WriteString(fs.Value.String())
	}
	out.WriteString(" in ")
	if fs.Iterable != nil {
		out.WriteString(fs.Iterable.String())
	}
	out.WriteString(" {\n")
	if fs.Body != nil {
		out.WriteString(fs.Body.String())
	}
	out.WriteString("\n}")
	return out.String()
}

// WhileStatement repeatedly evaluates Body while Condition is truthy.
type WhileStatement struct {
	Token     token.Token // the 'while' token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode() {}

func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }

func (ws *WhileStatement) Position() token.Position { return ws.Token.Position }

func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while ")
	if ws.Condition != nil {
		out.WriteString(ws.Condition.String())
	}
	out.WriteString(" {\n")
	if ws.Body != nil {
		out.WriteString(ws.Body.String())
	}
	out.WriteString("\n}")
	return out.String()
}
