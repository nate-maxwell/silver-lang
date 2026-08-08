package object

import "silver/ast"

// RuntimeErrorKind identifies one built-in runtime error struct. The table is
// the complete registry: adding an entry makes the nominal struct available to
// Silver code and allows evaluators and native builtins to construct it.
type RuntimeErrorKind string

const (
	RuntimeErrorKindRuntime      RuntimeErrorKind = "runtime"
	RuntimeErrorKindAssertion    RuntimeErrorKind = "assertion"
	RuntimeErrorKindType         RuntimeErrorKind = "type"
	RuntimeErrorKindValue        RuntimeErrorKind = "value"
	RuntimeErrorKindZeroDivision RuntimeErrorKind = "zero_division"
	RuntimeErrorKindName         RuntimeErrorKind = "name"
	RuntimeErrorKindAttribute    RuntimeErrorKind = "attribute"
	RuntimeErrorKindImport       RuntimeErrorKind = "import"
	RuntimeErrorKindSyntax       RuntimeErrorKind = "syntax"
	RuntimeErrorKindKey          RuntimeErrorKind = "key"
	RuntimeErrorKindIndex        RuntimeErrorKind = "index"
	RuntimeErrorKindTask         RuntimeErrorKind = "task"
)

var runtimeErrorStructNames = map[RuntimeErrorKind]string{
	RuntimeErrorKindRuntime:      "RuntimeError",
	RuntimeErrorKindAssertion:    "AssertionError",
	RuntimeErrorKindType:         "TypeError",
	RuntimeErrorKindValue:        "ValueError",
	RuntimeErrorKindZeroDivision: "ZeroDivisionError",
	RuntimeErrorKindName:         "NameError",
	RuntimeErrorKindAttribute:    "AttributeError",
	RuntimeErrorKindImport:       "ImportError",
	RuntimeErrorKindSyntax:       "SyntaxError",
	RuntimeErrorKindKey:          "KeyError",
	RuntimeErrorKindIndex:        "IndexError",
	RuntimeErrorKindTask:         "TaskError",
}

// Builtin nominal structs are predeclared alongside primitive types and native
// functions. Lexical bindings take precedence during identifier lookup.
var builtinStructDefinitions map[string]*Struct

func init() {
	environment := NewEnvironment()
	builtinStructDefinitions = map[string]*Struct{
		"IOError":          errorStructDefinition("IOError", environment),
		"FileNotFound":     errorStructDefinition("FileNotFound", environment),
		"PermissionDenied": errorStructDefinition("PermissionDenied", environment),
	}
	for _, name := range runtimeErrorStructNames {
		builtinStructDefinitions[name] = errorStructDefinition(name, environment)
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

// BuiltinStructDefinitionByName resolves File and built-in error structs.
func BuiltinStructDefinitionByName(name string) (*Struct, bool) {
	definition, ok := builtinStructDefinitions[name]
	return definition, ok
}

func runtimeErrorStructName(kind RuntimeErrorKind) (string, bool) {
	name, ok := runtimeErrorStructNames[kind]
	return name, ok
}

func isRuntimeErrorStruct(definition *Struct) bool {
	for _, name := range runtimeErrorStructNames {
		candidate, ok := builtinStructDefinitions[name]
		if ok && definition == candidate {
			return true
		}
	}
	return false
}
