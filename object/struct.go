package object

import (
	"silver/ast"
	"strings"
	"sync"
)

// Struct is the constructor and field layout bound by a struct declaration.
type Struct struct {
	Name           string
	Fields         []string
	FieldTypes     []*ast.TypeAnnotation
	EmbeddedFields []bool
	Env            *Environment
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
	value, _, _, ok := s.ResolveField(name)
	return value, ok
}

// ResolveField returns a direct field or a field promoted recursively through
// embedded struct values. Direct fields take precedence over promoted fields.
// Owner and index identify the instance and declaration that actually contain
// the value, which lets callers apply method binding and assignment correctly.
func (s *StructInstance) ResolveField(name string) (value Object, owner *StructInstance, index int, ok bool) {
	s.mu.RLock()
	value, exists := s.Values[name]
	if exists {
		fieldIndex := -1
		for index, fieldName := range s.Struct.Fields {
			if fieldName == name {
				fieldIndex = index
				break
			}
		}
		s.mu.RUnlock()
		return value, s, fieldIndex, true
	}
	for fieldIndex, fieldName := range s.Struct.Fields {
		if fieldName == name {
			s.mu.RUnlock()
			return nil, s, fieldIndex, false
		}
	}
	embedded := make([]Object, 0, len(s.Struct.EmbeddedFields))
	for fieldIndex, isEmbedded := range s.Struct.EmbeddedFields {
		if isEmbedded && fieldIndex < len(s.Struct.Fields) {
			embedded = append(embedded, s.Values[s.Struct.Fields[fieldIndex]])
		}
	}
	s.mu.RUnlock()

	for _, candidate := range embedded {
		if instance, isStruct := candidate.(*StructInstance); isStruct {
			if value, owner, index, ok = instance.ResolveField(name); ok {
				return value, owner, index, true
			}
		}
	}
	return nil, nil, -1, false
}

// Set replaces one field using synchronization suitable for task sharing.
func (s *StructInstance) Set(name string, value Object) {
	s.mu.Lock()
	s.Values[name] = value
	s.mu.Unlock()
}
