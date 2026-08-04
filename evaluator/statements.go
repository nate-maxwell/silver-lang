package evaluator

import (
	"silver/ast"
	"silver/object"
)

// evalBlockStatement evaluates a block until completion or until return/error
// control flow must propagate to an enclosing evaluator.
func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = e.Eval(statement, env)

		switch result.(type) {
		case *object.ReturnValue, *object.Error:
			return result
		}
	}

	return result
}
