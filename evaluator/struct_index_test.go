package evaluator

import (
	"silver/object"
	"strings"
	"testing"
)

const indexStructPrelude = `
struct Box {
	values: map
	get_item: call(self: Box, key: str) int
	set_item: call(self: Box, key: str, value: int)
}
let get_item = fn(self: Box, key: str) int { self.values[key] }
let set_item = fn(self: Box, key: str, value: int) { self.values[key] = value }
let box = Box{{"answer": 1}, get_item, set_item}
`

func TestStructIndexMethods(t *testing.T) {
	evaluated := testEval(indexStructPrelude + `
box["answer"] = 42
box["answer"]
`)
	testIntegerObject(t, evaluated, 42)
}

func TestStructIndexRequiresMappedMethod(t *testing.T) {
	for _, input := range []string{
		`struct Box {}
Box{}[0]`,
		`struct Box {}
let box = Box{}
box[0] = 1`,
	} {
		evaluated := testEval(input)
		err, ok := evaluated.(*object.Error)
		if !ok || !strings.Contains(err.MessageText(), "missing method") {
			t.Fatalf("result is %#v, want missing index method error", evaluated)
		}
	}
}

func TestStructIndexMethodMustBeCallable(t *testing.T) {
	evaluated := testEval(`struct Invalid { get_item: int }
Invalid{1}[0]`)
	err, ok := evaluated.(*object.Error)
	if !ok || !strings.Contains(err.MessageText(), `index method "get_item" on struct "Invalid" is not callable`) {
		t.Fatalf("result is %#v, want non-callable index method error", evaluated)
	}
}
