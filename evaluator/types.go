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

var primitiveTypes = map[string]object.ObjectType{
	"int":      object.INTEGER_OBJ,
	"float":    object.FLOAT_OBJ,
	"bool":     object.BOOLEAN_OBJ,
	"str":      object.STRING_OBJ,
	"null":     object.NULL_OBJ,
	"array":    object.ARRAY_OBJ,
	"hash":     object.HASH_OBJ,
	"function": object.FUNCTION_OBJ,
	"module":   object.MODULE_OBJ,
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
	object.FUNCTION_OBJ: "function",
	object.MODULE_OBJ:   "module",
}
