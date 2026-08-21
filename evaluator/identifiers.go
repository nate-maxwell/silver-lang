package evaluator

import (
	"fmt"
	"silver/ast"
	"silver/object"
)

// evalAssignment replaces the nearest existing lexical binding. Explicit type
// annotations from the original declaration remain enforced.
func (e *Evaluator) evalAssignment(node *ast.AssignmentStatement, env *object.Environment) object.Object {
	annotation, declarationEnv, ok := env.AssignmentTarget(node.Name.Value)
	if !ok {
		return newError(object.RuntimeErrorKindName, "identifier not found: %s", node.Name.Value)
	}

	value := e.Eval(node.Value, env)
	if isError(value) {
		return value
	}
	if err := e.requireType(annotation, value, declarationEnv, fmt.Sprintf("binding %q", node.Name.Value)); err != nil {
		return err
	}
	env.Assign(node.Name.Value, value)
	return NULL
}

// evalMemberAssignment mutates one field on an existing struct instance after
// enforcing the field's declared type.
func (e *Evaluator) evalMemberAssignment(node *ast.MemberAssignmentStatement, env *object.Environment) object.Object {
	target := e.Eval(node.Target.Object, env)
	if isError(target) {
		return target
	}
	instance, ok := target.(*object.StructInstance)
	if !ok {
		return newError(object.RuntimeErrorKindType, "member assignment not supported on %s", runtimeTypeName(target))
	}

	member := node.Target.Member.Value
	fieldIndex := -1
	for index, field := range instance.Struct.Fields {
		if field == member {
			fieldIndex = index
			break
		}
	}
	if fieldIndex < 0 {
		return newError(object.RuntimeErrorKindAttribute, "struct %q has no field %q", instance.Struct.Name, member)
	}

	value := e.Eval(node.Value, env)
	if isError(value) {
		return value
	}
	if err := e.requireType(instance.Struct.FieldTypes[fieldIndex], value, instance.Struct.Env, fmt.Sprintf("field %q", instance.Struct.Name+"."+member)); err != nil {
		return err
	}
	instance.Set(member, value)
	return NULL
}

// evalIndexAssignment mutates native arrays/maps directly or invokes set_item
// on structs. Native collection mutations are visible through aliases.
func (e *Evaluator) evalIndexAssignment(node *ast.IndexAssignmentStatement, env *object.Environment) object.Object {
	target := e.Eval(node.Target.Left, env)
	if isError(target) {
		return target
	}
	switch target.(type) {
	case *object.Array, *object.Map, *object.StructInstance:
	default:
		return newError(object.RuntimeErrorKindType, "index assignment not supported on %s", runtimeTypeName(target))
	}

	key := e.Eval(node.Target.Index, env)
	if isError(key) {
		return key
	}
	var hashable object.Hashable
	var arrayIndex int
	switch target := target.(type) {
	case *object.Array:
		var indexError *object.Error
		arrayIndex, indexError = requireArrayIndex(target, key)
		if indexError != nil {
			return indexError
		}
	case *object.Map:
		var ok bool
		hashable, ok = key.(object.Hashable)
		if !ok {
			return newError(object.RuntimeErrorKindType, "unusable as hash key: %s", key.Type())
		}
	}

	value := e.Eval(node.Value, env)
	if isError(value) {
		return value
	}

	switch target := target.(type) {
	case *object.Array:
		target.Elements[arrayIndex] = value
		return NULL
	case *object.Map:
		target.Set(hashable.HashKey(), object.MapPair{Key: key, Value: value})
		return NULL
	case *object.StructInstance:
		result := e.callStructIndexMethod(node, target, "set_item", []object.Object{key, value})
		if isError(result) {
			return result
		}
		return NULL
	}
	return newError(object.RuntimeErrorKindType, "index assignment not supported on %s", runtimeTypeName(target))
}

// evalMember resolves members on modules, enum namespaces, and structs.
func (e *Evaluator) evalMember(value object.Object, member string) object.Object {
	switch value := value.(type) {
	case *object.Module:
		export, ok := value.Get(member)
		if !ok {
			return newError(object.RuntimeErrorKindAttribute, "module %q has no exported member %q", value.Path, member)
		}
		return export
	case *object.Enum:
		enumValue, ok := value.Members[member]
		if !ok {
			return newError(object.RuntimeErrorKindAttribute, "enum %q has no member %q", value.Name, member)
		}
		return enumValue
	case *object.StructInstance:
		field, ok := value.Get(member)
		if !ok {
			return newError(object.RuntimeErrorKindAttribute, "struct %q has no field %q", value.Struct.Name, member)
		}
		for index, fieldName := range value.Struct.Fields {
			fieldType := value.Struct.FieldTypes[index]
			if fieldName != member || fieldType == nil || !fieldType.IsCallSignature() {
				continue
			}
			function, ok := field.(*object.Function)
			if !ok {
				return field
			}
			return &object.BoundMethod{Method: function, Receiver: value, Name: member}
		}
		return field
	default:
		return newError(object.RuntimeErrorKindType, "member access not supported on %s", value.Type())
	}
}

// evalIdentifier resolves lexical bindings and built-in type definitions.
func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if typeDefinition, ok := object.TypeDefinitionByName(node.Value); ok {
		return typeDefinition
	}
	if structDefinition, ok := object.BuiltinStructDefinitionByName(node.Value); ok {
		return structDefinition
	}
	return newError(object.RuntimeErrorKindName, "identifier not found: %s", node.Value)
}
