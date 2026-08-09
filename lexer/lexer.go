package lexer

import (
	"silver/token"
	"strings"
)

// Lexer converts Silver source text into a stream of positioned tokens. It
// scans bytes; line and column coordinates are one-based for diagnostics.
type Lexer struct {
	input        string // complete source text
	source       string // filename or synthetic source label
	position     int    // byte offset of ch
	readPosition int    // byte offset that readChar will consume next
	ch           byte   // byte currently under examination; zero means EOF
	line         int    // line containing ch
	column       int    // column containing ch
	nextLine     int    // line to assign to the next byte
	nextColumn   int    // column to assign to the next byte

	// template string trackers
	templateMode  templateMode
	templateDepth int
	templateOpen  token.Position
	templateStack []templateContext
}

// New constructs a lexer for in-memory input using <input> as its diagnostic
// source name.
func New(input string) *Lexer {
	return NewWithSource(input, "<input>")
}

// NewWithSource constructs a lexer whose tokens retain the source name used
// in diagnostics and tracebacks.
func NewWithSource(input, source string) *Lexer {
	l := &Lexer{
		input:      input,
		source:     source,
		nextLine:   1,
		nextColumn: 1,
	}
	l.readChar()
	return l
}

// readChar advances the scanner by one byte and updates the coordinates for
// both the current and next byte.
func (l *Lexer) readChar() {
	l.position = l.readPosition
	l.line = l.nextLine
	l.column = l.nextColumn

	if l.readPosition >= len(l.input) {
		l.ch = 0
		return
	}

	l.ch = l.input[l.readPosition]
	l.readPosition += 1
	if l.ch == '\n' {
		l.nextLine = l.line + 1
		l.nextColumn = 1
	} else {
		l.nextLine = l.line
		l.nextColumn = l.column + 1
	}
}

// NextToken returns the next non-comment, non-whitespace token. Identifier and
// number readers advance to the first byte after their token before returning.
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	if l.templateMode == templateText {
		return l.nextTemplateToken()
	}

	l.skipIgnoredCharacters()
	position := l.currentPosition()

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.EQ)
		} else {
			tok = newToken(token.ASSIGN, l.ch, position)
		}
	case '+':
		tok = newToken(token.PLUS, l.ch, position)
	case '-':
		tok = newToken(token.MINUS, l.ch, position)
	case '!':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.NOT_EQ)
		} else {
			tok = newToken(token.BANG, l.ch, position)
		}
	case '|':
		tok = newToken(token.PIPE, l.ch, position)
	case '*':
		if l.peekChar() == '*' {
			tok = l.makeTwoCharToken(token.POWER)
		} else {
			tok = newToken(token.ASTERISK, l.ch, position)
		}
	case '/':
		if l.peekChar() == '/' {
			tok = l.makeTwoCharToken(token.INT_DIV)
		} else {
			tok = newToken(token.SLASH, l.ch, position)
		}
	case '<':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.LTE)
		} else {
			tok = newToken(token.LT, l.ch, position)
		}
	case '>':
		if l.peekChar() == '=' {
			tok = l.makeTwoCharToken(token.GTE)
		} else {
			tok = newToken(token.GT, l.ch, position)
		}
	case ',':
		tok = newToken(token.COMMA, l.ch, position)
	case '(':
		tok = newToken(token.LPAREN, l.ch, position)
	case ')':
		tok = newToken(token.RPAREN, l.ch, position)
	case '{':
		tok = newToken(token.LBRACE, l.ch, position)
		if l.templateMode == templateExpression {
			l.templateDepth++
		}
	case '}':
		tok = newToken(token.RBRACE, l.ch, position)
		if l.templateMode == templateExpression {
			if l.templateDepth == 0 {
				l.templateMode = templateText
			} else {
				l.templateDepth--
			}
		}
	case '"':
		literal, diagnostic := l.readString()
		if diagnostic != "" {
			tok.Type = token.ILLEGAL
			tok.Literal = diagnostic
		} else {
			tok.Type = token.STRING
			tok.Literal = literal
		}
		tok.Position = position
	case '`':
		if !l.startsTripleBacktick() {
			tok = token.Token{Type: token.ILLEGAL, Literal: "template strings must start with three backticks", Position: position}
			break
		}
		l.beginTemplate(position)
		l.readChar()
		l.readChar()
		tok = token.Token{Type: token.TEMPLATE_START, Literal: "```", Position: position}
	case '[':
		tok = newToken(token.LBRACKET, l.ch, position)
	case ']':
		tok = newToken(token.RBRACKET, l.ch, position)
	case ':':
		tok = newToken(token.COLON, l.ch, position)
	case '.':
		tok = newToken(token.DOT, l.ch, position)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
		tok.Position = position
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			tok.Position = position
			return tok
		} else if isDigit(l.ch) {
			tok.Literal, tok.Type = l.readNumber()
			tok.Position = position
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch, position)
		}
	}

	l.readChar()
	return tok
}

// currentPosition snapshots the source coordinates at the current byte.
func (l *Lexer) currentPosition() token.Position {
	return token.Position{
		Source: l.source,
		Offset: l.position,
		Line:   l.line,
		Column: l.column,
	}
}

// skipIgnoredCharacters consumes whitespace and line comments before the next
// token. It loops because a comment's terminating newline may be followed by
// more whitespace or another comment.
func (l *Lexer) skipIgnoredCharacters() {
	for {
		l.skipWhitespace()
		if l.ch != '#' {
			return
		}
		l.skipLineComment()
	}
}

// skipLineComment consumes a # comment up to, but not including, its line
// ending. EOF also terminates a comment, allowing comments on the final line.
func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
}

// peekChar returns the next byte without advancing, or zero at EOF.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0 // Set to ascii null for EOF
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) peekSecondChar() byte {
	if l.readPosition+1 >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition+1]
}

func (l *Lexer) startsTripleBacktick() bool {
	return l.ch == '`' && l.peekChar() == '`' && l.peekSecondChar() == '`'
}

func (l *Lexer) beginTemplate(position token.Position) {
	l.templateStack = append(l.templateStack, templateContext{
		mode:  l.templateMode,
		depth: l.templateDepth,
		open:  l.templateOpen,
	})
	l.templateMode = templateText
	l.templateDepth = 0
	l.templateOpen = position
}

func (l *Lexer) endTemplate() {
	last := len(l.templateStack) - 1
	context := l.templateStack[last]
	l.templateStack = l.templateStack[:last]
	l.templateMode = context.mode
	l.templateDepth = context.depth
	l.templateOpen = context.open
}

// nextTemplateToken preserves template text byte-for-byte, except for doubled
// braces: {{ and }} spell literal braces without beginning interpolation.
func (l *Lexer) nextTemplateToken() token.Token {
	position := l.currentPosition()
	var literal strings.Builder

	for {
		switch {
		case l.ch == 0:
			if literal.Len() != 0 {
				return token.Token{Type: token.TEMPLATE_TEXT, Literal: literal.String(), Position: position}
			}
			opening := l.templateOpen
			l.endTemplate()
			return token.Token{Type: token.ILLEGAL, Literal: "unterminated template string literal", Position: opening}

		case l.startsTripleBacktick():
			if literal.Len() != 0 {
				return token.Token{Type: token.TEMPLATE_TEXT, Literal: literal.String(), Position: position}
			}
			l.readChar()
			l.readChar()
			l.readChar()
			l.endTemplate()
			return token.Token{Type: token.TEMPLATE_END, Literal: "```", Position: position}

		case l.ch == '{':
			if l.peekChar() == '{' {
				literal.WriteByte('{')
				l.readChar()
				l.readChar()
				continue
			}
			if literal.Len() != 0 {
				return token.Token{Type: token.TEMPLATE_TEXT, Literal: literal.String(), Position: position}
			}
			l.templateMode = templateExpression
			l.templateDepth = 0
			l.readChar()
			return token.Token{Type: token.LBRACE, Literal: "{", Position: position}

		case l.ch == '}':
			if l.peekChar() == '}' {
				literal.WriteByte('}')
				l.readChar()
				l.readChar()
				continue
			}
			if literal.Len() != 0 {
				return token.Token{Type: token.TEMPLATE_TEXT, Literal: literal.String(), Position: position}
			}
			l.readChar()
			return token.Token{Type: token.ILLEGAL, Literal: "unmatched } in template string literal", Position: position}

		default:
			literal.WriteByte(l.ch)
			l.readChar()
		}
	}
}

// makeTwoCharToken consumes the second byte of a two-character operator while
// retaining the first byte's position.
func (l *Lexer) makeTwoCharToken(tokenType token.TokenType) token.Token {
	position := l.currentPosition()
	ch := l.ch
	l.readChar()
	literal := string(ch) + string(l.ch)
	return token.Token{Type: tokenType, Literal: literal, Position: position}
}

// skipWhitespace advances past spaces, tabs, and line endings.
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

// newToken constructs a single-byte token at position.
func newToken(tokenType token.TokenType, ch byte, position token.Position) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch), Position: position}
}

// readIdentifier consumes a Silver identifier and returns its source text.
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}

// isLetter reports whether ch may begin a Silver identifier. Once an
// identifier has begun, readIdentifier also accepts decimal digits.
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// readString consumes bytes until an unescaped closing quote or EOF, then
// decodes escape sequences. It leaves ch on the closing quote so NextToken's
// final read advances to the following source byte.
func (l *Lexer) readString() (string, string) {
	position := l.position + 1
	escaped := false
	for {
		l.readChar()
		if l.ch == 0 {
			return "", "unterminated string literal"
		}
		if l.ch == '"' && !escaped {
			break
		}
		if escaped {
			escaped = false
		} else if l.ch == '\\' {
			escaped = true
		}
	}

	return decodeString(l.input[position:l.position])
}

// readNumber consumes an integer or decimal float. A dot is part of the number
// only when followed by a digit, preserving member access after integers.
func (l *Lexer) readNumber() (string, token.TokenType) {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}

	tokenType := token.TokenType(token.INT)
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = token.FLOAT
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	return l.input[position:l.position], tokenType
}

// isDigit reports whether ch is an ASCII decimal digit.
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
