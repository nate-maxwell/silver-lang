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
		Methods:    make(map[string]*object.Function),
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

// attachStructMethod registers a receiver function under its let-binding name.
func (e *Evaluator) attachStructMethod(function *object.Function) *object.Error {
	receiver, err := resolveMethodReceiver(function.ReceiverType, function.Env)
	if err != nil {
		return err
	}
	for _, field := range receiver.Fields {
		if field == function.Name {
			return newError("method %q conflicts with field %q on struct %q", function.Name, field, receiver.Name)
		}
	}
	if _, exists := receiver.Methods[function.Name]; exists {
		return newError("duplicate method %q on struct %q", function.Name, receiver.Name)
	}
	receiver.Methods[function.Name] = function
	return nil
}

// resolveMethodReceiver resolves and validates the nominal type inside fn[T].
func resolveMethodReceiver(annotation *ast.TypeAnnotation, env *object.Environment) (*object.Struct, *object.Error) {
	if annotation == nil {
		return nil, newError("method receiver type is missing")
	}
	if len(annotation.Parts) == 1 {
		if _, primitive := primitiveTypes[annotation.Parts[0]]; primitive {
			return nil, newError("method receiver %q must name a struct type", annotation.String())
		}
	}
	value, resolutionError := resolveNamedType(annotation, env)
	if resolutionError != "" {
		return nil, newError("%s", resolutionError)
	}
	receiver, ok := value.(*object.Struct)
	if !ok {
		return nil, newError("method receiver %q must name a struct type", annotation.String())
	}
	return receiver, nil
}
