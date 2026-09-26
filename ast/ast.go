package ast

import (
	"silver/token"
)

// Node is the common interface implemented by every syntax-tree node.
// Position identifies the construct's representative token for diagnostics.
// String reconstructs a readable form for debugging; it does not preserve
// source whitespace, comments, or all original spelling.
type Node interface {
	TokenLiteral() string
	Position() token.Position
	String() string
}

// Statement is a node that can appear in a program or block statement list.
type Statement interface {
	Node
	statementNode()
}
