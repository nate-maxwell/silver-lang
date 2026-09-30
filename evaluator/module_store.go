package evaluator

import (
	"silver/object"
	"silver/source"
)

// moduleStore belongs to one interpreter session, including every lazy
// template invocation. A nil entry marks an in-progress load; a non-nil entry
// retains the module and its exact nominal definitions. Like Eval, loading
// runs synchronously; recursive imports observe the active entry immediately.
type moduleStore struct {
	entries map[source.ModuleID]*object.Module
}

func newModuleStore() *moduleStore {
	return &moduleStore{entries: make(map[source.ModuleID]*object.Module)}
}

// load publishes only successful modules and clears loading state on every
// failed exit, allowing a later import to retry. path is for diagnostics only.
func (store *moduleStore) load(id source.ModuleID, path string, evaluate func() (*object.Module, *object.Error)) object.Object {
	if module, exists := store.entries[id]; exists {
		if module == nil {
			return newError(object.RuntimeErrorKindImport, "circular import detected while loading %q", path)
		}
		return module
	}
	// Install the sentinel before evaluating: recursive imports must detect
	// a cycle rather than see a partially initialized module and its types.
	store.entries[id] = nil
	defer func() {
		if store.entries[id] == nil {
			delete(store.entries, id)
		}
	}()

	module, failure := evaluate()
	if failure != nil {
		return failure
	}
	store.entries[id] = module
	return module
}
