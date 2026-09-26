package evaluator

import (
	"path/filepath"
	"silver/object"
	"testing"
)

func TestTypeBuiltinReturnsPrimitiveTypeValues(t *testing.T) {
	tests := []string{
		`core.type(1) == int`,
		`core.type(1.5) == float`,
		`core.type(True) == bool`,
		`core.type("silver") == str`,
		`core.type(if False { 1 }) == null`,
		`core.type([1, 2]) == array`,
		`core.type({"answer": 42}) == map`,
		`core.type(fn() {}) == call`,
		`core.type(core.len) == call`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			testBooleanObject(t, testEval(coreImport+input), true)
		})
	}
}

func TestTypeBuiltinReturnsInspectableTypeValue(t *testing.T) {
	evaluated := testEval(coreImport + `core.type(1)`)
	typeValue, ok := evaluated.(*object.TypeDefinition)
	if !ok {
		t.Fatalf("result is %T, want *object.TypeDefinition", evaluated)
	}
	if got, want := typeValue.Inspect(), "int"; got != want {
		t.Fatalf("type inspects as %q, want %q", got, want)
	}
}

func TestTypeBuiltinReturnsStructDefinition(t *testing.T) {
	for _, input := range []string{
		`type Point = struct { x: int }
core.type(Point) == Point`,
		`type Point = struct { x: int }
core.type(Point{1}) == Point`,
	} {
		testBooleanObject(t, testEval(coreImport+input), true)
	}
}

func TestTypeBuiltinReturnsEnumDefinition(t *testing.T) {
	for _, input := range []string{
		`type Color = enum { Red, Green }
core.type(Color) == Color`,
		`type Color = enum { Red, Green }
core.type(Color.Red) == Color`,
	} {
		testBooleanObject(t, testEval(coreImport+input), true)
	}
}

func TestTypeBuiltinIdentifiesCaughtStructError(t *testing.T) {
	evaluated := testEval(coreImport + `
type FileNotFound = struct { message: str }
let open = fn() str | FileNotFound {
	return FileNotFound{"missing"}
}
try {
	open()
} catch FileNotFound err {
	core.type(err) == FileNotFound
}
`)
	testBooleanObject(t, evaluated, true)
}

func TestTypeDefinitionsAreIdempotent(t *testing.T) {
	testBooleanObject(t, testEval(coreImport+`core.type(core.type(1)) == int`), true)
}

func TestTypeBuiltinReturnsModuleType(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	libraryPath := filepath.Join(dir, "library.slv")
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, libraryPath, `let answer = 42`)
	writeSilverFile(t, mainPath, `let core = import("core:core")
let library = import("fixture:library")
core.type(library) == module`)

	evaluated := New().EvalFile(mainPath, object.NewEnvironment())
	testBooleanObject(t, evaluated, true)
}

func TestTypeBuiltinChecksArity(t *testing.T) {
	for _, input := range []string{`core.type()`, `core.type(1, 2)`} {
		evaluated := testEval(coreImport + input)
		if _, ok := evaluated.(*object.Error); !ok {
			t.Fatalf("%s returned %T, want *object.Error", input, evaluated)
		}
	}
}
