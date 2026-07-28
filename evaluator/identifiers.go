package evaluator

import (
	"silver/ast"
	"silver/object"
)

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
	return newError("identifier not found: %s", node.Value)
}
