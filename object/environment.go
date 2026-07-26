package object

type Environment struct {
	store     map[string]Object
	outer     *Environment
	sourceDir string
}

func NewEnvironment() *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: nil}
}

func (e *Environment) Get(name string) (Object, bool) {
	obj, ok := e.store[name]
	if !ok && e.outer != nil {
		obj, ok = e.outer.Get(name)
	}
	return obj, ok
}

func (e *Environment) Set(name string, val Object) Object {
	e.store[name] = val
	return val
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

func (e *Environment) SetSourceDir(dir string) {
	e.sourceDir = dir
}

func (e *Environment) SourceDir() string {
	if e.sourceDir != "" {
		return e.sourceDir
	}
	if e.outer != nil {
		return e.outer.SourceDir()
	}
	return ""
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}
