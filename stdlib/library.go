// Package stdlib contains importable modules that ship with Silver. The
// evaluator depends only on the library exposed by this package.
package stdlib

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"silver/ast"
	"silver/object"
	"silver/packages"
	"silver/source"
	"strings"
)

// silverModuleFiles contains the Silver-authored portion of the standard
// library. Embedding the directory recursively allows each module to choose
// its own internal layout. Package manifests declare the public entry files.
//
//go:embed silver all:native
var silverModuleFiles embed.FS

// definition is the declarative form of one Go-backed module export.
type definition struct {
	name      string
	fn        object.BuiltinFunction
	value     object.Object
	signature *ast.TypeAnnotation
}

// Library contains the importable standard-library modules available to one
// evaluator.
type Library struct {
	modules        map[string]*object.Module
	sourceModules  map[string]SourceModule
	sourceFiles    map[string]SourceModule
	sourcePackages map[string][]SourceModule
}

// SourceModule is an embedded, Silver-authored standard-library module. The
// evaluator owns parsing and execution so stdlib does not depend on evaluator.
type SourceModule struct {
	Name       string
	Source     string
	SourceName string
	PackageID  string
	// NativeBindings supplies Go implementations to a Silver entry file in the
	// same package. The source's export declaration controls their visibility.
	NativeBindings map[string]object.Object
}

// New constructs Silver's standard library around the evaluator's configured
// output and canonical singleton values.
func New(out io.Writer, null *object.Null, trueValue, falseValue *object.Boolean) *Library {
	return NewWithStreams(nil, out, out, null, trueValue, falseValue)
}

// NewWithStreams constructs the standard library with explicit process
// streams. A nil input behaves as an empty stream; nil outputs are discarded.
func NewWithStreams(in io.Reader, out, errOut io.Writer, null *object.Null, trueValue, falseValue *object.Boolean) *Library {
	if in == nil {
		in = &emptyReader{}
	}
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}
	if null == nil {
		panic("stdlib: canonical null value must not be nil")
	}
	if trueValue == nil || falseValue == nil {
		panic("stdlib: canonical boolean values must not be nil")
	}

	return newLibrary(map[string][]definition{
		"arrays":      arraysDefinitions(null, trueValue, falseValue),
		"collections": collectionDefinitions(null),
		"core":        coreDefinitions(),
		"io":          ioDefinitions(in, out, errOut, null, trueValue, falseValue),
		"maps":        mapsDefinitions(null, trueValue, falseValue),
		"math":        mathDefinitions(),
		"networking":  networkingDefinitions(null),
		"random":      randomDefinitions(null),
		"regex":       regexDefinitions(null),
		"string":      stringDefinitions(trueValue, falseValue),
		"system":      systemDefinitions(null),
		"terminal":    terminalDefinitions(in, out, null),
		"time":        timeDefinitions(null, trueValue, falseValue),
	})
}

type emptyReader struct{}

func (*emptyReader) Read([]byte) (int, error) { return 0, io.EOF }

func newLibrary(definitions map[string][]definition) *Library {
	modules := make(map[string]*object.Module, len(definitions))
	for name, moduleDefinitions := range definitions {
		exports := make(map[string]object.Object, len(moduleDefinitions))
		environment := object.NewEnvironment()
		for _, definition := range moduleDefinitions {
			if definition.value != nil {
				environment.Set(definition.name, definition.value)
			}
		}
		for _, definition := range moduleDefinitions {
			if _, exists := exports[definition.name]; exists {
				panic(fmt.Sprintf("standard library module %q exports %q more than once", name, definition.name))
			}
			if definition.value != nil {
				exports[definition.name] = definition.value
				continue
			}
			signature, err := object.ResolveContract(definition.signature, environment)
			if err != nil {
				panic(fmt.Sprintf("standard library %s.%s: %s", name, definition.name, err.Inspect()))
			}
			exports[definition.name] = &object.Builtin{Fn: definition.fn, Signature: signature}
		}
		modules[name] = &object.Module{ID: source.BundledID(name), Path: name, Exports: exports}
	}

	library, err := loadPackageManifests(silverModuleFiles, modules)
	if err != nil {
		panic(fmt.Sprintf("stdlib: %s", err))
	}
	return library
}

// loadPackageManifests uses the same manifest schema as filesystem packages.
// A package can provide native bindings alone, or combine them with a Silver
// entry file under the same public import name.
func loadPackageManifests(filesystem fs.FS, nativeModules map[string]*object.Module) (*Library, error) {
	library := &Library{
		modules:        make(map[string]*object.Module),
		sourceModules:  make(map[string]SourceModule),
		sourceFiles:    make(map[string]SourceModule),
		sourcePackages: make(map[string][]SourceModule),
	}
	seenPackages := make(map[string]bool)
	seenSources := make(map[string]string)
	err := fs.WalkDir(filesystem, ".", func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "package.yaml" {
			return nil
		}
		manifest, err := packages.ReadManifestFS(filesystem, filename)
		if err != nil {
			return err
		}
		if seenPackages[manifest.Name()] {
			return fmt.Errorf("duplicate standard-library package %q", manifest.Name())
		}
		seenPackages[manifest.Name()] = true
		native := nativeModules[manifest.Name()]
		if native != nil && len(manifest.Members()) == 0 {
			library.modules[manifest.Name()] = native
			return nil
		}
		packageID := "stdlib:" + manifest.Path()
		for _, member := range manifest.Members() {
			input, err := fs.ReadFile(filesystem, member.Path())
			if err != nil {
				return err
			}
			module := SourceModule{
				Name:       sourceImportName(manifest.Name(), member.Declared()),
				Source:     string(input),
				SourceName: path.Join("stdlib", member.Path()),
				PackageID:  packageID,
			}
			if native != nil && module.Name == manifest.Name() {
				module.NativeBindings = native.Exports
			}
			if _, exists := library.sourceFiles[module.SourceName]; exists {
				return fmt.Errorf("embedded source %q belongs to multiple packages", module.SourceName)
			}
			if previous, exists := seenSources[module.Name]; exists {
				return fmt.Errorf("embedded sources %q and %q have the same module name %q", previous, module.SourceName, module.Name)
			}
			seenSources[module.Name] = module.SourceName
			library.sourceFiles[module.SourceName] = module
			library.sourcePackages[packageID] = append(library.sourcePackages[packageID], module)
		}
		for _, exported := range manifest.Exports() {
			module := library.sourceFiles[path.Join("stdlib", exported.Path())]
			if _, exists := library.sourceModules[module.Name]; exists {
				return fmt.Errorf("duplicate standard-library import %q", module.Name)
			}
			library.sourceModules[module.Name] = module
		}
		if native != nil {
			if _, ok := library.sourceModules[manifest.Name()]; !ok {
				return fmt.Errorf("standard-library package %q requires an exported %s.slv entry for its native bindings", manifest.Name(), manifest.Name())
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for name := range nativeModules {
		if !seenPackages[name] {
			return nil, fmt.Errorf("native standard-library module %q requires a package.yaml manifest", name)
		}
	}
	return library, nil
}

func sourceImportName(packageName, declared string) string {
	entry := strings.TrimSuffix(declared, path.Ext(declared))
	if entry == packageName {
		return packageName
	}
	return path.Join(packageName, entry)
}

// LookupModule returns the standard-library module with the given bare import
// name, such as "math".
func (l *Library) LookupModule(name string) (*object.Module, bool) {
	module, ok := l.modules[name]
	return module, ok
}

// LookupSourceModule returns an embedded Silver implementation registered for
// the bare import name. Evaluation remains the evaluator's responsibility.
func (l *Library) LookupSourceModule(name string) (source, sourceName string, ok bool) {
	module, ok := l.LookupSource(name)
	if !ok {
		return "", "", false
	}
	return module.Source, module.SourceName, true
}

// LookupSource returns a public entry declared by a bundled package manifest.
func (l *Library) LookupSource(name string) (SourceModule, bool) {
	// Keep the former native import as an alias of the complete public module.
	// Returning its canonical Name also preserves evaluator cache/type identity.
	if name == "_networking" {
		name = "networking"
	}
	module, ok := l.sourceModules[name]
	return module, ok
}

// LookupRelativeSource permits embedded members, including internal helpers,
// to be imported relative to another embedded source's directory.
func (l *Library) LookupRelativeSource(request, sourceDir string) (SourceModule, bool) {
	if path.IsAbs(request) || filepath.IsAbs(request) {
		return SourceModule{}, false
	}
	module, ok := l.sourceFiles[path.Join(sourceDir, request)]
	return module, ok
}

// SourceMembers returns every source sharing the bundled package's grammar.
func (l *Library) SourceMembers(packageID string) []SourceModule {
	return append([]SourceModule(nil), l.sourcePackages[packageID]...)
}

func requireArgumentCount(args []object.Object, want int) *object.Error {
	if len(args) == want {
		return nil
	}
	return newError(object.RuntimeErrorKindType, "wrong number of arguments. got=%d, want=%d", len(args), want)
}

func requireArray(name string, value object.Object) (*object.Array, *object.Error) {
	array, ok := value.(*object.Array)
	if !ok {
		return nil, newError(object.RuntimeErrorKindType, "argument to `%s` must be ARRAY, got %s", name, value.Type())
	}
	return array, nil
}

func newError(kind object.RuntimeErrorKind, format string, args ...interface{}) *object.Error {
	return object.NewError(kind, fmt.Sprintf(format, args...))
}
