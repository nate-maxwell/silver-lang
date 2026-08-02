package object

import (
	"silver/ast"
	"strings"
	"sync"
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
	mu     sync.RWMutex
}

// BoundMethod pairs a callable struct field with the instance that supplied
// it. Calling the method injects Receiver as the function's first argument.
type BoundMethod struct {
	Method   *Function
	Receiver *StructInstance
	Name     string
}

// Type makes a bound method usable anywhere a callable value is accepted.
func (bm *BoundMethod) Type() ObjectType { return FUNCTION_OBJ }

// Inspect delegates to the stored function.
func (bm *BoundMethod) Inspect() string { return bm.Method.Inspect() }

// Type returns the struct-value runtime tag.
func (s *StructInstance) Type() ObjectType { return STRUCT_VALUE_OBJ }

// Inspect renders a value using the struct's declared field order.
func (s *StructInstance) Inspect() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fields := make([]string, 0, len(s.Struct.Fields))
	for _, name := range s.Struct.Fields {
		fields = append(fields, name+": "+s.Values[name].Inspect())
	}
	return s.Struct.Name + " { " + strings.Join(fields, ", ") + " }"
}

// Get returns one field using synchronization suitable for task sharing.
func (s *StructInstance) Get(name string) (Object, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.Values[name]
	return value, ok
}

// Set replaces one field using synchronization suitable for task sharing.
func (s *StructInstance) Set(name string, value Object) {
	s.mu.Lock()
	s.Values[name] = value
	s.mu.Unlock()
}
