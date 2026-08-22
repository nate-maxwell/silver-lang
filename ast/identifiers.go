package ast

import "silver/token"

// Identifier names a variable, parameter, builtin, or module binding.
type Identifier struct {
	Token    token.Token // the token.IDENT token
	Value    string
	Type     *TypeAnnotation // optional declaration annotation
	Embedded bool            // struct field declared with :: instead of :
	Variadic bool            // final function parameter collects remaining arguments
}

// expressionNode marks Identifier as an Expression.
func (i *Identifier) expressionNode() {}

// TokenLiteral returns the identifier's source spelling.
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// Position returns the identifier's position.
func (i *Identifier) Position() token.Position {
	return i.Token.Position
}

// String returns the identifier's resolved source name.
func (i *Identifier) String() string { return i.Value }

// DeclarationString renders an identifier together with its optional type.
// Expression identifiers deliberately use String so annotations never leak
// into ordinary name references.
func (i *Identifier) DeclarationString() string {
	if i.Type == nil {
		if i.Variadic {
			return i.Value + "..."
		}
		return i.Value
	}
	suffix := ""
	if i.Variadic {
		suffix = "..."
	}
	if i.Embedded {
		return i.Value + ":: " + i.Type.String() + suffix
	}
	return i.Value + ": " + i.Type.String() + suffix
}
