package builtins

import "silver/object"

// mapDefinitions groups the native map operations. contains shares the
// collection implementation used by arrays so the global function can
// dispatch on either receiver type.
func mapDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	return []definition{
		{name: "get", fn: builtinMapGet(null)},
		{name: "set", fn: builtinMapSet},
		{name: "delete", fn: builtinMapDelete},
		{name: "values", fn: builtinMapValues},
		{name: "contains", fn: builtinContains(trueValue, falseValue)},
	}
}

// mapGlobalDefinitions omits contains because the array definition already
// registers the shared, type-dispatching implementation under that name.
func mapGlobalDefinitions(definitions []definition) []definition {
	globals := make([]definition, 0, len(definitions)-1)
	for _, definition := range definitions {
		if definition.name != "contains" {
			globals = append(globals, definition)
		}
	}
	return globals
}

// builtinMapGet returns the value associated with key, or null when no pair is
// present.
func builtinMapGet(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		mapping, err := requireMap("get", args[0])
		if err != nil {
			return err
		}
		key, keyError := requireHashKey(args[1])
		if keyError != nil {
			return keyError
		}
		pair, ok := mapping.Get(key)
		if !ok {
			return null
		}
		return pair.Value
	}
}

// builtinMapSet returns a copy containing the supplied key/value pair. An
// existing normalized key is replaced without mutating the input map.
func builtinMapSet(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 3); err != nil {
		return err
	}
	mapping, err := requireMap("set", args[0])
	if err != nil {
		return err
	}
	key, keyError := requireHashKey(args[1])
	if keyError != nil {
		return keyError
	}

	pairs := copyMapPairs(mapping)
	pairs[key] = object.HashPair{Key: args[1], Value: args[2]}
	return &object.Hash{Pairs: pairs}
}

// builtinMapDelete returns a copy without key. A missing key leaves the copied
// map unchanged.
func builtinMapDelete(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	mapping, err := requireMap("delete", args[0])
	if err != nil {
		return err
	}
	key, keyError := requireHashKey(args[1])
	if keyError != nil {
		return keyError
	}

	pairs := copyMapPairs(mapping)
	delete(pairs, key)
	return &object.Hash{Pairs: pairs}
}

// builtinMapValues returns the map values in the map's unspecified iteration
// order.
func builtinMapValues(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	mapping, err := requireMap("values", args[0])
	if err != nil {
		return err
	}

	pairs := mapping.Snapshot()
	values := make([]object.Object, 0, len(pairs))
	for _, pair := range pairs {
		values = append(values, pair.Value)
	}
	return &object.Array{Elements: values}
}

func requireMap(name string, value object.Object) (*object.Hash, *object.Error) {
	mapping, ok := value.(*object.Hash)
	if !ok {
		return nil, newError("argument to `%s` must be HASH, got %s", name, value.Type())
	}
	return mapping, nil
}

func requireHashKey(value object.Object) (object.HashKey, *object.Error) {
	hashable, ok := value.(object.Hashable)
	if !ok {
		return object.HashKey{}, newError("unusable as hash key: %s", value.Type())
	}
	return hashable.HashKey(), nil
}

func copyMapPairs(mapping *object.Hash) map[object.HashKey]object.HashPair {
	return mapping.Snapshot()
}
