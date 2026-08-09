package stdlib

import (
	stdjson "encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"silver/ast"
	"silver/object"
	"strconv"
	"strings"
	"unicode/utf8"
)

// jsonDefinitions groups the Python-shaped functions and error type exported
// by import("json").
func jsonDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	decodeErrorType := newJSONDecodeErrorDefinition()
	return []definition{
		{name: "JSONDecodeError", value: decodeErrorType},
		{name: "load", fn: builtinJSONLoad(decodeErrorType, null, trueValue, falseValue)},
		{name: "loads", fn: builtinJSONLoads(decodeErrorType, null, trueValue, falseValue)},
		{name: "dump", fn: builtinJSONDump(null)},
		{name: "dumps", fn: builtinJSONDumps},
	}
}

// newJSONDecodeErrorDefinition mirrors Python's useful JSONDecodeError
// attributes. message is Silver's conventional error text; the remaining
// fields have the same meaning as Python's msg, doc, pos, lineno, and colno.
func newJSONDecodeErrorDefinition() *object.Struct {
	environment := object.NewEnvironment()
	definition := &object.Struct{
		Name:   "JSONDecodeError",
		Fields: []string{"message", "msg", "doc", "pos", "lineno", "colno"},
		FieldTypes: []*ast.TypeAnnotation{
			namedType("str"),
			namedType("str"),
			namedType("str"),
			namedType("int"),
			namedType("int"),
			namedType("int"),
		},
		Env: environment,
	}
	environment.Set("JSONDecodeError", definition)
	return definition
}

func builtinJSONLoads(decodeErrorType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		document, ok := args[0].(*object.String)
		if !ok {
			return newError(object.RuntimeErrorKindType, "argument to `loads` must be STRING, got %s", args[0].Type())
		}
		return decodeJSON(decodeErrorType, document.Value, null, trueValue, falseValue)
	}
}

func builtinJSONLoad(decodeErrorType *object.Struct, null *object.Null, trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	loads := builtinJSONLoads(decodeErrorType, null, trueValue, falseValue)
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		contents := callJSONFileMethod("load", args[0], "read")
		if _, failed := contents.(*object.Error); failed {
			return contents
		}
		return loads(contents)
	}
}

func builtinJSONDumps(args ...object.Object) object.Object {
	if len(args) < 1 || len(args) > 2 {
		return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=1..2", len(args))
	}
	value, err := silverToJSON(args[0], make(map[object.Object]bool))
	if err != nil {
		return err
	}
	indent, indented, indentErr := jsonIndent(args, 1, "dumps")
	if indentErr != nil {
		return indentErr
	}
	var encoded []byte
	var goErr error
	if indented {
		encoded, goErr = stdjson.MarshalIndent(value, "", indent)
	} else {
		encoded, goErr = stdjson.Marshal(value)
	}
	if goErr != nil {
		return newError(object.RuntimeErrorKindValue, "could not encode JSON: %s", goErr)
	}
	return &object.String{Value: string(encoded)}
}

func builtinJSONDump(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if len(args) < 2 || len(args) > 3 {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=2..3", len(args))
		}
		dumpArgs := []object.Object{args[0]}
		if len(args) == 3 {
			dumpArgs = append(dumpArgs, args[2])
		}
		encoded := builtinJSONDumps(dumpArgs...)
		if _, failed := encoded.(*object.Error); failed {
			return encoded
		}
		result := callJSONFileMethod("dump", args[1], "write", encoded)
		if _, failed := result.(*object.Error); failed {
			return result
		}
		return null
	}
}

// jsonIndent implements Python's integer and string forms of indent. As in
// Python, zero and negative integers add line breaks without leading spaces.
func jsonIndent(args []object.Object, index int, function string) (string, bool, *object.Error) {
	if index >= len(args) {
		return "", false, nil
	}
	switch value := args[index].(type) {
	case *object.Integer:
		if value.Value <= 0 {
			return "", true, nil
		}
		const maximumJSONIndent = 1000
		if value.Value > maximumJSONIndent {
			return "", false, newError(object.RuntimeErrorKindValue, "argument %d to `%s` is too large (maximum %d)", index+1, function, maximumJSONIndent)
		}
		return strings.Repeat(" ", int(value.Value)), true, nil
	case *object.String:
		return value.Value, true, nil
	default:
		return "", false, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be INTEGER or STRING, got %s", index+1, function, args[index].Type())
	}
}

// callJSONFileMethod accepts Silver's File and other native file-like structs.
// Native method calls are made below the evaluator boundary, so declared I/O
// error structs must be wrapped explicitly for normal catch propagation.
func callJSONFileMethod(function string, file object.Object, method string, args ...object.Object) object.Object {
	instance, ok := file.(*object.StructInstance)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument to `%s` must be file-like, got %s", function, file.Type())
	}
	methodValue, ok := instance.Get(method)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument to `%s` must have a `%s` method", function, method)
	}
	builtin, ok := methodValue.(*object.Builtin)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument to `%s` must have a native `%s` method", function, method)
	}
	result := builtin.Fn(args...)
	if errorValue, ok := result.(*object.StructInstance); ok && errorValue.Struct != nil && strings.HasSuffix(errorValue.Struct.Name, "Error") {
		return &object.Error{Value: errorValue}
	}
	return result
}

func decodeJSON(decodeErrorType *object.Struct, document string, null *object.Null, trueValue, falseValue *object.Boolean) object.Object {
	decoder := stdjson.NewDecoder(strings.NewReader(document))
	decoder.UseNumber()

	var decoded interface{}
	if err := decoder.Decode(&decoded); err != nil {
		message, offset := jsonDecodeFailure(err)
		if offset < 0 {
			offset = len(document)
		}
		return newJSONDecodeError(decodeErrorType, message, document, offset)
	}

	end := int(decoder.InputOffset())
	for end < len(document) && isJSONWhitespace(document[end]) {
		end++
	}
	if end != len(document) {
		return newJSONDecodeError(decodeErrorType, "Extra data", document, end)
	}

	converted, err := jsonToSilver(decoded, null, trueValue, falseValue)
	if err != nil {
		return err
	}
	return converted
}

// jsonDecodeFailure returns a Python-style zero-based document position.
func jsonDecodeFailure(err error) (string, int) {
	if err == io.EOF {
		return "Expecting value", -1
	}
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return "Unexpected end of JSON input", -1
	}
	if syntaxError, ok := err.(*stdjson.SyntaxError); ok {
		position := int(syntaxError.Offset) - 1
		if position < 0 {
			position = 0
		}
		return syntaxError.Error(), position
	}
	return err.Error(), 0
}

func newJSONDecodeError(definition *object.Struct, msg, document string, bytePosition int) *object.Error {
	if bytePosition < 0 {
		bytePosition = 0
	}
	if bytePosition > len(document) {
		bytePosition = len(document)
	}
	prefix := document[:bytePosition]
	position := utf8.RuneCountInString(prefix)
	line := strings.Count(prefix, "\n") + 1
	columnPrefix := prefix
	if newline := strings.LastIndex(prefix, "\n"); newline >= 0 {
		columnPrefix = prefix[newline+1:]
	}
	column := utf8.RuneCountInString(columnPrefix) + 1
	message := fmt.Sprintf("%s: line %d column %d (char %d)", msg, line, column, position)
	return &object.Error{Value: &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{
			"message": &object.String{Value: message},
			"msg":     &object.String{Value: msg},
			"doc":     &object.String{Value: document},
			"pos":     &object.Integer{Value: int64(position)},
			"lineno":  &object.Integer{Value: int64(line)},
			"colno":   &object.Integer{Value: int64(column)},
		},
	}}
}

func isJSONWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

func jsonToSilver(value interface{}, null *object.Null, trueValue, falseValue *object.Boolean) (object.Object, *object.Error) {
	switch value := value.(type) {
	case nil:
		return null, nil
	case bool:
		if value {
			return trueValue, nil
		}
		return falseValue, nil
	case string:
		return &object.String{Value: value}, nil
	case stdjson.Number:
		literal := string(value)
		if strings.ContainsAny(literal, ".eE") {
			converted, err := strconv.ParseFloat(literal, 64)
			if err != nil {
				return nil, newError(object.RuntimeErrorKindValue, "JSON number %q cannot be represented as a float", literal)
			}
			return &object.Float{Value: converted}, nil
		}
		converted, err := strconv.ParseInt(literal, 10, 64)
		if err != nil {
			return nil, newError(object.RuntimeErrorKindValue, "JSON integer %q is outside Silver's integer range", literal)
		}
		return &object.Integer{Value: converted}, nil
	case []interface{}:
		elements := make([]object.Object, len(value))
		for index, element := range value {
			converted, err := jsonToSilver(element, null, trueValue, falseValue)
			if err != nil {
				return nil, err
			}
			elements[index] = converted
		}
		return &object.Array{Elements: elements}, nil
	case map[string]interface{}:
		pairs := make(map[object.HashKey]object.MapPair, len(value))
		for name, element := range value {
			converted, err := jsonToSilver(element, null, trueValue, falseValue)
			if err != nil {
				return nil, err
			}
			key := &object.String{Value: name}
			pairs[key.HashKey()] = object.MapPair{Key: key, Value: converted}
		}
		return &object.Map{Pairs: pairs}, nil
	default:
		return nil, newError(object.RuntimeErrorKindValue, "unsupported decoded JSON value %T", value)
	}
}

func silverToJSON(value object.Object, active map[object.Object]bool) (interface{}, *object.Error) {
	switch value := value.(type) {
	case *object.Null:
		return nil, nil
	case *object.Boolean:
		return value.Value, nil
	case *object.String:
		return value.Value, nil
	case *object.Integer:
		return value.Value, nil
	case *object.Float:
		if math.IsNaN(value.Value) || math.IsInf(value.Value, 0) {
			return nil, newError(object.RuntimeErrorKindValue, "out of range float values are not JSON compliant")
		}
		return value.Value, nil
	case *object.Array:
		if active[value] {
			return nil, newError(object.RuntimeErrorKindValue, "circular reference detected while encoding JSON")
		}
		active[value] = true
		defer delete(active, value)
		elements := make([]interface{}, len(value.Elements))
		for index, element := range value.Elements {
			converted, err := silverToJSON(element, active)
			if err != nil {
				return nil, err
			}
			elements[index] = converted
		}
		return elements, nil
	case *object.Map:
		if active[value] {
			return nil, newError(object.RuntimeErrorKindValue, "circular reference detected while encoding JSON")
		}
		active[value] = true
		defer delete(active, value)
		mapping := make(map[string]interface{}, value.Len())
		for _, pair := range value.Snapshot() {
			key, ok := pair.Key.(*object.String)
			if !ok {
				return nil, newError(object.RuntimeErrorKindType, "JSON map keys must be STRING, got %s", pair.Key.Type())
			}
			converted, err := silverToJSON(pair.Value, active)
			if err != nil {
				return nil, err
			}
			mapping[key.Value] = converted
		}
		return mapping, nil
	default:
		return nil, newError(object.RuntimeErrorKindType, "object of type %s is not JSON serializable", value.Type())
	}
}
