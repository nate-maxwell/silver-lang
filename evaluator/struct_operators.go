package evaluator

import (
	"silver/ast"
	"silver/object"
)

// structInfixOperatorMethods is the complete operator-overload registry.
// Adding or replacing an entry here changes only struct dispatch; primitive
// operator behavior remains in the existing type-specific evaluators.
var structInfixOperatorMethods = map[string]string{
	// Arithmetic operators
	"+": "add",
	"-": "sub",
	"*": "mul",
	"/": "div",

	"**": "pow",
	"//": "int_div",

	// Comparison operators
	"==": "eq",
	"!=": "not_eq",

	"<":  "lt",
	">":  "gt",
	"<=": "lte",
	">=": "gte",
}

// evalInfixExpression invokes an operator method when the left operand is a
// struct instance and otherwise preserves Silver's primitive dispatch.
func (e *Evaluator) evalInfixExpression(node *ast.InfixExpression, left, right object.Object) object.Object {
	instance, ok := left.(*object.StructInstance)
	if !ok {
		return evalInfixExpression(node.Operator, left, right)
	}

	methodName, ok := structInfixOperatorMethods[node.Operator]
	if !ok {
		return newError(object.RuntimeErrorKindType, "unknown operator: %s %s %s", left.Type(), node.Operator, right.Type())
	}
	method, exists := instance.Get(methodName)
	if !exists {
		return newError(
			object.RuntimeErrorKindAttribute,
			"operator %q is not defined for struct %q: missing method %q",
			node.Operator,
			instance.Struct.Name,
			methodName,
		)
	}

	var callable object.Object
	switch method := method.(type) {
	case *object.Function:
		callable = &object.BoundMethod{
			Method:   method,
			Receiver: instance,
			Name:     methodName,
		}
	case *object.BoundMethod:
		// Rebind stored bound methods to the struct currently participating in
		// the expression so copied operator fields cannot retain an old receiver.
		callable = &object.BoundMethod{
			Method:   method.Method,
			Receiver: instance,
			Name:     methodName,
		}
	default:
		return newError(
			object.RuntimeErrorKindType,
			"operator method %q on struct %q is not callable",
			methodName,
			instance.Struct.Name,
		)
	}

	result := e.applyFunction(callable, []object.Object{right})
	e.prependCallerFrame(result, node)
	return result
}
