package evaluator

import (
	"path/filepath"
	"silver/object"
	"testing"
)

func TestStructConstructionAndFieldAccess(t *testing.T) {
	evaluated := testEval(`
struct Point { x: int, y: int }
let point: Point = Point(3, 4);
point.x + point.y;
`)
	testIntegerObject(t, evaluated, 7)
}

func TestStructValueInspection(t *testing.T) {
	evaluated := testEval(`struct Person { name: string, age: int } Person("Ada", 36);`)
	value, ok := evaluated.(*object.StructInstance)
	if !ok {
		t.Fatalf("result is %T, want *object.StructInstance", evaluated)
	}
	if got, want := value.Inspect(), `Person { name: Ada, age: 36 }`; got != want {
		t.Fatalf("struct inspection is %q, want %q", got, want)
	}
}

func TestStructConstructorArity(t *testing.T) {
	evaluated := testEval(`struct Point { x: int, y: int } Point(1);`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `wrong number of arguments for struct Point. got=1, want=2`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestMissingStructField(t *testing.T) {
	evaluated := testEval(`struct Point { x: int } Point(1).y;`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `struct "Point" has no field "y"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructExportedFromModule(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library.lib")
	mainPath := filepath.Join(dir, "main.slvr")
	writeMonkeyFile(t, libraryPath, `struct Point { x: int, y: int }`)
	writeMonkeyFile(t, mainPath, `
let library = import("./library.lib");
let point: library.Point = library.Point(8, 9);
point.y;
`)

	evaluated := New().EvalFile(mainPath, object.NewEnvironment())
	testIntegerObject(t, evaluated, 9)
}

func TestStructFieldTypeMismatch(t *testing.T) {
	evaluated := testEval(`struct Person { name: string, age: int } Person("Ada", "old");`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for field "Person.age": expected int, got string`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}
