package lexer

import (
	"silver/token"
	"testing"
)

func TestTypeDeclarationKeyword(t *testing.T) {
	l := New("type Point = struct {}\ncore.type(Point)")
	for i, want := range []token.TokenType{
		token.TYPE, token.IDENT, token.ASSIGN, token.STRUCT, token.LBRACE, token.RBRACE,
		token.IDENT, token.DOT, token.TYPE, token.LPAREN, token.IDENT, token.RPAREN, token.EOF,
	} {
		if got := l.NextToken(); got.Type != want {
			t.Fatalf("token %d = %v, want %s", i, got, want)
		}
	}
}
