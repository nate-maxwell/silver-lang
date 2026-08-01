package evaluator

import (
	"path/filepath"
	"silver/object"
	"testing"
)

func TestTypeBuiltinReturnsPrimitiveTypeValues(t *testing.T) {
	tests := []string{
		`type(1) == int`,
		`type(1.5) == float`,
		`type(True) == bool`,
		`type("silver") == str`,
		`type(if (False) { 1 }) == null`,
		`type([1, 2]) == array`,
		`type({"answer": 42}) == hash`,
		`type(fn() {}) == call`,
		`type(len) == call`,
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			testBooleanObject(t, testEval(input), true)
		})
	}
}

func TestTypeBuiltinReturnsInspectableTypeValue(t *testing.T) {
	evaluated := testEval(`type(1)`)
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
		`struct Point { x: int }
type(Point) == Point`,
		`struct Point { x: int }
type(Point{1}) == Point`,
	} {
		testBooleanObject(t, testEval(input), true)
	}
}

func TestTypeBuiltinReturnsEnumDefinition(t *testing.T) {
	for _, input := range []string{
		`enum Color { Red, Green }
type(Color) == Color`,
		`enum Color { Red, Green }
type(Color.Red) == Color`,
	} {
		testBooleanObject(t, testEval(input), true)
	}
}

func TestTypeBuiltinIdentifiesStructValuedError(t *testing.T) {
	evaluated := testEval(`
struct FileNotFound { message: str }
let open = fn() str | FileNotFound {
	return FileNotFound{"missing"}
}
type(open()) == FileNotFound
`)
	testBooleanObject(t, evaluated, true)
}

func TestTypeDefinitionsAreIdempotent(t *testing.T) {
	testBooleanObject(t, testEval(`type(type(1)) == int`), true)
}

func TestTypeBuiltinReturnsModuleType(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library.silver")
	mainPath := filepath.Join(dir, "main.silver")
	writeMonkeyFile(t, libraryPath, `let answer = 42`)
	writeMonkeyFile(t, mainPath, `let library = import("./library.silver")
type(library) == module`)

	evaluated := New().EvalFile(mainPath, object.NewEnvironment())
	testBooleanObject(t, evaluated, true)
}

func TestTypeBuiltinChecksArity(t *testing.T) {
	for _, input := range []string{`type()`, `type(1, 2)`} {
		evaluated := testEval(input)
		if _, ok := evaluated.(*object.Error); !ok {
			t.Fatalf("%s returned %T, want *object.Error", input, evaluated)
		}
	}
}
