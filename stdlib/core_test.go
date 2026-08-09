package stdlib_test

import (
	"silver/object"
	"testing"
)

const coreImport = "let core = import(\"core\")\n"

func TestRangeBuiltin(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `core.range(1, 5)`, want: `[1, 2, 3, 4]`},
		{input: `core.range(-2, 2)`, want: `[-2, -1, 0, 1]`},
		{input: `core.range(3, 3)`, want: `[]`},
		{input: `core.range(3, 1)`, want: `[]`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(coreImport + tt.input)
			if got := result.Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLenBuiltinSupportsMaps(t *testing.T) {
	testIntegerObject(t, testEval(coreImport+`core.len({"one": 1, "two": 2})`), 2)
	testIntegerObject(t, testEval(coreImport+`core.len({})`), 0)
}

func TestCoreBuiltinErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `core.range(1.0, 2)`, message: "argument 1 to `range` must be INTEGER, got FLOAT"},
		{input: `core.range(1, 2.0)`, message: "argument 2 to `range` must be INTEGER, got FLOAT"},
		{input: `core.range(1)`, message: "wrong number of arguments. got=1, want=2"},
		{input: `core.range(0, 1000001)`, message: "range contains too many elements: 1000001 (maximum 1000000)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(coreImport + tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}
}
