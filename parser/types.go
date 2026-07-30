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
// any dot-qualified components. A call type may additionally declare a
// signature using call(parameterType, ...) returnType.
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
	if len(annotation.Parts) == 1 && annotation.Parts[0] == "call" && p.peekTokenIs(token.LPAREN) {
		annotation.ParameterTypes = make([]*ast.TypeAnnotation, 0)
		p.nextToken()
		if p.peekTokenIs(token.RPAREN) {
			p.nextToken()
		} else {
			parameterType := p.parseTypeAnnotation()
			if parameterType == nil {
				return nil
			}
			annotation.ParameterTypes = append(annotation.ParameterTypes, parameterType)
			for p.peekTokenIs(token.COMMA) {
				p.nextToken()
				parameterType = p.parseTypeAnnotation()
				if parameterType == nil {
					return nil
				}
				annotation.ParameterTypes = append(annotation.ParameterTypes, parameterType)
			}
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		annotation.ReturnType = p.parseTypeAnnotation()
		if annotation.ReturnType == nil {
			return nil
		}
	}
	return annotation
}
