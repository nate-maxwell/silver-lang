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

func TestStringToNumberConversions(t *testing.T) {
	integerTests := []struct {
		input string
		want  int64
	}{
		{`strings.to_int("42")`, 42},
		{`strings.to_int("-42")`, -42},
		{`strings.to_int("+42")`, 42},
		{`strings.to_int("-9223372036854775808")`, -9223372036854775808},
	}
	for _, test := range integerTests {
		t.Run(test.input, func(t *testing.T) {
			testIntegerObject(t, testEval(stringImport+test.input), test.want)
		})
	}

	floatTests := []struct {
		input string
		want  float64
	}{
		{`strings.to_float("42")`, 42},
		{`strings.to_float("-2.5")`, -2.5},
		{`strings.to_float("1.25e2")`, 125},
	}
	for _, test := range floatTests {
		t.Run(test.input, func(t *testing.T) {
			testFloatObject(t, testEval(stringImport+test.input), test.want)
		})
	}
}

func TestStringToBoolConversion(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`strings.to_bool("true")`, true},
		{`strings.to_bool("True")`, true},
		{`strings.to_bool("FALSE")`, false},
		{`strings.to_bool("false")`, false},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			testBooleanObject(t, testEval(stringImport+test.input), test.want)
		})
	}
}

func TestStringFromPrimitiveConversions(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`strings.from_int(42)`, "42"},
		{`strings.from_int(-42)`, "-42"},
		{`strings.from_float(42.0)`, "42.0"},
		{`strings.from_float(-2.5)`, "-2.5"},
		{`strings.from_bool(True)`, "true"},
		{`strings.from_bool(False)`, "false"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, ok := testEval(stringImport + test.input).(*object.String)
			if !ok || result.Value != test.want {
				t.Fatalf("result is %#v, want %q", result, test.want)
			}
		})
	}
}

func TestStringConversionSignatures(t *testing.T) {
	input := stringImport + coreImport + `let to_int: call(value: str) int | ValueError = strings.to_int
let from_int: call(value: int) str = strings.from_int
let to_float: call(value: str) float | ValueError = strings.to_float
let from_float: call(value: float) str = strings.from_float
let to_bool: call(value: str) bool | ValueError = strings.to_bool
let from_bool: call(value: bool) str = strings.from_bool
core.type(from_bool) == call`
	testBooleanObject(t, testEval(input), true)
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
		{input: `strings.to_int("12.5")`, message: `could not convert "12.5" to int`},
		{input: `strings.to_int("9223372036854775808")`, message: `could not convert "9223372036854775808" to int`},
		{input: `strings.to_float("1.2.3")`, message: `could not convert "1.2.3" to float`},
		{input: `strings.to_bool("yes")`, message: `could not convert "yes" to bool`},
		{input: `strings.to_int(1)`, message: "argument 1 to `to_int` must be STRING, got INTEGER"},
		{input: `strings.from_int("1")`, message: "argument 1 to `from_int` must be INTEGER, got STRING"},
		{input: `strings.to_float(1.0)`, message: "argument 1 to `to_float` must be STRING, got FLOAT"},
		{input: `strings.from_float(1)`, message: "argument 1 to `from_float` must be FLOAT, got INTEGER"},
		{input: `strings.to_bool(True)`, message: "argument 1 to `to_bool` must be STRING, got BOOLEAN"},
		{input: `strings.from_bool("true")`, message: "argument 1 to `from_bool` must be BOOLEAN, got STRING"},
		{input: `strings.from_int()`, message: "wrong number of arguments. got=0, want=1"},
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
