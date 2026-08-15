package evaluator

import (
	"silver/object"
	"testing"
)

func TestSwitchExpressionSelectsOneCase(t *testing.T) {
	input := `let label = switch 2 {
case 1:
    "one"
case 2:
    "two"
case 3:
    "three"
default:
    "other"
}
label`
	value, ok := testEval(input).(*object.String)
	if !ok || value.Value != "two" {
		t.Fatalf("switch returned %T (%v), want two", value, value)
	}
}

func TestSwitchStandaloneDoesNotFallThrough(t *testing.T) {
	input := `let calls = 0
switch "start" {
case "start":
    calls = calls + 1
case "stop":
    calls = calls + 10
default:
    calls = calls + 100
}
calls`
	testIntegerObject(t, testEval(input), 1)
}

func TestSwitchWithoutDefaultReturnsNull(t *testing.T) {
	testNullObject(t, testEval(`switch 10 {
case 1:
    "one"
}`))
}

func TestSwitchReevaluatesReachedCasesEveryTime(t *testing.T) {
	input := `let calls = 0
let candidate = fn(value: int) int {
    calls = calls + 1
    value
}
let classify = fn(value: int) str {
    switch value {
    case candidate(1):
        "one"
    case candidate(2):
        "two"
    default:
        "other"
    }
}
let first = classify(2)
let second = classify(2)
calls`
	testIntegerObject(t, testEval(input), 4)
}

func TestSwitchEvaluatesSubjectOnce(t *testing.T) {
	input := `let calls = 0
let subject = fn() int {
    calls = calls + 1
    2
}
switch subject() {
case 1:
    "one"
case 2:
    "two"
}
calls`
	testIntegerObject(t, testEval(input), 1)
}

func TestSwitchSkipsLaterCaseExpressionsAfterMatch(t *testing.T) {
	input := `let calls = 0
let candidate = fn(value: int) int {
    calls = calls + 1
    value
}
switch 1 {
case candidate(1):
    "one"
case candidate(2):
    "two"
}
calls`
	testIntegerObject(t, testEval(input), 1)
}

func TestSwitchUsesStructEqualitySemantics(t *testing.T) {
	input := operatorStructPrelude + `
switch number {
case 10:
    "ten"
default:
    "other"
}`
	value, ok := testEval(input).(*object.String)
	if !ok || value.Value != "ten" {
		t.Fatalf("switch returned %T (%v), want ten", value, value)
	}
}

func TestSwitchSupportsFirstClassTypeValues(t *testing.T) {
	input := coreImport + `let value = "silver"
switch core.type(value) {
case int:
    "integer"
case str:
    "string"
default:
    "other"
}`
	value, ok := testEval(input).(*object.String)
	if !ok || value.Value != "string" {
		t.Fatalf("switch returned %T (%v), want string", value, value)
	}
}
