package ast

import (
	"bytes"
	"silver/token"
)

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

	for index, s := range p.Statements {
		if index > 0 {
			out.WriteString("\n")
		}
		out.WriteString(s.String())
	}

	return out.String()
}
