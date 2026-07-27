package object

import (
	"bytes"
	"fmt"
	"strings"
)

// Hashable is implemented by runtime values that may be used as hash keys.
type Hashable interface {
	HashKey() HashKey
}

// HashPair retains both the original key object and its associated value. The
// original object is needed when rendering the hash.
type HashPair struct {
	Key   Object
	Value Object
}

// Hash stores pairs by their normalized HashKey.
type Hash struct {
	Pairs map[HashKey]HashPair
}

// HashKey combines a runtime type tag with a type-specific 64-bit payload so,
// for example, integer 1 and boolean true remain distinct keys.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Type returns the hash runtime tag.
func (h *Hash) Type() ObjectType { return HASH_OBJ }

// Inspect renders the hash's key/value pairs. Go map iteration means pair order
// is intentionally unspecified.
func (h *Hash) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range h.Pairs {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}
