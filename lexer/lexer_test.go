package lexer

import (
	"testing"

	"silver/token"
)

func TestNextToken(t *testing.T) {
	input := `let five = 5
	let ten = 10
	let add = fn(x, y) {
	x + y
	}
	let result = add(five, ten)
	!-/*5
	5 < 10 > 5

	if (5 < 10) {
		return True
		} else {
		return False
	}

	10 == 10
	10 != 9
	"foobar"
	"foo bar"
	[1, 2]
	{"foo": "bar"}
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "five"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.LET, "let"},
		{token.IDENT, "ten"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.LET, "let"},
		{token.IDENT, "add"},
		{token.ASSIGN, "="},
		{token.FUNCTION, "fn"},
		{token.LPAREN, "("},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.PLUS, "+"},
		{token.IDENT, "y"},
		{token.RBRACE, "}"},
		{token.LET, "let"},
		{token.IDENT, "result"},
		{token.ASSIGN, "="},
		{token.IDENT, "add"},
		{token.LPAREN, "("},
		{token.IDENT, "five"},
		{token.COMMA, ","},
		{token.IDENT, "ten"},
		{token.RPAREN, ")"},

		{token.BANG, "!"},
		{token.MINUS, "-"},
		{token.SLASH, "/"},
		{token.ASTERISK, "*"},
		{token.INT, "5"},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.GT, ">"},
		{token.INT, "5"},

		{token.IF, "if"},
		{token.LPAREN, "("},
		{token.INT, "5"},
		{token.LT, "<"},
		{token.INT, "10"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.TRUE, "True"},
		{token.RBRACE, "}"},
		{token.ELSE, "else"},
		{token.LBRACE, "{"},
		{token.RETURN, "return"},
		{token.FALSE, "False"},
		{token.RBRACE, "}"},

		{token.INT, "10"},
		{token.EQ, "=="},
		{token.INT, "10"},
		{token.INT, "10"},
		{token.NOT_EQ, "!="},
		{token.INT, "9"},
		{token.STRING, "foobar"},
		{token.STRING, "foo bar"},
		{token.LBRACKET, "["},
		{token.INT, "1"},
		{token.COMMA, ","},
		{token.INT, "2"},
		{token.RBRACKET, "]"},
		{token.LBRACE, "{"},
		{token.STRING, "foo"},
		{token.COLON, ":"},
		{token.STRING, "bar"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLineComments(t *testing.T) {
	input := `
	# A comment may occupy an entire line.
	let value = 10 # It may also follow other tokens.
	let half = value / 2
	# Slash operators must remain available.
	half # A comment at EOF does not need a trailing newline.`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.LET, "let"},
		{token.IDENT, "value"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.LET, "let"},
		{token.IDENT, "half"},
		{token.ASSIGN, "="},
		{token.IDENT, "value"},
		{token.SLASH, "/"},
		{token.INT, "2"},
		{token.IDENT, "half"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Type != tt.expectedType || tok.Literal != tt.expectedLiteral {
			t.Fatalf(
				"tests[%d] - wrong token. expected=(%q, %q), got=(%q, %q)",
				i,
				tt.expectedType,
				tt.expectedLiteral,
				tok.Type,
				tok.Literal,
			)
		}
	}
}

func TestTokenPositions(t *testing.T) {
	input := "let value = 1 # comment\n  value / 2"
	l := NewWithSource(input, "example.slv")

	tests := []struct {
		literal string
		line    int
		column  int
	}{
		{"let", 1, 1},
		{"value", 1, 5},
		{"=", 1, 11},
		{"1", 1, 13},
		{"value", 2, 3},
		{"/", 2, 9},
		{"2", 2, 11},
	}

	for i, tt := range tests {
		tok := l.NextToken()
		if tok.Literal != tt.literal {
			t.Fatalf("tests[%d] - literal is %q, want %q", i, tok.Literal, tt.literal)
		}
		if tok.Position.Source != "example.slv" || tok.Position.Line != tt.line || tok.Position.Column != tt.column {
			t.Fatalf("tests[%d] - position is %+v, want example.slv:%d:%d", i, tok.Position, tt.line, tt.column)
		}
	}
}

func TestIntegerDivideToken(t *testing.T) {
	l := NewWithSource("10 // 3 / 2", "divide.slv")
	want := []struct {
		tokenType token.TokenType
		literal   string
		column    int
	}{
		{token.INT, "10", 1},
		{token.INT_DIV, "//", 4},
		{token.INT, "3", 7},
		{token.SLASH, "/", 9},
		{token.INT, "2", 11},
		{token.EOF, "", 12},
	}
	for index, expected := range want {
		got := l.NextToken()
		if got.Type != expected.tokenType || got.Literal != expected.literal || got.Position.Column != expected.column {
			t.Fatalf("token %d is (%q, %q, column %d), want (%q, %q, column %d)",
				index, got.Type, got.Literal, got.Position.Column,
				expected.tokenType, expected.literal, expected.column)
		}
	}
}

func TestSemicolonIsIllegal(t *testing.T) {
	tok := New(";").NextToken()
	if tok.Type != token.ILLEGAL || tok.Literal != ";" {
		t.Fatalf("semicolon token is (%q, %q), want (ILLEGAL, %q)", tok.Type, tok.Literal, ";")
	}
}

func TestFloatTokens(t *testing.T) {
	input := `0.5 12.34 1.member 1. 1.5.6`
	tests := []struct {
		tokenType token.TokenType
		literal   string
	}{
		{token.FLOAT, "0.5"},
		{token.FLOAT, "12.34"},
		{token.INT, "1"},
		{token.DOT, "."},
		{token.IDENT, "member"},
		{token.INT, "1"},
		{token.DOT, "."},
		{token.FLOAT, "1.5"},
		{token.DOT, "."},
		{token.INT, "6"},
		{token.EOF, ""},
	}

	l := New(input)
	for i, want := range tests {
		got := l.NextToken()
		if got.Type != want.tokenType || got.Literal != want.literal {
			t.Fatalf("token %d is (%q, %q), want (%q, %q)", i, got.Type, got.Literal, want.tokenType, want.literal)
		}
	}
}

func TestIdentifiersMayContainDigits(t *testing.T) {
	l := New(`exp2 log1p value42 _2 2fast`)
	want := []struct {
		tokenType token.TokenType
		literal   string
	}{
		{token.IDENT, "exp2"},
		{token.IDENT, "log1p"},
		{token.IDENT, "value42"},
		{token.IDENT, "_2"},
		{token.INT, "2"},
		{token.IDENT, "fast"},
		{token.EOF, ""},
	}

	for index, expected := range want {
		got := l.NextToken()
		if got.Type != expected.tokenType || got.Literal != expected.literal {
			t.Fatalf("token %d is (%q, %q), want (%q, %q)", index, got.Type, got.Literal, expected.tokenType, expected.literal)
		}
	}
}

func TestPowerToken(t *testing.T) {
	l := NewWithSource("2 ** 3 * 4", "power.slv")
	want := []struct {
		tokenType token.TokenType
		literal   string
		column    int
	}{
		{token.INT, "2", 1},
		{token.POWER, "**", 3},
		{token.INT, "3", 6},
		{token.ASTERISK, "*", 8},
		{token.INT, "4", 10},
		{token.EOF, "", 11},
	}

	for i, expected := range want {
		got := l.NextToken()
		if got.Type != expected.tokenType || got.Literal != expected.literal || got.Position.Column != expected.column {
			t.Fatalf("token %d is (%q, %q, column %d), want (%q, %q, column %d)",
				i, got.Type, got.Literal, got.Position.Column,
				expected.tokenType, expected.literal, expected.column)
		}
	}
}

func TestInclusiveComparisonTokens(t *testing.T) {
	l := NewWithSource("left <= right >= value", "compare.slv")
	want := []struct {
		tokenType token.TokenType
		literal   string
		column    int
	}{
		{token.IDENT, "left", 1},
		{token.LTE, "<=", 6},
		{token.IDENT, "right", 9},
		{token.GTE, ">=", 15},
		{token.IDENT, "value", 18},
		{token.EOF, "", 23},
	}
	for index, expected := range want {
		got := l.NextToken()
		if got.Type != expected.tokenType || got.Literal != expected.literal || got.Position.Column != expected.column {
			t.Fatalf("token %d is (%q, %q, column %d), want (%q, %q, column %d)",
				index, got.Type, got.Literal, got.Position.Column,
				expected.tokenType, expected.literal, expected.column)
		}
	}
}

func TestPipeToken(t *testing.T) {
	l := New(`fn() str | FileNotFound | PermissionDenied {}`)
	want := []token.TokenType{
		token.FUNCTION, token.LPAREN, token.RPAREN, token.IDENT,
		token.PIPE, token.IDENT, token.PIPE, token.IDENT,
		token.LBRACE, token.RBRACE, token.EOF,
	}
	for index, expected := range want {
		if got := l.NextToken(); got.Type != expected {
			t.Fatalf("token %d has type %q, want %q", index, got.Type, expected)
		}
	}
}

func TestStructKeyword(t *testing.T) {
	l := New(`struct Point { x, y }`)
	want := []struct {
		tokenType token.TokenType
		literal   string
	}{
		{token.STRUCT, "struct"},
		{token.IDENT, "Point"},
		{token.LBRACE, "{"},
		{token.IDENT, "x"},
		{token.COMMA, ","},
		{token.IDENT, "y"},
		{token.RBRACE, "}"},
		{token.EOF, ""},
	}

	for i, expected := range want {
		got := l.NextToken()
		if got.Type != expected.tokenType || got.Literal != expected.literal {
			t.Fatalf("token %d is (%q, %q), want (%q, %q)", i, got.Type, got.Literal, expected.tokenType, expected.literal)
		}
	}
}

func TestDeferKeyword(t *testing.T) {
	tok := New("defer close()").NextToken()
	if tok.Type != token.DEFER || tok.Literal != "defer" {
		t.Fatalf("token is (%q, %q), want (DEFER, defer)", tok.Type, tok.Literal)
	}
}
