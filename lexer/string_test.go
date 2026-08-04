package lexer

import (
	"silver/token"
	"testing"
)

func TestStringEscapeSequences(t *testing.T) {
	input := `"quote: \" slash: \\ newline:\n tab:\t carriage:\r nul:\0 hex:\x41 unicode:\u263A wide:\U0001F642"`
	tok := New(input).NextToken()
	want := "quote: \" slash: \\ newline:\n tab:\t carriage:\r nul:\x00 hex:A unicode:☺ wide:🙂"
	if tok.Type != token.STRING || tok.Literal != want {
		t.Fatalf("token is (%q, %q), want (STRING, %q)", tok.Type, tok.Literal, want)
	}
}

func TestEscapedQuoteDoesNotEndString(t *testing.T) {
	l := New(`"say \"hello\"" + "again"`)
	want := []struct {
		tokenType token.TokenType
		literal   string
	}{
		{token.STRING, `say "hello"`},
		{token.PLUS, "+"},
		{token.STRING, "again"},
	}
	for index, expected := range want {
		got := l.NextToken()
		if got.Type != expected.tokenType || got.Literal != expected.literal {
			t.Fatalf("token %d is (%q, %q), want (%q, %q)", index, got.Type, got.Literal, expected.tokenType, expected.literal)
		}
	}
}

func TestMultilineStringRemainsSupported(t *testing.T) {
	tok := New("\"first\nsecond\"").NextToken()
	if tok.Type != token.STRING || tok.Literal != "first\nsecond" {
		t.Fatalf("token is (%q, %q), want multiline string", tok.Type, tok.Literal)
	}
}

func TestInvalidStringEscapesAreIllegal(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{`"bad\q"`, `unknown escape sequence \q in string literal`},
		{`"bad\xG0"`, `invalid hexadecimal digit 'G' in byte escape`},
		{`"bad\uD800"`, `invalid Unicode escape \uD800 in string literal`},
		{`"unfinished`, `unterminated string literal`},
	}
	for _, test := range tests {
		tok := New(test.input).NextToken()
		if tok.Type != token.ILLEGAL || tok.Literal != test.message {
			t.Errorf("token for %q is (%q, %q), want (ILLEGAL, %q)", test.input, tok.Type, tok.Literal, test.message)
		}
	}
}
