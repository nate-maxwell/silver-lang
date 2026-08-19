// Package formatter provides canonical whitespace formatting for Silver source.
package formatter

import (
	"errors"
	"fmt"
	"os"
	"silver/lexer"
	"silver/parser"
	"strings"
)

const indentWidth = 4

type elementKind uint8

const (
	wordElement elementKind = iota
	literalElement
	symbolElement
	commentElement
	newlineElement
)

type element struct {
	kind  elementKind
	value string
}

// Source formats one complete Silver source file. Invalid source is returned
// unchanged by callers because formatting only starts after a successful parse.
func Source(name string, source []byte) ([]byte, error) {
	p := parser.New(lexer.NewWithSource(string(source), name))
	p.ParseProgram()
	if diagnostics := p.Errors(); len(diagnostics) != 0 {
		return nil, errors.New(strings.Join(diagnostics, "\n"))
	}
	if len(source) == 0 {
		return []byte{}, nil
	}

	return render(scan(string(source))), nil
}

// File formats exactly one regular file in place and reports whether its
// contents changed.
func File(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("could not access %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("cannot format %q: not a regular file", path)
	}

	source, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("could not read %q: %w", path, err)
	}
	formatted, err := Source(path, source)
	if err != nil {
		return false, fmt.Errorf("could not format %q: %w", path, err)
	}
	if string(formatted) == string(source) {
		return false, nil
	}
	if err := os.WriteFile(path, formatted, info.Mode().Perm()); err != nil {
		return false, fmt.Errorf("could not write %q: %w", path, err)
	}
	return true, nil
}

func scan(source string) []element {
	elements := make([]element, 0, len(source)/2)
	for index := 0; index < len(source); {
		ch := source[index]
		switch {
		case ch == ' ' || ch == '\t':
			index++
		case ch == '\r' || ch == '\n':
			if ch == '\r' && index+1 < len(source) && source[index+1] == '\n' {
				index++
			}
			elements = append(elements, element{kind: newlineElement, value: "\n"})
			index++
		case ch == '#':
			start := index
			for index < len(source) && source[index] != '\r' && source[index] != '\n' {
				index++
			}
			elements = append(elements, element{kind: commentElement, value: strings.TrimRight(source[start:index], " \t")})
		case ch == '"':
			start := index
			index++
			escaped := false
			for index < len(source) {
				current := source[index]
				index++
				if current == '"' && !escaped {
					break
				}
				if escaped {
					escaped = false
				} else if current == '\\' {
					escaped = true
				}
			}
			elements = append(elements, element{kind: literalElement, value: source[start:index]})
		case strings.HasPrefix(source[index:], "```"):
			start := index
			index = scanTemplate(source, index)
			elements = append(elements, element{kind: literalElement, value: source[start:index]})
		case isWordStart(ch):
			start := index
			for index < len(source) && isWordPart(source[index]) {
				index++
			}
			elements = append(elements, element{kind: wordElement, value: source[start:index]})
		case isDigit(ch):
			start := index
			for index < len(source) && isDigit(source[index]) {
				index++
			}
			if index+1 < len(source) && source[index] == '.' && isDigit(source[index+1]) {
				index++
				for index < len(source) && isDigit(source[index]) {
					index++
				}
			}
			elements = append(elements, element{kind: literalElement, value: source[start:index]})
		default:
			value := string(ch)
			if index+1 < len(source) && isDoubleSymbol(source[index:index+2]) {
				value = source[index : index+2]
				index++
			}
			elements = append(elements, element{kind: symbolElement, value: value})
			index++
		}
	}
	return elements
}

func render(elements []element) []byte {
	var output strings.Builder
	line := make([]element, 0, 16)
	depth := 0
	pendingBlank := false
	wroteLine := false
	activeCases := make(map[int]bool)

	flush := func() {
		if len(line) == 0 {
			if wroteLine {
				pendingBlank = true
			}
			return
		}
		if pendingBlank {
			output.WriteByte('\n')
			pendingBlank = false
		}

		leadingClosers := 0
		for leadingClosers < len(line) && line[leadingClosers].kind == symbolElement && isClosing(line[leadingClosers].value) {
			leadingClosers++
		}
		lineDepth := depth - leadingClosers
		if lineDepth < 0 {
			lineDepth = 0
		}
		firstWord := firstWord(line)
		caseClause := firstWord == "case" || firstWord == "default"
		extra := 0
		for caseDepth := range activeCases {
			if lineDepth >= caseDepth && !(caseClause && lineDepth == caseDepth) {
				extra++
			}
		}
		output.WriteString(strings.Repeat(" ", (lineDepth+extra)*indentWidth))
		output.WriteString(formatLine(line))
		output.WriteByte('\n')
		wroteLine = true

		for _, item := range line {
			if item.kind != symbolElement {
				continue
			}
			switch item.value {
			case "{", "(", "[":
				depth++
			case "}", ")", "]":
				if depth > 0 {
					depth--
				}
				for caseDepth := range activeCases {
					if caseDepth > depth {
						delete(activeCases, caseDepth)
					}
				}
			}
		}
		if caseClause {
			activeCases[depth] = true
		}
		line = line[:0]
	}

	for _, item := range elements {
		if item.kind == newlineElement {
			flush()
			continue
		}
		line = append(line, item)
	}
	flush()
	return []byte(output.String())
}

func formatLine(line []element) string {
	var output strings.Builder
	blockBraces := findBlockBraces(line)
	for index, current := range line {
		if index > 0 && needsSpace(line, index, blockBraces) {
			output.WriteByte(' ')
		}
		output.WriteString(current.value)
	}
	return output.String()
}

func needsSpace(line []element, index int, blockBraces map[int]bool) bool {
	previous, current := line[index-1], line[index]
	if current.kind == commentElement {
		return true
	}
	if previous.kind == commentElement {
		return false
	}
	if current.kind == symbolElement {
		switch current.value {
		case ",", ")", "]", ".", ":":
			return false
		case "}":
			return blockBraces[index] && previous.value != "{"
		case "(":
			return isOperator(previous.value) || previous.kind == wordElement && spacedBeforeParen(previous.value)
		case "[":
			return isOperator(previous.value) || previous.kind == wordElement && statementKeyword(previous.value)
		case "{":
			if blockBraces[index] {
				return previous.value != "{"
			}
			return isOperator(previous.value)
		case "!":
			return previous.kind == wordElement || previous.kind == literalElement || isClosing(previous.value)
		case "-":
			if unaryAt(line, index) {
				return previous.kind == wordElement && statementKeyword(previous.value)
			}
			return true
		default:
			return isOperator(current.value)
		}
	}
	if previous.kind == symbolElement {
		switch previous.value {
		case "(", "[", ".":
			return false
		case ")", "]", "}":
			return true
		case "{":
			return blockBraces[index-1]
		case "!":
			return false
		case "-":
			return !unaryAt(line, index-1)
		case ",", ":":
			return true
		default:
			return isOperator(previous.value)
		}
	}
	return true
}

func findBlockBraces(line []element) map[int]bool {
	result := make(map[int]bool)
	lastOpen := -1
	controlSeen := false
	declaration := false
	for index, item := range line {
		if item.kind == wordElement {
			switch item.value {
			case "fn", "if", "else", "for", "while", "try", "catch", "switch":
				controlSeen = true
			case "struct", "enum":
				declaration = firstWord(line) == item.value
			}
		}
		if item.kind == symbolElement && item.value == "{" {
			lastOpen = index
		}
	}
	if lastOpen >= 0 && (controlSeen || declaration) {
		result[lastOpen] = true
		for index, item := range line {
			if item.kind == symbolElement && item.value == "}" && index > lastOpen {
				result[index] = true
			}
		}
	}
	return result
}

func unaryAt(line []element, index int) bool {
	if index == 0 {
		return true
	}
	previous := line[index-1]
	if previous.kind == symbolElement {
		return isOperator(previous.value) || previous.value == "(" || previous.value == "[" || previous.value == "{" || previous.value == "," || previous.value == ":"
	}
	return previous.kind == wordElement && statementKeyword(previous.value)
}

func firstWord(line []element) string {
	for _, item := range line {
		if item.kind == wordElement {
			return item.value
		}
		if item.kind != commentElement {
			break
		}
	}
	return ""
}

func spacedBeforeParen(word string) bool {
	switch word {
	case "if", "for", "while", "switch", "return", "assert", "defer", "task", "collect":
		return true
	default:
		return false
	}
}

func statementKeyword(word string) bool {
	switch word {
	case "let", "return", "assert", "defer", "if", "else", "for", "in", "while", "task", "collect", "switch", "case", "catch":
		return true
	default:
		return false
	}
}

func isOperator(value string) bool {
	switch value {
	case "=", "+", "-", "*", "**", "/", "//", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||", "|":
		return true
	default:
		return false
	}
}

func isClosing(value string) bool {
	return value == ")" || value == "]" || value == "}"
}

func isDoubleSymbol(value string) bool {
	switch value {
	case "==", "!=", "&&", "||", "**", "//", "<=", ">=":
		return true
	default:
		return false
	}
}

func scanTemplate(source string, start int) int {
	index := start + 3
	inExpression := false
	braceDepth := 0
	for index < len(source) {
		if !inExpression {
			switch {
			case strings.HasPrefix(source[index:], "```"):
				return index + 3
			case strings.HasPrefix(source[index:], "{{"), strings.HasPrefix(source[index:], "}}"):
				index += 2
			case source[index] == '{':
				inExpression = true
				braceDepth = 0
				index++
			default:
				index++
			}
			continue
		}

		switch {
		case source[index] == '"':
			index++
			escaped := false
			for index < len(source) {
				current := source[index]
				index++
				if current == '"' && !escaped {
					break
				}
				if escaped {
					escaped = false
				} else if current == '\\' {
					escaped = true
				}
			}
		case strings.HasPrefix(source[index:], "```"):
			index = scanTemplate(source, index)
		case source[index] == '#':
			for index < len(source) && source[index] != '\r' && source[index] != '\n' {
				index++
			}
		case source[index] == '{':
			braceDepth++
			index++
		case source[index] == '}':
			if braceDepth == 0 {
				inExpression = false
			} else {
				braceDepth--
			}
			index++
		default:
			index++
		}
	}
	return index
}

func isWordStart(ch byte) bool {
	return ch == '_' || ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func isWordPart(ch byte) bool { return isWordStart(ch) || isDigit(ch) }

func isDigit(ch byte) bool { return ch >= '0' && ch <= '9' }
