package object

// Enum is the namespace bound by an enum declaration. Each member points to a
// singleton EnumValue created when the declaration is evaluated.
type Enum struct {
	Name    string
	Members map[string]*EnumValue
}

// Type returns the enum namespace runtime tag.
func (e *Enum) Type() ObjectType { return ENUM_OBJ }

// Inspect returns a compact enum namespace description.
func (e *Enum) Inspect() string { return "<enum " + e.Name + ">" }

// EnumValue is one singleton member of an Enum. HashID is assigned by the
// evaluator and uniquely identifies this value within an execution session.
type EnumValue struct {
	EnumName string
	Member   string
	HashID   uint64
}

// Type returns the enum-value runtime tag.
func (e *EnumValue) Type() ObjectType { return ENUM_VALUE_OBJ }

// Inspect renders a qualified value such as Direction.North.
func (e *EnumValue) Inspect() string { return e.EnumName + "." + e.Member }

// HashKey lets enum values serve as hash keys without conflating declarations
// that happen to use the same names.
func (e *EnumValue) HashKey() HashKey {
	return HashKey{Type: e.Type(), Value: e.HashID}
}
