package evaluator

import (
	"os"
	"silver/ast"
	"silver/object"
	"silver/packages"
	"silver/parser"
	"silver/source"
	"silver/stdlib"
	"sync"
)

// packageState contains evaluator-owned data derived from an immutable
// package manifest. It is shared by evaluator forks.
type packageState struct {
	mu       sync.Mutex
	prepared bool
	programs map[source.ModuleID]*ast.Program
	parseErr *object.Error
}

type packageStateSet struct {
	mu     sync.Mutex
	values map[packageStateKey]*packageState
}

type packageStateKey struct {
	manifest  *packages.Manifest
	bundledID string
}

func newPackageStateSet() *packageStateSet {
	return &packageStateSet{values: make(map[packageStateKey]*packageState)}
}

func (states *packageStateSet) forPackage(key packageStateKey) *packageState {
	states.mu.Lock()
	defer states.mu.Unlock()
	state := states.values[key]
	if state == nil {
		state = &packageState{programs: make(map[source.ModuleID]*ast.Program)}
		states.values[key] = state
	}
	return state
}

func (e *Evaluator) refreshPackageIndex() error {
	return e.packages.Refresh(os.Getenv(importPathEnvironment))
}

// preparePackage parses every member with one operator registry. This makes
// operator spellings visible throughout their package without leaking them
// into standalone sources or other packages.
func (e *Evaluator) preparePackage(manifest *packages.Manifest) (*packageState, *object.Error) {
	state := e.packageStates.forPackage(packageStateKey{manifest: manifest})
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.prepared {
		return state, state.parseErr
	}
	state.prepared = true

	registry := e.operatorScope(manifest.ID()).registry
	members := manifest.Members()
	inputs := make(map[source.ModuleID][]byte, len(members))
	for _, exported := range members {
		input, err := os.ReadFile(exported.Path())
		if err != nil {
			state.parseErr = newError(object.RuntimeErrorKindImport, "could not read %q: %s", exported.Path(), err)
			return state, state.parseErr
		}
		inputs[source.FileID(exported.Path())] = input
		for _, declaration := range parser.DiscoverOperatorDeclarations(string(input), exported.Path()) {
			if message := registry.Predefine(declaration); message != "" {
				state.parseErr = newError(
					object.RuntimeErrorKindSyntax,
					"could not parse %q:\n%s:%d:%d: %s",
					exported.Path(),
					declaration.Position.Source,
					declaration.Position.Line,
					declaration.Position.Column,
					message,
				)
				return state, state.parseErr
			}
		}
	}
	for _, exported := range members {
		input := inputs[source.FileID(exported.Path())]
		program, parseError := ParseSourceWithRegistry(exported.Path(), input, registry)
		if parseError != nil {
			state.parseErr = parseError
			return state, parseError
		}
		state.programs[source.FileID(exported.Path())] = program
	}
	return state, nil
}

// prepareSourcePackage gives each embedded package its own grammar, including
// declarations in internal members and in members imported later at runtime.
func (e *Evaluator) prepareSourcePackage(module stdlib.SourceModule) (*packageState, *object.Error) {
	state := e.packageStates.forPackage(packageStateKey{bundledID: module.PackageID})
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.prepared {
		return state, state.parseErr
	}
	state.prepared = true
	registry := e.operatorScope(module.PackageID).registry
	members := e.standardLibrary.SourceMembers(module.PackageID)
	for _, member := range members {
		for _, declaration := range parser.DiscoverOperatorDeclarations(member.Source, member.SourceName) {
			if message := registry.Predefine(declaration); message != "" {
				state.parseErr = newError(object.RuntimeErrorKindSyntax, "could not parse %q:\n%s:%d:%d: %s",
					member.SourceName, declaration.Position.Source, declaration.Position.Line, declaration.Position.Column, message)
				return state, state.parseErr
			}
		}
	}
	for _, member := range members {
		program, parseError := ParseSourceWithRegistry(member.SourceName, []byte(member.Source), registry)
		if parseError != nil {
			state.parseErr = parseError
			return state, parseError
		}
		state.programs[source.BundledID(member.Name)] = program
	}
	return state, nil
}
