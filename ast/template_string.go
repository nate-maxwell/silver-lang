package ast

import (
	"silver/token"
	"strings"
)

// TemplateStringPart is either literal text or one delayed Silver expression.
// Exactly one field is populated for every part.
type TemplateStringPart struct {
	Text       string
	Expression Expression
}

// TemplateStringLiteral is a triple-backtick string whose interpolations are
// evaluated only when the resulting TemplateString's eval field is called.
type TemplateStringLiteral struct {
	Token token.Token
	Parts []TemplateStringPart
}

func (ts *TemplateStringLiteral) expressionNode() {}

func (ts *TemplateStringLiteral) TokenLiteral() string { return ts.Token.Literal }

func (ts *TemplateStringLiteral) Position() token.Position { return ts.Token.Position }

func (ts *TemplateStringLiteral) String() string {
	var result strings.Builder
	result.WriteString("```")
	for _, part := range ts.Parts {
		if part.Expression != nil {
			result.WriteByte('{')
			result.WriteString(part.Expression.String())
			result.WriteByte('}')
			continue
		}
		text := strings.ReplaceAll(part.Text, "{", "{{")
		text = strings.ReplaceAll(text, "}", "}}")
		result.WriteString(text)
	}
	result.WriteString("```")
	return result.String()
}
