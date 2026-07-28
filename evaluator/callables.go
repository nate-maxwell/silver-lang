package evaluator

import (
	"fmt"
	"silver/object"
)

// applyFunction invokes either a Silver closure or native builtin. Silver calls
// create a lexical child environment and a named traceback context.
func (e *Evaluator) applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		if len(args) != len(fn.Parameters) {
			return newError("wrong number of arguments. got=%d, want=%d", len(args), len(fn.Parameters))
		}
		for i, parameter := range fn.Parameters {
			if err := e.requireType(parameter.Type, args[i], fn.Env, fmt.Sprintf("parameter %q", parameter.Value)); err != nil {
				return err
			}
		}
		extendedEnv := extendFunctionEnv(fn, args)
		name := fn.Name
		if name == "" {
			name = "<anonymous>"
		}
		e.pushContext(name)
		defer e.popContext()
		evaluated := unwrapReturnValue(e.Eval(fn.Body, extendedEnv))
		if isError(evaluated) {
			return evaluated
		}
		if err := e.requireType(fn.ReturnType, evaluated, fn.Env, fmt.Sprintf("return value of %q", name)); err != nil {
			return err
		}
		return evaluated

	case *object.Builtin:
		return fn.Fn(args...)

	case *object.Struct:
		if len(args) != len(fn.Fields) {
			return newError("wrong number of arguments for struct %s. got=%d, want=%d", fn.Name, len(args), len(fn.Fields))
		}
		values := make(map[string]object.Object, len(fn.Fields))
		for i, field := range fn.Fields {
			if err := e.requireType(fn.FieldTypes[i], args[i], fn.Env, fmt.Sprintf("field %q", fn.Name+"."+field)); err != nil {
				return err
			}
			values[field] = args[i]
		}
		return &object.StructInstance{Struct: fn, Values: values}

	default:
		return newError("not a function: %s", fn.Type())
	}
}

// extendFunctionEnv binds evaluated arguments to parameters in a child of the
// function's captured lexical environment. Arity is validated by applyFunction.
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

// unwrapReturnValue removes the evaluator's internal function-return wrapper.
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}
