package stdlib

import "silver/object"

// collectionDefinitions provides mutable deque and stack operations over
// arrays.
func collectionDefinitions(null *object.Null) []definition {
	defaultDictType := newDefaultDictStructDefinition()
	return []definition{
		{name: "DefaultDict", value: defaultDictType},
		{name: "deque", fn: newSequence("deque")},
		{name: "defaultdict", fn: newDefaultDict(defaultDictType, null)},
		{name: "stack", fn: newSequence("stack")},
		{name: "append", fn: collectionAppend(null)},
		{name: "appendleft", fn: collectionAppendLeft(null)},
		{name: "clear", fn: collectionClear(null)},
		{name: "copy", fn: collectionCopy},
		{name: "count", fn: collectionCount},
		{name: "extend", fn: collectionExtend(null)},
		{name: "extendleft", fn: collectionExtendLeft(null)},
		{name: "index", fn: collectionIndex},
		{name: "insert", fn: collectionInsert(null)},
		{name: "peek", fn: collectionPeek},
		{name: "pop", fn: collectionPop},
		{name: "popleft", fn: collectionPopLeft},
		{name: "push", fn: collectionAppend(null)},
		{name: "remove", fn: collectionRemove(null)},
		{name: "reverse", fn: collectionReverse(null)},
		{name: "rotate", fn: collectionRotate(null)},
	}
}

// newSequence constructs a deque or stack. Both use Silver arrays so they can
// be indexed, iterated, printed, and passed to core.len without adapters.
func newSequence(name string) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if len(args) > 1 {
			return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=0 or 1", len(args))
		}
		if len(args) == 0 {
			return &object.Array{Elements: []object.Object{}}
		}
		values, err := requireArray(name, args[0])
		if err != nil {
			return err
		}
		return copyArray(values)
	}
}

func collectionAppend(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("append", args, 2)
		if err != nil {
			return err
		}
		values.Elements = append(values.Elements, args[1])
		return null
	}
}

func collectionAppendLeft(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("appendleft", args, 2)
		if err != nil {
			return err
		}
		elements := make([]object.Object, len(values.Elements)+1)
		elements[0] = args[1]
		copy(elements[1:], values.Elements)
		values.Elements = elements
		return null
	}
}

func collectionClear(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("clear", args, 1)
		if err != nil {
			return err
		}
		for index := range values.Elements {
			values.Elements[index] = nil
		}
		values.Elements = []object.Object{}
		return null
	}
}

func collectionCopy(args ...object.Object) object.Object {
	values, err := collectionAndArity("copy", args, 1)
	if err != nil {
		return err
	}
	return copyArray(values)
}

func collectionCount(args ...object.Object) object.Object {
	values, err := collectionAndArity("count", args, 2)
	if err != nil {
		return err
	}
	var count int64
	for _, element := range values.Elements {
		if objectsEqual(element, args[1]) {
			count++
		}
	}
	return &object.Integer{Value: count}
}

func collectionExtend(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, other, err := twoCollections("extend", args)
		if err != nil {
			return err
		}
		// Copy first so extending a collection with itself is well-defined.
		values.Elements = append(values.Elements, append([]object.Object(nil), other.Elements...)...)
		return null
	}
}

func collectionExtendLeft(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, other, err := twoCollections("extendleft", args)
		if err != nil {
			return err
		}
		added := append([]object.Object(nil), other.Elements...)
		elements := make([]object.Object, 0, len(values.Elements)+len(added))
		for index := len(added) - 1; index >= 0; index-- {
			elements = append(elements, added[index])
		}
		elements = append(elements, values.Elements...)
		values.Elements = elements
		return null
	}
}

func collectionIndex(args ...object.Object) object.Object {
	values, err := collectionAndArity("index", args, 2)
	if err != nil {
		return err
	}
	for index, element := range values.Elements {
		if objectsEqual(element, args[1]) {
			return &object.Integer{Value: int64(index)}
		}
	}
	return newError(object.RuntimeErrorKindValue, "%s is not in the collection", args[1].Inspect())
}

func collectionInsert(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("insert", args, 3)
		if err != nil {
			return err
		}
		index, ok := args[1].(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "index argument to `insert` must be INTEGER, got %s", args[1].Type())
		}
		position := normalizedInsertIndex(index.Value, len(values.Elements))
		values.Elements = append(values.Elements, nil)
		copy(values.Elements[position+1:], values.Elements[position:])
		values.Elements[position] = args[2]
		return null
	}
}

func collectionPeek(args ...object.Object) object.Object {
	values, err := collectionAndArity("peek", args, 1)
	if err != nil {
		return err
	}
	if len(values.Elements) == 0 {
		return newError(object.RuntimeErrorKindIndex, "peek from an empty stack")
	}
	return values.Elements[len(values.Elements)-1]
}

func collectionPop(args ...object.Object) object.Object {
	values, err := collectionAndArity("pop", args, 1)
	if err != nil {
		return err
	}
	if len(values.Elements) == 0 {
		return newError(object.RuntimeErrorKindIndex, "pop from an empty collection")
	}
	last := len(values.Elements) - 1
	value := values.Elements[last]
	values.Elements[last] = nil
	values.Elements = values.Elements[:last]
	return value
}

func collectionPopLeft(args ...object.Object) object.Object {
	values, err := collectionAndArity("popleft", args, 1)
	if err != nil {
		return err
	}
	if len(values.Elements) == 0 {
		return newError(object.RuntimeErrorKindIndex, "popleft from an empty deque")
	}
	value := values.Elements[0]
	copy(values.Elements, values.Elements[1:])
	last := len(values.Elements) - 1
	values.Elements[last] = nil
	values.Elements = values.Elements[:last]
	return value
}

func collectionRemove(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("remove", args, 2)
		if err != nil {
			return err
		}
		for index, element := range values.Elements {
			if !objectsEqual(element, args[1]) {
				continue
			}
			copy(values.Elements[index:], values.Elements[index+1:])
			last := len(values.Elements) - 1
			values.Elements[last] = nil
			values.Elements = values.Elements[:last]
			return null
		}
		return newError(object.RuntimeErrorKindValue, "%s is not in the collection", args[1].Inspect())
	}
}

func collectionReverse(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("reverse", args, 1)
		if err != nil {
			return err
		}
		for left, right := 0, len(values.Elements)-1; left < right; left, right = left+1, right-1 {
			values.Elements[left], values.Elements[right] = values.Elements[right], values.Elements[left]
		}
		return null
	}
}

func collectionRotate(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("rotate", args, 2)
		if err != nil {
			return err
		}
		amount, ok := args[1].(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "rotation argument to `rotate` must be INTEGER, got %s", args[1].Type())
		}
		length := len(values.Elements)
		if length == 0 {
			return null
		}
		shift := int(amount.Value % int64(length))
		if shift < 0 {
			shift += length
		}
		if shift != 0 {
			elements := make([]object.Object, 0, length)
			elements = append(elements, values.Elements[length-shift:]...)
			elements = append(elements, values.Elements[:length-shift]...)
			values.Elements = elements
		}
		return null
	}
}

func collectionAndArity(name string, args []object.Object, want int) (*object.Array, *object.Error) {
	if err := requireArgumentCount(args, want); err != nil {
		return nil, err
	}
	return requireArray(name, args[0])
}

func twoCollections(name string, args []object.Object) (*object.Array, *object.Array, *object.Error) {
	values, err := collectionAndArity(name, args, 2)
	if err != nil {
		return nil, nil, err
	}
	other, otherErr := requireArray(name, args[1])
	if otherErr != nil {
		return nil, nil, otherErr
	}
	return values, other, nil
}

func copyArray(values *object.Array) *object.Array {
	return &object.Array{Elements: append([]object.Object(nil), values.Elements...)}
}

func normalizedInsertIndex(index int64, length int) int {
	if index < 0 {
		index += int64(length)
		if index < 0 {
			return 0
		}
	}
	if index > int64(length) {
		return length
	}
	return int(index)
}
