package lexer

import (
	"fmt"
	"silver/token"
	"strings"
	"unicode/utf8"
)

// decodeString translates Silver's conventional single-character, byte, and
// Unicode escape sequences into their runtime representation.
func decodeString(raw string) (string, string) {
	var out strings.Builder

	for index := 0; index < len(raw); index++ {
		if raw[index] != '\\' {
			out.WriteByte(raw[index])
			continue
		}
		if index+1 >= len(raw) {
			return "", "unterminated escape sequence in string literal"
		}

		index++
		switch escaped := raw[index]; escaped {
		case '\\', '"', '\'', '/':
			out.WriteByte(escaped)
		case '0':
			out.WriteByte(0)
		case 'a':
			out.WriteByte('\a')
		case 'b':
			out.WriteByte('\b')
		case 'f':
			out.WriteByte('\f')
		case 'n':
			out.WriteByte('\n')
		case 'r':
			out.WriteByte('\r')
		case 't':
			out.WriteByte('\t')
		case 'v':
			out.WriteByte('\v')
		case 'x':
			value, next, diagnostic := decodeHexEscape(raw, index+1, 2, "byte")
			if diagnostic != "" {
				return "", diagnostic
			}
			out.WriteByte(byte(value))
			index = next - 1
		case 'u':
			value, next, diagnostic := decodeHexEscape(raw, index+1, 4, "Unicode")
			if diagnostic != "" {
				return "", diagnostic
			}
			if !validUnicodeCodePoint(value) {
				return "", fmt.Sprintf("invalid Unicode escape \\u%04X in string literal", value)
			}
			out.WriteRune(rune(value))
			index = next - 1
		case 'U':
			value, next, diagnostic := decodeHexEscape(raw, index+1, 8, "Unicode")
			if diagnostic != "" {
				return "", diagnostic
			}
			if !validUnicodeCodePoint(value) {
				return "", fmt.Sprintf("invalid Unicode escape \\U%08X in string literal", value)
			}
			out.WriteRune(rune(value))
			index = next - 1
		default:
			return "", fmt.Sprintf("unknown escape sequence \\%c in string literal", escaped)
		}
	}
	return out.String(), ""
}

func decodeHexEscape(raw string, start, digits int, kind string) (uint32, int, string) {
	if start+digits > len(raw) {
		return 0, start, fmt.Sprintf("incomplete %s escape in string literal", kind)
	}
	var value uint32
	for index := start; index < start+digits; index++ {
		digit, ok := hexDigit(raw[index])
		if !ok {
			return 0, start, fmt.Sprintf("invalid hexadecimal digit %q in %s escape", raw[index], kind)
		}
		value = value*16 + uint32(digit)
	}
	return value, start + digits, ""
}

func hexDigit(ch byte) (byte, bool) {
	switch {
	case '0' <= ch && ch <= '9':
		return ch - '0', true
	case 'a' <= ch && ch <= 'f':
		return ch - 'a' + 10, true
	case 'A' <= ch && ch <= 'F':
		return ch - 'A' + 10, true
	default:
		return 0, false
	}
}

func validUnicodeCodePoint(value uint32) bool {
	return value <= utf8.MaxRune && !(0xD800 <= value && value <= 0xDFFF)
}

/* ----------------------------------------------------------------------------------------------------------
Template Strings
---------------------------------------------------------------------------------------------------------- */

type templateMode uint8

const (
	_ templateMode = iota
	templateText
	templateExpression
)

type templateContext struct {
	mode  templateMode
	depth int
	open  token.Position
}
