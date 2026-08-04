package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalTryExpression handles every error by matching the nominal struct it
// carries, including the built-in Error struct used for runtime faults.
func (e *Evaluator) evalTryExpression(expression *ast.TryExpression, env *object.Environment) object.Object {
	for _, clause := range expression.Catches {
		if err := e.validateErrorTypeAnnotation(clause.ErrorType, env); err != nil {
			return err
		}
	}

	result := e.Eval(expression.Body, env)
	error, ok := result.(*object.Error)
	if !ok {
		return result
	}

	for _, clause := range expression.Catches {
		matches, resolutionError := typeMatches(clause.ErrorType, error.Value, env)
		if resolutionError != "" {
			return newError(object.RuntimeErrorKindName, "%s", resolutionError)
		}
		if !matches {
			continue
		}

		catchEnv := object.NewEnclosedEnvironment(env)
		catchEnv.SetTyped(clause.Binding.Value, error.Value, clause.ErrorType)
		return e.Eval(clause.Body, catchEnv)
	}

	return error
}
