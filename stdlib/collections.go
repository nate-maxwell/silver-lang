package stdlib

import (
	"silver/ast"
	"silver/object"
)

// collectionDefinitions provides mutable deque and stack operations over
// arrays.
func collectionDefinitions(null *object.Null) []definition {
	defaultMapType := newDefaultMapStructDefinition()
	dequeType := newSequenceStructDefinition("Deque")
	stackType := newSequenceStructDefinition("Stack")
	return []definition{
		{name: "DefaultMap", value: defaultMapType},
		{name: "Deque", value: dequeType},
		{name: "Stack", value: stackType},
		{name: "deque", fn: newDeque(dequeType, null)},
		{name: "defaultmap", fn: newDefaultMap(defaultMapType, null)},
		{name: "stack", fn: newStack(stackType, null)},
		{name: "clear", fn: collectionClear(null)},
		{name: "copy", fn: collectionCopy(null)},
		{name: "count", fn: collectionCount},
		{name: "extend", fn: collectionExtend(null)},
		{name: "extendleft", fn: collectionExtendLeft(null)},
		{name: "index", fn: collectionIndex},
		{name: "insert", fn: collectionInsert(null)},
		{name: "popleft", fn: collectionPopLeft},
		{name: "remove", fn: collectionRemove(null)},
		{name: "reverse", fn: collectionReverse(null)},
		{name: "rotate", fn: collectionRotate(null)},
	}
}

func newSequenceStructDefinition(name string) *object.Struct {
	environment := object.NewEnvironment()
	definition := &object.Struct{
		Name:       name,
		Fields:     []string{"values"},
		FieldTypes: []*ast.TypeAnnotation{namedType("array")},
		Env:        environment,
	}
	environment.Set(name, definition)
	return definition
}

// newDeque constructs an empty bounded deque.
func newDeque(definition *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		maxLength, ok := args[0].(*object.Integer)
		if !ok {
			return newError(object.RuntimeErrorKindType, "argument to `deque` must be INTEGER, got %s", args[0].Type())
		}
		if maxLength.Value < 0 {
			return newError(object.RuntimeErrorKindValue, "argument to `deque` must be nonnegative")
		}
		values := &object.Array{Elements: []object.Object{}}
		return newSequenceInstance(definition, values, maxLength, null)
	}
}

// newStack constructs an empty stack.
func newStack(definition *object.Struct, null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		values := &object.Array{Elements: []object.Object{}}
		return newSequenceInstance(definition, values, nil, null)
	}
}

// newSequenceInstance supplies the operations shared by deque and stack
// instances. A deque always carries its required maximum length.
func newSequenceInstance(definition *object.Struct, values *object.Array, maxLength *object.Integer, null *object.Null) *object.StructInstance {
	instance := &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{"values": values},
	}
	if definition.Name == "Deque" {
		instance.Values[dequeMaxLengthField] = maxLength
		addDequeMethods(instance, values, maxLength.Value, null)
	} else {
		addStackMethods(instance, values, null)
	}
	addSequenceIndexMethods(instance, values, null)
	return instance
}

// addDequeMethods keeps deque mutation attached to the deque instance.
func addDequeMethods(instance *object.StructInstance, values *object.Array, maxLength int64, null *object.Null) {
	instance.Values["append"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		appendDeque(values, args[0], maxLength)
		return null
	}}
	instance.Values["appendleft"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		appendLeftDeque(values, args[0], maxLength)
		return null
	}}
	instance.Values["pop"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		return popCollection(values)
	}}
}

// addStackMethods exposes stack operations only on values made by stack().
func addStackMethods(instance *object.StructInstance, values *object.Array, null *object.Null) {
	instance.Values["push"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		values.Elements = append(values.Elements, args[0])
		return null
	}}
	instance.Values["peek"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		if len(values.Elements) == 0 {
			return newError(object.RuntimeErrorKindIndex, "peek from an empty stack")
		}
		return values.Elements[len(values.Elements)-1]
	}}
	instance.Values["pop"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		return popCollection(values)
	}}
}

func addSequenceIndexMethods(instance *object.StructInstance, values *object.Array, null *object.Null) {
	instance.Values["get_item"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		index, err := collectionArrayIndex(args[0], len(values.Elements))
		if err != nil {
			return err
		}
		return values.Elements[index]
	}}
	instance.Values["set_item"] = &object.Builtin{Fn: func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		index, err := collectionArrayIndex(args[0], len(values.Elements))
		if err != nil {
			return err
		}
		values.Elements[index] = args[1]
		return null
	}}
}

func collectionArrayIndex(value object.Object, length int) (int, *object.Error) {
	index, ok := value.(*object.Integer)
	if !ok {
		return 0, newError(object.RuntimeErrorKindType, "collection index must be INTEGER, got %s", value.Type())
	}
	if index.Value < 0 || index.Value >= int64(length) {
		return 0, newError(object.RuntimeErrorKindIndex, "collection index out of range: %d", index.Value)
	}
	return int(index.Value), nil
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

func collectionCopy(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		values, err := collectionAndArity("copy", args, 1)
		if err != nil {
			return err
		}
		result := copyArray(values)
		if instance, ok := args[0].(*object.StructInstance); ok && (instance.Struct.Name == "Deque" || instance.Struct.Name == "Stack") {
			var maxLength *object.Integer
			if instance.Struct.Name == "Deque" {
				maxLength = instance.Values[dequeMaxLengthField].(*object.Integer)
			}
			return newSequenceInstance(instance.Struct, result, maxLength, null)
		}
		return result
	}
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
		added := append([]object.Object(nil), other.Elements...)
		if maxLength, bounded := dequeMaxLength(args[0]); bounded {
			for _, value := range added {
				appendDeque(values, value, maxLength)
			}
		} else {
			values.Elements = append(values.Elements, added...)
		}
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
		if maxLength, bounded := dequeMaxLength(args[0]); bounded {
			for _, value := range added {
				appendLeftDeque(values, value, maxLength)
			}
		} else {
			elements := make([]object.Object, 0, len(values.Elements)+len(added))
			for index := len(added) - 1; index >= 0; index-- {
				elements = append(elements, added[index])
			}
			elements = append(elements, values.Elements...)
			values.Elements = elements
		}
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
		if maxLength, bounded := dequeMaxLength(args[0]); bounded && int64(len(values.Elements)) >= maxLength {
			return newError(object.RuntimeErrorKindIndex, "insert into a deque at its maximum length")
		}
		position := normalizedInsertIndex(index.Value, len(values.Elements))
		values.Elements = append(values.Elements, nil)
		copy(values.Elements[position+1:], values.Elements[position:])
		values.Elements[position] = args[2]
		return null
	}
}

func popCollection(values *object.Array) object.Object {
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
	return requireCollectionArray(name, args[0])
}

func requireCollectionArray(name string, value object.Object) (*object.Array, *object.Error) {
	if array, ok := value.(*object.Array); ok {
		return array, nil
	}
	if instance, ok := value.(*object.StructInstance); ok {
		if stored, exists := instance.Get("values"); exists {
			if array, ok := stored.(*object.Array); ok {
				return array, nil
			}
		}
	}
	return nil, newError(object.RuntimeErrorKindType, "argument to `%s` must be a collection, got %s", name, value.Type())
}

func twoCollections(name string, args []object.Object) (*object.Array, *object.Array, *object.Error) {
	values, err := collectionAndArity(name, args, 2)
	if err != nil {
		return nil, nil, err
	}
	other, otherErr := requireCollectionArray(name, args[1])
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

func dequeMaxLength(value object.Object) (int64, bool) {
	instance, ok := value.(*object.StructInstance)
	if !ok || instance.Struct.Name != "Deque" {
		return 0, false
	}
	maxLength, ok := instance.Values[dequeMaxLengthField].(*object.Integer)
	if !ok {
		return 0, false
	}
	return maxLength.Value, true
}

const dequeMaxLengthField = "<max_len>"

func appendDeque(values *object.Array, value object.Object, maxLength int64) {
	if maxLength == 0 {
		return
	}
	if int64(len(values.Elements)) >= maxLength {
		copy(values.Elements, values.Elements[1:])
		values.Elements[len(values.Elements)-1] = value
		return
	}
	values.Elements = append(values.Elements, value)
}

func appendLeftDeque(values *object.Array, value object.Object, maxLength int64) {
	if maxLength == 0 {
		return
	}
	if int64(len(values.Elements)) >= maxLength {
		copy(values.Elements[1:], values.Elements[:len(values.Elements)-1])
		values.Elements[0] = value
		return
	}
	elements := make([]object.Object, len(values.Elements)+1)
	elements[0] = value
	copy(elements[1:], values.Elements)
	values.Elements = elements
}
