package evaluator

import (
	"math"
	"silver/object"
)

// constantPool interns immutable scalar literals for one evaluator session.
// A session spans its entry file and every imported module, while separate
// evaluators can release their constants independently.
type constantPool struct {
	integers map[int64]*object.Integer
	floats   map[uint64]*object.Float
	strings  map[string]*object.String
}

func newConstantPool() *constantPool {
	return &constantPool{
		integers: make(map[int64]*object.Integer),
		floats:   make(map[uint64]*object.Float),
		strings:  make(map[string]*object.String),
	}
}

func (pool *constantPool) integer(value int64) *object.Integer {
	if constant, ok := pool.integers[value]; ok {
		return constant
	}
	constant := &object.Integer{Value: value}
	pool.integers[value] = constant
	return constant
}

func (pool *constantPool) float(value float64) *object.Float {
	// Bits preserve distinctions such as positive and negative zero and do not
	// collapse different NaN payloads.
	key := math.Float64bits(value)
	if constant, ok := pool.floats[key]; ok {
		return constant
	}
	constant := &object.Float{Value: value}
	pool.floats[key] = constant
	return constant
}

func (pool *constantPool) string(value string) *object.String {
	if constant, ok := pool.strings[value]; ok {
		return constant
	}
	constant := &object.String{Value: value}
	pool.strings[value] = constant
	return constant
}
