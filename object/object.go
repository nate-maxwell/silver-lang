package object

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"silver/ast"
	"strings"
)

// ObjectType is a stable runtime tag used for type checks and diagnostics.
type ObjectType string

// Runtime object tags. These strings are user-visible in evaluator errors, so
// changing them is an observable language change.
const (
	INTEGER_OBJ      = "INTEGER"
	FLOAT_OBJ        = "FLOAT"
	BOOLEAN_OBJ      = "BOOLEAN"
	STRING_OBJ       = "STRING"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTINT_OBJ     = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
	MODULE_OBJ       = "MODULE"
	ENUM_OBJ         = "ENUM"
	ENUM_VALUE_OBJ   = "ENUM_VALUE"
)

// Object is any value that can exist at runtime in a Silver program.
type Object interface {
	Type() ObjectType
	Inspect() string
}

// Module represents one evaluated source file and its exported top-level
// bindings. Path is the canonical path used by the evaluator's module cache.
type Module struct {
	Path    string
	Exports map[string]Object
}

// Type returns the module runtime tag.
func (m *Module) Type() ObjectType { return MODULE_OBJ }

// Inspect returns a compact module description for diagnostics.
func (m *Module) Inspect() string { return "<module " + m.Path + ">" }

// BuiltinFunction is the Go calling convention used by native Silver
// functions. Language errors are returned as Error objects rather than Go
// errors so they propagate through normal evaluation.
type BuiltinFunction func(args ...Object) Object

// Function is a Silver closure. Env captures the lexical environment active at
// declaration time, and Name is assigned when the closure is bound by let for
// use in tracebacks.
type Function struct {
	Name       string
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

// Type returns the user-defined function runtime tag.
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

// Inspect renders a readable representation of the closure's signature and
// body.
func (f *Function) Inspect() string {
	var out bytes.Buffer

	params := []string{}
	for _, p := range f.Parameters {
		params = append(params, p.String())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") {\n")
	out.WriteString(f.Body.String())
	out.WriteString("\n}")

	return out.String()
}

// Builtin wraps a Go function so it can participate in Silver calls.
type Builtin struct {
	Fn BuiltinFunction
}

// Type returns the native function runtime tag.
func (b *Builtin) Type() ObjectType { return BUILTINT_OBJ }

// Inspect returns a generic description because builtin names live in the
// evaluator registry rather than on the value itself.
func (b *Builtin) Inspect() string { return "builtin function" }

// HashKey converts a boolean to the stable 0/1 representation used as a hash
// table key.
func (b *Boolean) HashKey() HashKey {
	var value uint64

	if b.Value {
		value = 1
	} else {
		value = 0
	}

	return HashKey{Type: b.Type(), Value: value}
}

// ReturnValue is an internal control-flow wrapper used to carry a value out of
// nested evaluator blocks. It is unwrapped at the function boundary.
type ReturnValue struct {
	Value Object
}

// Type returns the internal return-wrapper tag.
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Inspect delegates to the wrapped language value.
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// Integer stores a signed 64-bit Silver integer.
type Integer struct {
	Value int64
}

// Inspect returns the base-10 representation of the integer.
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Type returns the integer runtime tag.
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// HashKey uses the integer bits directly as the hash payload.
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}

// Boolean stores a Silver truth value. The evaluator normally reuses singleton
// instances so equality can compare boolean object identity.
type Boolean struct {
	Value bool
}

// Type returns the boolean runtime tag.
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// Inspect returns True's or False's Go-style lowercase representation.
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// String stores an immutable Silver string value.
type String struct {
	Value string
}

// Type returns the string runtime tag.
func (s *String) Type() ObjectType { return STRING_OBJ }

// Inspect returns the raw string value without adding quotes.
func (s *String) Inspect() string { return s.Value }

// HashKey computes a stable FNV-1a digest for use in a Silver hash.
func (s *String) HashKey() HashKey {
	h := fnv.New64a()
	h.Write([]byte(s.Value))

	return HashKey{Type: s.Type(), Value: h.Sum64()}
}

// Array stores an ordered sequence of runtime objects.
type Array struct {
	Elements []Object
}

// Type returns the array runtime tag.
func (ao *Array) Type() ObjectType { return ARRAY_OBJ }

// Inspect renders the elements as a comma-separated bracketed list.
func (ao *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, e := range ao.Elements {
		elements = append(elements, e.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")

	return out.String()
}

// Null is the absence-of-value object. It carries no per-instance state.
type Null struct{}

// Type returns the null runtime tag.
func (n *Null) Type() ObjectType { return NULL_OBJ }

// Inspect returns Silver's null literal spelling.
func (n *Null) Inspect() string { return "null" }

// Hashable is implemented by runtime values that may be used as hash keys.
type Hashable interface {
	HashKey() HashKey
}

// HashPair retains both the original key object and its associated value. The
// original object is needed when rendering the hash.
type HashPair struct {
	Key   Object
	Value Object
}

// Hash stores pairs by their normalized HashKey.
type Hash struct {
	Pairs map[HashKey]HashPair
}

// HashKey combines a runtime type tag with a type-specific 64-bit payload so,
// for example, integer 1 and boolean true remain distinct keys.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Type returns the hash runtime tag.
func (h *Hash) Type() ObjectType { return HASH_OBJ }

// Inspect renders the hash's key/value pairs. Go map iteration means pair order
// is intentionally unspecified.
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")

	return out.String()
}
