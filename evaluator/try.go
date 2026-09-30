package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalTryExpression handles every error by matching the nominal struct it
// carries, including the built-in Error struct used for runtime faults.
func (e *Evaluator) evalTryExpression(expression *ast.TryExpression, env *object.Environment) object.Object {
	// Resolve catches before executing the body so its assignments cannot
	// change which nominal types the handlers recognize.
	contracts := make([]*object.Contract, len(expression.Catches))
	for index, clause := range expression.Catches {
		contract, err := object.ResolveErrorContract(clause.ErrorType, env)
		if err != nil {
			return err
		}
		contracts[index] = contract
	}

	result := e.Eval(expression.Body, env)
	error, ok := result.(*object.Error)
	if !ok {
		return result
	}

	for index, clause := range expression.Catches {
		if !typeMatches(contracts[index], error.Value) {
			continue
		}

		// Bind the carried struct, not the control-flow wrapper, in a child
		// scope. Handler defers still belong to the surrounding invocation.
		catchEnv := object.NewEnclosedEnvironment(env)
		catchEnv.SetTyped(clause.Binding.Value, error.Value, contracts[index])
		return e.Eval(clause.Body, catchEnv)
	}

	return error
}
