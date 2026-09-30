// The DefaultMap implementation of the map data structure found in the
// collections standard library module.

package stdlib

import (
	"silver/ast"
	"silver/lexer"
	"silver/object"
	"silver/parser"
)

// newDefaultMapStructDefinition creates the nominal layout shared by instances.
// get_item is installed as a Silver closure so it can invoke the user's factory
// through the evaluator without introducing an stdlib-to-evaluator dependency.
func newDefaultMapStructDefinition() *object.Struct {
	definition := &object.Struct{
		Name: "DefaultMap",
		Fields: []string{
			"values",
			"factory",
		},
		FieldTypes: []*object.Contract{
			object.MustResolveContract(namedType("map")),
			object.MustResolveContract(namedType("call")),
		},
	}
	return definition
}

// newDefaultMap accepts a Silver factory to be called with no arguments on a
// missing key. get_item retains successful results in the exposed values map;
// factory arity is checked when that call actually runs.
func newDefaultMap(definition *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}

		returnType, errorTypes, factoryEnvironment, ok := defaultFactoryResult(args[0])
		if !ok {
			return newError(object.RuntimeErrorKindType, "default factory must be a Silver function, got %s", args[0].Type())
		}

		mapping := &object.Map{Pairs: make(map[object.HashKey]object.MapPair)}

		values := map[string]object.Object{
			"values":  mapping,
			"factory": args[0],
		}
		instance := &object.StructInstance{Struct: definition, Values: values}
		// Reuse the parsed body but capture this factory's resolved result and
		// error contracts. The template's textual int return type is a parser
		// placeholder, not the contract imposed on the user's factory result.
		getter := &object.Function{
			Name:           "get_item",
			Parameters:     defaultMapGetTemplate.Parameters,
			ParameterTypes: make([]*object.Contract, len(defaultMapGetTemplate.Parameters)),
			ReturnType:     returnType,
			ErrorTypes:     errorTypes,
			Body:           defaultMapGetTemplate.Body,
			Env:            factoryEnvironment,
		}
		values["get_item"] = &object.BoundMethod{Method: getter, Receiver: instance, Name: "get_item"}
		values["set_item"] = &object.Builtin{Fn: func(setArgs ...object.Object) object.Object {
			if err := requireArgumentCount(setArgs, 2); err != nil {
				return err
			}
			key, keyError := requireHashKey(setArgs[0])
			if keyError != nil {
				return keyError
			}
			mapping.Set(key, object.MapPair{Key: setArgs[0], Value: setArgs[1]})
			return null
		}}
		return instance
	}
}

func defaultFactoryResult(factory object.Object) (*object.Contract, []*object.Contract, *object.Environment, bool) {
	switch factory := factory.(type) {
	case *object.Function:
		return factory.ReturnType, factory.ErrorTypes, factory.Env, true
	case *object.BoundMethod:
		return factory.Method.ReturnType, factory.Method.ErrorTypes, factory.Method.Env, true
	default:
		return nil, nil, nil, false
	}
}

var defaultMapGetTemplate = parseDefaultMapGetTemplate()

// parseDefaultMapGetTemplate parses trusted implementation source once during
// Go initialization. The body checks membership, not a nullable lookup, so a
// stored null does not trigger the factory again. A factory failure propagates
// before the assignment, leaving the key absent for a later retry.
func parseDefaultMapGetTemplate() *ast.FunctionLiteral {
	const source = `fn(self, key) int {
	let maps = import("core:maps")
	if maps.contains(self.values, key) {
		return self.values[key]
	}
	let value = self.factory()
	self.values[key] = value
	return value
}`
	p := parser.New(lexer.NewWithSource(source, "<stdlib collections.DefaultMap.get_item>"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 || len(program.Statements) != 1 {
		panic("stdlib: invalid DefaultMap.get_item implementation")
	}
	statement, ok := program.Statements[0].(*ast.ExpressionStatement)
	if !ok {
		panic("stdlib: DefaultMap.get_item is not an expression")
	}
	function, ok := statement.Expression.(*ast.FunctionLiteral)
	if !ok {
		panic("stdlib: DefaultMap.get_item is not a function")
	}
	return function
}
