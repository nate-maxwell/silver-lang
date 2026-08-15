package ast

import (
	"bytes"
	"silver/token"
)

// SwitchCase pairs a value expression with the block selected when the
// switch value compares equal to it.
type SwitchCase struct {
	Token token.Token // the 'case' token
	Value Expression
	Body  *BlockStatement
}

// SwitchExpression evaluates its value once, then compares it with case
// values in source order. Default is optional.
type SwitchExpression struct {
	Token   token.Token // the 'switch' token
	Value   Expression
	Cases   []*SwitchCase
	Default *BlockStatement
}

func (se *SwitchExpression) expressionNode() {}

func (se *SwitchExpression) TokenLiteral() string { return se.Token.Literal }

func (se *SwitchExpression) Position() token.Position { return se.Token.Position }

func (se *SwitchExpression) String() string {
	var out bytes.Buffer
	out.WriteString("switch ")
	if se.Value != nil {
		out.WriteString(se.Value.String())
	}
	out.WriteString(" {")
	for _, switchCase := range se.Cases {
		out.WriteString("\ncase ")
		if switchCase.Value != nil {
			out.WriteString(switchCase.Value.String())
		}
		out.WriteString(":\n")
		if switchCase.Body != nil {
			out.WriteString(switchCase.Body.String())
		}
	}
	if se.Default != nil {
		out.WriteString("\ndefault:\n")
		out.WriteString(se.Default.String())
	}
	out.WriteString("\n}")
	return out.String()
}
