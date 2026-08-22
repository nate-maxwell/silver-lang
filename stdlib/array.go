package stdlib

import (
	"silver/ast"
	"silver/object"
	"sort"
)

// arrayDefinitions groups the native array operations. Keeping this list next
// to their implementations makes their Silver names easy to audit.
func arrayDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	return []definition{
		{name: "append", fn: builtinAppend},
		{name: "of", fn: builtinArrayOf, signature: &ast.TypeAnnotation{Parts: []string{"call"}, ParameterNames: []string{"values"}, ParameterTypes: []*ast.TypeAnnotation{nil}, Variadic: true, ReturnType: namedType("array")}},
		{name: "contains", fn: builtinArrayContains(trueValue, falseValue)},
		{name: "first", fn: builtinFirst(null)},
		{name: "last", fn: builtinLast(null)},
		{name: "remove", fn: builtinRemove(null)},
		{name: "rest", fn: builtinRest(null)},
		{name: "reverse", fn: builtinReverse},
		{name: "sort", fn: builtinSort},
	}
}

// builtinArrayOf collects variadic call arguments into an ordinary array.
// It gives Silver code a way to inspect an optional argument pack.
func builtinArrayOf(args ...object.Object) object.Object {
	elements := append([]object.Object(nil), args...)
	return &object.Array{Elements: elements}
}

func builtinArrayContains(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		array, err := requireArray("contains", args[0])
		if err != nil {
			return err
		}
		for _, element := range array.Elements {
			if objectsEqual(element, args[1]) {
				return trueValue
			}
		}
		return falseValue
	}
}

func objectsEqual(left, right object.Object) bool {
	if isNumber(left) && isNumber(right) {
		if numberIsNaN(left) || numberIsNaN(right) {
			return false
		}
		return compareNumbers(left, right) == 0
	}
	if left.Type() != right.Type() {
		return false
	}
	switch left := left.(type) {
	case *object.String:
		return left.Value == right.(*object.String).Value
	case *object.Boolean:
		return left.Value == right.(*object.Boolean).Value
	default:
		return left == right
	}
}

// builtinFirst creates a builtin that returns the first array element, or null
// for an empty array.
func builtinFirst(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("first", args[0])
		if err != nil {
			return err
		}
		if len(array.Elements) == 0 {
			return null
		}
		return array.Elements[0]
	}
}

// builtinLast creates a builtin that returns the last array element, or null
// for an empty array.
func builtinLast(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("last", args[0])
		if err != nil {
			return err
		}
		if len(array.Elements) == 0 {
			return null
		}
		return array.Elements[len(array.Elements)-1]
	}
}

// builtinRemove returns a copy without the element at index. Like array
// indexing, an index outside the array produces null.
func builtinRemove(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		array, err := requireArray("remove", args[0])
		if err != nil {
			return err
		}
		index, ok := args[1].(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "index argument to `remove` must be INTEGER, got %s", args[1].Type())
		}
		if index.Value < 0 || index.Value >= int64(len(array.Elements)) {
			return null
		}

		elements := make([]object.Object, 0, len(array.Elements)-1)
		elements = append(elements, array.Elements[:index.Value]...)
		elements = append(elements, array.Elements[index.Value+1:]...)
		return &object.Array{Elements: elements}
	}
}

// builtinRest creates a builtin that returns a new array containing every
// element except the first. Empty input produces null.
func builtinRest(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("rest", args[0])
		if err != nil {
			return err
		}
		if len(array.Elements) == 0 {
			return null
		}

		elements := make([]object.Object, len(array.Elements)-1)
		copy(elements, array.Elements[1:])
		return &object.Array{Elements: elements}
	}
}

// builtinAppend returns a new array with the supplied value appended. The input
// array is not mutated.
func builtinAppend(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	array, err := requireArray("append", args[0])
	if err != nil {
		return err
	}

	elements := make([]object.Object, len(array.Elements)+1)
	copy(elements, array.Elements)
	elements[len(array.Elements)] = args[1]
	return &object.Array{Elements: elements}
}

// builtinReverse returns a reversed copy and leaves the input untouched.
func builtinReverse(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray("reverse", args[0])
	if err != nil {
		return err
	}

	elements := make([]object.Object, len(array.Elements))
	for index, element := range array.Elements {
		elements[len(elements)-1-index] = element
	}
	return &object.Array{Elements: elements}
}

// builtinSort returns an ascending, stable-sorted copy. Numeric arrays may mix
// integers and floats; otherwise every element must be a string.
func builtinSort(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray("sort", args[0])
	if err != nil {
		return err
	}

	elements := append([]object.Object(nil), array.Elements...)
	if len(elements) == 0 {
		return &object.Array{Elements: elements}
	}

	numeric := isNumber(elements[0])
	strings := elements[0].Type() == object.STRING_OBJ
	if !numeric && !strings {
		return newError(object.RuntimeErrorKindType, "argument to `sort` contains unsortable type %s", elements[0].Type())
	}
	for _, element := range elements[1:] {
		if numeric && !isNumber(element) || strings && element.Type() != object.STRING_OBJ {
			return newError(object.RuntimeErrorKindType, "argument to `sort` must contain only numbers or only strings, got %s and %s", elements[0].Type(), element.Type())
		}
	}
	if len(elements) == 1 {
		return &object.Array{Elements: elements}
	}

	sort.SliceStable(elements, func(left, right int) bool {
		if numeric {
			return compareNumbers(elements[left], elements[right]) < 0
		}
		return elements[left].(*object.String).Value < elements[right].(*object.String).Value
	})
	return &object.Array{Elements: elements}
}
