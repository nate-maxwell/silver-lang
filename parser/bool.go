package parser

import (
	"silver/ast"
	"silver/token"
)

// parseBoolean maps the True and False token types to a boolean AST value.
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}
