package evaluator

import (
	"silver/object"
	"testing"
)

func TestArrayBuiltinMethods(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "append", input: `[1, 2].append(3)`, want: `[1, 2, 3]`},
		{name: "first", input: `[1, 2].first()`, want: `1`},
		{name: "last", input: `[1, 2].last()`, want: `2`},
		{name: "remove", input: `[1, 2, 3].remove(1)`, want: `[1, 3]`},
		{name: "rest", input: `[1, 2, 3].rest()`, want: `[2, 3]`},
		{name: "reverse", input: `[1, 2, 3].reverse()`, want: `[3, 2, 1]`},
		{name: "numeric sort", input: `[3, 1.5, 2, -1].sort()`, want: `[-1, 1.5, 2, 3]`},
		{name: "string sort", input: `["pear", "apple", "orange"].sort()`, want: `[apple, orange, pear]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(tt.input)
			if got := result.Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}
}

func TestArrayContainsUsesScalarEquality(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: `[1, 2, 3].contains(2)`, want: true},
		{input: `[1.0].contains(1)`, want: true},
		{input: `["silver"].contains("sil" + "ver")`, want: true},
		{input: `[1, 2, 3].contains(4)`, want: false},
		{input: `let nested = [1]
[nested].contains(nested)`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestArrayTransformationsDoNotMutateInput(t *testing.T) {
	result := testEval(`let values = [3, 1, 2]
let removed = values.remove(0)
let reversed = values.reverse()
let sorted = values.sort()
[values, removed, reversed, sorted]`)

	if got, want := result.Inspect(), `[[3, 1, 2], [1, 2], [2, 1, 3], [1, 2, 3]]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestArrayBuiltinBoundaryAndTypeErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
	}{
		{name: "remove index type", input: `[1].remove("0")`, message: "index argument to `remove` must be INTEGER, got STRING"},
		{name: "sort unsupported type", input: `[True].sort()`, message: "argument to `sort` contains unsortable type BOOLEAN"},
		{name: "sort mixed types", input: `[1, "two"].sort()`, message: "argument to `sort` must contain only numbers or only strings, got INTEGER and STRING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := testEval(tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.Message != tt.message {
				t.Fatalf("error is %q, want %q", result.Message, tt.message)
			}
		})
	}

	testNullObject(t, testEval(`[1].remove(-1)`))
	testNullObject(t, testEval(`[1].remove(1)`))
}

func TestNewArrayBuiltinsRemainCallableGlobally(t *testing.T) {
	testBooleanObject(t, testEval(`contains([1, 2], 2)`), true)
	if got, want := testEval(`remove([1, 2, 3], 1)`).Inspect(), `[1, 3]`; got != want {
		t.Fatalf("remove result is %q, want %q", got, want)
	}
	if got, want := testEval(`reverse([1, 2, 3])`).Inspect(), `[3, 2, 1]`; got != want {
		t.Fatalf("reverse result is %q, want %q", got, want)
	}
	if got, want := testEval(`sort([3, 1, 2])`).Inspect(), `[1, 2, 3]`; got != want {
		t.Fatalf("sort result is %q, want %q", got, want)
	}
}
