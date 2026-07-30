package evaluator

import (
	"silver/object"
	"testing"
)

func TestTypedLetBinding(t *testing.T) {
	testIntegerObject(t, testEval("let age: int = 36\nage"), 36)
}

func TestTypedLetBindingRejectsMismatch(t *testing.T) {
	evaluated := testEval(`let age: int = "thirty-six"`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for binding "age": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestTypedFunctionParameter(t *testing.T) {
	evaluated := testEval("let double = fn(value: int) int = { value * 2 }\ndouble(\"two\")")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for parameter "value": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestTypedFunctionReturnValue(t *testing.T) {
	evaluated := testEval("let label = fn() str = { 42 }\nlabel()")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for return value of "label": expected str, got int`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestUnknownType(t *testing.T) {
	evaluated := testEval(`let value: Missing = 1`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `unknown type "Missing"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStringTypeNameIsRejected(t *testing.T) {
	evaluated := testEval(`let value: string = "old spelling"`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `unknown type "string"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestEnumTypeAnnotation(t *testing.T) {
	evaluated := testEval("enum Direction { North, South }\nlet direction: Direction = Direction.North\ndirection")
	value, ok := evaluated.(*object.EnumValue)
	if !ok || value.Member != "North" {
		t.Fatalf("result is %#v, want Direction.North", evaluated)
	}
}

func TestBareReturnUsesNull(t *testing.T) {
	evaluated := testEval("let noValue = fn() null = {\nreturn\n}\nnoValue()")
	testNullObject(t, evaluated)
}

func TestFunctionWithoutReturnTypeReturnsNull(t *testing.T) {
	evaluated := testEval("let noValue = fn() = {\nreturn 42\n}\nnoValue()")
	testNullObject(t, evaluated)
}

func TestTypedHigherOrderFunction(t *testing.T) {
	evaluated := testEval(`
let apply = fn(operation: call(int) int, value: int) int = {
	return operation(value)
}
apply(fn(value: int) int = { value * 2 }, 21)
`)
	testIntegerObject(t, evaluated, 42)
}

func TestCallSignatureRejectsWrongArgumentSignature(t *testing.T) {
	evaluated := testEval(`
let apply = fn(operation: call(int) int, value: int) int = {
	return operation(value)
}
apply(fn(value: str) int = { 1 }, 21)
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for parameter "operation": expected call(int) int, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestCallSignatureRejectsWrongArity(t *testing.T) {
	evaluated := testEval(`
let apply = fn(operation: call(int) int, value: int) int = {
	return operation(value)
}
apply(fn(left: int, right: int) int = { left + right }, 21)
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for parameter "operation": expected call(int) int, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestCallSignatureReturnTypeAndClosure(t *testing.T) {
	evaluated := testEval(`
let makeAdder = fn(amount: int) call(int) int = {
	return fn(value: int) int = { value + amount }
}
let addTwo: call(int) int = makeAdder(2)
addTwo(40)
`)
	testIntegerObject(t, evaluated, 42)
}

func TestCallSignatureRejectsWrongReturnedFunction(t *testing.T) {
	evaluated := testEval(`
let makeIdentity = fn() call(int) int = {
	return fn(value: str) int = { 1 }
}
makeIdentity()
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.Message, `type mismatch for return value of "makeIdentity": expected call(int) int, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}
