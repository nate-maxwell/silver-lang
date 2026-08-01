package evaluator

import (
	"silver/ast"
	"silver/object"
)

func (e *Evaluator) evalForStatement(statement *ast.ForStatement, env *object.Environment) object.Object {
	iterable := e.Eval(statement.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	switch collection := iterable.(type) {
	case *object.Array:
		if statement.Value != nil {
			return newError("array for loop requires one binding")
		}
		for _, element := range collection.Elements {
			env.Set(statement.Key.Value, element)
			if result := e.Eval(statement.Body, env); loopMustStop(result) {
				return result
			}
		}
	case *object.Hash:
		if statement.Value == nil {
			return newError("map for loop requires key and value bindings")
		}
		for _, pair := range collection.Pairs {
			env.Set(statement.Key.Value, pair.Key)
			env.Set(statement.Value.Value, pair.Value)
			if result := e.Eval(statement.Body, env); loopMustStop(result) {
				return result
			}
		}
	default:
		return newError("not iterable: %s", runtimeTypeName(iterable))
	}

	return NULL
}

func (e *Evaluator) evalWhileStatement(statement *ast.WhileStatement, env *object.Environment) object.Object {
	for {
		condition := e.Eval(statement.Condition, env)
		if isError(condition) {
			return condition
		}
		if !isTruthy(condition) {
			return NULL
		}
		if result := e.Eval(statement.Body, env); loopMustStop(result) {
			return result
		}
	}
}

func loopMustStop(result object.Object) bool {
	if result == nil {
		return false
	}
	return result.Type() == object.RETURN_VALUE_OBJ || result.Type() == object.ERROR_OBJ
}
