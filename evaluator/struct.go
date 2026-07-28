package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalStructStatement creates a callable struct definition and binds it in
// the current environment.
func (e *Evaluator) evalStructStatement(node *ast.StructStatement, env *object.Environment) object.Object {
	fields := make([]string, 0, len(node.Fields))
	fieldTypes := make([]*ast.TypeAnnotation, 0, len(node.Fields))
	seen := make(map[string]bool, len(node.Fields))
	for _, field := range node.Fields {
		if seen[field.Value] {
			return newError("duplicate struct field %q", field.Value)
		}
		seen[field.Value] = true
		fields = append(fields, field.Value)
		fieldTypes = append(fieldTypes, field.Type)
	}

	definition := &object.Struct{
		Name:       node.Name.Value,
		Fields:     fields,
		FieldTypes: fieldTypes,
		Env:        env,
	}
	env.Set(node.Name.Value, definition)
	for _, fieldType := range fieldTypes {
		if err := e.validateTypeAnnotation(fieldType, env); err != nil {
			return err
		}
	}
	return nil
}
