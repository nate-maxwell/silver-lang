package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalStructType creates a callable struct definition and binds it in
// the current environment.
func (e *Evaluator) evalStructType(name string, node *ast.StructTypeLiteral, env *object.Environment) object.Object {
	fields := make([]string, 0, len(node.Fields))
	fieldTypes := make([]*object.Contract, len(node.Fields))
	embeddedFields := make([]bool, 0, len(node.Fields))
	seen := make(map[string]bool, len(node.Fields))
	for _, field := range node.Fields {
		if seen[field.Value] {
			return newError(object.RuntimeErrorKindValue, "duplicate struct field %q", field.Value)
		}
		seen[field.Value] = true
		fields = append(fields, field.Value)
		embeddedFields = append(embeddedFields, field.Embedded)
	}

	definition := &object.Struct{
		Name:           name,
		PackageID:      env.PackageID(),
		Fields:         fields,
		FieldTypes:     fieldTypes,
		EmbeddedFields: embeddedFields,
	}
	// Publish the identity before resolving fields so self-referential types
	// and method signatures can name this exact definition. A resolution error
	// propagates with the partially initialized binding still in the environment.
	env.Set(name, definition)
	for index, field := range node.Fields {
		contract, err := object.ResolveContract(field.Type, env)
		if err != nil {
			return err
		}
		fieldTypes[index] = contract
		if embeddedFields[index] {
			if contract == nil {
				return newError(object.RuntimeErrorKindType, "embedded field %q must have a struct type", name+"."+fields[index])
			}
			if _, ok := contract.Definition.(*object.Struct); !ok {
				return newError(object.RuntimeErrorKindType, "embedded field %q must have a struct type", name+"."+fields[index])
			}
		}
	}
	return nil
}
