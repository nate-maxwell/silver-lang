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
		case *object.ReturnValue, *object.Break, *object.Continue, *object.Error:
			return result
		}
	}

	return result
}

// runDefers invokes a scope's captured calls in last-in-first-out order. All
// deferred calls run even if one fails; as with a panic from a Go defer, a
// later failure while unwinding replaces the result seen so far.
func (e *Evaluator) runDefers(env *object.Environment, result object.Object) object.Object {
	deferred := env.TakeDefers()
	var deferredFailure *object.Error
	for index := len(deferred) - 1; index >= 0; index-- {
		call := deferred[index]
		callResult := e.applyFunction(call.Function, call.Arguments)
		e.prependCallerFrame(callResult, call.Call)
		if failure, ok := callResult.(*object.Error); ok {
			failure.SetOrigin(e.traceFrame(call.Call))
			deferredFailure = failure
		}
	}
	if deferredFailure != nil {
		return deferredFailure
	}
	return result
}
