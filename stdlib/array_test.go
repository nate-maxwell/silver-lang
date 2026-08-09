package stdlib_test

import (
	"silver/object"
	"testing"
)

const arrayImport = "let arrays = import(\"array\")\n"

func TestArrayModuleFunctions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "append", input: `arrays.append([1, 2], 3)`, want: `[1, 2, 3]`},
		{name: "first", input: `arrays.first([1, 2])`, want: `1`},
		{name: "last", input: `arrays.last([1, 2])`, want: `2`},
		{name: "remove", input: `arrays.remove([1, 2, 3], 1)`, want: `[1, 3]`},
		{name: "rest", input: `arrays.rest([1, 2, 3])`, want: `[2, 3]`},
		{name: "reverse", input: `arrays.reverse([1, 2, 3])`, want: `[3, 2, 1]`},
		{name: "numeric sort", input: `arrays.sort([3, 1.5, 2, -1])`, want: `[-1, 1.5, 2, 3]`},
		{name: "string sort", input: `arrays.sort(["pear", "apple", "orange"])`, want: `[apple, orange, pear]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := testEval(arrayImport + tt.input)
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
		{input: `arrays.contains([1, 2, 3], 2)`, want: true},
		{input: `arrays.contains([1.0], 1)`, want: true},
		{input: `arrays.contains(["silver"], "sil" + "ver")`, want: true},
		{input: `arrays.contains([1, 2, 3], 4)`, want: false},
		{input: `let nested = [1]
arrays.contains([nested], nested)`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(arrayImport+tt.input), tt.want)
		})
	}
}

func TestArrayTransformationsDoNotMutateInput(t *testing.T) {
	result := testEval(arrayImport + `let values = [3, 1, 2]
let removed = arrays.remove(values, 0)
let reversed = arrays.reverse(values)
let sorted = arrays.sort(values)
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
		{name: "remove index type", input: `arrays.remove([1], "0")`, message: "index argument to `remove` must be INTEGER, got STRING"},
		{name: "sort unsupported type", input: `arrays.sort([True])`, message: "argument to `sort` contains unsortable type BOOLEAN"},
		{name: "sort mixed types", input: `arrays.sort([1, "two"])`, message: "argument to `sort` must contain only numbers or only strings, got INTEGER and STRING"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := testEval(arrayImport + tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}

	testNullObject(t, testEval(arrayImport+`arrays.remove([1], -1)`))
	testNullObject(t, testEval(arrayImport+`arrays.remove([1], 1)`))
}

func TestArrayFunctionsRequireImport(t *testing.T) {
	for _, input := range []string{`append([1], 2)`, `[1].append(2)`} {
		if _, ok := testEval(input).(*object.Error); !ok {
			t.Fatalf("%s did not require the array module", input)
		}
	}
}
