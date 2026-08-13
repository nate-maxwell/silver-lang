package stdlib_test

import (
	"silver/object"
	"strconv"
	"strings"
	"testing"
)

const randomImport = `let random = import("random")
`

func TestRandomAndSeed(t *testing.T) {
	result, ok := testEval(randomImport + `random.random()`).(*object.Float)
	if !ok || result.Value < 0 || result.Value >= 1 {
		t.Fatalf("random result is %T (%v), want a float in [0, 1)", result, result)
	}

	sequence := `random.seed(12345)
[random.random(), random.random(), random.random()]`
	first := testEval(randomImport + sequence).(*object.Array)
	second := testEval(randomImport + sequence).(*object.Array)
	for index := range first.Elements {
		left := first.Elements[index].(*object.Float).Value
		right := second.Elements[index].(*object.Float).Value
		if left != right {
			t.Fatalf("seeded values differ at index %d: %g != %g", index, left, right)
		}
	}

	testNullObject(t, testEval(randomImport+`random.seed(1)`))
}

func TestRandomIntInclusive(t *testing.T) {
	testIntegerObject(t, testEval(randomImport+`random.randint(7, 7)`), 7)

	result := testEval(randomImport + `random.seed(99)
random.randint(-9223372036854775807 - 1, 9223372036854775807)`)
	if _, ok := result.(*object.Integer); !ok {
		t.Fatalf("full-range randint is %T (%v), want integer", result, result)
	}

	for seed := int64(0); seed < 100; seed++ {
		input := randomImport + `random.seed(` + strconv.FormatInt(seed, 10) + `)
random.randint(-3, 4)`
		value := testEval(input).(*object.Integer).Value
		if value < -3 || value > 4 {
			t.Fatalf("randint returned %d outside [-3, 4]", value)
		}
	}
}

func TestRandomCollectionOperations(t *testing.T) {
	element := testEval(randomImport + `random.seed(5)
random.randelem(["red", "green", "blue"])`)
	if value, ok := element.(*object.String); !ok || value.Value != "red" && value.Value != "green" && value.Value != "blue" {
		t.Fatalf("randelem result is %T (%v), want an input element", element, element)
	}

	key := testEval(randomImport + `random.seed(5)
random.randkey({"red": 1, "green": 2, "blue": 3})`)
	if value, ok := key.(*object.String); !ok || value.Value != "red" && value.Value != "green" && value.Value != "blue" {
		t.Fatalf("randkey result is %T (%v), want an input key", key, key)
	}

	shuffled := testEval(randomImport + `let values = [1, 2, 3, 4, 5]
random.seed(5)
random.shuffle(values)
values`).(*object.Array)
	seen := make(map[int64]bool, len(shuffled.Elements))
	for _, element := range shuffled.Elements {
		seen[element.(*object.Integer).Value] = true
	}
	for value := int64(1); value <= 5; value++ {
		if !seen[value] {
			t.Fatalf("shuffle result %s is missing %d", shuffled.Inspect(), value)
		}
	}

	testNullObject(t, testEval(randomImport+`let values = [1, 2, 3]
random.shuffle(values)`))
}

func TestRandomErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `random.seed(1.5)`, message: "argument 1 to `seed` must be INTEGER, got FLOAT"},
		{input: `random.randint(1.5, 2)`, message: "argument 1 to `randint` must be INTEGER, got FLOAT"},
		{input: `random.randint(1, 2.5)`, message: "argument 2 to `randint` must be INTEGER, got FLOAT"},
		{input: `random.randint(2, 1)`, message: "lower bound to `randint` must not exceed upper bound"},
		{input: `random.randelem({})`, message: "argument to `randelem` must be ARRAY, got MAP"},
		{input: `random.randelem([])`, message: "cannot choose an element from an empty array"},
		{input: `random.randkey([])`, message: "argument to `randkey` must be MAP, got ARRAY"},
		{input: `random.randkey({})`, message: "cannot choose a key from an empty map"},
		{input: `random.shuffle({})`, message: "argument to `shuffle` must be ARRAY, got MAP"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := testEval(randomImport + test.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T (%v), want error", result, result)
			}
			if !strings.Contains(result.MessageText(), test.message) {
				t.Fatalf("error is %q, want it to contain %q", result.MessageText(), test.message)
			}
		})
	}
}
