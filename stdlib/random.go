package stdlib

import (
	"math/rand"
	"silver/object"
	"sync"
	"time"
)

// randomGenerator owns the state shared by the functions in one random
// module. rand.Rand is not safe for concurrent use, so calls are serialized.
type randomGenerator struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// randomDefinitions contains the functions exported by import("random").
func randomDefinitions(null *object.Null) []definition {
	generator := &randomGenerator{rng: rand.New(rand.NewSource(time.Now().UnixNano()))}
	return []definition{
		{name: "random", fn: generator.random},
		{name: "seed", fn: generator.seed(null)},
		{name: "randint", fn: generator.randint},
		{name: "randelem", fn: generator.randelem},
		{name: "randkey", fn: generator.randkey},
		{name: "shuffle", fn: generator.shuffle(null)},
	}
}

func (g *randomGenerator) random(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	g.mu.Lock()
	value := g.rng.Float64()
	g.mu.Unlock()
	return &object.Float{Value: value}
}

func (g *randomGenerator) seed(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		value, err := requireRandomInteger("seed", 0, args[0])
		if err != nil {
			return err
		}
		g.mu.Lock()
		g.rng.Seed(value)
		g.mu.Unlock()
		return null
	}
}

func (g *randomGenerator) randint(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	lower, err := requireRandomInteger("randint", 0, args[0])
	if err != nil {
		return err
	}
	upper, err := requireRandomInteger("randint", 1, args[1])
	if err != nil {
		return err
	}
	if lower > upper {
		return newError(object.RuntimeErrorKindValue, "lower bound to `randint` must not exceed upper bound")
	}

	g.mu.Lock()
	value := int64(uint64(lower) + g.boundedUint64Locked(uint64(upper)-uint64(lower)+1))
	g.mu.Unlock()
	return &object.Integer{Value: value}
}

func (g *randomGenerator) randelem(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	array, err := requireArray("randelem", args[0])
	if err != nil {
		return err
	}
	if len(array.Elements) == 0 {
		return newError(object.RuntimeErrorKindIndex, "cannot choose an element from an empty array")
	}

	g.mu.Lock()
	index := g.boundedUint64Locked(uint64(len(array.Elements)))
	g.mu.Unlock()
	return array.Elements[index]
}

func (g *randomGenerator) randkey(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	mapping, err := requireMap("randkey", args[0])
	if err != nil {
		return err
	}
	pairs := mapping.Snapshot()
	if len(pairs) == 0 {
		return newError(object.RuntimeErrorKindKey, "cannot choose a key from an empty map")
	}

	g.mu.Lock()
	index := g.boundedUint64Locked(uint64(len(pairs)))
	g.mu.Unlock()
	for _, pair := range pairs {
		if index == 0 {
			return pair.Key
		}
		index--
	}
	panic("stdlib: random map index out of range")
}

func (g *randomGenerator) shuffle(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 1); err != nil {
			return err
		}
		array, err := requireArray("shuffle", args[0])
		if err != nil {
			return err
		}

		g.mu.Lock()
		g.rng.Shuffle(len(array.Elements), func(left, right int) {
			array.Elements[left], array.Elements[right] = array.Elements[right], array.Elements[left]
		})
		g.mu.Unlock()
		return null
	}
}

// boundedUint64Locked returns a value in [0, bound). A zero bound represents
// the full uint64 range, which lets randint cover Silver's entire int range.
// Rejection sampling avoids modulo bias for all other bounds.
func (g *randomGenerator) boundedUint64Locked(bound uint64) uint64 {
	if bound == 0 {
		return g.rng.Uint64()
	}
	minimum := -bound % bound
	for {
		value := g.rng.Uint64()
		if value >= minimum {
			return value % bound
		}
	}
}

func requireRandomInteger(name string, index int, value object.Object) (int64, *object.Error) {
	integer, ok := value.(*object.Integer)
	if !ok {
		return 0, newError(object.RuntimeErrorKindType, "argument %d to `%s` must be INTEGER, got %s", index+1, name, value.Type())
	}
	return integer.Value, nil
}
