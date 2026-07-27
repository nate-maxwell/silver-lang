package object

import "fmt"

// Integer stores a signed 64-bit Silver integer.
type Integer struct {
	Value int64
}

// Inspect returns the base-10 representation of the integer.
func (i *Integer) Inspect() string { return fmt.Sprintf("%d", i.Value) }

// Type returns the integer runtime tag.
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

// HashKey uses the integer bits directly as the hash payload.
func (i *Integer) HashKey() HashKey {
	return HashKey{Type: i.Type(), Value: uint64(i.Value)}
}
