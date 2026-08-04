package object

import (
	"silver/ast"
	"sync"
)

// Environment is a lexical scope. outer links closures and nested calls to
// their parent scope, while sourceDir supplies the base for relative imports.
type Environment struct {
	mu        sync.RWMutex
	store     map[string]Object // bindings defined directly in this scope
	types     map[string]*ast.TypeAnnotation
	outer     *Environment // enclosing lexical scope, if any
	sourceDir string       // directory of the source file being evaluated
	tasks     []*Task      // tasks launched directly in this lexical scope
}

// NewEnvironment constructs an empty top-level environment.
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]Object),
		types: make(map[string]*ast.TypeAnnotation),
	}
}

// Get resolves name in the current scope, then walks outward through enclosing
// scopes.
func (e *Environment) Get(name string) (Object, bool) {
	e.mu.RLock()
	obj, ok := e.store[name]
	outer := e.outer
	e.mu.RUnlock()
	if !ok && outer != nil {
		obj, ok = outer.Get(name)
	}
	return obj, ok
}

// Set creates or replaces an untyped binding in the current scope.
func (e *Environment) Set(name string, val Object) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.store[name] = val
	delete(e.types, name)
}

// SetTyped creates or replaces a binding and records its explicit type, if
// any, for later assignment checks.
func (e *Environment) SetTyped(name string, val Object, annotation *ast.TypeAnnotation) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.store[name] = val
	if annotation == nil {
		delete(e.types, name)
	} else {
		e.types[name] = annotation
	}
}

// AssignmentTarget finds the nearest lexical binding and its declared type.
func (e *Environment) AssignmentTarget(name string) (*ast.TypeAnnotation, *Environment, bool) {
	e.mu.RLock()
	if _, ok := e.store[name]; ok {
		annotation := e.types[name]
		e.mu.RUnlock()
		return annotation, e, true
	}
	outer := e.outer
	e.mu.RUnlock()
	if outer != nil {
		return outer.AssignmentTarget(name)
	}
	return nil, nil, false
}

// Assign replaces the nearest existing lexical binding. Callers resolve the
// target first, so a missing name is not an evaluator-visible condition here.
func (e *Environment) Assign(name string, val Object) {
	e.mu.Lock()
	if _, ok := e.store[name]; ok {
		e.store[name] = val
		e.mu.Unlock()
		return
	}
	outer := e.outer
	e.mu.Unlock()
	if outer != nil {
		outer.Assign(name, val)
	}
}

// Bindings returns a copy of the names defined directly in this environment.
// Enclosing scopes are intentionally excluded.
func (e *Environment) Bindings() map[string]Object {
	e.mu.RLock()
	defer e.mu.RUnlock()
	bindings := make(map[string]Object, len(e.store))
	for name, value := range e.store {
		bindings[name] = value
	}
	return bindings
}

// SetSourceDir records the directory used to resolve relative module imports.
func (e *Environment) SetSourceDir(dir string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.sourceDir = dir
}

// SourceDir returns this scope's source directory, inheriting it from an outer
// scope when necessary.
func (e *Environment) SourceDir() string {
	e.mu.RLock()
	if e.sourceDir != "" {
		dir := e.sourceDir
		e.mu.RUnlock()
		return dir
	}
	outer := e.outer
	e.mu.RUnlock()
	if outer != nil {
		return outer.SourceDir()
	}
	return ""
}

// RegisterTask associates a launched task with this scope for exit-time
// diagnostics and cleanup.
func (e *Environment) RegisterTask(task *Task) {
	e.mu.Lock()
	e.tasks = append(e.tasks, task)
	e.mu.Unlock()
}

// Tasks returns a stable snapshot of tasks launched in this scope.
func (e *Environment) Tasks() []*Task {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return append([]*Task(nil), e.tasks...)
}

// NewEnclosedEnvironment constructs a child lexical scope linked to outer.
func NewEnclosedEnvironment(outer *Environment) *Environment {
	env := NewEnvironment()
	env.outer = outer
	return env
}
