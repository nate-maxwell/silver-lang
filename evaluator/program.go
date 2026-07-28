package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalProgram evaluates top-level statements in order, stopping immediately on
// a return or error object.
func (e *Evaluator) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = e.Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}
