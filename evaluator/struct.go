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
	embeddedFields := make([]bool, 0, len(node.Fields))
	seen := make(map[string]bool, len(node.Fields))
	for _, field := range node.Fields {
		if seen[field.Value] {
			return newError(object.RuntimeErrorKindValue, "duplicate struct field %q", field.Value)
		}
		seen[field.Value] = true
		fields = append(fields, field.Value)
		fieldTypes = append(fieldTypes, field.Type)
		embeddedFields = append(embeddedFields, field.Embedded)
	}

	definition := &object.Struct{
		Name:           node.Name.Value,
		Fields:         fields,
		FieldTypes:     fieldTypes,
		EmbeddedFields: embeddedFields,
		Env:            env,
	}
	env.Set(node.Name.Value, definition)
	for index, fieldType := range fieldTypes {
		if err := e.validateTypeAnnotation(fieldType, env); err != nil {
			return err
		}
		if embeddedFields[index] {
			if len(fieldType.Parts) == 1 {
				if _, isPrimitive := object.TypeDefinitionByName(fieldType.String()); isPrimitive {
					return newError(object.RuntimeErrorKindType, "embedded field %q must have a struct type", node.Name.Value+"."+fields[index])
				}
			}
			fieldTypeValue, resolutionError := resolveNamedType(fieldType, env)
			if resolutionError != "" {
				return newError(object.RuntimeErrorKindName, "%s", resolutionError)
			}
			if _, ok := fieldTypeValue.(*object.Struct); !ok {
				return newError(object.RuntimeErrorKindType, "embedded field %q must have a struct type", node.Name.Value+"."+fields[index])
			}
		}
	}
	return nil
}
