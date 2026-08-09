package stdlib_test

import (
	"bytes"
	"silver/evaluator"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"strings"
	"testing"
)

func TestSilverTestingModuleRunsTestsAndSuites(t *testing.T) {
	result, output := evalTesting(t, `
let t = import("testing")
let strings = import("string")
let math = import("math")

let add = fn(left: int, right: int) int { left + right }

t.run("add function", fn() {
    let result = add(2, 3)
    assert result == 5
})

t.run("string operations", fn() {
    assert strings.upper("hello") == "HELLO"
})

t.suite("math module", fn() {
    t.run("sqrt", fn() { assert math.sqrt(4.0) == 2.0 })
    t.run("abs", fn() { assert math.abs(-5) == 5 })
})

t.summary()
`)

	summary, ok := result.(*object.StructInstance)
	if !ok || summary.Struct.Name != "Summary" {
		t.Fatalf("result is %T (%v), want testing.Summary", result, result)
	}
	assertStructInteger(t, summary, "total", 4)
	assertStructInteger(t, summary, "passed", 4)
	assertStructInteger(t, summary, "failed", 0)

	for _, line := range []string{
		"PASS add function",
		"PASS string operations",
		"PASS math module / sqrt",
		"PASS math module / abs",
	} {
		if !strings.Contains(output, line) {
			t.Fatalf("output does not contain %q:\n%s", line, output)
		}
	}
}

func TestSilverTestingModuleRecordsAssertionHelpersAndReports(t *testing.T) {
	result, output := evalTesting(t, `
let t = import("testing")

t.run("passing helper", fn() {
    t.is_true(True, "value should be true")
    t.is_false(False, "value should be false")
    t.not_equal(1, 2, "values should differ")
})

t.run("failing helper", fn() {
    t.equal(1, 2, "numbers differ")
})

t.report()
`)
	testBooleanObject(t, result, false)

	for _, line := range []string{
		"PASS passing helper",
		"FAIL failing helper",
		"2 tests: 1 passed, 1 failed",
		"Failed tests:",
		"failing helper: numbers differ",
	} {
		if !strings.Contains(output, line) {
			t.Fatalf("output does not contain %q:\n%s", line, output)
		}
	}
}

func TestSilverTestingModuleExposesResultsAndReset(t *testing.T) {
	result, _ := evalTesting(t, `
let t = import("testing")
t.run("before reset", fn() { assert False, "broken" })
let recorded = t.results()
assert recorded[0].name == "before reset"
assert recorded[0].passed == False
assert recorded[0].message == "broken"
t.reset()
t.summary()
`)

	summary, ok := result.(*object.StructInstance)
	if !ok || summary.Struct.Name != "Summary" {
		t.Fatalf("result is %T (%v), want testing.Summary", result, result)
	}
	assertStructInteger(t, summary, "total", 0)
	assertStructInteger(t, summary, "passed", 0)
	assertStructInteger(t, summary, "failed", 0)
}

func TestSilverTestingSuiteRestoresNestedNames(t *testing.T) {
	_, output := evalTesting(t, `
let t = import("testing")
t.suite("outer", fn() {
    t.suite("inner", fn() {
        t.run("nested", fn() {})
    })
    t.run("sibling", fn() {})
})
t.run("top level", fn() {})
`)

	for _, line := range []string{
		"PASS outer / inner / nested",
		"PASS outer / sibling",
		"PASS top level",
	} {
		if !strings.Contains(output, line) {
			t.Fatalf("output does not contain %q:\n%s", line, output)
		}
	}
}

func evalTesting(t *testing.T, input string) (object.Object, string) {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	var output bytes.Buffer
	result := evaluator.NewWithOutput(&output).Eval(program, object.NewEnvironment())
	if failure, ok := result.(*object.Error); ok {
		t.Fatalf("evaluation failed:\n%s", failure.Inspect())
	}
	return result, output.String()
}

func assertStructInteger(t *testing.T, instance *object.StructInstance, field string, want int64) {
	t.Helper()
	value, ok := instance.Get(field)
	if !ok {
		t.Fatalf("struct %s has no field %q", instance.Struct.Name, field)
	}
	integer, ok := value.(*object.Integer)
	if !ok || integer.Value != want {
		t.Fatalf("field %q is %T (%v), want %d", field, value, value, want)
	}
}
