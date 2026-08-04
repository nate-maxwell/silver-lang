package object

import (
	"fmt"
	"strings"
)

// TraceFrame is one user-visible execution frame. Frames store structured data,
// so alternate frontends can render tracebacks differently without parsing a
// preformatted error string.
type TraceFrame struct {
	Source   string // filename or synthetic input label
	Line     int    // one-based source line
	Column   int    // one-based byte column
	Function string // active Silver function, or <module>
}

// IsValid reports whether the frame has displayable source coordinates.
func (f TraceFrame) IsValid() bool {
	return f.Line > 0 && f.Column > 0
}

// Error is the one failure object used by the evaluator. Value is always
// an error-struct instance: either a declared Silver error alternative or the
// built-in Error struct used for ordinary runtime faults.
type Error struct {
	Value  *StructInstance
	Frames []TraceFrame
}

// NewError creates an error containing the built-in Error struct.
func NewError(message string) *Error {
	definition, ok := BuiltinStructDefinitionByName("Error")
	if !ok {
		panic("object: built-in Error struct is not registered")
	}
	return &Error{
		Value: &StructInstance{
			Struct: definition,
			Values: map[string]Object{"message": &String{Value: message}},
		},
	}
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Inspect renders a Python-style traceback followed by the error struct's
// nominal type and message. The complete struct remains available for typed
// catch binding and field access.
func (e *Error) Inspect() string {
	ending := e.ending()
	if len(e.Frames) == 0 {
		return ending
	}

	var out strings.Builder
	out.WriteString("Traceback (most recent call last):\n")
	for _, frame := range e.Frames {
		function := frame.Function
		if function == "" {
			function = "<module>"
		}
		fmt.Fprintf(
			&out,
			"  File \"%s\", line %d, column %d, in %s\n",
			frame.Source,
			frame.Line,
			frame.Column,
			function,
		)
	}
	out.WriteString(ending)
	return out.String()
}

func (e *Error) ending() string {
	name := "Error"
	detail := e.MessageText()
	if e.Value != nil && e.Value.Struct != nil {
		name = e.Value.Struct.Name
		if detail == "" {
			detail = e.Value.Inspect()
		}
	}
	if detail == "" {
		return name
	}
	return name + ": " + detail
}

// MessageText returns the conventional message field from the carried error
// struct. It exists for Go frontends and tests; the struct remains the source
// of truth rather than duplicating message state on Error.
func (e *Error) MessageText() string {
	if e == nil || e.Value == nil {
		return ""
	}
	message, ok := e.Value.Get("message")
	if !ok {
		return ""
	}
	text, ok := message.(*String)
	if !ok {
		return ""
	}
	return text.Value
}

// IsRuntimeError reports whether this error contains Silver's built-in,
// unchecked Error struct rather than a declared callable error alternative.
func (e *Error) IsRuntimeError() bool {
	if e == nil || e.Value == nil {
		return false
	}
	definition, ok := BuiltinStructDefinitionByName("Error")
	return ok && e.Value.Struct == definition
}

// SetOrigin records the innermost location that produced an error.
// Propagating AST nodes call this freely; only the first valid origin is kept.
func (e *Error) SetOrigin(frame TraceFrame) {
	if len(e.Frames) == 0 && frame.IsValid() {
		e.Frames = append(e.Frames, frame)
	}
}

// PrependFrame adds an outer caller so frames remain ordered like Python
// tracebacks: outermost call first, error origin last.
func (e *Error) PrependFrame(frame TraceFrame) {
	if frame.IsValid() {
		e.Frames = append([]TraceFrame{frame}, e.Frames...)
	}
}

// HasTraceback reports whether an origin or caller frame has been attached.
func (e *Error) HasTraceback() bool { return len(e.Frames) != 0 }
