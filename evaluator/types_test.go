package evaluator

import (
	"silver/object"
	"testing"
)

func TestTypedLetBinding(t *testing.T) {
	testIntegerObject(t, testEval(`let age: int = 36; age;`), 36)
}

func TestTypedLetBindingRejectsMismatch(t *testing.T) {
	evaluated := testEval(`let age: int = "thirty-six";`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for binding "age": expected int, got string`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestTypedFunctionParameter(t *testing.T) {
	evaluated := testEval(`let double = fn(value: int): int { value * 2; }; double("two");`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for parameter "value": expected int, got string`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestTypedFunctionReturnValue(t *testing.T) {
	evaluated := testEval(`let label = fn(): string { 42; }; label();`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for return value of "label": expected string, got int`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestUnknownType(t *testing.T) {
	evaluated := testEval(`let value: Missing = 1;`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `unknown type "Missing"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestEnumTypeAnnotation(t *testing.T) {
	evaluated := testEval(`enum Direction { North, South } let direction: Direction = Direction.North; direction;`)
	value, ok := evaluated.(*object.EnumValue)
	if !ok || value.Member != "North" {
		t.Fatalf("result is %#v, want Direction.North", evaluated)
	}
}
