package object

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
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
	mu    sync.RWMutex
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
	for _, pair := range h.Snapshot() {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// Get returns the pair stored for a normalized key.
func (h *Hash) Get(key HashKey) (HashPair, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	pair, ok := h.Pairs[key]
	return pair, ok
}

// Len returns the number of pairs currently stored in the map.
func (h *Hash) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Pairs)
}

// Set creates or replaces one pair. Synchronization makes map assignment safe
// when aliases are shared by concurrent tasks.
func (h *Hash) Set(key HashKey, pair HashPair) {
	h.mu.Lock()
	if h.Pairs == nil {
		h.Pairs = make(map[HashKey]HashPair)
	}
	h.Pairs[key] = pair
	h.mu.Unlock()
}

// Snapshot returns a shallow copy suitable for iteration without holding a
// lock while user code runs.
func (h *Hash) Snapshot() map[HashKey]HashPair {
	h.mu.RLock()
	defer h.mu.RUnlock()
	pairs := make(map[HashKey]HashPair, len(h.Pairs))
	for key, pair := range h.Pairs {
		pairs[key] = pair
	}
	return pairs
}
