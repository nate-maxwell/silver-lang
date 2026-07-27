package object

import (
	"bytes"
	"strings"
)

// Array stores an ordered sequence of runtime objects.
type Array struct {
	Elements []Object
}

// Type returns the array runtime tag.
func (a *Array) Type() ObjectType { return ARRAY_OBJ }

// Inspect renders the elements as a comma-separated bracketed list.
func (a *Array) Inspect() string {
	var out bytes.Buffer

	elements := []string{}
	for _, element := range a.Elements {
		elements = append(elements, element.Inspect())
	}

	out.WriteString("[")
	out.WriteString(strings.Join(elements, ", "))
	out.WriteString("]")
	return out.String()
}
