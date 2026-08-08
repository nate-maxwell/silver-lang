package lexer

import (
	"silver/token"
	"testing"
)

func TestAssertKeyword(t *testing.T) {
	got := New("assert asserted").NextToken()
	if got.Type != token.ASSERT || got.Literal != "assert" {
		t.Fatalf("token is %#v, want ASSERT", got)
	}
	got = New("asserted").NextToken()
	if got.Type != token.IDENT {
		t.Fatalf("token is %#v, want IDENT", got)
	}
}
