package evaluator

import (
	"silver/object"
	"testing"
)

func TestAssertPassesForTruthyValues(t *testing.T) {
	testIntegerObject(t, testEval("assert True\nassert 0\n42"), 42)
}

func TestAssertRaisesAssertionError(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: "assert False", message: ""},
		{input: `assert False, "must be true"`, message: "must be true"},
		{input: "assert False, 42", message: "42"},
	}
	for _, test := range tests {
		err := assertErrorStruct(t, testEval(test.input), "AssertionError")
		if got := err.MessageText(); got != test.message {
			t.Fatalf("message for %q is %q, want %q", test.input, got, test.message)
		}
	}
}

func TestAssertMessageIsLazy(t *testing.T) {
	testIntegerObject(t, testEval("assert True, missing\n42"), 42)
	assertErrorStruct(t, testEval("assert False, missing"), "NameError")
}

func TestAssertionErrorCanBeCaught(t *testing.T) {
	result := testEval(`try {
    assert False, "broken"
} catch AssertionError err {
    err.message
}`)
	message, ok := result.(*object.String)
	if !ok || message.Value != "broken" {
		t.Fatalf("result is %#v, want caught assertion message", result)
	}
}
