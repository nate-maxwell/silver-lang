package stdlib

import "silver/object"

// mapDefinitions groups the functions exported by import("map").
func mapDefinitions(null *object.Null, trueValue, falseValue *object.Boolean) []definition {
	return []definition{
		{name: "get", fn: builtinMapGet(null)},
		{name: "set", fn: builtinMapSet},
		{name: "delete", fn: builtinMapDelete},
		{name: "values", fn: builtinMapValues},
		{name: "contains", fn: builtinMapContains(trueValue, falseValue)},
	}
}

func builtinMapContains(trueValue, falseValue *object.Boolean) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		mapping, err := requireMap("contains", args[0])
		if err != nil {
			return err
		}
		key, keyError := requireHashKey(args[1])
		if keyError != nil {
			return keyError
		}
		if _, ok := mapping.Get(key); ok {
			return trueValue
		}
		return falseValue
	}
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
	pairs[key] = object.MapPair{Key: args[1], Value: args[2]}
	return &object.Map{Pairs: pairs, DefaultFactory: mapping.DefaultFactory}
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
	return &object.Map{Pairs: pairs, DefaultFactory: mapping.DefaultFactory}
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

func requireMap(name string, value object.Object) (*object.Map, *object.Error) {
	mapping, ok := value.(*object.Map)
	if !ok {
		return nil, newError(object.RuntimeErrorKindType, "argument to `%s` must be MAP, got %s", name, value.Type())
	}
	return mapping, nil
}

func requireHashKey(value object.Object) (object.HashKey, *object.Error) {
	hashable, ok := value.(object.Hashable)
	if !ok {
		return object.HashKey{}, newError(object.RuntimeErrorKindType, "unusable as hash key: %s", value.Type())
	}
	return hashable.HashKey(), nil
}

func copyMapPairs(mapping *object.Map) map[object.HashKey]object.MapPair {
	return mapping.Snapshot()
}
