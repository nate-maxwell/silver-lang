package evaluator

import (
	"fmt"
	"silver/object"
)

// applyFunction invokes either a Silver closure or native builtin. Silver calls
// create a lexical child environment and a named traceback context.
// Arguments are already evaluated and variadic packs expanded by the caller.
// Native builtins validate their own inputs; their declared error alternatives
// are converted here from ordinary struct values into propagating errors.
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
		if matchesDeclaredError(fn.Signature.ErrorTypes, result) {
			return &object.Error{Value: result.(*object.StructInstance)}
		}
		return result

	case *object.Struct:
		return e.applyStruct(fn, args)

	default:
		return newError(object.RuntimeErrorKindType, "not a function: %s", fn.Type())
	}
}

// applyStruct validates positional values against the declaration's captured
// field contracts before publishing an instance. It is shared by brace syntax
// and calls to struct constructors; field order follows definition.Fields.
func (e *Evaluator) applyStruct(definition *object.Struct, values []object.Object) object.Object {
	if len(values) != len(definition.Fields) {
		return newError(object.RuntimeErrorKindType, "wrong number of arguments for struct %s. got=%d, want=%d", definition.Name, len(values), len(definition.Fields))
	}
	fields := make(map[string]object.Object, len(definition.Fields))
	for index, field := range definition.Fields {
		if err := e.requireType(definition.FieldTypes[index], values[index], fmt.Sprintf("field %q", definition.Name+"."+field)); err != nil {
			return err
		}
		fields[field] = values[index]
	}
	return &object.StructInstance{Struct: definition, Values: fields}
}

// applyUserFunction owns a Silver invocation's lifetime: bind arguments, run
// the body and defers, then validate the outgoing result or failure. Runtime
// faults can always propagate; user error structs must occur in ErrorTypes.
func (e *Evaluator) applyUserFunction(fn *object.Function, args []object.Object, contextName string) object.Object {
	boundArgs, err := e.bindFunctionArguments(fn, args)
	if err != nil {
		return err
	}
	extendedEnv := extendFunctionEnv(fn, boundArgs)
	if contextName == "" {
		contextName = "<anonymous>"
	}
	e.pushContext(contextName)
	defer e.popContext()
	evaluated := e.Eval(fn.Body, extendedEnv)
	// Unwind while the function context is still active. A failing defer can
	// replace a pending return or error, and that replacement must pass through
	// the same declared-error checks as a failure from the body.
	evaluated = e.runDefers(extendedEnv, evaluated)
	if err, ok := evaluated.(*object.Error); ok {
		if err.IsRuntimeError() {
			return err
		}
		if !matchesDeclaredError(fn.ErrorTypes, err.Value) {
			return newError(
				object.RuntimeErrorKindRuntime,
				"error %s escaped %q but is not declared in its return union",
				err.Value.Struct.Name,
				contextName,
			)
		}
		return err
	}
	returned, didReturn := evaluated.(*object.ReturnValue)
	if didReturn {
		evaluated = returned.Value
	} else {
		// Evaluating a function body does not implicitly return the value of its
		// final statement. Only a return statement supplies a result.
		evaluated = NULL
	}
	// A completely omitted return declaration denotes a void function. A leading
	// pipe instead declares null success plus one or more
	// struct error alternatives, so its actual result must escape.
	if fn.ReturnType == nil && len(fn.ErrorTypes) == 0 && !fn.Operator {
		return NULL
	}
	if !fn.Operator || fn.ReturnType != nil {
		if err := e.requireReturnType(fn.ReturnType, fn.ErrorTypes, evaluated, fmt.Sprintf("return value of %q", contextName)); err != nil {
			return err
		}
	}
	if matchesDeclaredError(fn.ErrorTypes, evaluated) {
		error := &object.Error{Value: evaluated.(*object.StructInstance)}
		error.SetOrigin(e.traceFrame(fn.Body))
		return error
	}
	return evaluated
}

// bindFunctionArguments applies ordinary positional binding first. When a
// destructurable value does not satisfy the parameter at its position, its
// named fields or exports are offered to the remaining parameters. A value
// that satisfies its parameter is therefore kept intact.
// The returned slice has one entry per parameter, with a VariadicArguments
// pack in the final slot when needed. No call environment is mutated until
// every argument and required parameter has passed validation.
func (e *Evaluator) bindFunctionArguments(fn *object.Function, args []object.Object) ([]object.Object, *object.Error) {
	variadicIndex := -1
	if len(fn.Parameters) > 0 && fn.Parameters[len(fn.Parameters)-1].Variadic {
		variadicIndex = len(fn.Parameters) - 1
	}
	if variadicIndex < 0 && len(args) > len(fn.Parameters) {
		return nil, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", len(args), len(fn.Parameters))
	}

	bound := make([]object.Object, len(fn.Parameters))
	// Destructuring can fill later slots while leaving earlier ones vacant.
	// Track occupancy separately so the next argument fills the first gap.
	assigned := make([]bool, len(fn.Parameters))
	boundCount := 0
	variadicArguments := make([]object.Object, 0)

	for _, argument := range args {
		parameterIndex := nextUnassignedParameter(assigned)
		if parameterIndex == len(fn.Parameters) {
			return nil, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", boundCount+1, len(fn.Parameters))
		}

		parameter := fn.Parameters[parameterIndex]
		contract := fn.ParameterTypes[parameterIndex]
		if parameter.Variadic {
			// Leave this slot unassigned until all arguments have been visited,
			// so each remaining positional value is checked as a pack element.
			if !typeMatches(contract, argument) {
				return nil, e.requireType(contract, argument, fmt.Sprintf("parameter %q", parameter.Value))
			}
			variadicArguments = append(variadicArguments, argument)
			continue
		}
		if typeMatches(contract, argument) {
			bound[parameterIndex] = argument
			assigned[parameterIndex] = true
			boundCount++
			continue
		}

		destructurable, ok := argument.(object.Destructurable)
		if !ok {
			return nil, e.requireType(contract, argument, fmt.Sprintf("parameter %q", parameter.Value))
		}

		extracted := 0
		for index := parameterIndex; index < len(fn.Parameters); index++ {
			if assigned[index] {
				continue
			}
			candidate := fn.Parameters[index]
			if candidate.Variadic {
				continue
			}
			fieldValue, ok := destructurable.Get(candidate.Value)
			if !ok {
				continue
			}
			if err := e.requireType(fn.ParameterTypes[index], fieldValue, fmt.Sprintf("parameter %q", candidate.Value)); err != nil {
				return nil, err
			}
			bound[index] = fieldValue
			assigned[index] = true
			boundCount++
			extracted++
		}
		if extracted == 0 {
			return nil, e.requireType(contract, argument, fmt.Sprintf("parameter %q", parameter.Value))
		}
	}

	if variadicIndex >= 0 {
		bound[variadicIndex] = &object.VariadicArguments{Elements: variadicArguments}
		assigned[variadicIndex] = true
	}

	fixedParameterCount := len(fn.Parameters)
	if variadicIndex >= 0 {
		fixedParameterCount--
	}
	if boundCount != fixedParameterCount {
		if variadicIndex >= 0 {
			return nil, newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want>=%d", boundCount, fixedParameterCount)
		}
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

// extendFunctionEnv binds evaluated arguments to parameters in a child of the
// function's captured lexical environment with its own deferred-call lifetime.
// args must be the complete parameter-aligned result of bindFunctionArguments.
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewFunctionEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		contract := fn.ParameterTypes[i]
		if param.Variadic {
			// The annotation constrains each incoming argument, not the bound
			// argument pack itself.
			contract = nil
		}
		env.SetTyped(param.Value, args[i], contract)
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
