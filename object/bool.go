package object

import "fmt"

// Boolean stores a Silver truth value. The evaluator normally reuses singleton
// instances so equality can compare boolean object identity.
type Boolean struct {
	Value bool
}

// Type returns the boolean runtime tag.
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

// Inspect returns True's or False's Go-style lowercase representation.
func (b *Boolean) Inspect() string { return fmt.Sprintf("%t", b.Value) }

// HashKey converts a boolean to the stable 0/1 representation used as a hash
// table key.
func (b *Boolean) HashKey() HashKey {
	var value uint64
	if b.Value {
		value = 1
	}
	return HashKey{Type: b.Type(), Value: value}
}
