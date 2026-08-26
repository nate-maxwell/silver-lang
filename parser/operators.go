package parser

import (
	"fmt"
	"silver/lexer"
	"silver/token"
	"strconv"
)

// InfixRegistry holds the symbolic operators known to an interpreter or
// standalone parser session. Evaluators share one registry across files,
// imports, tasks, and REPL submissions.
type InfixRegistry struct {
	precedences map[string]int
	definitions map[string]token.Position
}

// OperatorDeclaration is the grammar portion of an operator statement found
// during a package pre-scan.
type OperatorDeclaration struct {
	Symbol       string
	BindingPower int
	Position     token.Position
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
		precedences: make(map[string]int),
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

func (r *InfixRegistry) define(symbol string, power int, position token.Position) string {
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
	r.precedences[symbol] = power
	r.definitions[symbol] = position
	return ""
}

// Predefine installs a discovered package operator before full parsing. A
// later parse of the declaration at the same source position is idempotent.
func (r *InfixRegistry) Predefine(declaration OperatorDeclaration) string {
	return r.define(declaration.Symbol, declaration.BindingPower, declaration.Position)
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
		if tokens[cursor].Type == token.FUNCTION || tokens[cursor].Type == token.INT {
			continue
		}
		symbol := tokens[cursor].Literal
		if !isOperatorFragment(symbol) {
			continue
		}
		previous := tokens[cursor]
		cursor++
		for cursor < len(tokens) && tokens[cursor].Type != token.FUNCTION && tokens[cursor].Type != token.INT {
			part := tokens[cursor]
			if !isOperatorFragment(part.Literal) || part.Position.Source != previous.Position.Source || part.Position.Offset != previous.Position.Offset+len(previous.Literal) {
				break
			}
			symbol += part.Literal
			previous = part
			cursor++
		}
		power := SUM
		if cursor < len(tokens) && tokens[cursor].Type == token.INT {
			parsed, err := strconv.Atoi(tokens[cursor].Literal)
			if err != nil || parsed <= LOWEST {
				continue
			}
			power = parsed
			cursor++
		}
		if cursor < len(tokens) && tokens[cursor].Type == token.FUNCTION {
			declarations = append(declarations, OperatorDeclaration{Symbol: symbol, BindingPower: power, Position: current.Position})
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
// It checks the InfixRegistry first, then the precedences map.
func (p *Parser) peekPrecedence() int {
	if precedence, ok := p.operators.precedences[p.peekToken.Literal]; ok {
		return precedence
	}
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// curPrecedence returns the binding power of the current token.
func (p *Parser) curPrecedence() int {
	if precedence, ok := p.operators.precedences[p.curToken.Literal]; ok {
		return precedence
	}
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}
