package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalEnumType creates the enum namespace and one singleton value per
// member, then binds the namespace in the current environment.
func (e *Evaluator) evalEnumType(name string, node *ast.EnumTypeLiteral, env *object.Environment) object.Object {
	enum := &object.Enum{Name: name}
	members := make(map[string]*object.EnumValue, len(node.Members))
	for _, member := range node.Members {
		if _, exists := members[member.Value]; exists {
			return newError(object.RuntimeErrorKindValue, "duplicate enum member %q", member.Value)
		}
		hashID := e.nextEnumValueID.Add(1)
		members[member.Value] = &object.EnumValue{
			EnumName: name,
			Member:   member.Value,
			HashID:   hashID,
			Enum:     enum,
		}
	}

	enum.Members = members
	env.Set(name, enum)
	return nil
}
