package parser

import (
	"fmt"
	"silver/ast"
	"silver/object"
	"silver/token"
)

// parseMapLiteral parses comma-separated key:value expressions between braces,
// including empty maps and an optional trailing comma. It reports duplicate
// literal keys at the repeated key's position. Keys that require evaluation are
// checked for hashability and duplicates by the evaluator.
func (p *Parser) parseMapLiteral() ast.Expression {
	mapLiteral := &ast.MapLiteral{Token: p.curToken}
	mapLiteral.Pairs = make(map[ast.Expression]ast.Expression)
	// Pairs uses AST node identity, so equal literals occupy separate entries.
	// Track runtime hashes within this literal to recognize equivalent keys,
	// including integer/float equivalents such as 1 and 1.0 or 0 and -0.0.
	seen := make(map[object.HashKey]bool)

	for !p.peekTokenIs(token.RBRACE) {
		p.nextToken()
		key := p.parseExpression(LOWEST)
		if literal := literalMapKey(key); literal != nil {
			hash := literal.HashKey()
			if seen[hash] {
				p.addError(key.Position(), fmt.Sprintf("duplicate map key %s", key.String()))
			}
			// Keep parsing after a duplicate so the remaining pairs and following
			// statements can still be parsed. Callers must check p.Errors().
			seen[hash] = true
		}

		if !p.expectPeek(token.COLON) {
			return nil
		}

		p.nextToken()
		value := p.parseExpression(LOWEST)
		mapLiteral.Pairs[key] = value

		if !p.peekTokenIs(token.RBRACE) && !p.expectPeek(token.COMMA) {
			return nil
		}
	}

	if !p.expectPeek(token.RBRACE) {
		return nil
	}
	return mapLiteral
}

// literalMapKey converts scalar literals and negated numeric literals to their
// runtime representations, reusing HashKey's normalization rules. It does not
// resolve names, call functions, or evaluate general expressions. A nil result
// leaves key validation to the evaluator; it does not mean the key is invalid.
func literalMapKey(expression ast.Expression) object.Hashable {
	switch node := expression.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.Boolean:
		return &object.Boolean{Value: node.Value}
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}
	case *ast.PrefixExpression:
		// The parser represents a negative number as unary minus applied to a
		// literal, rather than storing the sign in the numeric literal itself.
		if node.Operator == "-" {
			switch value := literalMapKey(node.Right).(type) {
			case *object.Integer:
				return &object.Integer{Value: -value.Value}
			case *object.Float:
				return &object.Float{Value: -value.Value}
			}
		}
	}
	return nil
}
