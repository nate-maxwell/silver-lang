package evaluator

import (
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
let point: Point = Point{3, 4}
point.x + point.y
`)
	testIntegerObject(t, evaluated, 7)
}

func TestStructValueInspection(t *testing.T) {
	evaluated := testEval(`
struct Person {
	name: str
	age: int
}
Person{"Ada", 36}
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
Point{1}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `wrong number of arguments for struct Point. got=1, want=2`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestMissingStructField(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
}
Point{1}.y
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `struct "Point" has no field "y"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructExportedFromModule(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library.slv")
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, libraryPath, `
struct Point {
	x: int
	y: int
}
`)
	writeSilverFile(t, mainPath, `
let library = import("./library.slv")
let point: library.Point = library.Point{8, 9}
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
Person{"Ada", "old"}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for field "Person.age": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestFunctionDestructuresMismatchedStructArgument(t *testing.T) {
	evaluated := testEval(`
struct Location {
	x: float
	y: float
	z: float
}
struct Rotation {
	x: float
	y: float
	z: float
}
struct Scale {
	x: float
	y: float
	z: float
}
struct Transform {
	location: Location
	rotation: Rotation
	scale: Scale
}
let move = fn(x: float, y: float, z: float) Location {
	return Location{x + 5.0, y + 5.0, z + 5.0}
}
let grow = fn(scale: Scale, amount: float) Scale {
	return Scale{scale.x + amount, scale.y + amount, scale.z + amount}
}
let actor = Transform{
	Location{0.0, 0.0, 0.0},
	Rotation{0.0, 0.0, 0.0},
	Scale{1.0, 1.0, 1.0}
}
let moved: Location = move(actor.location)
let grown: Scale = grow(actor.scale, 5.0)
moved.x + moved.y + moved.z + grown.x + grown.y + grown.z
`)
	testFloatObject(t, evaluated, 33.0)
}

func TestMatchingStructArgumentIsNotDestructured(t *testing.T) {
	evaluated := testEval(`
struct Wrapped {
	wrapped: int
}
let read = fn(wrapped: Wrapped) int { wrapped.wrapped }
read(Wrapped{7})
`)
	testIntegerObject(t, evaluated, 7)
}

func TestDestructuringCombinesWithPositionalArguments(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
	y: int
}
let combine = fn(offset: int, x: int, y: int) int {
	offset + x * 10 + y
}
combine(100, Point{2, 3})
`)
	testIntegerObject(t, evaluated, 123)
}

func TestDestructuringCombinesMultipleStructArguments(t *testing.T) {
	evaluated := testEval(`
struct Position {
	x: int
	y: int
}
struct Size {
	width: int
	height: int
}
let encode = fn(x: int, y: int, width: int, height: int) int {
	x * 1000 + y * 100 + width * 10 + height
}
encode(Position{1, 2}, Size{3, 4})
`)
	testIntegerObject(t, evaluated, 1234)
}

func TestDestructuringMatchesFieldNamesNotDeclarationOrder(t *testing.T) {
	evaluated := testEval(`
struct Coordinates {
	z: int
	x: int
	y: int
}
let encode = fn(x: int, y: int, z: int) int { x * 100 + y * 10 + z }
encode(Coordinates{3, 1, 2})
`)
	testIntegerObject(t, evaluated, 123)
}

func TestDestructuredFieldMustMatchParameterType(t *testing.T) {
	evaluated := testEval(`
struct Payload {
	value: str
}
let consume = fn(value: int) int { value }
consume(Payload{"wrong"})
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "value": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestDestructuringCanBindLaterNamedParameters(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
}
let combine = fn(value: int, x: int) int { value * 10 + x }
combine(Point{2}, 3)
`)
	testIntegerObject(t, evaluated, 32)
}

func TestStructWithoutMatchingFieldsKeepsTypeMismatch(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
}
let consume = fn(value: int) int { value }
consume(Point{1})
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "value": expected int, got Point`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructLiteralRejectsNonStructType(t *testing.T) {
	evaluated := testEval(`
let value = 1
value{}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `not a struct: int`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructMemberAssignment(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
	y: int
}
let point = Point{3, 4}
point.x = 8
point.x + point.y
`)
	testIntegerObject(t, evaluated, 12)
}

func TestNestedStructMemberAssignment(t *testing.T) {
	evaluated := testEval(`
struct Location {
	x: int
}
struct Actor {
	location: Location
}
let move = fn(x: int) Location { Location{x + 1} }
let actor = Actor{Location{4}}
actor.location = move(actor.location)
actor.location.x
`)
	testIntegerObject(t, evaluated, 5)
}

func TestStructMemberAssignmentChecksFieldType(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
}
let point = Point{1}
point.x = "wrong"
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for field "Point.x": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStructMemberAssignmentMutatesAliases(t *testing.T) {
	evaluated := testEval(`
struct Point {
	x: int
}
let first = Point{1}
let second = first
first.x = 9
second.x
`)
	testIntegerObject(t, evaluated, 9)
}
