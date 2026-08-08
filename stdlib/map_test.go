package stdlib_test

import (
	"silver/object"
	"testing"
)

const mapImport = "let maps = import(\"map\")\n"

func TestMapModuleFunctions(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{name: "get", input: `maps.get({"one": 1}, "one")`, want: 1},
		{name: "set new key", input: `maps.get(maps.set({"one": 1}, "two", 2), "two")`, want: 2},
		{name: "set existing key", input: `maps.get(maps.set({"one": 1}, "one", 2), "one")`, want: 2},
		{name: "delete", input: `maps.get(maps.delete({"one": 1, "two": 2}, "one"), "two")`, want: 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(mapImport+tt.input), tt.want)
		})
	}

	testNullObject(t, testEval(mapImport+`maps.get({"one": 1}, "missing")`))
	testNullObject(t, testEval(mapImport+`maps.get(maps.delete({"one": 1}, "missing"), "missing")`))
}

func TestMapContainsUsesNormalizedKeys(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{input: `maps.contains({"one": 1}, "one")`, want: true},
		{input: `maps.contains({"one": 1}, "two")`, want: false},
		{input: `maps.contains({1: "one"}, 1.0)`, want: true},
		{input: `maps.contains({True: 1}, True)`, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(mapImport+tt.input), tt.want)
		})
	}
}

func TestMapValuesReturnsEveryValue(t *testing.T) {
	result := testEval(mapImport + `maps.values({"one": 1, "two": 2, "three": 3})`)
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
let changed = maps.set(original, "two", 2)
maps.contains(original, "two") == False`,
			want: true,
		},
		{
			name: "delete",
			input: `let original = {"one": 1}
let changed = maps.delete(original, "one")
maps.contains(original, "one")`,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBooleanObject(t, testEval(mapImport+tt.input), tt.want)
		})
	}
}

func TestMapBuiltinErrors(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
	}{
		{name: "get receiver", input: `maps.get(1, "one")`, message: "argument to `get` must be MAP, got INTEGER"},
		{name: "set receiver", input: `maps.set(1, "one", 1)`, message: "argument to `set` must be MAP, got INTEGER"},
		{name: "delete receiver", input: `maps.delete(1, "one")`, message: "argument to `delete` must be MAP, got INTEGER"},
		{name: "values receiver", input: `maps.values(1)`, message: "argument to `values` must be MAP, got INTEGER"},
		{name: "get key", input: `maps.get({}, [])`, message: "unusable as hash key: ARRAY"},
		{name: "set key", input: `maps.set({}, [], 1)`, message: "unusable as hash key: ARRAY"},
		{name: "delete key", input: `maps.delete({}, [])`, message: "unusable as hash key: ARRAY"},
		{name: "contains key", input: `maps.contains({}, [])`, message: "unusable as hash key: ARRAY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := testEval(mapImport + tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}
}

func TestMapFunctionsRequireImport(t *testing.T) {
	for _, input := range []string{`get({}, "key")`, `{}.get("key")`} {
		if _, ok := testEval(input).(*object.Error); !ok {
			t.Fatalf("%s did not require the map module", input)
		}
	}
}
