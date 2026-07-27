package object

import (
	"math"
	"strconv"
	"strings"
)

// Exact float boundaries for values representable as signed 64-bit integers.
const (
	minInt64AsFloat        = -9223372036854775808.0
	maxInt64ExclusiveFloat = 9223372036854775808.0
)

// Float stores an IEEE-754 double-precision Silver number.
type Float struct {
	Value float64
}

// Type returns the float runtime tag.
func (f *Float) Type() ObjectType { return FLOAT_OBJ }

// Inspect renders a compact decimal representation while retaining a decimal
// marker for integral floats such as 1.0.
func (f *Float) Inspect() string {
	literal := strconv.FormatFloat(f.Value, 'g', -1, 64)
	if !math.IsInf(f.Value, 0) && !math.IsNaN(f.Value) && !strings.ContainsAny(literal, ".eE") {
		literal += ".0"
	}
	return literal
}

// HashKey makes numerically equal integers and integral floats share a key.
// Other floats use their canonical IEEE-754 bits; both positive and negative
// zero normalize to the integer zero key.
func (f *Float) HashKey() HashKey {
	if f.Value >= minInt64AsFloat && f.Value < maxInt64ExclusiveFloat && math.Trunc(f.Value) == f.Value {
		return HashKey{Type: INTEGER_OBJ, Value: uint64(int64(f.Value))}
	}
	return HashKey{Type: FLOAT_OBJ, Value: math.Float64bits(f.Value)}
}
