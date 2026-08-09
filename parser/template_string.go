package parser

import (
	"silver/ast"
	"silver/token"
)

func (p *Parser) parseTemplateStringLiteral() ast.Expression {
	literal := &ast.TemplateStringLiteral{Token: p.curToken}

	for !p.peekTokenIs(token.TEMPLATE_END) {
		if p.peekTokenIs(token.EOF) {
			p.addError(literal.Position(), "unterminated template string literal")
			return literal
		}
		p.nextToken()

		switch p.curToken.Type {
		case token.TEMPLATE_TEXT:
			literal.Parts = append(literal.Parts, ast.TemplateStringPart{Text: p.curToken.Literal})

		case token.LBRACE:
			if p.peekTokenIs(token.RBRACE) {
				p.addError(p.curToken.Position, "template interpolation requires an expression")
				p.nextToken()
				continue
			}
			p.nextToken()
			expression := p.parseNestedExpression(LOWEST)
			if expression != nil {
				literal.Parts = append(literal.Parts, ast.TemplateStringPart{Expression: expression})
			}
			if !p.expectPeek(token.RBRACE) {
				return literal
			}

		case token.ILLEGAL:
			p.addError(p.curToken.Position, p.curToken.Literal)
			return literal

		default:
			p.addError(p.curToken.Position, "expected template text or interpolation")
			return literal
		}
	}

	p.nextToken()
	return literal
}
