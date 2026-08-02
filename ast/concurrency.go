package ast

import (
	"silver/token"
	"strings"
)

// TaskExpression launches a zero-argument callable reference concurrently.
type TaskExpression struct {
	Token token.Token
	Work  Expression
}

func (te *TaskExpression) expressionNode()          {}
func (te *TaskExpression) TokenLiteral() string     { return te.Token.Literal }
func (te *TaskExpression) Position() token.Position { return te.Token.Position }
func (te *TaskExpression) String() string {
	if te.Work != nil {
		return "task " + te.Work.String()
	}
	return "task"
}

// CollectExpression waits for named task handles. The identifiers are kept in
// the AST because their source names become the result struct's field names.
type CollectExpression struct {
	Token   token.Token
	Handles []*Identifier
}

func (ce *CollectExpression) expressionNode()          {}
func (ce *CollectExpression) TokenLiteral() string     { return ce.Token.Literal }
func (ce *CollectExpression) Position() token.Position { return ce.Token.Position }
func (ce *CollectExpression) String() string {
	names := make([]string, len(ce.Handles))
	for index, handle := range ce.Handles {
		names[index] = handle.String()
	}
	return "collect " + strings.Join(names, ", ")
}
