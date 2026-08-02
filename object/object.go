package object

import (
	"bytes"
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
	STRUCT_OBJ       = "STRUCT"
	STRUCT_VALUE_OBJ = "STRUCT_VALUE"
	TYPE_OBJ         = "TYPE"
	TASK_OBJ         = "TASK"
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
// declaration time, and Name is assigned when the closure is bound by let.
type Function struct {
	Name       string
	Parameters []*ast.Identifier
	ReturnType *ast.TypeAnnotation
	ErrorTypes []*ast.TypeAnnotation
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
		params = append(params, p.DeclarationString())
	}

	out.WriteString("fn")
	out.WriteString("(")
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(")")
	if f.ReturnType != nil {
		out.WriteString(" ")
		out.WriteString(f.ReturnType.String())
	}
	for _, errorType := range f.ErrorTypes {
		out.WriteString(" | ")
		out.WriteString(errorType.String())
	}
	out.WriteString(" {\n")
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

// ReturnValue is an internal control-flow wrapper used to carry a value out of
// nested evaluator blocks. It is unwrapped at the function boundary.
type ReturnValue struct {
	Value Object
}

// Type returns the internal return-wrapper tag.
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

// Inspect delegates to the wrapped language value.
func (rv *ReturnValue) Inspect() string { return rv.Value.Inspect() }

// Null is the absence-of-value object. It carries no per-instance state.
type Null struct{}

// Type returns the null runtime tag.
func (n *Null) Type() ObjectType { return NULL_OBJ }

// Inspect returns Silver's null literal spelling.
func (n *Null) Inspect() string { return "null" }
