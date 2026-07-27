package lexer

import "silver/token"

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
