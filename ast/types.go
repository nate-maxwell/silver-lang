package ast

import (
	"silver/token"
	"strings"
)

// TypeAnnotation is a primitive or qualified nominal type name appearing
// after a declaration colon, such as int or geometry.Point.
type TypeAnnotation struct {
	Token token.Token // the first identifier in the type name
	Parts []string
}

// Position returns the first type-name component's position.
func (ta *TypeAnnotation) Position() token.Position { return ta.Token.Position }

// String renders the complete qualified type name.
func (ta *TypeAnnotation) String() string { return strings.Join(ta.Parts, ".") }
