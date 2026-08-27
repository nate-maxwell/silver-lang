package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"silver/ast"
	"silver/astcache"
	"silver/object"
	"silver/packages"
	"silver/parser"
	"sort"
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
	var declarations []parser.OperatorDeclaration
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
			declarations = append(declarations, declaration)
		}
	}
	cacheContext := packageCacheContext(declarations)
	for _, exported := range exports {
		input := inputs[packagePathKey(exported.Path())]
		if program, ok := astcache.LoadWithContext(exported.Path(), input, cacheContext); ok {
			state.programs[packagePathKey(exported.Path())] = program
			continue
		}
		program, parseError := parseSourceWithRegistry(exported.Path(), input, registry)
		if parseError != nil {
			state.parseErr = parseError
			return state, parseError
		}
		state.programs[packagePathKey(exported.Path())] = program
		// Cache writes are optional; read-only packages must remain importable.
		_ = astcache.StoreWithContext(exported.Path(), input, cacheContext, program)
	}
	return state, nil
}

// packageCacheContext identifies the grammar shared by a manifest's exports.
// An empty context deliberately remains compatible with ordinary file caches.
func packageCacheContext(declarations []parser.OperatorDeclaration) []byte {
	if len(declarations) == 0 {
		return nil
	}
	sort.Slice(declarations, func(left, right int) bool {
		if declarations[left].Symbol != declarations[right].Symbol {
			return declarations[left].Symbol < declarations[right].Symbol
		}
		return declarations[left].BindingPower < declarations[right].BindingPower
	})
	var context strings.Builder
	context.WriteString("package-operators-v1\n")
	for _, declaration := range declarations {
		fmt.Fprintf(&context, "%d:%s:%d\n", len(declaration.Symbol), declaration.Symbol, declaration.BindingPower)
	}
	return []byte(context.String())
}

func packagePathKey(path string) string {
	key := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return key
}
