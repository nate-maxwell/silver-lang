package evaluator

import (
	"silver/object"
	"testing"
)

func TestVariableAssignment(t *testing.T) {
	result := testEval("let value = 1\nvalue = 42\nvalue")
	testIntegerObject(t, result, 42)
}

func TestAssignmentUpdatesEnclosingBinding(t *testing.T) {
	result := testEval(`
let value = 1
let update = fn() { value = 42 }
update()
value
`)
	testIntegerObject(t, result, 42)
}

func TestAssignmentPreservesDeclaredType(t *testing.T) {
	result := testEval(`
let value: int = 1
value = "wrong"
`)
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if got, want := err.Message, `type mismatch for binding "value": expected int, got str`; got != want {
		t.Fatalf("error is %q, want %q", got, want)
	}
}

func TestAssignmentRequiresExistingBinding(t *testing.T) {
	result := testEval("missing = 1")
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if got, want := err.Message, "identifier not found: missing"; got != want {
		t.Fatalf("error is %q, want %q", got, want)
	}
}
