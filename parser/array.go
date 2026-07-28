package parser

import (
	"silver/ast"
	"silver/token"
)

// parseArrayLiteral parses a bracket-delimited expression list.
func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}
