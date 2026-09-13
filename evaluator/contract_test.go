package evaluator

import (
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestContractsRetainDeclaredTypes(t *testing.T) {
	for name, input := range map[string]string{
		"binding": `type A = struct { x: int }
type B = struct { x: int }
let T = A
let value: T = A{42}
T = B
value = value
value.x`,
		"parameter shadows type": `type Box = struct { x: int }
let f = fn(Box, value: Box) int {
    value = value
    return value.x
}
f(1, Box{42})`,
		"closure assignment": `type Box = struct { x: int }
let value: Box = Box{42}
let update = fn(Box) { value = value }
Box = 1
update(2)
value.x`,
		"initializer rebinds type": `let T = int
let initialize = fn() int {
    T = str
    return 42
}
let value: T = initialize()
value = value
value`,
		"function entry and return": `type Box = struct { x: int }
let saved = Box{42}
let f = fn(value: Box) Box {
    value = value
    return value
}
Box = 1
f(saved).x`,
		"return body rebinds type": `let T = int
let f = fn() T {
    T = str
    return 42
}
f()`,
		"destructuring": `type Box = struct { x: int }
type Args = struct { Box, value: Box }
let args = Args{1, Box{42}}
let f = fn(Box: int, value: Box) int {
    value = value
    return value.x
}
Box = False
f(args)`,
		"variadic": `type Box = struct { x: int }
let saved = Box{42}
let first = fn(value: Box) int { return value.x }
let f = fn(values: Box...) int { return first(values) }
Box = str
f(saved)`,
		"field construction and mutation": `type Box = struct { x: int }
type Holder = struct { value: Box }
let saved = Box{42}
Box = 1
let holder = Holder{saved}
holder.value = saved
holder.value.x`,
		"embedded field": `type Box = struct { x: int }
type Holder = struct { value :: Box }
let saved = Box{42}
Box = 1
let holder = Holder{saved}
holder.value = saved
holder.x`,
		"recursive fields and method receiver": `type Box = struct { children: array[Box], read: call(self: Box) int }
let read = fn(self: Box) int { return 42 }
let saved = Box{[], read}
Box = 1
saved.children = [saved]
saved.read = read
saved.read()`,
		"callable field alias": `type Read = call(self: any) int
type Box = struct { read: Read }
let read = fn(self) int { return 42 }
let box = Box{read}
Read = int
box.read = read
box.read()`,
		"nested callable contracts": `type Box = struct { x: int }
let value = Box{42}
let read = fn(values: array[Box]) array[Box] { return values }
let callbacks: array[call(array[Box]) array[Box]] = [read]
Box = 1
callbacks = callbacks
callbacks[0]([value])[0].x`,
		"declared error and propagation": `type Failure = struct { code: int }
let Original = Failure
let failure = Failure{42}
let inner = fn() | Failure { return failure }
let outer = fn() | Failure { inner() }
Failure = 1
try { outer() } catch Original err { err.code }`,
		"catch type and binding": `type Failure = struct { code: int }
let fail = fn() | Failure { return Failure{42} }
try { fail() } catch Failure Failure {
    Failure = Failure
    Failure.code
}`,
		"catch resolves before body": `type Failure = struct { code: int }
let failure = Failure{42}
let fail = fn() | Failure {
    Failure = 1
    return failure
}
try { fail() } catch Failure err {
    err = err
    err.code
}`,
		"fresh contracts per function evaluation": `type A = struct { x: int }
type B = struct { x: int }
let make = fn(T) call { return fn(value: T) T { return value } }
let a = make(A)
let b = make(B)
a(A{20}).x + b(B{22}).x`,
		"fresh contracts per struct evaluation": `type A = struct { x: int }
type B = struct { x: int }
let make = fn(T) any {
    type Holder = struct { value: T }
    return Holder
}
let HA = make(A)
let HB = make(B)
HA{A{20}}.value.x + HB{B{22}}.value.x`,
		"default map factory return": `type Box = struct { x: int }
let saved = Box{42}
let factory = fn() Box { return saved }
Box = 1
let values = import("collections").defaultmap(factory)
values[0].x`,
		"native nominal return": `let time = import("time")
let now: call() time.Time = time.now
time = 1
now = now
now()
42`,
		"native error shadowing": `let string = import("string")
let BuiltinError = ValueError
let codepoint: call(str) int | ValueError = string.codepoint
let ValueError = 1
codepoint = codepoint
try { codepoint("") } catch BuiltinError err { 42 }`,
	} {
		t.Run(name, func(t *testing.T) {
			value := evalInput(t, New(), object.NewEnvironment(), input)
			if err, ok := value.(*object.Error); ok {
				t.Fatal(err.Inspect())
			}
			testIntegerObject(t, value, 42)
		})
	}
}

func TestContractsRejectReboundTypes(t *testing.T) {
	for _, test := range []struct{ name, input, message string }{
		{"binding", `type A = struct { x: int }
type B = struct { x: int }
let T = A
let value: T = A{1}
T = B
value = B{2}`, `binding "value": expected T, got B`},
		{"parameter entry", `type A = struct { x: int }
type B = struct { x: int }
let T = A
let f = fn(value: T) {}
T = B
f(B{2})`, `parameter "value": expected T, got B`},
		{"parameter assignment", `type A = struct { x: int }
type B = struct { x: int }
let f = fn(A, value: A) { value = B{2} }
f(B, A{1})`, `binding "value": expected A, got B`},
		{"return", `type A = struct { x: int }
type B = struct { x: int }
let T = A
let f = fn() T { return B{2} }
T = B
f()`, `return value of "f": expected T, got B`},
		{"field construction", `type A = struct { x: int }
type B = struct { x: int }
let T = A
type Holder = struct { value: T }
T = B
Holder{B{2}}`, `field "Holder.value": expected T, got B`},
		{"field assignment", `type A = struct { x: int }
type B = struct { x: int }
let T = A
type Holder = struct { value: T }
let holder = Holder{A{1}}
T = B
holder.value = B{2}`, `field "Holder.value": expected T, got B`},
		{"enum", `type A = enum { One }
type B = enum { One }
let T = A
let value: T = A.One
T = B
value = B.One`, `binding "value": expected T, got B`},
		{"alias", `type T = int
let value: T = 1
type T = str
value = "bad"`, `binding "value": expected T, got str`},
		{"nested array", `let T = int
let value: array[array[T]] = [[1]]
T = str
value = [["bad"]]`, `binding "value": expected array[array[T]], got array`},
		{"callable parameter identity", `type A = struct { x: int }
type B = struct { x: int }
let T = A
let f = fn(value: T) {}
T = B
let callback: call(B) = f`, `binding "callback": expected call(B), got call`},
		{"callable return identity", `type A = struct { x: int }
type B = struct { x: int }
let T = A
let f = fn() T {}
T = B
let callback: call() B = f`, `binding "callback": expected call() B, got call`},
		{"callable error identity", `type A = struct { code: int }
type B = struct { code: int }
let T = A
let f = fn() | T {}
T = B
let callback: call() | B = f`, `binding "callback": expected call() | B, got call`},
		{"error return", `type A = struct { code: int }
type B = struct { code: int }
let T = A
let f = fn() | T { return B{2} }
T = B
f()`, `return value of "f": expected null | T, got B`},
		{"native error identity", `let codepoint = import("string").codepoint
type ValueError = struct { message: str }
let f: call(str) int | ValueError = codepoint`, `binding "f": expected call(str) int | ValueError, got call`},
	} {
		t.Run(test.name, func(t *testing.T) {
			value := evalInput(t, New(), object.NewEnvironment(), test.input)
			err, ok := value.(*object.Error)
			if !ok {
				t.Fatalf("result = %v, want TypeError containing %q", value, test.message)
			}
			definition, _ := object.BuiltinStructDefinitionByName("TypeError")
			if err.Value.Struct != definition || !strings.Contains(err.MessageText(), test.message) {
				t.Fatalf("error = %s, want TypeError containing %q", err.Inspect(), test.message)
			}
		})
	}
}

func TestQualifiedContractsSurviveModuleRebinding(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "box.slv"), `export { Box }
type Box = struct { x: int }`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `let library = import("./box.slv")
let value: library.Box = library.Box{42}
type Boxes = array[library.Box]
type Holder = struct { value: library.Box }
let f = fn(value: library.Box) library.Box {
    value = value
    return value
}
library = 1
value = value
let values: Boxes = [value]
Holder{f(values[0])}.value.x`)
	for range 2 { // Resolved contracts belong to evaluations, not cached ASTs.
		value := New().EvalFile(mainPath, object.NewEnvironment())
		if err, ok := value.(*object.Error); ok {
			t.Fatal(err.Inspect())
		}
		testIntegerObject(t, value, 42)
	}
}
