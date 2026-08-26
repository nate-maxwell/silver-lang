package evaluator

import (
	"os"
	"path/filepath"
	"runtime"
	"silver/ast"
	"silver/object"
	"silver/packages"
	"silver/parser"
	"strings"
	"sync"
)

// packageState contains evaluator-owned data derived from an immutable
// package manifest. It is shared by evaluator forks.
type packageState struct {
	mu       sync.Mutex
	prepared bool
	programs map[string]*ast.Program
	parseErr *object.Error
}

type packageStateSet struct {
	mu     sync.Mutex
	values map[*packages.Manifest]*packageState
}

func newPackageStateSet() *packageStateSet {
	return &packageStateSet{values: make(map[*packages.Manifest]*packageState)}
}

func (states *packageStateSet) forManifest(manifest *packages.Manifest) *packageState {
	states.mu.Lock()
	defer states.mu.Unlock()
	state := states.values[manifest]
	if state == nil {
		state = &packageState{programs: make(map[string]*ast.Program)}
		states.values[manifest] = state
	}
	return state
}

func (e *Evaluator) refreshPackageIndex() error {
	return e.packages.Refresh(os.Getenv(importPathEnvironment))
}

// preparePackage parses every export with one operator registry. This makes
// operator spellings visible throughout their package without leaking them
// into standalone sources or other packages.
func (e *Evaluator) preparePackage(manifest *packages.Manifest) (*packageState, *object.Error) {
	state := e.packageStates.forManifest(manifest)
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.prepared {
		return state, state.parseErr
	}
	state.prepared = true

	registry := e.operatorScope(manifest.ID()).registry
	exports := manifest.Exports()
	inputs := make(map[string][]byte, len(exports))
	for _, exported := range exports {
		input, err := os.ReadFile(exported.Path())
		if err != nil {
			state.parseErr = newError(object.RuntimeErrorKindImport, "could not read %q: %s", exported.Path(), err)
			return state, state.parseErr
		}
		inputs[packagePathKey(exported.Path())] = input
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
	for _, exported := range exports {
		input := inputs[packagePathKey(exported.Path())]
		program, parseError := parseSourceWithRegistry(exported.Path(), input, registry)
		if parseError != nil {
			state.parseErr = parseError
			return state, parseError
		}
		state.programs[packagePathKey(exported.Path())] = program
	}
	return state, nil
}

func packagePathKey(path string) string {
	key := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return key
}
