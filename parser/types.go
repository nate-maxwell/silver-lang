package parser

import (
	"silver/ast"
	"silver/token"
)

// parseDeclarationIdentifier parses the optional : type suffix on the
// current identifier. It leaves the final type component as the current token.
func (p *Parser) parseDeclarationIdentifier() *ast.Identifier {
	identifier := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	if p.peekTokenIs(token.COLON) {
		p.nextToken()
		identifier.Type = p.parseTypeAnnotation()
	}
	return identifier
}

// parseTypeAnnotation parses the identifier following the current token and
// any dot-qualified components. Type names are resolved by the evaluator.
func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	annotation := &ast.TypeAnnotation{
		Token: p.curToken,
		Parts: []string{p.curToken.Literal},
	}
	for p.peekTokenIs(token.DOT) {
		p.nextToken()
		if !p.expectPeek(token.IDENT) {
			return nil
		}
		annotation.Parts = append(annotation.Parts, p.curToken.Literal)
	}
	return annotation
}
