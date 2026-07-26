package object

import (
	"fmt"
	"strings"
)

// TraceFrame is one user-visible execution frame. Frames store structured data
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

// Error is a Silver runtime failure. Message remains separate from Frames so
// frontends can render structured traceback data in their preferred format.
type Error struct {
	Message string
	Frames  []TraceFrame
}

// Type returns the runtime error tag.
func (e *Error) Type() ObjectType { return ERROR_OBJ }

// Inspect renders a Python-style traceback followed by the error message. An
// error without frames retains the compact legacy ERROR form.
func (e *Error) Inspect() string {
	if len(e.Frames) == 0 {
		return "ERROR: " + e.Message
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
	out.WriteString("ERROR: ")
	out.WriteString(e.Message)
	return out.String()
}

// SetOrigin records the innermost location that produced an error. Propagating
// AST nodes call this freely; only the first valid origin is retained.
func (e *Error) SetOrigin(frame TraceFrame) {
	if len(e.Frames) == 0 && frame.IsValid() {
		e.Frames = append(e.Frames, frame)
	}
}

// PrependFrame adds an outer caller so frames remain ordered like Python
// tracebacks: outermost call first, error origin last.
func (e *Error) PrependFrame(frame TraceFrame) {
	if !frame.IsValid() {
		return
	}
	e.Frames = append([]TraceFrame{frame}, e.Frames...)
}

// HasTraceback reports whether an origin or caller frame has been attached.
func (e *Error) HasTraceback() bool {
	return len(e.Frames) != 0
}
