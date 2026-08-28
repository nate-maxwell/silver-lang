package parser

import (
	"fmt"
	"silver/lexer"
	"silver/token"
)

// InfixRegistry holds the symbolic operators known to an interpreter or
// standalone parser session. Evaluators share one registry across files,
// imports, tasks, and REPL submissions.
type InfixRegistry struct {
	definitions map[string]token.Position
}

// OperatorDeclaration is the grammar portion of an operator statement found
// during a package pre-scan.
type OperatorDeclaration struct {
	Symbol   string
	Position token.Position
}

var definedLanguageInfixOperators = map[string]bool{
	"=": true, ".": true, ":": true, "...": true,
	"||": true, "&&": true, "==": true, "!=": true,
	"<": true, ">": true, "<=": true, ">=": true,
	"+": true, "-": true, "*": true, "/": true,
	"//": true, "%": true, "**": true,
}

func NewInfixRegistry() *InfixRegistry {
	return &InfixRegistry{
		definitions: make(map[string]token.Position),
	}
}

// HasUserOperators reports whether this session's grammar has been extended.
func (r *InfixRegistry) HasUserOperators() bool {
	return len(r.definitions) != 0
}

// Has reports whether symbol is a user-defined operator in this registry.
// Evaluators use this to distinguish package-local operator identities from
// the operators built into the language.
func (r *InfixRegistry) Has(symbol string) bool {
	_, ok := r.definitions[symbol]
	return ok
}

func (r *InfixRegistry) define(symbol string, position token.Position) string {
	if definedLanguageInfixOperators[symbol] {
		return fmt.Sprintf("operator %q is already defined by the language", symbol)
	}
	if previous, exists := r.definitions[symbol]; exists {
		if previous.Source == position.Source && previous.Offset == position.Offset {
			return ""
		}
		message := fmt.Sprintf("operator %q is already defined", symbol)
		if previous.IsValid() {
			message += fmt.Sprintf(" at %s:%d:%d", previous.Source, previous.Line, previous.Column)
		}
		return message
	}
	r.definitions[symbol] = position
	return ""
}

// Predefine installs a discovered package operator before full parsing. A
// later parse of the declaration at the same source position is idempotent.
func (r *InfixRegistry) Predefine(declaration OperatorDeclaration) string {
	return r.define(declaration.Symbol, declaration.Position)
}

// DiscoverOperatorDeclarations performs a lexical pre-scan without parsing
// expressions. It lets every file in a package see every package operator,
// independent of manifest ordering, while strings, comments, and template
// text remain opaque to discovery.
func DiscoverOperatorDeclarations(input, source string) []OperatorDeclaration {
	l := lexer.NewWithSource(input, source)
	var tokens []token.Token
	for {
		current := l.NextToken()
		tokens = append(tokens, current)
		if current.Type == token.EOF {
			break
		}
	}

	var declarations []OperatorDeclaration
	for index, current := range tokens {
		if current.Type != token.OPERATOR || index+1 >= len(tokens) {
			continue
		}
		cursor := index + 1
		if tokens[cursor].Type == token.FUNCTION || tokens[cursor].Type == token.ASSIGN {
			continue
		}
		var symbol string
		var previous token.Token
		for cursor < len(tokens) {
			part := tokens[cursor]
			if part.Type == token.ASSIGN && cursor+1 < len(tokens) && tokens[cursor+1].Type == token.FUNCTION {
				cursor++
				break
			}
			if !isOperatorFragment(part.Literal) || symbol != "" &&
				(part.Position.Source != previous.Position.Source || part.Position.Offset != previous.Position.Offset+len(previous.Literal)) {
				break
			}
			symbol += part.Literal
			previous = part
			cursor++
		}
		if symbol != "" && cursor < len(tokens) && tokens[cursor].Type == token.FUNCTION {
			declarations = append(declarations, OperatorDeclaration{Symbol: symbol, Position: current.Position})
		}
	}
	return declarations
}

// precedences assigns binding power to infix token types. Tokens absent from
// the table use LOWEST, which terminates the current Pratt-parser expression.
var precedences = map[token.TokenType]int{
	token.OR:       OR,
	token.AND:      AND,
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.INT_DIV:  PRODUCT,
	token.MODULO:   PRODUCT,
	token.ASTERISK: PRODUCT,
	token.POWER:    POWER,
	token.LPAREN:   CALL,
	token.LBRACE:   CALL,
	token.LBRACKET: INDEX,
	token.DOT:      MEMBER,
}

// peekPrecedence returns the binding power of the lookahead token.
// User-defined operators always use the language's lowest infix power.
func (p *Parser) peekPrecedence() int {
	if p.operators.Has(p.peekToken.Literal) {
		return CUSTOM
	}
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence returns the binding power of the current token.
func (p *Parser) curPrecedence() int {
	if p.operators.Has(p.curToken.Literal) {
		return CUSTOM
	}
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
