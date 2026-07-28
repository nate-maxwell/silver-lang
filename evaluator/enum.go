package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalEnumStatement creates the enum namespace and one singleton value per
// member, then binds the namespace in the current environment.
func (e *Evaluator) evalEnumStatement(node *ast.EnumStatement, env *object.Environment) object.Object {
	enum := &object.Enum{Name: node.Name.Value}
	members := make(map[string]*object.EnumValue, len(node.Members))
	for _, member := range node.Members {
		if _, exists := members[member.Value]; exists {
			return newError("duplicate enum member %q", member.Value)
		}
		e.nextEnumValueID++
		members[member.Value] = &object.EnumValue{
			EnumName: node.Name.Value,
			Member:   member.Value,
			HashID:   e.nextEnumValueID,
			Enum:     enum,
		}
	}

	enum.Members = members
	env.Set(node.Name.Value, enum)
	return nil
}
