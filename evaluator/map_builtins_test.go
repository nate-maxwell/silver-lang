package evaluator

import (
	"silver/object"
	"testing"
)

func TestMapBuiltinMethods(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "get", input: `{"one": 1}.get("one")`, want: 1},
		{name: "set new key", input: `{"one": 1}.set("two", 2).get("two")`, want: 2},
		{name: "set existing key", input: `{"one": 1}.set("one", 2).get("one")`, want: 2},
		{name: "delete", input: `{"one": 1, "two": 2}.delete("one").get("two")`, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}

	testNullObject(t, testEval(`{"one": 1}.get("missing")`))
	testNullObject(t, testEval(`{"one": 1}.delete("missing").get("missing")`))
}

func TestMapContainsUsesNormalizedKeys(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: `{"one": 1}.contains("one")`, want: true},
		{input: `{"one": 1}.contains("two")`, want: false},
		{input: `{1: "one"}.contains(1.0)`, want: true},
		{input: `{True: 1}.contains(True)`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestMapValuesReturnsEveryValue(t *testing.T) {
	result := testEval(`{"one": 1, "two": 2, "three": 3}.values()`)
	values, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("result is %T, want *object.Array", result)
	}
	if len(values.Elements) != 3 {
		t.Fatalf("values length is %d, want 3", len(values.Elements))
	}
	seen := make(map[int64]bool, 3)
	for _, element := range values.Elements {
		integer, ok := element.(*object.Integer)
		if !ok {
			t.Fatalf("value is %T, want *object.Integer", element)
		}
		seen[integer.Value] = true
	}
	for _, want := range []int64{1, 2, 3} {
		if !seen[want] {
			t.Fatalf("values result does not contain %d", want)
		}
	}
}

func TestMapSetAndDeleteDoNotMutateInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name: "set",
			input: `let original = {"one": 1}
let changed = original.set("two", 2)
original.contains("two") == False`,
			want: true,
		},
		{
			name: "delete",
			input: `let original = {"one": 1}
let changed = original.delete("one")
original.contains("one")`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBooleanObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestMapBuiltinErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
	}{
		{name: "get receiver", input: `get(1, "one")`, message: "argument to `get` must be MAP, got INTEGER"},
		{name: "set receiver", input: `set(1, "one", 1)`, message: "argument to `set` must be MAP, got INTEGER"},
		{name: "delete receiver", input: `delete(1, "one")`, message: "argument to `delete` must be MAP, got INTEGER"},
		{name: "values receiver", input: `values(1)`, message: "argument to `values` must be MAP, got INTEGER"},
		{name: "get key", input: `{}.get([])`, message: "unusable as hash key: ARRAY"},
		{name: "set key", input: `{}.set([], 1)`, message: "unusable as hash key: ARRAY"},
		{name: "delete key", input: `{}.delete([])`, message: "unusable as hash key: ARRAY"},
		{name: "contains key", input: `{}.contains([])`, message: "unusable as hash key: ARRAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := testEval(tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}
}

func TestMapBuiltinsRemainCallableGlobally(t *testing.T) {
	testIntegerObject(t, testEval(`get({"one": 1}, "one")`), 1)
	testIntegerObject(t, testEval(`set({}, "one", 1).get("one")`), 1)
	testBooleanObject(t, testEval(`delete({"one": 1}, "one").contains("one")`), false)
	testBooleanObject(t, testEval(`contains({"one": 1}, "one")`), true)
	if result, ok := testEval(`values({"one": 1})`).(*object.Array); !ok || len(result.Elements) != 1 {
		t.Fatalf("values global result is %T (%+v), want one-element array", result, result)
	}
}
