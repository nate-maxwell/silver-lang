package object

import "silver/ast"

// Environment is a lexical scope. outer links closures and nested calls to
// their parent scope, while sourceDir supplies the base for relative imports.
type Environment struct {
	store     map[string]Object // bindings defined directly in this scope
	types     map[string]*ast.TypeAnnotation
	outer     *Environment // enclosing lexical scope, if any
	sourceDir string       // directory of the source file being evaluated
}

// NewEnvironment constructs an empty top-level environment.
func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, types: make(map[string]*ast.TypeAnnotation), outer: nil}
}

// Get resolves name in the current scope, then walks outward through enclosing
// scopes.
func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

// Set creates or replaces a binding in the current scope and returns val for
// evaluator convenience.
func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	delete(e.types, name)
	return val
}

// SetTyped creates or replaces a binding and records its explicit type, if
// any, for later assignment checks.
func (e *Environment) SetTyped(name string, val Object, annotation *ast.TypeAnnotation) Object {
	e.store[name] = val
	if annotation == nil {
		delete(e.types, name)
	} else {
		e.types[name] = annotation
	}
	return val
}

// AssignmentTarget finds the nearest lexical binding and its declared type.
func (e *Environment) AssignmentTarget(name string) (*ast.TypeAnnotation, *Environment, bool) {
	if _, ok := e.store[name]; ok {
		return e.types[name], e, true
	}
	if e.outer != nil {
		return e.outer.AssignmentTarget(name)
	}
	return nil, nil, false
}

// Assign replaces the nearest existing lexical binding.
func (e *Environment) Assign(name string, val Object) bool {
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		return true
	}
	if e.outer != nil {
		return e.outer.Assign(name, val)
	}
	return false
}

// Bindings returns a copy of the names defined directly in this environment.
// Enclosing scopes are intentionally excluded.
func (e *Environment) Bindings() map[string]Object {
	bindings := make(map[string]Object, len(e.store))
	for name, value := range e.store {
		bindings[name] = value
	}
	return bindings
}

// SetSourceDir records the directory used to resolve relative module imports.
func (e *Environment) SetSourceDir(dir string) {
	e.sourceDir = dir
}

// SourceDir returns this scope's source directory, inheriting it from an outer
// scope when necessary.
func (e *Environment) SourceDir() string {
	if e.sourceDir != "" {
		return e.sourceDir
	}
	if e.outer != nil {
		return e.outer.SourceDir()
	}
	return ""
}

// NewEnclosedEnvironment constructs a child lexical scope linked to outer.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}
