package stdlib

import (
	"silver/ast"
	"silver/lexer"
	"silver/object"
	"silver/parser"
)

// DefaultMap is a standard-library-owned map-like struct. Its get_item method
// is a Silver closure so it can invoke a user-supplied Silver factory.
func newDefaultMapStructDefinition() *object.Struct {
	environment := object.NewEnvironment()
	definition := &object.Struct{
		Name: "DefaultMap",
		Fields: []string{
			"values",
			"factory",
		},
		FieldTypes: []*ast.TypeAnnotation{
			namedType("map"),
			namedType("call"),
		},
		Env: environment,
	}
	environment.Set("DefaultMap", definition)
	return definition
}

// newDefaultMap accepts a zero-argument Silver factory and an optional
// initial map. Missing values are created by DefaultMap.get and retained in
// the ordinary map exposed by DefaultMap.values.
func newDefaultMap(definition *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if len(args) < 1 || len(args) > 2 {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=1 or 2", len(args))
		}

		returnType, errorTypes, factoryEnvironment, ok := defaultFactoryResult(args[0])
		if !ok {
			return newError(object.RuntimeErrorKindType, "default factory must be a Silver function, got %s", args[0].Type())
		}

		pairs := make(map[object.HashKey]object.MapPair)
		if len(args) == 2 {
			initial, err := requireMap("defaultmap", args[1])
			if err != nil {
				return err
			}
			pairs = initial.Snapshot()
		}
		mapping := &object.Map{Pairs: pairs}

		values := map[string]object.Object{
			"values":  mapping,
			"factory": args[0],
		}
		instance := &object.StructInstance{Struct: definition, Values: values}
		getter := &object.Function{
			Name:       "get_item",
			Parameters: defaultMapGetTemplate.Parameters,
			ReturnType: returnType,
			ErrorTypes: errorTypes,
			Body:       defaultMapGetTemplate.Body,
			Env:        factoryEnvironment,
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

func defaultFactoryResult(factory object.Object) (*ast.TypeAnnotation, []*ast.TypeAnnotation, *object.Environment, bool) {
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

func parseDefaultMapGetTemplate() *ast.FunctionLiteral {
	const source = `fn(self, key) int {
	let maps = import("map")
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
