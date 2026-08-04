package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalProgram evaluates top-level statements in order, stopping immediately on
// a return or error object.
func (e *Evaluator) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	defer e.finishTasks(env)
	var result object.Object

	for _, statement := range program.Statements {
		result = e.Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			deferredResult := e.runDefers(env, result)
			if returned, ok := deferredResult.(*object.ReturnValue); ok {
				return returned.Value
			}
			return deferredResult
		case *object.Error:
			return e.runDefers(env, result)
		}
	}

	return e.runDefers(env, result)
}
