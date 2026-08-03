package evaluator

import (
	"silver/object"
	"testing"
)

func TestHashIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`{"foo": 5}["foo"]`, 5},
		{`{"foo": 5}["bar"]`, nil},
		{"let key = \"foo\"\n{\"foo\": 5}[key]", 5},
		{`{}["foo"]`, nil},
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
	result, ok := evaluated.(*object.Hash)
	if !ok {
		t.Fatalf("Eval didn't return Hash. got=%T (%+v)", evaluated, evaluated)
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
		t.Fatalf("Hash has wrong num of pairs. got=%d", len(result.Pairs))
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
		{input: `let values = []
values[0] = 1`, message: "index assignment not supported on array"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.Message != tt.message {
				t.Fatalf("error is %q, want %q", result.Message, tt.message)
			}
		})
	}
}
