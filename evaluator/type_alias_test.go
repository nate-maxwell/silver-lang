package evaluator

import (
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestTypeAliasPipeline(t *testing.T) {
	testIntegerObject(t, testEval(`
type Token = struct { text: str }
type Node = struct { value: int }
type lex = call(str) array[Token]
type parse = call(array[Token]) array[Node]
type eval = call(array[Node]) int
let parse_program = fn(l: lex, p: parse, e: eval) int {
	return e(p(l("42")))
}
let lexer = fn(source: str) array[Token] { return [Token{source}] }
let parser = fn(tokens: array[Token]) array[Node] { return [Node{42}] }
let evaluator = fn(nodes: array[Node]) int { return nodes[0].value }
parse_program(lexer, parser, evaluator)
`), 42)
}

func TestTypeAliasContracts(t *testing.T) {
	for name, input := range map[string]string{
		"primitive chain": `type Number = int
type Count = Number
let value: Count = 42
value`,
		"primitive value": `let Number = int
let value: Number = 42
value`,
		"return alias": `type Number = int
let produce = fn() Number { return 42 }
let callback: call() int = produce
callback()`,
		"captured binding": `type Number = int
type Numbers = array[Number]
type Number = str
let values: Numbers = [42]
values[0]`,
		"self rebind captures previous": `type Value = int
type Value = array[Value]
let values: Value = [42]
values[0]`,
		"nested arrays": `let values: array[array[int]] = [[], [42]]
values[1][0]`,
		"array of callables": `type Operation = call(int) int
let operations: array[Operation] = [fn(value: int) int { return value }]
operations[0](42)`,
		"variadic": `type Sum = call(values: int...) int
let add = fn(left: int, right: int) int { return left + right }
let sum: Sum = fn(values: int...) int { return add(values) }
sum(20, 22)`,
		"method receiver": `type Scale = struct { x: int }
type Grow = call(scale: Scale, amount: int)
type Transform = struct { scale: Scale, grow: Grow }
let grow = fn(scale: Scale, amount: int) { scale.x = scale.x + amount }
let actor = Transform{Scale{40}, grow}
actor.grow(2)
actor.scale.x`,
		"embedded struct": `type Item = struct { value: int }
type ItemType = Item
type Box = struct { item :: ItemType }
Box{Item{42}}.value`,
		"destructuring": `type Number = int
type Reader = call() Number
type Provider = struct { read: Reader }
let use = fn(read: Reader) int { return read() }
use(Provider{fn() int { return 42 }})`,
		"nominal capture": `type Item = struct { value: int }
type ItemType = Item
let item = Item{42}
type Item = struct { value: int }
let saved: ItemType = item
saved.value`,
		"errors": `type Failure = struct { value: int }
type ErrorType = Failure
type Reader = call() int | ErrorType
let read: Reader = fn() int | Failure { return Failure{42} }
try { read() } catch ErrorType err { err.value }`,
		"null return alias": `type Nothing = null
type Action = call() Nothing
let run: Action = fn() {}
run()
42`,
		"native callable": `type Codepoint = call(str) int | ValueError
let codepoint: Codepoint = import("string").codepoint
codepoint("*")`,
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

func TestTypeAliasAndArrayContractFailures(t *testing.T) {
	for _, test := range []struct{ input, message string }{
		{`type MissingType = Missing`, `unknown type "Missing"`},
		{`type T = array[Missing]`, `unknown type "Missing"`},
		{`type T = call() Missing`, `unknown type "Missing"`},
		{`type T = T`, `unknown type "T"`},
		{"let value = 1\ntype T = value", `does not name a value type`},
		{"type T = call(int) str\nlet f: T = fn(x: str) str { return x }", `expected T, got call`},
		{"type T = int\nlet value: T = 1\nvalue = \"bad\"", `expected T, got str`},
		{`let values: array[int] = [1, "bad"]`, `expected array[int], got array`},
		{`let values: array[array[int]] = [[1], ["bad"]]`, `expected array[array[int]], got array`},
		{"let values: array[int] = []\nvalues = [False]", `expected array[int], got array`},
		{"let use = fn(values: array[int]) {}\nuse([\"bad\"])", `parameter "values"`},
		{"let produce = fn() array[int] { return [\"bad\"] }\nproduce()", `return value of "produce"`},
		{"type Box = struct { values: array[int] }\nBox{[\"bad\"]}", `field "Box.values"`},
		{"type Box = struct { values: array[int] }\nlet box = Box{[]}\nbox.values = [\"bad\"]", `field "Box.values"`},
		{`type T = call() int | array[str]`, `must be a struct`},
		{"type T = str\ntype U = call() int | T", `must be a struct`},
		{"type Left = enum { One }\ntype Right = enum { One }\ntype Values = array[Left]\nlet values: Values = [Right.One]", `expected Values, got array`},
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

func TestArrayCallableCompatibility(t *testing.T) {
	for _, test := range []struct {
		expected, actual string
		accepted         bool
	}{
		{"call(array[int]) array[str]", "fn(x: array[int]) array[str]", true},
		{"call(array[int]) array[str]", "fn(x: array) array[str]", true},
		{"call(array[int]) array[str]", "fn(x: array[any]) array[str]", true},
		{"call(array[int]) array", "fn(x: array[int]) array[str]", true},
		{"call(array[int]) array[any]", "fn(x: array[int]) array[str]", true},
		{"call(array) array[str]", "fn(x: array[int]) array[str]", false},
		{"call(array[int]) array[str]", "fn(x: array[str]) array[str]", false},
		{"call(array[int]) array[str]", "fn(x: array[int]) array", false},
		{"call(array[int]) array[str]", "fn(x: array[int]) array[int]", false},
	} {
		t.Run(test.expected+" / "+test.actual, func(t *testing.T) {
			value := testEval("type Contract = " + test.expected + "\nlet f: Contract = " + test.actual + " {}\nf")
			_, accepted := value.(*object.Function)
			if accepted != test.accepted {
				t.Fatalf("accepted = %t, want %t: %v", accepted, test.accepted, value)
			}
		})
	}
}

func TestExportedTypeAlias(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	libraryPath, mainPath := filepath.Join(dir, "library.slv"), filepath.Join(dir, "main.slv")
	writeSilverFile(t, libraryPath, `export { Read, make }
type Item = struct { value: int }
type Items = array[Item]
type Read = call() Items
let make = fn() Items { return [Item{42}] }
`)
	writeSilverFile(t, mainPath, `let library = import("./library.slv")
type Reader = library.Read
let read: Reader = library.make
read()[0].value
`)
	testIntegerObject(t, New().EvalFile(mainPath, object.NewEnvironment()), 42)
}

func TestTypeAliasIsATypeValue(t *testing.T) {
	testBooleanObject(t, testEval(`type Alias = array[str]
import("core").type(Alias) == Alias`), true)
}
