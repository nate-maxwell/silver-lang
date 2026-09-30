package object

import (
	"fmt"
	"silver/ast"
	"strings"
)

// Contract is a resolved annotation. Definition retains the exact primitive or
// nominal type value; composite contracts retain resolved children. Contracts
// are immutable after construction and never consult a lexical environment.
type Contract struct {
	Name           string // source spelling, retained for diagnostics
	Definition     Object
	ElementType    *Contract
	ParameterNames []string
	ParameterTypes []*Contract
	HasSignature   bool
	Variadic       bool
	ReturnType     *Contract
	ErrorTypes     []*Contract
}

func (c *Contract) String() string        { return c.Name }
func (c *Contract) IsCallSignature() bool { return c != nil && c.HasSignature }

// ResolveContract resolves all dependencies now, including qualified names and
// aliases. A nil annotation denotes an unannotated boundary.
func ResolveContract(annotation *ast.TypeAnnotation, env *Environment) (*Contract, *Error) {
	if annotation == nil {
		return nil, nil
	}
	contract := &Contract{Name: annotation.String()}
	if annotation.ElementType != nil {
		element, err := ResolveContract(annotation.ElementType, env)
		if err != nil {
			return nil, err
		}
		contract.ElementType = element
		return contract, nil
	}
	if annotation.IsCallSignature() {
		contract.HasSignature = true
		contract.Variadic = annotation.Variadic
		contract.ParameterNames = append([]string(nil), annotation.ParameterNames...)
		contract.ParameterTypes = make([]*Contract, len(annotation.ParameterTypes))
		for index, parameter := range annotation.ParameterTypes {
			resolved, err := ResolveContract(parameter, env)
			if err != nil {
				return nil, err
			}
			contract.ParameterTypes[index] = resolved
		}
		result, err := ResolveContract(annotation.ReturnType, env)
		if err != nil {
			return nil, err
		}
		contract.ReturnType = result
		contract.ErrorTypes = make([]*Contract, len(annotation.ErrorTypes))
		for index, failure := range annotation.ErrorTypes {
			resolved, err := ResolveErrorContract(failure, env)
			if err != nil {
				return nil, err
			}
			contract.ErrorTypes[index] = resolved
		}
		return contract, nil
	}
	if len(annotation.Parts) == 1 {
		// Primitive spellings are resolved independently of value bindings.
		// Qualified and nominal names below use the declaration environment.
		if definition, ok := TypeDefinitionByName(annotation.Parts[0]); ok {
			contract.Definition = definition
			return contract, nil
		}
	}
	value, err := resolveContractName(annotation, env)
	if err != nil {
		return nil, err
	}
	switch value := value.(type) {
	case *TypeAlias:
		// Preserve the annotation's spelling without changing the shared alias.
		*contract = *value.Contract
		contract.Name = annotation.String()
	case *Struct, *Enum, *TypeDefinition:
		contract.Definition = value
	default:
		return nil, NewError(RuntimeErrorKindType, fmt.Sprintf("%q does not name a value type", annotation.String()))
	}
	return contract, nil
}

// ResolveErrorContract requires a nominal struct for each failure alternative.
func ResolveErrorContract(annotation *ast.TypeAnnotation, env *Environment) (*Contract, *Error) {
	if annotation == nil {
		return nil, NewError(RuntimeErrorKindType, "error return type must be a struct")
	}
	contract, err := ResolveContract(annotation, env)
	if err != nil {
		return nil, err
	}
	if _, ok := contract.Definition.(*Struct); !ok {
		return nil, NewError(RuntimeErrorKindType, fmt.Sprintf("error return type %q must be a struct", contract.String()))
	}
	return contract, nil
}

// MustResolveContract constructs native contracts from trusted declarations.
// Native names resolve against built-in definitions, independent of user scopes.
func MustResolveContract(annotation *ast.TypeAnnotation) *Contract {
	contract, err := ResolveContract(annotation, nil)
	if err != nil {
		panic(err.Inspect())
	}
	return contract
}

// resolveContractName looks up the first component lexically, then falls back
// to built-in structs. Further components traverse module exports only; type
// annotations do not evaluate arbitrary member-access expressions.
func resolveContractName(annotation *ast.TypeAnnotation, env *Environment) (Object, *Error) {
	if len(annotation.Parts) == 0 {
		return nil, NewError(RuntimeErrorKindName, "empty type annotation")
	}
	var value Object
	var ok bool
	if env != nil {
		value, ok = env.Get(annotation.Parts[0])
	}
	if !ok {
		value, ok = BuiltinStructDefinitionByName(annotation.Parts[0])
	}
	if !ok {
		return nil, NewError(RuntimeErrorKindName, fmt.Sprintf("unknown type %q", annotation.String()))
	}
	for _, part := range annotation.Parts[1:] {
		module, isModule := value.(*Module)
		if !isModule {
			name, known := RuntimeTypeName(value.Type())
			if !known {
				name = strings.ToLower(string(value.Type()))
				switch value := value.(type) {
				case *StructInstance:
					name = value.Struct.Name
				case *EnumValue:
					name = value.EnumName
				}
			}
			return nil, NewError(RuntimeErrorKindName, fmt.Sprintf("cannot resolve type %q through %s", annotation.String(), name))
		}
		value, ok = module.Get(part)
		if !ok {
			return nil, NewError(RuntimeErrorKindName, fmt.Sprintf("unknown type %q", annotation.String()))
		}
	}
	return value, nil
}
