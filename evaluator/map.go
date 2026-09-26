package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalMapIndexExpression normalizes a hashable key and returns KeyError when
// the key is absent. The map get builtin remains the nullable lookup API.
func evalMapIndexExpression(mapping, index object.Object) object.Object {
	mapObject := mapping.(*object.Map)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError(object.RuntimeErrorKindType, "unusable as hash key: %s", index.Type())
	}

	pair, ok := mapObject.Get(key.HashKey())
	if !ok {
		return newError(object.RuntimeErrorKindKey, "key not found: %s", index.Inspect())
	}
	return pair.Value
}

// evalMapLiteral evaluates key/value pairs and rejects keys that do not
// implement object.Hashable or collide after normalization. Unlike the parser's
// literal-only check, this catches equal computed keys. Pair traversal follows
// the AST's unordered map; within a pair the key is checked before its value.
func (e *Evaluator) evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.MapPair)

	for keyNode, valueNode := range node.Pairs {
		key := e.evalValue(keyNode, env)
		if isError(key) {
			return key
		}

		hashable, ok := key.(object.Hashable)
		if !ok {
			return newError(object.RuntimeErrorKindType, "unusable as hash key: %s", key.Type())
		}
		hash := hashable.HashKey()
		if _, exists := pairs[hash]; exists {
			return newError(object.RuntimeErrorKindValue, "duplicate map key %s", key.Inspect())
		}

		value := e.evalValue(valueNode, env)
		if isError(value) {
			return value
		}

		pairs[hash] = object.MapPair{Key: key, Value: value}
	}
	return &object.Map{Pairs: pairs}
}
