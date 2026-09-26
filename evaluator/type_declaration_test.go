package evaluator

import (
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestTypeDeclarations(t *testing.T) {
	testBooleanObject(t, testEval(`type Point = struct { x: float, y: float }
type Direction = enum { North, East, South, West }
type Names = array[str]
type Transform = call(Point) Point

let origin: Point = Point{0.0, 0.0}
let heading: Direction = Direction.North
let names: Names = ["Ada"]
let transform: Transform = fn(point: Point) Point { return point }
let core = import("core:core")
transform(origin).x == origin.x && heading == Direction.North && names[0] == "Ada" && core.type(origin) == Point && core.type(heading) == Direction
`), true)
}

func TestTypeDeclarationContracts(t *testing.T) {
	for name, input := range map[string]string{
		"primitive and alias chain": `type Number = int
type Count = Number
let value: Count = 42
value`,
		"captured dependencies": `type Number = int
type Numbers = array[Number]
type Number = str
let values: Numbers = [42]
values[0]`,
		"local scope": `type Number = str
let produce = fn() int {
    type Number = int
    let value: Number = 42
    return value
}
let outer: Number = "still a string"
produce()`,
		"self reference and bound method": `type Point = struct { x: int, move: call(self: Point, amount: int) }
let move = fn(self: Point, amount: int) { self.x = self.x + amount }
let point = Point{40, move}
point.move(2)
point.x`,
		"embedding": `type Point = struct { x: int }
type Actor = struct { point :: Point }
Actor{Point{42}}.x`,
		"errors and signatures": `type Failure = struct { code: int }
type Read = call() int | Failure
let read: Read = fn() int | Failure { return Failure{42} }
try { read() } catch Failure err { err.code }`,
		"variadic contract": `type Sum = call(values: int...) int
let add = fn(left: int, right: int) int { return left + right }
let sum: Sum = fn(values: int...) int { return add(values) }
sum(20, 22)`,
		"nominal capture": `type Item = struct { value: int }
type Saved = Item
let original = Item{42}
type Item = struct { value: int }
let saved: Saved = original
saved.value`,
	} {
		t.Run(name, func(t *testing.T) {
			value := testEval(input)
			if err, ok := value.(*object.Error); ok {
				t.Fatal(err.Inspect())
			}
			testIntegerObject(t, value, 42)
		})
	}
}

func TestTypeDeclarationFailures(t *testing.T) {
	for _, test := range []struct{ input, message string }{
		{`type T = Missing`, `unknown type "Missing"`},
		{`type T = array[Missing]`, `unknown type "Missing"`},
		{`type T = call() Missing`, `unknown type "Missing"`},
		{`type T = T`, `unknown type "T"`},
		{"let value = 42\ntype T = value", "does not name a value type"},
		{"type T = struct { x: Missing }", `unknown type "Missing"`},
		{"type T = call() int | str", "must be a struct"},
		{"type Names = array[str]\nlet names: Names = [42]", "expected Names, got array"},
		{"type T = call(int) int\nlet f: T = fn(x: str) str { return x }", "expected T, got call"},
		{"type Number = int\nlet n: Number = 1\nn = \"bad\"", "expected Number, got str"},
		{"type Left = struct { x: int }\ntype Right = struct { x: int }\nlet p: Left = Right{1}", "expected Left, got Right"},
		{"type Left = enum { One }\ntype Right = enum { One }\nlet p: Left = Right.One", "expected Left, got Right"},
		{"type T = struct {}\nlet old = T{}\ntype T = struct {}\nlet p: T = old", "expected T, got T"},
	} {
		t.Run(test.input, func(t *testing.T) {
			value := testEval(test.input)
			err, ok := value.(*object.Error)
			if !ok || !strings.Contains(err.MessageText(), test.message) {
				t.Fatalf("result = %v, want error containing %q", value, test.message)
			}
		})
	}
}

func TestExportedTypeDeclarations(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	libraryPath, mainPath := filepath.Join(dir, "library.slv"), filepath.Join(dir, "main.slv")
	writeSilverFile(t, libraryPath, `export { Point, Direction, Names, Transform, Read }
type Point = struct { x: float, y: float }
type Direction = enum { North, South }
type Names = array[str]
type Transform = call(Point) Point
type Read = call() array[Point]
`)
	writeSilverFile(t, mainPath, `let library = import("fixture:library")
type Names = library.Names
let names: Names = ["Ada"]
let point: library.Point = library.Point{0.0, 0.0}
let heading: library.Direction = library.Direction.North
let transform: library.Transform = fn(p: library.Point) library.Point { return p }
let read: library.Read = fn() array[library.Point] { return [point] }
names[0] == "Ada" && heading == library.Direction.North && transform(point).x == read()[0].x
`)
	testBooleanObject(t, New().EvalFile(mainPath, object.NewEnvironment()), true)
}

func TestTypeInspectionMember(t *testing.T) {
	testBooleanObject(t, testEval(`let core = import("core:core")
let inspect_type = core.type
inspect_type(42) == int
`), true)
}
