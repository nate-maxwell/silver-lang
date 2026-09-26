package ast

import (
	"bytes"
	"silver/token"
	"strings"
)

// TypeAnnotation is a primitive, alias, or qualified nominal type name, such
// as int or geometry.Point, optionally with an array element contract.
// Callable signatures may name parameters, and an omitted
// ReturnType means null; ErrorTypes are the struct error alternatives
// following it. A non-nil ParameterTypes slice distinguishes a signature,
// including call(), from the broad call type.
type TypeAnnotation struct {
	Token          token.Token // the first identifier in the type name
	Parts          []string
	ElementType    *TypeAnnotation // array[T]; nil denotes the broad array type
	ParameterNames []string
	ParameterTypes []*TypeAnnotation
	Variadic       bool // final callable parameter accepts zero or more values
	ReturnType     *TypeAnnotation
	ErrorTypes     []*TypeAnnotation
}

// Position returns the first type-name component's position.
func (ta *TypeAnnotation) Position() token.Position { return ta.Token.Position }

// IsCallSignature reports whether the annotation includes a callable
// parameter and return signature rather than naming the broad call type.
func (ta *TypeAnnotation) IsCallSignature() bool {
	return ta != nil && len(ta.Parts) == 1 && ta.Parts[0] == "call" && ta.ParameterTypes != nil
}

// String renders the complete qualified type or callable signature.
func (ta *TypeAnnotation) String() string {
	if !ta.IsCallSignature() {
		if ta.ElementType != nil {
			return strings.Join(ta.Parts, ".") + "[" + ta.ElementType.String() + "]"
		}
		return strings.Join(ta.Parts, ".")
	}

	var out bytes.Buffer
	out.WriteString("call(")
	for index, parameterType := range ta.ParameterTypes {
		if index > 0 {
			out.WriteString(", ")
		}
		if index < len(ta.ParameterNames) && ta.ParameterNames[index] != "" {
			out.WriteString(ta.ParameterNames[index])
			if parameterType != nil {
				out.WriteString(": ")
			}
		}
		if parameterType != nil {
			out.WriteString(parameterType.String())
		}
		if ta.Variadic && index == len(ta.ParameterTypes)-1 {
			out.WriteString("...")
		}
	}
	out.WriteString(")")
	if ta.ReturnType != nil {
		out.WriteString(" ")
		out.WriteString(ta.ReturnType.String())
	}
	for _, errorType := range ta.ErrorTypes {
		out.WriteString(" | ")
		out.WriteString(errorType.String())
	}
	return out.String()
}

func (ta *TypeAnnotation) typeExpressionNode()  {}
func (ta *TypeAnnotation) TokenLiteral() string { return ta.Token.Literal }
