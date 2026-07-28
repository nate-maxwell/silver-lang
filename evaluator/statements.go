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

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}
