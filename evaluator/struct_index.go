package evaluator

import (
	"silver/ast"
	"silver/object"
)

// callStructIndexMethod resolves and invokes get_item or set_item. Silver
// functions receive the indexed instance as self; native closures may capture
// their instance and are invoked directly.
func (e *Evaluator) callStructIndexMethod(node ast.Node, instance *object.StructInstance, methodName string, args []object.Object) object.Object {
	method, exists := instance.Get(methodName)
	if !exists {
		return newError(
			object.RuntimeErrorKindAttribute,
			"indexing is not defined for struct %q: missing method %q",
			instance.Struct.Name,
			methodName,
		)
	}

	var callable object.Object
	switch method := method.(type) {
	case *object.Function:
		callable = &object.BoundMethod{Method: method, Receiver: instance, Name: methodName}
	case *object.BoundMethod:
		callable = &object.BoundMethod{Method: method.Method, Receiver: instance, Name: methodName}
	case *object.Builtin:
		callable = method
	default:
		return newError(
			object.RuntimeErrorKindType,
			"index method %q on struct %q is not callable",
			methodName,
			instance.Struct.Name,
		)
	}

	result := e.applyFunction(callable, args)
	e.prependCallerFrame(result, node)
	return result
}
