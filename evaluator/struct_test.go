package evaluator

import (
	"bytes"
	"path/filepath"
	"silver/object"
	"testing"
)

func TestStructConstructionAndFieldAccess(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
	y: int
}
let point: Point = Point(3, 4)
point.x + point.y
`)
	testIntegerObject(t, evaluated, 7)
}

func TestStructMethodBindsSelf(t *testing.T) {
	evaluated := testEval(`
struct Person {
	name: str
	age: int
	numbers: array
}

let shout = fn[Person]() str = {
	self.name
}

let person = Person("Ada", 36, [1, 2, 3])
person.shout()
`)
	value, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("result is %T, want *object.String", evaluated)
	}
	if value.Value != "Ada" {
		t.Fatalf("result is %q, want Ada", value.Value)
	}
}

func TestStructMethodAcceptsParameters(t *testing.T) {
	evaluated := testEval(`
struct Person {
	age: int
}
let ageAfter = fn[Person](years: int) int = {
	self.age + years
}
Person(36).ageAfter(4)
`)
	testIntegerObject(t, evaluated, 40)
}

func TestStructMethodCanCallBuiltin(t *testing.T) {
	var output bytes.Buffer
	evaluated := evalInput(t, NewWithOutput(&output), object.NewEnvironment(), `
struct Person {
	name: str
}
let greet = fn[Person]() = {
	print("hello")
}
Person("Ada").greet()
`)
	testNullObject(t, evaluated)
	if got, want := output.String(), "hello\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestStructMethodRequiresReceiver(t *testing.T) {
	evaluated := testEval(`
struct Person {
	name: str
}
let shout = fn[Person]() str = {
	self.name
}
shout()
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `method "shout" requires a receiver`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestMethodReceiverMustBeStruct(t *testing.T) {
	evaluated := testEval(`let invalid = fn[int]() = {}`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `method receiver "int" must name a struct type`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructValueInspection(t *testing.T) {
	evaluated := testEval(`
struct Person {
	name: str
	age: int
}
Person("Ada", 36)
`)
	value, ok := evaluated.(*object.StructInstance)
	if !ok {
		t.Fatalf("result is %T, want *object.StructInstance", evaluated)
	}
	if got, want := value.Inspect(), `Person { name: Ada, age: 36 }`; got != want {
		t.Fatalf("struct inspection is %q, want %q", got, want)
	}
}

func TestStructConstructorArity(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
	y: int
}
Point(1)
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `wrong number of arguments for struct Point. got=1, want=2`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestMissingStructField(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
}
Point(1).y
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `struct "Point" has no field or method "y"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructExportedFromModule(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library.lib")
	mainPath := filepath.Join(dir, "main.slvr")
	writeMonkeyFile(t, libraryPath, `
struct Point {
	x: int
	y: int
}
`)
	writeMonkeyFile(t, mainPath, `
let library = import("./library.lib")
let point: library.Point = library.Point(8, 9)
point.y
`)

	evaluated := New().EvalFile(mainPath, object.NewEnvironment())
	testIntegerObject(t, evaluated, 9)
}

func TestStructFieldTypeMismatch(t *testing.T) {
	evaluated := testEval(`
struct Person {
	name: str
	age: int
}
Person("Ada", "old")
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for field "Person.age": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructMethodCannotConflictWithField(t *testing.T) {
	evaluated := testEval(`
struct Person {
	name: str
}
let name = fn[Person]() = {}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `method "name" conflicts with field "name" on struct "Person"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestDuplicateStructMethodIsRejected(t *testing.T) {
	evaluated := testEval(`
struct Person {}
let greet = fn[Person]() = {}
let greet = fn[Person]() = {}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `duplicate method "greet" on struct "Person"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructMethodExportedFromModule(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "people.lib")
	mainPath := filepath.Join(dir, "main.slvr")
	writeMonkeyFile(t, libraryPath, `
struct Person {
	name: str
}
let shout = fn[Person]() str = {
	self.name
}
`)
	writeMonkeyFile(t, mainPath, `
let people = import("./people.lib")
people.Person("Ada").shout()
`)

	evaluated := New().EvalFile(mainPath, object.NewEnvironment())
	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "Ada" {
		t.Fatalf("result is %#v, want Ada", evaluated)
	}
}
