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
		"ConnectionError":  errorStructDefinition("ConnectionError", environment),
		"ListenError":      errorStructDefinition("ListenError", environment),
		"ReadError":        errorStructDefinition("ReadError", environment),
		"WriteError":       errorStructDefinition("WriteError", environment),
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
	builtinStructDefinitions["IOStream"] = &Struct{
		Name:   "IOStream",
		Fields: []string{"name", "read", "write"},
		FieldTypes: []*ast.TypeAnnotation{
			namedAnnotation("str"),
			callAnnotation(nil, nil, namedAnnotation("str"), "IOError"),
			callAnnotation([]string{"data"}, []*ast.TypeAnnotation{namedAnnotation("str")}, nil, "IOError"),
		},
		Env: environment,
	}
	builtinStructDefinitions["ReadFromResult"] = &Struct{
		Name:   "ReadFromResult",
		Fields: []string{"data", "address"},
		FieldTypes: []*ast.TypeAnnotation{
			namedAnnotation("str"),
			namedAnnotation("str"),
		},
		Env: environment,
	}
	builtinStructDefinitions["Connection"] = &Struct{
		Name:   "Connection",
		Fields: []string{"address", "read", "write", "write_to", "read_from", "close"},
		FieldTypes: []*ast.TypeAnnotation{
			namedAnnotation("str"),
			callAnnotation([]string{"bytes"}, []*ast.TypeAnnotation{namedAnnotation("int")}, namedAnnotation("str"), "ReadError"),
			callAnnotation([]string{"data"}, []*ast.TypeAnnotation{namedAnnotation("str")}, nil, "WriteError"),
			callAnnotation([]string{"data", "address"}, []*ast.TypeAnnotation{namedAnnotation("str"), namedAnnotation("str")}, nil, "WriteError"),
			callAnnotation([]string{"bytes"}, []*ast.TypeAnnotation{namedAnnotation("int")}, namedAnnotation("ReadFromResult"), "ReadError"),
			callAnnotation(nil, nil, nil, "ConnectionError"),
		},
		Env: environment,
	}
	builtinStructDefinitions["Listener"] = &Struct{
		Name:   "Listener",
		Fields: []string{"address", "accept", "close"},
		FieldTypes: []*ast.TypeAnnotation{
			namedAnnotation("str"),
			callAnnotation(nil, nil, namedAnnotation("Connection"), "ConnectionError"),
			callAnnotation(nil, nil, nil, "ConnectionError"),
		},
		Env: environment,
	}
	builtinStructDefinitions["TemplateString"] = &Struct{
		Name:       "TemplateString",
		Fields:     []string{"eval"},
		FieldTypes: []*ast.TypeAnnotation{callAnnotation(nil, nil, namedAnnotation("str"))},
		Env:        environment,
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

// BuiltinStructDefinitionByName resolves native standard-library structs and
// built-in error structs.
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
