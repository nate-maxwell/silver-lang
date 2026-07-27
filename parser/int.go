package parser

import (
	"fmt"
	"silver/ast"
	"strconv"
)

// parseIntegerLiteral converts the current token into a signed 64-bit value.
func (p *Parser) parseIntegerLiteral() ast.Expression {
	literal := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		message := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.addError(p.curToken.Position, message)
		return nil
	}

	literal.Value = value
	return literal
}
