package object

import "hash/fnv"

// String stores an immutable Silver string value.
type String struct {
	Value string
}

// Type returns the string runtime tag.
func (s *String) Type() ObjectType { return STRING_OBJ }

// Inspect returns the raw string value without adding quotes.
func (s *String) Inspect() string { return s.Value }

// HashKey computes a stable FNV-1a digest for use in a Silver hash.
func (s *String) HashKey() HashKey {
	hash := fnv.New64a()
	hash.Write([]byte(s.Value))
	return HashKey{Type: s.Type(), Value: hash.Sum64()}
}
