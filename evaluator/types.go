package evaluator

import (
	"silver/ast"
	"silver/object"
	"strings"
)

// evalTypeStatement binds a first-class type value. Structs and enums create
// nominal definitions; aliases capture an existing resolved contract without
// introducing a new nominal identity.
func (e *Evaluator) evalTypeStatement(node *ast.TypeStatement, env *object.Environment) object.Object {
	switch value := node.Value.(type) {
	case *ast.StructTypeLiteral:
		return e.evalStructType(node.Name.Value, value, env)
	case *ast.EnumTypeLiteral:
		return e.evalEnumType(node.Name.Value, value, env)
	case *ast.TypeAnnotation:
		alias := e.evalTypeAlias(value, env)
		if isError(alias) {
			return alias
		}
		env.Set(node.Name.Value, alias)
		return nil
	default:
		return newError(object.RuntimeErrorKindType, "invalid type declaration %q", node.Name.Value)
	}
}

func (e *Evaluator) evalTypeAlias(annotation *ast.TypeAnnotation, env *object.Environment) object.Object {
	contract, err := object.ResolveContract(annotation, env)
	if err != nil {
		return err
	}
	return &object.TypeAlias{Contract: contract}
}

// inferredBindingContract captures the initializer's runtime type without
// inspecting collection contents or callable signatures. Nominal values retain
// their exact definition, independently of later changes to its lexical name.
func inferredBindingContract(value object.Object) *object.Contract {
	contract := &object.Contract{Name: runtimeTypeName(value)}
	switch value := value.(type) {
	case *object.StructInstance:
		contract.Definition = value.Struct
	case *object.EnumValue:
		contract.Definition = value.Enum
	default:
		if definition, ok := object.TypeDefinitionByName(contract.Name); ok {
			contract.Definition = definition
		} else {
			// Type definitions and other first-class values may have runtime
			// categories without a corresponding source-level annotation.
			contract.Definition = &object.TypeDefinition{Name: contract.Name, RuntimeType: value.Type()}
		}
	}
	return contract
}

// requireType checks the contract captured by a declaration. No names are
// resolved here, including when a parameter or captured binding is reassigned.
func (e *Evaluator) requireType(contract *object.Contract, value object.Object, subject string) *object.Error {
	if typeMatches(contract, value) {
		return nil
	}
	return newError(object.RuntimeErrorKindType, "type mismatch for %s: expected %s, got %s", subject, contract.String(), runtimeTypeName(value))
}

// requireReturnType accepts the success type or a declared struct failure.
// A nil success contract denotes null.
func (e *Evaluator) requireReturnType(success *object.Contract, errorTypes []*object.Contract, value object.Object, subject string) *object.Error {
	if success == nil && value == NULL || success != nil && typeMatches(success, value) {
		return nil
	}
	if matchesDeclaredError(errorTypes, value) {
		return nil
	}
	return newError(object.RuntimeErrorKindType, "type mismatch for %s: expected %s, got %s", subject, returnTypesString(success, errorTypes), runtimeTypeName(value))
}

func matchesDeclaredError(errorTypes []*object.Contract, value object.Object) bool {
	for _, errorType := range errorTypes {
		if typeMatches(errorType, value) {
			return true
		}
	}
	return false
}

func returnTypesString(success *object.Contract, errorTypes []*object.Contract) string {
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

// typeMatches checks resolved primitive, nominal, array, and callable contracts.
// A nil contract is an unannotated boundary. Typed arrays are checked against
// their current elements; matching does not attach a contract to the array or
// constrain later mutation through other aliases.
func typeMatches(contract *object.Contract, value object.Object) bool {
	if contract == nil {
		return true
	}
	if contract.ElementType != nil {
		array, ok := value.(*object.Array)
		if !ok {
			return false
		}
		for _, element := range array.Elements {
			if !typeMatches(contract.ElementType, element) {
				return false
			}
		}
		return true
	}
	if contract.IsCallSignature() {
		switch value := value.(type) {
		case *object.Function:
			return runtimeFunctionMatches(contract, value)
		case *object.BoundMethod:
			return runtimeFunctionMatches(contract, value.Method)
		case *object.Builtin:
			return contractAssignable(contract, value.Signature)
		default:
			return false
		}
	}
	if value == nil {
		return false
	}
	switch expected := contract.Definition.(type) {
	case *object.TypeDefinition:
		if expected.Name == "any" {
			return true
		}
		return value.Type() == expected.RuntimeType || expected.RuntimeType == object.FUNCTION_OBJ && value.Type() == object.BUILTIN_OBJ
	case *object.Struct:
		actual, ok := value.(*object.StructInstance)
		return ok && actual.Struct == expected
	case *object.Enum:
		actual, ok := value.(*object.EnumValue)
		return ok && actual.Enum == expected
	}
	return false
}

// runtimeFunctionMatches uses the function's resolved signature, independently
// of its body environment. Untyped parameters accept every input type.
func runtimeFunctionMatches(expected *object.Contract, actual *object.Function) bool {
	if len(expected.ParameterTypes) != len(actual.Parameters) {
		return false
	}
	actualVariadic := len(actual.Parameters) > 0 && actual.Parameters[len(actual.Parameters)-1].Variadic
	if expected.Variadic != actualVariadic {
		return false
	}
	for index, expectedParameter := range expected.ParameterTypes {
		if index < len(expected.ParameterNames) && expected.ParameterNames[index] != "" && expected.ParameterNames[index] != actual.Parameters[index].Value {
			return false
		}
		actualParameter := actual.ParameterTypes[index]
		if actualParameter == nil {
			continue
		}
		// The supplied function must accept everything the expected signature
		// permits callers to pass, so parameter assignability is reversed.
		if !contractAssignable(actualParameter, expectedParameter) {
			return false
		}
	}
	return callReturnsAssignable(expected.ReturnType, expected.ErrorTypes, actual.ReturnType, actual.ErrorTypes)
}

// contractAssignable reports whether every source value is accepted by target.
// Function parameters are contravariant and returns are covariant; nominal
// contracts compare definition identities.
func contractAssignable(target, source *object.Contract) bool {
	if target == nil || source == nil {
		return false
	}
	if isPrimitiveContract(target, "any") {
		return true
	}
	if isPrimitiveContract(source, "any") {
		return false
	}
	if target.ElementType != nil {
		return source.ElementType != nil && contractAssignable(target.ElementType, source.ElementType)
	}
	if source.ElementType != nil {
		return isPrimitiveContract(target, "array")
	}
	if target.IsCallSignature() {
		if !source.IsCallSignature() || target.Variadic != source.Variadic || len(target.ParameterTypes) != len(source.ParameterTypes) {
			return false
		}
		for index := range target.ParameterTypes {
			if index < len(target.ParameterNames) && target.ParameterNames[index] != "" {
				if index >= len(source.ParameterNames) || target.ParameterNames[index] != source.ParameterNames[index] {
					return false
				}
			}
			if source.ParameterTypes[index] == nil {
				continue
			}
			if !contractAssignable(source.ParameterTypes[index], target.ParameterTypes[index]) {
				return false
			}
		}
		return callReturnsAssignable(target.ReturnType, target.ErrorTypes, source.ReturnType, source.ErrorTypes)
	}
	if source.IsCallSignature() {
		return isPrimitiveContract(target, "call")
	}
	return target.Definition == source.Definition
}

// callReturnAssignable interprets omitted callable results as null, rather
// than the unannotated boundary that nil represents for a parameter.
func callReturnAssignable(target, source *object.Contract) bool {
	if target == nil && source == nil {
		return true
	}
	if target == nil {
		return isPrimitiveContract(source, "null")
	}
	if source == nil {
		return isPrimitiveContract(target, "null")
	}
	return contractAssignable(target, source)
}

// callReturnsAssignable requires a compatible success result and ensures every
// failure the supplied callable can declare is allowed by the target. The
// supplied callable may declare fewer failures; alternative order is irrelevant.
func callReturnsAssignable(targetSuccess *object.Contract, targetErrors []*object.Contract, sourceSuccess *object.Contract, sourceErrors []*object.Contract) bool {
	if !callReturnAssignable(targetSuccess, sourceSuccess) {
		return false
	}
	for _, sourceError := range sourceErrors {
		accepted := false
		for _, targetError := range targetErrors {
			if contractAssignable(targetError, sourceError) {
				accepted = true
				break
			}
		}
		if !accepted {
			return false
		}
	}
	return true
}

func isPrimitiveContract(contract *object.Contract, name string) bool {
	if contract == nil {
		return false
	}
	definition, ok := contract.Definition.(*object.TypeDefinition)
	return ok && definition.Name == name
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
	if name, ok := object.RuntimeTypeName(value.Type()); ok {
		return name
	}
	return strings.ToLower(string(value.Type()))
}
