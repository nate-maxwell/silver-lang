package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalHashIndexExpression normalizes a hashable key and returns KeyError when
// the key is absent. The map get builtin remains the nullable lookup API.
func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError(object.RuntimeErrorKindType, "unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Get(key.HashKey())
	if !ok {
		return newError(object.RuntimeErrorKindKey, "key not found: %s", index.Inspect())
	}
	return pair.Value
}

// evalMapLiteral evaluates key/value pairs and rejects keys that do not
// implement object.Hashable.
func (e *Evaluator) evalMapLiteral(node *ast.MapLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := e.Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError(object.RuntimeErrorKindType, "unusable as hash key: %s", key.Type())
		}

		value := e.Eval(valueNode, env)
		if isError(value) {
			return value
		}

		pairs[hashKey.HashKey()] = object.HashPair{Key: key, Value: value}
	}
	return &object.Hash{Pairs: pairs}
}
