package evaluator

import (
	"silver/object"
	"testing"
)

func TestVariadicParameterForwardsIndividualArguments(t *testing.T) {
	evaluated := testEval(`
let concatenate = fn(first: str, second: str, third: str) str {
	first + second + third
}
let forward = fn(parts: str...) str { concatenate(parts) }
forward("one", "two", "three")
`)
	testStringObject(t, evaluated, "onetwothree")
}

func TestVariadicParameterAcceptsNoArguments(t *testing.T) {
	evaluated := testEval(`
let answer = fn() int { 42 }
let forward = fn(parts: str...) int { answer(parts) }
forward()
`)
	testIntegerObject(t, evaluated, 42)
}

func TestVariadicParameterChecksEveryArgumentType(t *testing.T) {
	evaluated := testEval(`
let gather = fn(parts: str...) { }
gather("one", 2, "three")
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "parts": expected str, got int`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestVariadicFunctionRequiresFixedArguments(t *testing.T) {
	evaluated := testEval(`
let gather = fn(prefix: str, parts: str...) { }
gather()
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), "wrong number of arguments. got=0, want>=1"; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestVariadicCallSignatureCompatibility(t *testing.T) {
	evaluated := testEval(`
let concatenate = fn(first: str, second: str) str { first + second }
let gather: call(parts: str...) str = fn(parts: str...) str {
	concatenate(parts)
}
gather("one", "two")
`)
	testStringObject(t, evaluated, "onetwo")
}

func TestVariadicCallSignatureRejectsFixedFunction(t *testing.T) {
	evaluated := testEval(`let gather: call(str...) = fn(value: str) {}`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for binding "gather": expected call(str...), got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}
