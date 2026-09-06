package evaluator

import (
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestTypeAliasPipeline(t *testing.T) {
	testIntegerObject(t, testEval(`
struct Token { text: str }
struct Node { value: int }
let lex = <call(str) array[Token]>
let parse = <call(array[Token]) array[Node]>
let eval = <call(array[Node]) int>
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
		"primitive chain": `let Number = <int>
let Count = <Number>
let value: Count = 42
value`,
		"primitive value": `let Number = int
let value: Number = 42
value`,
		"return alias": `let Number = <int>
let produce = fn() Number { return 42 }
let callback: call() int = produce
callback()`,
		"captured binding": `let Number = <int>
let Numbers = <array[Number]>
Number = <str>
let values: Numbers = [42]
values[0]`,
		"self rebind captures previous": `let Value = <int>
Value = <array[Value]>
let values: Value = [42]
values[0]`,
		"nested arrays": `let values: array[array[int]] = [[], [42]]
values[1][0]`,
		"array of callables": `let Operation = <call(int) int>
let operations: array[Operation] = [fn(value: int) int { return value }]
operations[0](42)`,
		"variadic": `let Sum = <call(values: int...) int>
let add = fn(left: int, right: int) int { return left + right }
let sum: Sum = fn(values: int...) int { return add(values) }
sum(20, 22)`,
		"method receiver": `struct Scale { x: int }
let Grow = <call(scale: Scale, amount: int)>
struct Transform { scale: Scale, grow: Grow }
let grow = fn(scale: Scale, amount: int) { scale.x = scale.x + amount }
let actor = Transform{Scale{40}, grow}
actor.grow(2)
actor.scale.x`,
		"embedded struct": `struct Item { value: int }
let ItemType = <Item>
struct Box { item :: ItemType }
Box{Item{42}}.value`,
		"destructuring": `let Number = <int>
let Reader = <call() Number>
struct Provider { read: Reader }
let use = fn(read: Reader) int { return read() }
use(Provider{fn() int { return 42 }})`,
		"nominal capture": `struct Item { value: int }
let ItemType = <Item>
let item = Item{42}
struct Item { value: int }
let saved: ItemType = item
saved.value`,
		"errors": `struct Failure { value: int }
let ErrorType = <Failure>
let Reader = <call() int | ErrorType>
let read: Reader = fn() int | Failure { return Failure{42} }
try { read() } catch ErrorType err { err.value }`,
		"null return alias": `let Nothing = <null>
let Action = <call() Nothing>
let run: Action = fn() {}
run()
42`,
		"native callable": `let Codepoint = <call(str) int | ValueError>
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
		{`let MissingType = <Missing>`, `unknown type "Missing"`},
		{`let T = <array[Missing]>`, `unknown type "Missing"`},
		{`let T = <call() Missing>`, `unknown type "Missing"`},
		{`let T = <T>`, `unknown type "T"`},
		{"let value = 1\nlet T = <value>", `does not name a value type`},
		{"let T = <call(int) str>\nlet f: T = fn(x: str) str { return x }", `expected T, got call`},
		{"let T = <int>\nlet value: T = 1\nvalue = \"bad\"", `expected T, got str`},
		{`let values: array[int] = [1, "bad"]`, `expected array[int], got array`},
		{`let values: array[array[int]] = [[1], ["bad"]]`, `expected array[array[int]], got array`},
		{"let values: array[int] = []\nvalues = [False]", `expected array[int], got array`},
		{"let use = fn(values: array[int]) {}\nuse([\"bad\"])", `parameter "values"`},
		{"let produce = fn() array[int] { return [\"bad\"] }\nproduce()", `return value of "produce"`},
		{"struct Box { values: array[int] }\nBox{[\"bad\"]}", `field "Box.values"`},
		{"struct Box { values: array[int] }\nlet box = Box{[]}\nbox.values = [\"bad\"]", `field "Box.values"`},
		{`let T = <call() int | array[str]>`, `must be a struct`},
		{"let T = <str>\nlet U = <call() int | T>", `must be a struct`},
		{"enum Left { One }\nenum Right { One }\nlet Values = <array[Left]>\nlet values: Values = [Right.One]", `expected Values, got array`},
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
			value := testEval("let Contract = <" + test.expected + ">\nlet f: Contract = " + test.actual + " {}\nf")
			_, accepted := value.(*object.Function)
			if accepted != test.accepted {
				t.Fatalf("accepted = %t, want %t: %v", accepted, test.accepted, value)
			}
		})
	}
}

func TestExportedTypeAlias(t *testing.T) {
	dir := t.TempDir()
	libraryPath, mainPath := filepath.Join(dir, "library.slv"), filepath.Join(dir, "main.slv")
	writeSilverFile(t, libraryPath, `export { Read, make }
struct Item { value: int }
let Items = <array[Item]>
let Read = <call() Items>
let make = fn() Items { return [Item{42}] }
`)
	writeSilverFile(t, mainPath, `let library = import("./library.slv")
let Reader = <library.Read>
let read: Reader = library.make
read()[0].value
`)
	for range 2 { // Exercise both source parsing and cached AST evaluation.
		testIntegerObject(t, New().EvalFile(mainPath, object.NewEnvironment()), 42)
	}
}

func TestTypeAliasIsATypeValue(t *testing.T) {
	testBooleanObject(t, testEval(`let Alias = <array[str]>
import("core").type(Alias) == Alias`), true)
}
