package object

import (
	"silver/ast"
	"strings"
)

// Struct is the constructor and field layout bound by a struct declaration.
type Struct struct {
	Name       string
	Fields     []string
	FieldTypes []*ast.TypeAnnotation
	Env        *Environment
}

// Type returns the struct-constructor runtime tag.
func (s *Struct) Type() ObjectType { return STRUCT_OBJ }

// Inspect returns a compact struct type description.
func (s *Struct) Inspect() string { return "<struct " + s.Name + ">" }

// StructInstance stores the values of one constructed struct in field-name
// form. Struct retains their stable declaration order for inspection.
type StructInstance struct {
	Struct *Struct
	Values map[string]Object
}

// Type returns the struct-value runtime tag.
func (s *StructInstance) Type() ObjectType { return STRUCT_VALUE_OBJ }

// Inspect renders a value using the struct's declared field order.
func (s *StructInstance) Inspect() string {
	fields := make([]string, 0, len(s.Struct.Fields))
	for _, name := range s.Struct.Fields {
		fields = append(fields, name+": "+s.Values[name].Inspect())
	}
	return s.Struct.Name + " { " + strings.Join(fields, ", ") + " }"
}
