package evaluator

import (
	"silver/object"
	"strings"
	"testing"
)

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"foo": 5}["foo"]`, 5},
		{"let key = \"foo\"\n{\"foo\": 5}[key]", 5},
		{`{5: 5}[5]`, 5},
		{`{True: 5}[True]`, 5},
		{`{False: 5}[False]`, 5},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestMapLiteralRejectsDuplicateComputedKeys(t *testing.T) {
	for _, input := range []string{
		"let key = 1\nlet t = {key: 10, 1: 20}",
		"let key = 1.0\nlet t = {1: 10, key: 20}",
		"let key = True\nlet t = {key: 10, True: 20}",
		"let key = \"same\"\nlet t = {key: 10, key: 20}",
		`let t = {"sa" + "me": 10, "same": 20}`,
		`let t = {1 + 1: 10, 2: 20}`,
		"let key = fn() int { return 1 }\nlet t = {key(): 10, key(): 20}",
	} {
		t.Run(input, func(t *testing.T) {
			program, parseErr := ParseSource("duplicate.slv", []byte(input))
			if parseErr != nil {
				t.Fatalf("parse error: %s", parseErr.Inspect())
			}
			err := assertErrorStruct(t, New().Eval(program, object.NewEnvironment()), "ValueError")
			if !strings.Contains(err.MessageText(), "duplicate map key") {
				t.Fatalf("unexpected error: %s", err.MessageText())
			}
		})
	}
}

func TestMapLiteralEvaluatesRepeatedCallsAsDistinctKeys(t *testing.T) {
	program, parseErr := ParseSource("distinct.slv", []byte(`
let count = 0
let next = fn() int {
    count = count + 1
    return count
}
{next(): 10, next(): 10}
`))
	if parseErr != nil {
		t.Fatalf("parse error: %s", parseErr.Inspect())
	}
	result := New().Eval(program, object.NewEnvironment())
	mapping, ok := result.(*object.Map)
	if !ok || mapping.Len() != 2 {
		t.Fatalf("result is %#v, want map with two distinct keys", result)
	}
	for _, key := range []int64{1, 2} {
		pair, ok := mapping.Get((&object.Integer{Value: key}).HashKey())
		if !ok {
			t.Fatalf("map is missing key %d", key)
		}
		testIntegerObject(t, pair.Value, 10)
	}
}

func TestMissingMapIndexProducesKeyError(t *testing.T) {
	for _, input := range []string{`{"foo": 5}["bar"]`, `{}["foo"]`} {
		err, ok := testEval(input).(*object.Error)
		if !ok || err.Value.Struct.Name != "KeyError" {
			t.Fatalf("%s returned %#v, want KeyError", input, err)
		}
	}
}

func TestMapLiterals(t *testing.T) {
	input := `let two = "two"
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 // 2,
		4: 4,
		True: 5,
		False: 6
	}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Map)
	if !ok {
		t.Fatalf("Eval didn't return Map. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.HashKey]int64{
		(&object.String{Value: "one"}).HashKey():   1,
		(&object.String{Value: "two"}).HashKey():   2,
		(&object.String{Value: "three"}).HashKey(): 3,
		(&object.Integer{Value: 4}).HashKey():      4,
		TRUE.HashKey():                             5,
		FALSE.HashKey():                            6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("Map has wrong num of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]
		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testIntegerObject(t, pair.Value, expectedValue)
	}
}

func TestMapIndexAssignmentCreatesAndReplacesEntries(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int64
	}{
		{
			name: "create",
			input: `let values = {}
values["answer"] = 42
values["answer"]`,
			want: 42,
		},
		{
			name: "replace",
			input: `let values = {"answer": 1}
values["answer"] = 42
values["answer"]`,
			want: 42,
		},
		{
			name: "normalized numeric key",
			input: `let values = {1: 1}
values[1.0] = 42
values[1]`,
			want: 42,
		},
		{
			name: "nested map",
			input: `let values = {"nested": {}}
values["nested"]["answer"] = 42
values["nested"]["answer"]`,
			want: 42,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testIntegerObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestMapIndexAssignmentMutatesAliases(t *testing.T) {
	result := testEval(`let original = {}
let alias = original
alias["answer"] = 42
original["answer"]`)
	testIntegerObject(t, result, 42)
}

func TestMapIndexAssignmentErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `let values = {}
values[[]] = 1`, message: "unusable as hash key: ARRAY"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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
