package evaluator

import (
	"fmt"
	"silver/ast"
	"silver/object"
	"strings"
)

// requireType enforces an explicit annotation. Declarations without an
// annotation retain Silver's existing inference-compatible behavior.
func (e *Evaluator) requireType(annotation *ast.TypeAnnotation, value object.Object, env *object.Environment, subject string) *object.Error {
	if annotation == nil {
		return nil
	}

	matches, resolutionError := typeMatches(annotation, value, env)
	if resolutionError != "" {
		return newError(object.RuntimeErrorKindName, "%s", resolutionError)
	}
	if matches {
		return nil
	}
	return newError(object.RuntimeErrorKindType, "type mismatch for %s: expected %s, got %s", subject, annotation.String(), runtimeTypeName(value))
}

// validateTypeAnnotation rejects unknown names at declaration time, even when
// the declared function or struct is never called.
func (e *Evaluator) validateTypeAnnotation(annotation *ast.TypeAnnotation, env *object.Environment) *object.Error {
	if annotation == nil {
		return nil
	}
	if annotation.IsCallSignature() {
		for _, parameterType := range annotation.ParameterTypes {
			if err := e.validateTypeAnnotation(parameterType, env); err != nil {
				return err
			}
		}
		if annotation.ReturnType != nil {
			if err := e.validateTypeAnnotation(annotation.ReturnType, env); err != nil {
				return err
			}
		}
		for _, errorType := range annotation.ErrorTypes {
			if err := e.validateErrorTypeAnnotation(errorType, env); err != nil {
				return err
			}
		}
		return nil
	}
	if len(annotation.Parts) == 1 {
		if _, ok := primitiveTypes[annotation.String()]; ok {
			return nil
		}
	}
	value, resolutionError := resolveNamedType(annotation, env)
	if resolutionError != "" {
		return newError(object.RuntimeErrorKindName, "%s", resolutionError)
	}
	switch value.(type) {
	case *object.Struct, *object.Enum:
		return nil
	default:
		return newError(object.RuntimeErrorKindType, "%q does not name a value type", annotation.String())
	}
}

// validateErrorTypeAnnotation enforces that every failure alternative is a
// nominal struct type. Returned instances are wrapped for unwinding only at a
// callable boundary, leaving structs ordinary values everywhere else.
func (e *Evaluator) validateErrorTypeAnnotation(annotation *ast.TypeAnnotation, env *object.Environment) *object.Error {
	if annotation == nil {
		return newError(object.RuntimeErrorKindType, "error return type must be a struct")
	}
	if err := e.validateTypeAnnotation(annotation, env); err != nil {
		return err
	}
	_, primitive := primitiveTypes[annotation.String()]
	if annotation.IsCallSignature() || len(annotation.Parts) == 1 && primitive {
		return newError(object.RuntimeErrorKindType, "error return type %q must be a struct", annotation.String())
	}
	value, resolutionError := resolveNamedType(annotation, env)
	if resolutionError != "" {
		return newError(object.RuntimeErrorKindName, "%s", resolutionError)
	}
	if _, ok := value.(*object.Struct); !ok {
		return newError(object.RuntimeErrorKindType, "error return type %q must be a struct", annotation.String())
	}
	return nil
}

// requireReturnType accepts the declared success type or any declared struct
// error type. A nil success annotation in a union denotes null.
func (e *Evaluator) requireReturnType(success *ast.TypeAnnotation, errorTypes []*ast.TypeAnnotation, value object.Object, env *object.Environment, subject string) *object.Error {
	if success == nil {
		if value == NULL {
			return nil
		}
	} else {
		matches, resolutionError := typeMatches(success, value, env)
		if resolutionError != "" {
			return newError(object.RuntimeErrorKindName, "%s", resolutionError)
		}
		if matches {
			return nil
		}
	}
	for _, errorType := range errorTypes {
		matches, resolutionError := typeMatches(errorType, value, env)
		if resolutionError != "" {
			return newError(object.RuntimeErrorKindName, "%s", resolutionError)
		}
		if matches {
			return nil
		}
	}
	return newError(object.RuntimeErrorKindType, "type mismatch for %s: expected %s, got %s", subject, returnTypesString(success, errorTypes), runtimeTypeName(value))
}

// matchesDeclaredError reports whether value is one of a callable's declared
// struct failure alternatives.
func matchesDeclaredError(errorTypes []*ast.TypeAnnotation, value object.Object, env *object.Environment) (bool, *object.Error) {
	for _, errorType := range errorTypes {
		matches, resolutionError := typeMatches(errorType, value, env)
		if resolutionError != "" {
			return false, newError(object.RuntimeErrorKindName, "%s", resolutionError)
		}
		if matches {
			return true, nil
		}
	}
	return false, nil
}

// matchesBuiltinDeclaredError uses the native definition's nominal identity,
// so a user binding that shadows names such as FileNotFound cannot turn a
// builtin error back into an ordinary return value.
func matchesBuiltinDeclaredError(errorTypes []*ast.TypeAnnotation, value object.Object) bool {
	instance, ok := value.(*object.StructInstance)
	if !ok {
		return false
	}
	for _, errorType := range errorTypes {
		if errorType == nil || len(errorType.Parts) != 1 {
			continue
		}
		definition, ok := object.BuiltinStructDefinitionByName(errorType.Parts[0])
		if ok && instance.Struct == definition {
			return true
		}
	}
	return false
}

func returnTypesString(success *ast.TypeAnnotation, errorTypes []*ast.TypeAnnotation) string {
	parts := make([]string, 0, len(errorTypes)+1)
	if success == nil {
		parts = append(parts, "null")
	} else {
		parts = append(parts, success.String())
	}
	for _, errorType := range errorTypes {
		parts = append(parts, errorType.String())
	}
	return strings.Join(parts, " | ")
}

// typeMatches handles primitive aliases directly and resolves all other names
// as nominal struct or enum types in the declaration's lexical environment.
func typeMatches(annotation *ast.TypeAnnotation, value object.Object, env *object.Environment) (bool, string) {
	if annotation.IsCallSignature() {
		switch value := value.(type) {
		case *object.Function:
			return runtimeFunctionMatches(annotation, value, env)
		case *object.BoundMethod:
			return runtimeFunctionMatches(annotation, value.Method, env)
		case *object.Builtin:
			if value.Signature == nil {
				return false, ""
			}
			return annotationAssignable(annotation, value.Signature, env, env)
		default:
			return false, ""
		}
	}

	name := annotation.String()
	if len(annotation.Parts) == 1 {
		if expected, ok := primitiveTypes[name]; ok {
			if expected == object.FUNCTION_OBJ && value != nil && value.Type() == object.BUILTINT_OBJ {
				return true, ""
			}
			return value != nil && value.Type() == expected, ""
		}
	}

	expected, err := resolveNamedType(annotation, env)
	if err != "" {
		return false, err
	}
	switch expected := expected.(type) {
	case *object.Struct:
		actual, ok := value.(*object.StructInstance)
		return ok && actual.Struct == expected, ""
	case *object.Enum:
		actual, ok := value.(*object.EnumValue)
		return ok && actual.Enum == expected, ""
	default:
		return false, fmt.Sprintf("%q does not name a value type", name)
	}
}

// runtimeFunctionMatches checks a closure against a concrete call signature.
// Untyped parameters accept every input type, while an
// omitted return annotation has Silver's guaranteed null result.
func runtimeFunctionMatches(expected *ast.TypeAnnotation, actual *object.Function, expectedEnv *object.Environment) (bool, string) {
	if len(expected.ParameterTypes) != len(actual.Parameters) {
		return false, ""
	}
	for index, expectedParameter := range expected.ParameterTypes {
		if index < len(expected.ParameterNames) && expected.ParameterNames[index] != "" && expected.ParameterNames[index] != actual.Parameters[index].Value {
			return false, ""
		}
		actualParameter := actual.Parameters[index].Type
		if actualParameter == nil {
			continue
		}
		matches, resolutionError := annotationAssignable(actualParameter, expectedParameter, actual.Env, expectedEnv)
		if resolutionError != "" || !matches {
			return matches, resolutionError
		}
	}

	return callReturnsAssignable(expected.ReturnType, expected.ErrorTypes, actual.ReturnType, actual.ErrorTypes, expectedEnv, actual.Env)
}

// annotationAssignable reports whether every value described by source is
// accepted by target. Function parameters are contravariant and returns are
// covariant; all other Silver types currently require nominal equality.
func annotationAssignable(target, source *ast.TypeAnnotation, targetEnv, sourceEnv *object.Environment) (bool, string) {
	if target == nil || source == nil {
		return false, ""
	}
	if target.IsCallSignature() {
		if !source.IsCallSignature() || len(target.ParameterTypes) != len(source.ParameterTypes) {
			return false, ""
		}
		for index := range target.ParameterTypes {
			if index < len(target.ParameterNames) && target.ParameterNames[index] != "" {
				if index >= len(source.ParameterNames) || target.ParameterNames[index] != source.ParameterNames[index] {
					return false, ""
				}
			}
			matches, resolutionError := annotationAssignable(
				source.ParameterTypes[index], target.ParameterTypes[index], sourceEnv, targetEnv,
			)
			if resolutionError != "" || !matches {
				return matches, resolutionError
			}
		}
		return callReturnsAssignable(target.ReturnType, target.ErrorTypes, source.ReturnType, source.ErrorTypes, targetEnv, sourceEnv)
	}
	if source.IsCallSignature() {
		return isPrimitiveAnnotation(target, "call"), ""
	}

	if len(target.Parts) == 1 {
		if targetType, ok := primitiveTypes[target.String()]; ok {
			if len(source.Parts) != 1 {
				return false, ""
			}
			sourceType, ok := primitiveTypes[source.String()]
			return ok && targetType == sourceType, ""
		}
	}
	if len(source.Parts) == 1 {
		if _, primitive := primitiveTypes[source.String()]; primitive {
			return false, ""
		}
	}

	targetType, resolutionError := resolveNamedType(target, targetEnv)
	if resolutionError != "" {
		return false, resolutionError
	}
	sourceType, resolutionError := resolveNamedType(source, sourceEnv)
	if resolutionError != "" {
		return false, resolutionError
	}
	return targetType == sourceType, ""
}

// callReturnAssignable treats an omitted call-signature or function return as
// null while preserving ordinary annotation assignability for explicit types.
func callReturnAssignable(target, source *ast.TypeAnnotation, targetEnv, sourceEnv *object.Environment) (bool, string) {
	if target == nil && source == nil {
		return true, ""
	}
	if target == nil {
		return isPrimitiveAnnotation(source, "null"), ""
	}
	if source == nil {
		return isPrimitiveAnnotation(target, "null"), ""
	}
	return annotationAssignable(target, source, targetEnv, sourceEnv)
}

// callReturnsAssignable is covariant across callable results. The source's
// success value must fit the target success type, and every source error must
// be included in the target's accepted error alternatives.
func callReturnsAssignable(targetSuccess *ast.TypeAnnotation, targetErrors []*ast.TypeAnnotation, sourceSuccess *ast.TypeAnnotation, sourceErrors []*ast.TypeAnnotation, targetEnv, sourceEnv *object.Environment) (bool, string) {
	matches, resolutionError := callReturnAssignable(targetSuccess, sourceSuccess, targetEnv, sourceEnv)
	if resolutionError != "" || !matches {
		return matches, resolutionError
	}
	for _, sourceError := range sourceErrors {
		accepted := false
		for _, targetError := range targetErrors {
			matches, resolutionError = annotationAssignable(targetError, sourceError, targetEnv, sourceEnv)
			if resolutionError != "" {
				return false, resolutionError
			}
			if matches {
				accepted = true
				break
			}
		}
		if !accepted {
			return false, ""
		}
	}
	return true, ""
}

func isPrimitiveAnnotation(annotation *ast.TypeAnnotation, name string) bool {
	return annotation != nil && !annotation.IsCallSignature() && len(annotation.Parts) == 1 && annotation.Parts[0] == name
}

var primitiveTypes = map[string]object.ObjectType{
	"int":    object.INTEGER_OBJ,
	"float":  object.FLOAT_OBJ,
	"bool":   object.BOOLEAN_OBJ,
	"str":    object.STRING_OBJ,
	"null":   object.NULL_OBJ,
	"array":  object.ARRAY_OBJ,
	"hash":   object.HASH_OBJ,
	"call":   object.FUNCTION_OBJ,
	"module": object.MODULE_OBJ,
}

// resolveNamedType follows module members in a qualified annotation and
// returns the declaration object represented by the final component.
func resolveNamedType(annotation *ast.TypeAnnotation, env *object.Environment) (object.Object, string) {
	if len(annotation.Parts) == 0 {
		return nil, "empty type annotation"
	}
	value, ok := env.Get(annotation.Parts[0])
	if !ok {
		value, ok = object.BuiltinStructDefinitionByName(annotation.Parts[0])
	}
	if !ok {
		return nil, fmt.Sprintf("unknown type %q", annotation.String())
	}
	for _, part := range annotation.Parts[1:] {
		module, ok := value.(*object.Module)
		if !ok {
			return nil, fmt.Sprintf("cannot resolve type %q through %s", annotation.String(), runtimeTypeName(value))
		}
		value, ok = module.Exports[part]
		if !ok {
			return nil, fmt.Sprintf("unknown type %q", annotation.String())
		}
	}
	return value, ""
}

// runtimeTypeName produces source-level names for diagnostics.
func runtimeTypeName(value object.Object) string {
	if value == nil {
		return "nothing"
	}
	switch value := value.(type) {
	case *object.StructInstance:
		return value.Struct.Name
	case *object.EnumValue:
		return value.EnumName
	}
	if name, ok := sourceTypeNames[value.Type()]; ok {
		return name
	}
	return strings.ToLower(string(value.Type()))
}

var sourceTypeNames = map[object.ObjectType]string{
	object.INTEGER_OBJ:  "int",
	object.FLOAT_OBJ:    "float",
	object.BOOLEAN_OBJ:  "bool",
	object.STRING_OBJ:   "str",
	object.NULL_OBJ:     "null",
	object.ARRAY_OBJ:    "array",
	object.HASH_OBJ:     "hash",
	object.FUNCTION_OBJ: "call",
	object.MODULE_OBJ:   "module",
	object.TYPE_OBJ:     "type",
	object.TASK_OBJ:     "task",
}
