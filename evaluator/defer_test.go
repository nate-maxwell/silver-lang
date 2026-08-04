package evaluator

import (
	"bytes"
	"silver/object"
	"testing"
)

func TestDeferredCallsRunLIFOOnReturn(t *testing.T) {
	var out bytes.Buffer
	input := `let record = fn(value: str) { print(value) }
let run = fn() int {
    defer record("first")
    defer record("second")
    return 7
}
run()`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	testIntegerObject(t, result, 7)
	if got, want := out.String(), "second\nfirst\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestDeferCapturesArgumentsAndCallableImmediately(t *testing.T) {
	var out bytes.Buffer
	input := `let value = "before"
let action = fn(text: str) { print("old " + text) }
defer action(value)
value = "after"
action = fn(text: str) { print("new " + text) }`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	testNullObject(t, result)
	if got, want := out.String(), "old before\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestDeferredCallsRunWhileUnwindingError(t *testing.T) {
	var out bytes.Buffer
	input := `let run = fn() {
    defer print("cleanup")
    missing_name
}
run()`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	if _, ok := result.(*object.Error); !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if got, want := out.String(), "cleanup\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestAllDeferredCallsRunAndLaterFailureWins(t *testing.T) {
	var out bytes.Buffer
	input := `let fail_first = fn() {
    print("first")
    missing_first
}
let fail_second = fn() {
    print("second")
    missing_second
}
let run = fn() {
    defer fail_first()
    defer fail_second()
    missing_body
}
run()`

	result, ok := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input).(*object.Error)
	if !ok {
		t.Fatalf("result is not *object.Error")
	}
	if got, want := result.MessageText(), "identifier not found: missing_first"; got != want {
		t.Fatalf("error is %q, want %q", got, want)
	}
	if got, want := out.String(), "second\nfirst\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestDeferredCallsUseFunctionScope(t *testing.T) {
	var out bytes.Buffer
	input := `let run = fn() {
    if True { defer print("cleanup") }
    print("body")
}
run()`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	testNullObject(t, result)
	if got, want := out.String(), "body\ncleanup\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}
