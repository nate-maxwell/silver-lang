package token

// TokenType identifies a lexical category such as an identifier, operator, or
// keyword.
type TokenType string

// Position identifies the start of a token in its source. Line and Column are
// one-based for display, while Offset is a zero-based byte offset useful for
// tooling that needs to index the original source.
type Position struct {
	Source string
	Offset int
	Line   int
	Column int
}

// IsValid reports whether the position contains displayable source
// coordinates. Manually constructed AST nodes may have a zero position.
func (p Position) IsValid() bool {
	return p.Line > 0 && p.Column > 0
}

// Token is one lexical unit produced by the lexer. Position points to the
// first byte of Literal in the original source.
type Token struct {
	Type     TokenType
	Literal  string
	Position Position
}

// Token type constants are grouped by their lexical role. Their string values
// are also used in parser diagnostics.
const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + literals
	IDENT          = "IDENT"          // add, foobar, x, y, ...
	INT            = "INT"            // 123456
	FLOAT          = "FLOAT"          // 123.456
	STRING         = "STRING"         // "foobar" / "foo bar"
	TEMPLATE_START = "TEMPLATE_START" // ```
	TEMPLATE_TEXT  = "TEMPLATE_TEXT"  // uninterpolated template contents
	TEMPLATE_END   = "TEMPLATE_END"   // ```

	// Operators
	ASSIGN = "="
	BANG   = "!"
	PIPE   = "|"
	AND    = "&&"
	OR     = "||"

	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	POWER    = "**"
	SLASH    = "/"
	INT_DIV  = "//"
	MODULO   = "%"

	LT     = "<"
	GT     = ">"
	LTE    = "<="
	GTE    = ">="
	EQ     = "=="
	NOT_EQ = "!="

	// Delimiters
	COMMA = ","
	COLON = ":"
	EMBED = "::"
	DOT   = "."

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	// Keywords
	FUNCTION = "FUNCTION"
	LET      = "LET"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
	ASSERT   = "ASSERT"
	DEFER    = "DEFER"
	IMPORT   = "IMPORT"
	EXPORT   = "EXPORT"
	ENUM     = "ENUM"
	STRUCT   = "STRUCT"
	FOR      = "FOR"
	IN       = "IN"
	WHILE    = "WHILE"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
	TASK     = "TASK"
	COLLECT  = "COLLECT"
	TRY      = "TRY"
	CATCH    = "CATCH"
	SWITCH   = "SWITCH"
	CASE     = "CASE"
	DEFAULT  = "DEFAULT"
)

// keywords maps reserved source words to their specialized token types. Any
// word absent from this table is treated as an identifier.
var keywords = map[string]TokenType{
	"fn":       FUNCTION,
	"let":      LET,
	"True":     TRUE,
	"False":    FALSE,
	"if":       IF,
	"else":     ELSE,
	"return":   RETURN,
	"assert":   ASSERT,
	"defer":    DEFER,
	"import":   IMPORT,
	"export":   EXPORT,
	"enum":     ENUM,
	"struct":   STRUCT,
	"for":      FOR,
	"in":       IN,
	"while":    WHILE,
	"break":    BREAK,
	"continue": CONTINUE,
	"task":     TASK,
	"collect":  COLLECT,
	"try":      TRY,
	"catch":    CATCH,
	"switch":   SWITCH,
	"case":     CASE,
	"default":  DEFAULT,
}

// LookupIdent classifies an identifier as a reserved keyword or a normal
// user-defined name.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
