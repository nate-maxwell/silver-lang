package ast

import (
	"silver/token"
)

// Node is the common interface implemented by every syntax-tree node.
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
