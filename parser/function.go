package parser

import "silver/ast"
import "silver/token"

// parseFunctionLiteral parses an fn expression, including its parameter list,
// optional return type, and brace-delimited body.
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()
	if p.peekTokenIs(token.IDENT) {
		lit.ReturnType = p.parseTypeAnnotation()
		if lit.ReturnType == nil {
			return nil
		}
	}
	lit.ErrorTypes = p.parseErrorTypeAlternatives()
	if lit.ErrorTypes == nil && p.curTokenIs(token.PIPE) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	previousLoopDepth := p.loopDepth
	p.loopDepth = 0
	lit.Body = p.parseBlockStatement()
	p.loopDepth = previousLoopDepth

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
