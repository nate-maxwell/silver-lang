package parser

import (
	"fmt"
	"silver/ast"
	"strconv"
)

// parseFloatLiteral converts the current decimal token to an IEEE-754
// double-precision value.
func (p *Parser) parseFloatLiteral() ast.Expression {
	literal := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		message := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.addError(p.curToken.Position, message)
		return nil
	}

	literal.Value = value
	return literal
}
