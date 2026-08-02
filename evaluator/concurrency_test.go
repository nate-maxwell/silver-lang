package evaluator

import (
	"bytes"
	"silver/object"
	"strings"
	"testing"
)

func TestTaskCollectReturnsNamedNonNullResults(t *testing.T) {
	result := testEval(`
let double = fn(value: int) int { value * 2 }
let a = task(double(2))
let b = task { double(5) }
let c = task(print("side effect"))
let results = collect(a, b, c)
results
`)
	collected, ok := result.(*object.StructInstance)
	if !ok {
		t.Fatalf("result is %T, want collect struct", result)
	}
	a, ok := collected.Get("a")
	if !ok {
		t.Fatal("collect result has no a field")
	}
	b, ok := collected.Get("b")
	if !ok {
		t.Fatal("collect result has no b field")
	}
	testIntegerObject(t, a, 4)
	testIntegerObject(t, b, 10)
	if _, ok := collected.Get("c"); ok {
		t.Fatal("null-returning task unexpectedly produced c field")
	}
}

func TestTaskPreservesErrorStructAsFieldValue(t *testing.T) {
	result := testEval(`
struct FileNotFound { message: str }
let read = fn() str | FileNotFound { FileNotFound{"missing"} }
let a = task(read())
let results = collect(a)
type(results.a) == FileNotFound
`)
	testBooleanObject(t, result, true)
}

func TestTaskHoldsRuntimeErrorUntilCollect(t *testing.T) {
	var out bytes.Buffer
	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), `
let fail = fn() int { 1 + True }
let bad = task(fail())
print("continued")
collect(bad)
`)
	err, ok := result.(*object.Error)
	if !ok || !strings.Contains(err.Message, "type mismatch") {
		t.Fatalf("result is %#v, want retained task runtime error", result)
	}
	if !strings.Contains(out.String(), "continued") {
		t.Fatalf("current scope did not continue before collect: %q", out.String())
	}
}

func TestUncollectedTaskWarnsAndIsJoinedAtScopeExit(t *testing.T) {
	var out, warnings bytes.Buffer
	result := evalInput(t, NewWithWriters(&out, &warnings), object.NewEnvironment(), `
let work = fn() { print("finished") }
let abandoned = task(work())
42
`)
	testIntegerObject(t, result, 42)
	if got := out.String(); !strings.Contains(got, "finished") {
		t.Fatalf("scope exited before task output completed: %q", got)
	}
	if got := warnings.String(); !strings.Contains(got, `task handle "abandoned" was never collected`) {
		t.Fatalf("missing uncollected-task warning: %q", got)
	}
}
