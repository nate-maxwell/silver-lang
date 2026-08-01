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
// signature using call([name:] parameterType, ...) [returnType].
func (p *Parser) parseTypeAnnotation() *ast.TypeAnnotation {
	if !p.expectPeek(token.IDENT) {
		return nil
	}
	return p.parseTypeAnnotationFromCurrent()
}

// parseTypeAnnotationFromCurrent parses a type whose first identifier is the
// current token. This lets named call parameters distinguish name: Type from
// an unnamed Type without requiring more lexer lookahead.
func (p *Parser) parseTypeAnnotationFromCurrent() *ast.TypeAnnotation {
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
		annotation.ParameterNames = make([]string, 0)
		p.nextToken()
		if p.peekTokenIs(token.RPAREN) {
			p.nextToken()
		} else {
			for {
				p.nextToken()
				if !p.curTokenIs(token.IDENT) {
					p.peekError(token.IDENT)
					return nil
				}

				parameterName := ""
				var parameterType *ast.TypeAnnotation
				if p.peekTokenIs(token.COLON) {
					parameterName = p.curToken.Literal
					p.nextToken()
					parameterType = p.parseTypeAnnotation()
				} else {
					parameterType = p.parseTypeAnnotationFromCurrent()
				}
				if parameterType == nil {
					return nil
				}
				annotation.ParameterNames = append(annotation.ParameterNames, parameterName)
				annotation.ParameterTypes = append(annotation.ParameterTypes, parameterType)

				if !p.peekTokenIs(token.COMMA) {
					break
				}
				p.nextToken()
			}
			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}

		if p.peekTokenIs(token.IDENT) && !p.lineBreakBeforePeek() {
			annotation.ReturnType = p.parseTypeAnnotation()
		}
		annotation.ErrorTypes = p.parseErrorTypeAlternatives()
	}
	return annotation
}

// parseErrorTypeAlternatives consumes zero or more `| StructType` suffixes.
func (p *Parser) parseErrorTypeAlternatives() []*ast.TypeAnnotation {
	errorTypes := make([]*ast.TypeAnnotation, 0)
	for p.peekTokenIs(token.PIPE) {
		p.nextToken()
		errorType := p.parseTypeAnnotation()
		if errorType == nil {
			return nil
		}
		errorTypes = append(errorTypes, errorType)
	}
	return errorTypes
}
