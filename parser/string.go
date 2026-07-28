package parser

import "silver/ast"

// parseStringLiteral builds a literal from the lexer's unquoted token value.
func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}
