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
		return newError("identifier not found: %s", node.Name.Value)
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
		return newError("member assignment not supported on %s", runtimeTypeName(target))
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
		return newError("struct %q has no field %q", instance.Struct.Name, member)
	}

	value := e.Eval(node.Value, env)
	if isError(value) {
		return value
	}
	if err := e.requireType(instance.Struct.FieldTypes[fieldIndex], value, instance.Struct.Env, fmt.Sprintf("field %q", instance.Struct.Name+"."+member)); err != nil {
		return err
	}
	instance.Values[member] = value
	return NULL
}

// evalMember resolves members on modules, enum namespaces, and struct values.
func evalMember(value object.Object, member string) object.Object {
	switch value := value.(type) {
	case *object.Module:
		export, ok := value.Exports[member]
		if !ok {
			return newError("module %q has no member %q", value.Path, member)
		}
		return export
	case *object.Enum:
		enumValue, ok := value.Members[member]
		if !ok {
			return newError("enum %q has no member %q", value.Name, member)
		}
		return enumValue
	case *object.StructInstance:
		field, ok := value.Values[member]
		if !ok {
			return newError("struct %q has no field %q", value.Struct.Name, member)
		}
		for index, fieldName := range value.Struct.Fields {
			if fieldName != member || !value.Struct.FieldTypes[index].IsCallSignature() {
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
		return newError("member access not supported on %s", value.Type())
	}
}

// evalIdentifier resolves lexical bindings before falling back to the native
// builtin registry.
func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := e.builtins.Lookup(node.Value); ok {
		return builtin
	}
	if typeDefinition, ok := object.TypeDefinitionByName(node.Value); ok {
		return typeDefinition
	}
	return newError("identifier not found: %s", node.Value)
}
