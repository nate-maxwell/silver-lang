package stdlib_test

import (
	"silver/object"
	"testing"
)

const stringImport = `let strings = import("string")
`

func TestStringTransformFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `strings.capitalize("sILVER LANGUAGE")`, want: "Silver language"},
		{input: `strings.lower("Silver")`, want: "silver"},
		{input: `strings.lstrip("  silver  ")`, want: "silver  "},
		{input: `strings.removeprefix("silver-lang", "silver-")`, want: "lang"},
		{input: `strings.removesuffix("main.slv", ".slv")`, want: "main"},
		{input: `strings.repeat("go", 3)`, want: "gogogo"},
		{input: `strings.replace("one fish two fish", "fish", "bird")`, want: "one bird two bird"},
		{input: `strings.reverse("silver")`, want: "revlis"},
		{input: `strings.rstrip("  silver  ")`, want: "  silver"},
		{input: `strings.strip("  silver  ")`, want: "silver"},
		{input: `strings.swapcase("Silver 2")`, want: "sILVER 2"},
		{input: `strings.title("the silver language")`, want: "The Silver Language"},
		{input: `strings.upper("Silver")`, want: "SILVER"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(stringImport + tt.input).(*object.String)
			if !ok {
				t.Fatalf("result is %T (%v), want *object.String", result, result)
			}
			if result.Value != tt.want {
				t.Fatalf("result is %q, want %q", result.Value, tt.want)
			}
		})
	}
}

func TestStringSearchFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{input: `strings.compare("apple", "banana")`, want: -1},
		{input: `strings.compare("same", "same")`, want: 0},
		{input: `strings.count("banana", "an")`, want: 2},
		{input: `strings.find("banana", "na")`, want: 2},
		{input: `strings.find("banana", "pear")`, want: -1},
		{input: `strings.rfind("banana", "na")`, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(stringImport+tt.input), tt.want)
		})
	}
}

func TestStringPredicateFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: `strings.contains("silver", "ilve")`, want: true},
		{input: `strings.endswith("main.slv", ".slv")`, want: true},
		{input: `strings.equal_fold("Silver", "sILVer")`, want: true},
		{input: `strings.startswith("silver", "sil")`, want: true},
		{input: `strings.isalnum("Silver2")`, want: true},
		{input: `strings.isalpha("Silver")`, want: true},
		{input: `strings.isascii("Silver 2")`, want: true},
		{input: `strings.isdecimal("2026")`, want: true},
		{input: `strings.isdigit("2026")`, want: true},
		{input: `strings.islower("silver 2")`, want: true},
		{input: `strings.isnumeric("2026")`, want: true},
		{input: `strings.isprintable("Silver 2")`, want: true},
		{input: `strings.isspace(" \t\n")`, want: true},
		{input: `strings.istitle("Silver Language")`, want: true},
		{input: `strings.isupper("SILVER 2")`, want: true},
		{input: `strings.isalpha("Silver2")`, want: false},
		{input: `strings.islower("SILVER")`, want: false},
		{input: `strings.isprintable("line\n")`, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(stringImport+tt.input), tt.want)
		})
	}
}

func TestStringArrayFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `strings.chars("go")`, want: `[g, o]`},
		{input: `strings.fields(" one\ttwo  three ")`, want: `[one, two, three]`},
		{input: `strings.split("one,two,three", ",")`, want: `[one, two, three]`},
		{input: `strings.splitlines("one\ntwo\r\nthree")`, want: `[one, two, three]`},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(stringImport + tt.input)
			if got := result.Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}

	joined, ok := testEval(stringImport + `strings.join(["one", "two", "three"], " | ")`).(*object.String)
	if !ok || joined.Value != "one | two | three" {
		t.Fatalf("join result is %T (%v), want %q", joined, joined, "one | two | three")
	}
}

func TestStringFunctionErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `strings.lower(1)`, message: "argument 1 to `lower` must be STRING, got INTEGER"},
		{input: `strings.contains("one", 1)`, message: "argument 2 to `contains` must be STRING, got INTEGER"},
		{input: `strings.join("one", ",")`, message: "argument to `join` must be ARRAY, got STRING"},
		{input: `strings.join(["one", 2], ",")`, message: "element 2 of argument 1 to `join` must be STRING, got INTEGER"},
		{input: `strings.repeat("one", 1.5)`, message: "argument 2 to `repeat` must be INTEGER, got FLOAT"},
		{input: `strings.repeat("one", -1)`, message: "argument 2 to `repeat` must be nonnegative"},
		{input: `strings.repeat("one", 1000000)`, message: "result of `repeat` exceeds 1000000 bytes"},
		{input: `strings.split("one", "")`, message: "separator argument to `split` must not be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(stringImport + tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}
}
