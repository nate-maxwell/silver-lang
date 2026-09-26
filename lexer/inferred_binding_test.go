package lexer

import (
	"silver/token"
	"testing"
)

func TestInferredBindingToken(t *testing.T) {
	l := New("let value:=32\nlet typed: int = 1\n: :: :=")
	for _, want := range []struct {
		kind token.TokenType
		text string
	}{
		{token.LET, "let"}, {token.IDENT, "value"}, {token.INFER_ASSIGN, ":="}, {token.INT, "32"},
		{token.LET, "let"}, {token.IDENT, "typed"}, {token.COLON, ":"}, {token.IDENT, "int"},
		{token.ASSIGN, "="}, {token.INT, "1"}, {token.COLON, ":"}, {token.EMBED, "::"},
		{token.INFER_ASSIGN, ":="}, {token.EOF, ""},
	} {
		got := l.NextToken()
		if got.Type != want.kind || got.Literal != want.text {
			t.Fatalf("token = %s %q, want %s %q", got.Type, got.Literal, want.kind, want.text)
		}
	}
}
