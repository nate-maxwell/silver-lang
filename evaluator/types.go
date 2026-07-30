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
		return newError("%s", resolutionError)
	}
	if matches {
		return nil
	}
	return newError("type mismatch for %s: expected %s, got %s", subject, annotation.String(), runtimeTypeName(value))
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
		if annotation.ReturnType == nil {
			return newError("call type %q is missing a return type", annotation.String())
		}
		return e.validateTypeAnnotation(annotation.ReturnType, env)
	}
	if len(annotation.Parts) == 1 {
		if _, ok := primitiveTypes[annotation.String()]; ok {
			return nil
		}
	}
	value, resolutionError := resolveNamedType(annotation, env)
	if resolutionError != "" {
		return newError("%s", resolutionError)
	}
	switch value.(type) {
	case *object.Struct, *object.Enum:
		return nil
	default:
		return newError("%q does not name a value type", annotation.String())
	}
}

// typeMatches handles primitive aliases directly and resolves all other names
// as nominal struct or enum types in the declaration's lexical environment.
func typeMatches(annotation *ast.TypeAnnotation, value object.Object, env *object.Environment) (bool, string) {
	if annotation.IsCallSignature() {
		var function *object.Function
		switch value := value.(type) {
		case *object.Function:
			function = value
		case *object.BoundMethod:
			function = value.Method
		default:
			return false, ""
		}
		return runtimeFunctionMatches(annotation, function, env)
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

// runtimeFunctionMatches checks a closure or bound method against a concrete
// function signature. Untyped parameters accept every input type, while an
// omitted return annotation has Silver's guaranteed null result.
func runtimeFunctionMatches(expected *ast.TypeAnnotation, actual *object.Function, expectedEnv *object.Environment) (bool, string) {
	if len(expected.ParameterTypes) != len(actual.Parameters) {
		return false, ""
	}
	for index, expectedParameter := range expected.ParameterTypes {
		actualParameter := actual.Parameters[index].Type
		if actualParameter == nil {
			continue
		}
		matches, resolutionError := annotationAssignable(actualParameter, expectedParameter, actual.Env, expectedEnv)
		if resolutionError != "" || !matches {
			return matches, resolutionError
		}
	}

	if actual.ReturnType == nil {
		return isPrimitiveAnnotation(expected.ReturnType, "null"), ""
	}
	return annotationAssignable(expected.ReturnType, actual.ReturnType, expectedEnv, actual.Env)
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
			matches, resolutionError := annotationAssignable(
				source.ParameterTypes[index], target.ParameterTypes[index], sourceEnv, targetEnv,
			)
			if resolutionError != "" || !matches {
				return matches, resolutionError
			}
		}
		return annotationAssignable(target.ReturnType, source.ReturnType, targetEnv, sourceEnv)
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
}
