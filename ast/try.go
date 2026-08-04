package ast

import (
	"bytes"
	"silver/token"
)

// CatchClause handles one declared struct-valued failure and explicitly names
// the local variable that receives the caught struct.
type CatchClause struct {
	Token     token.Token // the catch token
	ErrorType *TypeAnnotation
	Binding   *Identifier
	Body      *BlockStatement
}

// TryExpression evaluates Body and handles the first matching typed failure.
type TryExpression struct {
	Token   token.Token // the try token
	Body    *BlockStatement
	Catches []*CatchClause
}

func (te *TryExpression) expressionNode() {}

func (te *TryExpression) TokenLiteral() string { return te.Token.Literal }

func (te *TryExpression) Position() token.Position { return te.Token.Position }

func (te *TryExpression) String() string {
	var out bytes.Buffer
	out.WriteString("try {")
	out.WriteString(te.Body.String())
	out.WriteString("}")
	for _, clause := range te.Catches {
		out.WriteString(" catch ")
		out.WriteString(clause.ErrorType.String())
		out.WriteString(" ")
		out.WriteString(clause.Binding.String())
		out.WriteString(" {")
		out.WriteString(clause.Body.String())
		out.WriteString("}")
	}
	return out.String()
}
