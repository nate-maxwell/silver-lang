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

// MapPair retains both the original key object and its associated value. The
// original object is needed when rendering the map.
type MapPair struct {
	Key   Object
	Value Object
}

// Map stores pairs by their normalized HashKey.
type Map struct {
	Pairs map[HashKey]MapPair
	mu    sync.RWMutex
}

// HashKey combines a runtime type tag with a type-specific 64-bit payload so,
// for example, integer 1 and boolean true remain distinct keys.
type HashKey struct {
	Type  ObjectType
	Value uint64
}

// Type returns the map runtime tag.
func (m *Map) Type() ObjectType { return MAP_OBJ }

// Inspect renders the map's key/value pairs. Go map iteration means pair order
// is intentionally unspecified.
func (m *Map) Inspect() string {
	var out bytes.Buffer

	pairs := []string{}
	for _, pair := range m.Snapshot() {
		pairs = append(pairs, fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect()))
	}

	out.WriteString("{")
	out.WriteString(strings.Join(pairs, ", "))
	out.WriteString("}")
	return out.String()
}

// Get returns the pair stored for a normalized key.
func (m *Map) Get(key HashKey) (MapPair, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pair, ok := m.Pairs[key]
	return pair, ok
}

// Len returns the number of pairs currently stored in the map.
func (m *Map) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.Pairs)
}

// Set creates or replaces one pair. Synchronization makes map assignment safe
// when aliases are shared by concurrent tasks.
func (m *Map) Set(key HashKey, pair MapPair) {
	m.mu.Lock()
	if m.Pairs == nil {
		m.Pairs = make(map[HashKey]MapPair)
	}
	m.Pairs[key] = pair
	m.mu.Unlock()
}

// Snapshot returns a shallow copy suitable for iteration without holding a
// lock while user code runs.
func (m *Map) Snapshot() map[HashKey]MapPair {
	m.mu.RLock()
	defer m.mu.RUnlock()
	pairs := make(map[HashKey]MapPair, len(m.Pairs))
	for key, pair := range m.Pairs {
		pairs[key] = pair
	}
	return pairs
}
