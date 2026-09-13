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
}

// Builtin nominal structs are predeclared alongside primitive types and native
// functions. Lexical bindings take precedence during identifier lookup.
var builtinStructDefinitions map[string]*Struct

func init() {
	builtinStructDefinitions = map[string]*Struct{
		"IOError":          errorStructDefinition("IOError"),
		"FileNotFound":     errorStructDefinition("FileNotFound"),
		"PermissionDenied": errorStructDefinition("PermissionDenied"),
		"ConnectionError":  errorStructDefinition("ConnectionError"),
		"ListenError":      errorStructDefinition("ListenError"),
		"ReadError":        errorStructDefinition("ReadError"),
		"WriteError":       errorStructDefinition("WriteError"),
	}
	for _, name := range runtimeErrorStructNames {
		builtinStructDefinitions[name] = errorStructDefinition(name)
	}
	builtinStructDefinitions["File"] = &Struct{
		Name:   "File",
		Fields: []string{"path", "read", "write", "close"},
		FieldTypes: []*Contract{
			MustResolveContract(namedAnnotation("str")),
			MustResolveContract(callAnnotation(nil, nil, namedAnnotation("str"), "IOError")),
			MustResolveContract(callAnnotation([]string{"contents"}, []*ast.TypeAnnotation{namedAnnotation("str")}, nil, "IOError")),
			MustResolveContract(callAnnotation(nil, nil, nil, "IOError")),
		},
	}
	builtinStructDefinitions["IOStream"] = &Struct{
		Name:   "IOStream",
		Fields: []string{"name", "read", "read_line", "write"},
		FieldTypes: []*Contract{
			MustResolveContract(namedAnnotation("str")),
			MustResolveContract(callAnnotation(nil, nil, namedAnnotation("str"), "IOError")),
			MustResolveContract(callAnnotation(nil, nil, namedAnnotation("str"), "IOError")),
			MustResolveContract(callAnnotation([]string{"data"}, []*ast.TypeAnnotation{namedAnnotation("str")}, nil, "IOError")),
		},
	}
	builtinStructDefinitions["ReadFromResult"] = &Struct{
		Name:   "ReadFromResult",
		Fields: []string{"data", "address"},
		FieldTypes: []*Contract{
			MustResolveContract(namedAnnotation("str")),
			MustResolveContract(namedAnnotation("str")),
		},
	}
	builtinStructDefinitions["Connection"] = &Struct{
		Name:   "Connection",
		Fields: []string{"address", "read", "write", "write_to", "read_from", "close"},
		FieldTypes: []*Contract{
			MustResolveContract(namedAnnotation("str")),
			MustResolveContract(callAnnotation([]string{"bytes"}, []*ast.TypeAnnotation{namedAnnotation("int")}, namedAnnotation("str"), "ReadError")),
			MustResolveContract(callAnnotation([]string{"data"}, []*ast.TypeAnnotation{namedAnnotation("str")}, nil, "WriteError")),
			MustResolveContract(callAnnotation([]string{"data", "address"}, []*ast.TypeAnnotation{namedAnnotation("str"), namedAnnotation("str")}, nil, "WriteError")),
			MustResolveContract(callAnnotation([]string{"bytes"}, []*ast.TypeAnnotation{namedAnnotation("int")}, namedAnnotation("ReadFromResult"), "ReadError")),
			MustResolveContract(callAnnotation(nil, nil, nil, "ConnectionError")),
		},
	}
	builtinStructDefinitions["Listener"] = &Struct{
		Name:   "Listener",
		Fields: []string{"address", "accept", "close"},
		FieldTypes: []*Contract{
			MustResolveContract(namedAnnotation("str")),
			MustResolveContract(callAnnotation(nil, nil, namedAnnotation("Connection"), "ConnectionError")),
			MustResolveContract(callAnnotation(nil, nil, nil, "ConnectionError")),
		},
	}
	builtinStructDefinitions["TemplateString"] = &Struct{
		Name:       "TemplateString",
		Fields:     []string{"eval"},
		FieldTypes: []*Contract{MustResolveContract(callAnnotation(nil, nil, namedAnnotation("str")))},
	}
}

func errorStructDefinition(name string) *Struct {
	return &Struct{
		Name:       name,
		Fields:     []string{"message"},
		FieldTypes: []*Contract{MustResolveContract(namedAnnotation("str"))},
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
