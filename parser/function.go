package parser

import "silver/ast"
import "silver/token"

// parseFunctionLitearl parses an fn expression, including its parameter list
// and brace-delimited body.
func (p *Parser) parseFunctionLitearl() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		lit.ReturnType = p.parseTypeAnnotation()
		if lit.ReturnType == nil {
			return nil
		}
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters parses zero or more comma-separated parameter names.
func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	ident := p.parseDeclarationIdentifier()
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		ident := p.parseDeclarationIdentifier()
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}
