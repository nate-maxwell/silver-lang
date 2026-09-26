package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalForStatement evaluates the iterable once and consumes break/continue at
// this loop boundary. Returns and errors propagate to the enclosing caller.
// Arrays bind one element; maps bind a key and value from a shallow snapshot.
func (e *Evaluator) evalForStatement(statement *ast.ForStatement, env *object.Environment) object.Object {
	iterable := e.evalValue(statement.Iterable, env)
	if isError(iterable) {
		return iterable
	}

	switch collection := iterable.(type) {
	case *object.Array:
		if statement.Value != nil {
			return newError(object.RuntimeErrorKindValue, "array for loop requires one binding")
		}
		for _, element := range collection.Elements {
			result := e.evalForIteration(statement, env, element, nil)
			switch result.(type) {
			case *object.Break:
				return NULL
			case *object.Continue:
				continue
			case *object.ReturnValue, *object.Error:
				return result
			}
		}
	case *object.Map:
		if statement.Value == nil {
			return newError(object.RuntimeErrorKindValue, "map for loop requires key and value bindings")
		}
		// Snapshot before running the body so inserts and deletes do not change
		// which entries this iteration visits or require holding the map lock.
		for _, pair := range collection.Snapshot() {
			result := e.evalForIteration(statement, env, pair.Key, pair.Value)
			switch result.(type) {
			case *object.Break:
				return NULL
			case *object.Continue:
				continue
			case *object.ReturnValue, *object.Error:
				return result
			}
		}
	default:
		return newError(object.RuntimeErrorKindType, "not iterable: %s", runtimeTypeName(iterable))
	}

	return NULL
}

// evalForIteration gives each iteration fresh lexical bindings for its loop
// variables and body declarations, including those captured by closures.
func (e *Evaluator) evalForIteration(statement *ast.ForStatement, env *object.Environment, key, value object.Object) object.Object {
	iterationEnv := object.NewEnclosedEnvironment(env)
	iterationEnv.Set(statement.Key.Value, key)
	if statement.Value != nil {
		iterationEnv.Set(statement.Value.Value, value)
	}
	return e.Eval(statement.Body, iterationEnv)
}

// evalWhileStatement rechecks the condition before every iteration and runs
// the body in the supplied environment. Unlike for, it introduces no fresh
// iteration bindings; break/continue stop here while returns/errors escape.
func (e *Evaluator) evalWhileStatement(statement *ast.WhileStatement, env *object.Environment) object.Object {
	for {
		condition := e.evalValue(statement.Condition, env)
		if isError(condition) {
			return condition
		}
		if !isTruthy(condition) {
			return NULL
		}
		result := e.Eval(statement.Body, env)
		switch result.(type) {
		case *object.Break:
			return NULL
		case *object.Continue:
			continue
		case *object.ReturnValue, *object.Error:
			return result
		}
	}
}
