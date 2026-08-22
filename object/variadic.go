package object

// VariadicArguments is an internal argument pack bound to a variadic
// parameter. When used as a call argument, its elements contribute individual
// positional arguments rather than becoming one collection value.
type VariadicArguments struct {
	Elements []Object
}

// Type returns the internal argument-pack runtime tag.
func (v *VariadicArguments) Type() ObjectType { return VARIADIC_OBJ }

// Inspect keeps escaped argument packs identifiable in diagnostics. Normal
// calls expand a pack before a callee can inspect it.
func (v *VariadicArguments) Inspect() string { return "<variadic arguments>" }
