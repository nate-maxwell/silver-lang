package evaluator

import (
	"fmt"
	"silver/ast"
	"silver/object"
)

// applyFunction invokes either a Silver closure or native builtin. Silver calls
// create a lexical child environment and a named traceback context.
func (e *Evaluator) applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		return e.applyUserFunction(fn, args, fn.Name)

	case *object.BoundMethod:
		boundArgs := make([]object.Object, 0, len(args)+1)
		boundArgs = append(boundArgs, fn.Receiver)
		boundArgs = append(boundArgs, args...)
		return e.applyUserFunction(fn.Method, boundArgs, fn.Receiver.Struct.Name+"."+fn.Name)

	case *object.Builtin:
		result := fn.Fn(args...)
		if isError(result) || fn.Signature == nil || len(fn.Signature.ErrorTypes) == 0 {
			return result
		}
		if matchesBuiltinDeclaredError(fn.Signature.ErrorTypes, result) {
			return &object.Error{Value: result.(*object.StructInstance)}
		}
		return result

	case *object.Struct:
		return e.applyStruct(fn, args)

	default:
		return newError(object.RuntimeErrorKindType, "not a function: %s", fn.Type())
	}
}

func (e *Evaluator) applyStruct(definition *object.Struct, values []object.Object) object.Object {
	if len(values) != len(definition.Fields) {
		return newError(object.RuntimeErrorKindType, "wrong number of arguments for struct %s. got=%d, want=%d", definition.Name, len(values), len(definition.Fields))
	}
	fields := make(map[string]object.Object, len(definition.Fields))
	for index, field := range definition.Fields {
		if err := e.requireType(definition.FieldTypes[index], values[index], definition.Env, fmt.Sprintf("field %q", definition.Name+"."+field)); err != nil {
			return err
		}
		fields[field] = values[index]
	}
	return &object.StructInstance{Struct: definition, Values: fields}
}

// applyUserFunction validates and invokes a closure.
func (e *Evaluator) applyUserFunction(fn *object.Function, args []object.Object, contextName string) object.Object {
	boundArgs, err := e.bindFunctionArguments(fn, args)
	if err != nil {
		return err
	}
	extendedEnv := extendFunctionEnv(fn, boundArgs)
	defer e.finishTasks(extendedEnv)
	if contextName == "" {
		contextName = "<anonymous>"
	}
	e.pushContext(contextName)
	defer e.popContext()
	evaluated := e.Eval(fn.Body, extendedEnv)
	if error, ok := evaluated.(*object.Error); ok {
		if error.IsRuntimeError() {
			return error
		}
		matches, matchError := matchesDeclaredError(fn.ErrorTypes, error.Value, fn.Env)
		if matchError != nil {
			return matchError
		}
		if !matches {
			return newError(
				object.RuntimeErrorKindRuntime,
				"error %s escaped %q but is not declared in its return union",
				error.Value.Struct.Name,
				contextName,
			)
		}
		return error
	}
	// A completely omitted return declaration denotes a void function. A leading
	// pipe instead declares null success plus one or more
	// struct error alternatives, so its actual result must escape.
	if fn.ReturnType == nil && len(fn.ErrorTypes) == 0 {
		return NULL
	}
	evaluated = unwrapReturnValue(evaluated)
	if evaluated == nil {
		evaluated = NULL
	}
	if err := e.requireReturnType(fn.ReturnType, fn.ErrorTypes, evaluated, fn.Env, fmt.Sprintf("return value of %q", contextName)); err != nil {
		return err
	}
	if matches, err := matchesDeclaredError(fn.ErrorTypes, evaluated, fn.Env); err != nil {
		return err
	} else if matches {
		error := &object.Error{Value: evaluated.(*object.StructInstance)}
		error.SetOrigin(e.traceFrame(fn.Body))
		return error
	}
	return evaluated
}

// bindFunctionArguments applies ordinary positional binding first. When a
// struct value does not satisfy the parameter at its position, its fields are
// offered to the remaining parameters by name. A matching struct parameter is
// therefore kept intact instead of being destructured.
func (e *Evaluator) bindFunctionArguments(fn *object.Function, args []object.Object) ([]object.Object, *object.Error) {
	if len(args) > len(fn.Parameters) {
		return nil, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", len(args), len(fn.Parameters))
	}

	bound := make([]object.Object, len(fn.Parameters))
	assigned := make([]bool, len(fn.Parameters))
	boundCount := 0

	for _, argument := range args {
		parameterIndex := nextUnassignedParameter(assigned)
		if parameterIndex == len(fn.Parameters) {
			return nil, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", boundCount+1, len(fn.Parameters))
		}

		parameter := fn.Parameters[parameterIndex]
		matches, resolutionError := parameterTypeMatches(parameter, argument, fn.Env)
		if resolutionError != "" {
			return nil, newError(object.RuntimeErrorKindName, "%s", resolutionError)
		}
		if matches {
			bound[parameterIndex] = argument
			assigned[parameterIndex] = true
			boundCount++
			continue
		}

		structValue, ok := argument.(*object.StructInstance)
		if !ok {
			return nil, e.parameterTypeError(parameter, argument, fn.Env)
		}

		extracted := 0
		for index := parameterIndex; index < len(fn.Parameters); index++ {
			if assigned[index] {
				continue
			}
			candidate := fn.Parameters[index]
			fieldValue, ok := structValue.Get(candidate.Value)
			if !ok {
				continue
			}
			if err := e.requireType(candidate.Type, fieldValue, fn.Env, fmt.Sprintf("parameter %q", candidate.Value)); err != nil {
				return nil, err
			}
			bound[index] = fieldValue
			assigned[index] = true
			boundCount++
			extracted++
		}
		if extracted == 0 {
			return nil, e.parameterTypeError(parameter, argument, fn.Env)
		}
	}

	if boundCount != len(fn.Parameters) {
		return nil, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", boundCount, len(fn.Parameters))
	}
	return bound, nil
}

func nextUnassignedParameter(assigned []bool) int {
	for index, isAssigned := range assigned {
		if !isAssigned {
			return index
		}
	}
	return len(assigned)
}

func parameterTypeMatches(parameter *ast.Identifier, argument object.Object, env *object.Environment) (bool, string) {
	if parameter.Type == nil {
		return true, ""
	}
	return typeMatches(parameter.Type, argument, env)
}

func (e *Evaluator) parameterTypeError(parameter *ast.Identifier, argument object.Object, env *object.Environment) *object.Error {
	return e.requireType(parameter.Type, argument, env, fmt.Sprintf("parameter %q", parameter.Value))
}

// extendFunctionEnv binds evaluated arguments to parameters in a child of the
// function's captured lexical environment. Arity is validated by applyFunction.
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.SetTyped(param.Value, args[i], param.Type)
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
