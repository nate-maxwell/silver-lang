package object

import "silver/ast"

// Builtin nominal structs are fallback definitions, like primitive types and
// builtin functions. User bindings with the same name still take precedence.
var builtinStructDefinitions map[string]*Struct

func init() {
	environment := NewEnvironment()
	builtinStructDefinitions = map[string]*Struct{
		"IOError":          errorStructDefinition("IOError", environment),
		"FileNotFound":     errorStructDefinition("FileNotFound", environment),
		"PermissionDenied": errorStructDefinition("PermissionDenied", environment),
	}
	builtinStructDefinitions["File"] = &Struct{
		Name:   "File",
		Fields: []string{"path", "read", "write", "close"},
		FieldTypes: []*ast.TypeAnnotation{
			namedAnnotation("str"),
			callAnnotation(nil, nil, namedAnnotation("str"), "IOError"),
			callAnnotation([]string{"contents"}, []*ast.TypeAnnotation{namedAnnotation("str")}, nil, "IOError"),
			callAnnotation(nil, nil, nil, "IOError"),
		},
		Env: environment,
	}
	for name, definition := range builtinStructDefinitions {
		environment.Set(name, definition)
	}
}

func errorStructDefinition(name string, environment *Environment) *Struct {
	return &Struct{
		Name:       name,
		Fields:     []string{"message"},
		FieldTypes: []*ast.TypeAnnotation{namedAnnotation("str")},
		Env:        environment,
	}
}

func namedAnnotation(name string) *ast.TypeAnnotation {
	return &ast.TypeAnnotation{Parts: []string{name}}
}

func callAnnotation(parameterNames []string, parameterTypes []*ast.TypeAnnotation, returnType *ast.TypeAnnotation, errorNames ...string) *ast.TypeAnnotation {
	errors := make([]*ast.TypeAnnotation, len(errorNames))
	for index, name := range errorNames {
		errors[index] = namedAnnotation(name)
	}
	if parameterTypes == nil {
		parameterTypes = []*ast.TypeAnnotation{}
	}
	if parameterNames == nil {
		parameterNames = []string{}
	}
	return &ast.TypeAnnotation{
		Parts:          []string{"call"},
		ParameterNames: parameterNames,
		ParameterTypes: parameterTypes,
		ReturnType:     returnType,
		ErrorTypes:     errors,
	}
}

// BuiltinStructDefinitionByName resolves File and the native I/O error types.
func BuiltinStructDefinitionByName(name string) (*Struct, bool) {
	definition, ok := builtinStructDefinitions[name]
	return definition, ok
}
