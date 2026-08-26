package evaluator

import (
	"silver/ast"
	"silver/object"
)

// structOperatorVisible preserves the identity of package-local operators.
// Built-in operators are language-wide, while a custom spelling overloads a
// struct only when the expression and struct declaration belong to the same
// package. Thus another package's identically-spelled operator cannot invoke
// an imported struct's field accidentally.
func (e *Evaluator) structOperatorVisible(symbol string, instance *object.StructInstance, env *object.Environment) bool {
	packageID := env.PackageID()
	if !e.operatorScope(packageID).registry.Has(symbol) {
		return true
	}
	return instance.Struct.PackageID == packageID
}

// evalInfixExpression dispatches eligible struct values through their exact
// symbolic field and otherwise retains primitive operator behavior.
func (e *Evaluator) evalInfixExpression(node *ast.InfixExpression, left, right object.Object, env *object.Environment) object.Object {
	if instance, ok := left.(*object.StructInstance); ok && e.structOperatorVisible(node.Operator, instance, env) {
		return e.evalStructInfixExpression(node, instance, right)
	}
	return evalInfixExpression(node.Operator, left, right)
}

// evalStructInfixExpression invokes the field whose name exactly matches the
// operator spelling.
func (e *Evaluator) evalStructInfixExpression(node *ast.InfixExpression, instance *object.StructInstance, right object.Object) object.Object {
	method, exists := instance.Get(node.Operator)
	if !exists {
		return newError(
			object.RuntimeErrorKindAttribute,
			"operator %q is not defined for struct %q: missing field %q",
			node.Operator,
			instance.Struct.Name,
			node.Operator,
		)
	}

	var callable object.Object
	switch method := method.(type) {
	case *object.Function:
		callable = &object.BoundMethod{
			Method:   method,
			Receiver: instance,
			Name:     node.Operator,
		}
	case *object.BoundMethod:
		// Rebind stored bound methods to the struct currently participating in
		// the expression so copied operator fields cannot retain an old receiver.
		callable = &object.BoundMethod{
			Method:   method.Method,
			Receiver: instance,
			Name:     node.Operator,
		}
	default:
		return newError(
			object.RuntimeErrorKindType,
			"operator field %q on struct %q is not callable",
			node.Operator,
			instance.Struct.Name,
		)
	}

	result := e.applyFunction(callable, []object.Object{right})
	e.prependCallerFrame(result, node)
	return result
}
