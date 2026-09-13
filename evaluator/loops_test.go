package evaluator

import (
	"silver/object"
	"testing"
)

func TestForLoopIteratesArray(t *testing.T) {
	result := testEval(`
let total = 0
for number in [1, 2, 3, 4] {
    total = total + number
}
total
`)
	testIntegerObject(t, result, 10)
}

func TestForLoopIteratesMapKeysAndValues(t *testing.T) {
	result := testEval(`
let total = 0
let entries = {1: 10, 2: 20, 3: 30}
for customKey, customValue in entries {
    total = total + customKey + customValue
}
total
`)
	testIntegerObject(t, result, 66)
}

func TestForLoopShadowsEnclosingBindings(t *testing.T) {
	for name, input := range map[string]string{
		"array": `let n: int = 1
for n in ["oops"] {
    assert n == "oops"
    n = False
    assert n == False
}
n`,
		"map": `let key: int = 1
let value: int = 2
let entries = {"key": "value"}
for key, value in entries {
    assert key == "key"
    assert value == "value"
    key = False
    value = False
}
key + value - 2`,
		"nested loops": `let n: int = 1
let total = 0
for n in [2, 3] {
    let entries = {4: 5}
    for n, value in entries {
        assert n == 4
    }
    total = total + n
}
assert total == 5
n`,
		"iterable uses enclosing binding": `let values = [1, 2, 3]
let total = 0
for values in values {
    total = total + values
}
assert total == 6
values[0]`,
	} {
		t.Run(name, func(t *testing.T) {
			result := evalInput(t, New(), object.NewEnvironment(), input)
			testIntegerObject(t, result, 1)
		})
	}
}

func TestForLoopPreservesEnclosingBindingType(t *testing.T) {
	for name, loop := range map[string]string{
		"array":     `for n in ["oops"] {}`,
		"map key":   "let entries = {\"oops\": False}\nfor n, value in entries {}",
		"map value": "let entries = {False: \"oops\"}\nfor key, n in entries {}",
	} {
		t.Run(name, func(t *testing.T) {
			engine, env := New(), object.NewEnvironment()
			result := evalInput(t, engine, env, "let n: int = 1\n"+loop+"\nn")
			testIntegerObject(t, result, 1)

			result = evalInput(t, engine, env, "n = False")
			err, ok := result.(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if got, want := err.MessageText(), `type mismatch for binding "n": expected int, got bool`; got != want {
				t.Fatalf("error is %q, want %q", got, want)
			}
		})
	}
}

func TestForLoopBindingsDoNotLeak(t *testing.T) {
	for name, input := range map[string]string{
		"array":       "for local in [1] {}\nlocal",
		"map key":     "let entries = {1: 2}\nfor local, value in entries {}\nlocal",
		"map value":   "let entries = {1: 2}\nfor key, local in entries {}\nlocal",
		"declaration": "for n in [1] { let local = n }\nlocal",
		"break":       "for local in [1] { break }\nlocal",
		"continue":    "for local in [1] { continue }\nlocal",
		"empty array": "for local in [] {}\nlocal",
		"empty map":   "let entries = {}\nfor key, local in entries {}\nlocal",
	} {
		t.Run(name, func(t *testing.T) {
			result := evalInput(t, New(), object.NewEnvironment(), input)
			err, ok := result.(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if got, want := err.MessageText(), "identifier not found: local"; got != want {
				t.Fatalf("error is %q, want %q", got, want)
			}
		})
	}
}

func TestForLoopClosuresCaptureEachIteration(t *testing.T) {
	for name, input := range map[string]string{
		"array": `let callbacks = {}
for n in [1, 2, 3] {
    let offset = n * 2
    callbacks[n] = fn() int { return n + offset }
    n = n + 10
}
for n in [100] {}
assert callbacks[1]() == 13
assert callbacks[2]() == 16
assert callbacks[3]() == 19`,
		"map": `let callbacks = {}
let entries = {1: 10, 2: 20}
for key, value in entries {
    callbacks[key] = fn() int { return key * 100 + value }
    key = key + 10
    value = value + 1
}
assert callbacks[1]() == 1111
assert callbacks[2]() == 1221`,
	} {
		t.Run(name, func(t *testing.T) {
			result := evalInput(t, New(), object.NewEnvironment(), input)
			testNullObject(t, result)
		})
	}
}

func TestWhileLoopReevaluatesCondition(t *testing.T) {
	result := testEval(`
let count = 0
while count < 5 {
    count = count + 1
}
count
`)
	testIntegerObject(t, result, 5)
}

func TestReturnPropagatesOutOfLoop(t *testing.T) {
	result := testEval(`
let find = fn(values) int {
    for value in values {
        if value > 2 { return value }
    }
    return 0
}
find([1, 2, 3, 4])
`)
	testIntegerObject(t, result, 3)
}

func TestContinueSkipsRestOfForIteration(t *testing.T) {
	result := testEval(`
let total = 0
for number in [1, 2, 3, 4, 5] {
    if number % 2 == 0 { continue }
    total = total + number
}
total
`)
	testIntegerObject(t, result, 9)
}

func TestBreakExitsWhileLoop(t *testing.T) {
	result := testEval(`
let count = 0
while True {
    count = count + 1
    if count == 3 { break }
}
count
`)
	testIntegerObject(t, result, 3)
}

func TestLoopControlTargetsNearestLoop(t *testing.T) {
	result := testEval(`
let visits = 0
for outer in [1, 2, 3] {
    for inner in [1, 2, 3] {
        if inner == 2 { break }
        visits = visits + 1
    }
}
visits
`)
	testIntegerObject(t, result, 3)
}

func TestContinueReevaluatesWhileCondition(t *testing.T) {
	result := testEval(`
let count = 0
let total = 0
while count < 5 {
    count = count + 1
    if count < 3 { continue }
    total = total + count
}
total
`)
	testIntegerObject(t, result, 12)
}

func TestMapLoopConsumesBreakAndContinue(t *testing.T) {
	result := testEval(`
let visits = 0
let entries = {1: 10, 2: 20}
for key, value in entries {
    visits = visits + 1
    continue
    visits = 100
}
let breaks = 0
for key, value in entries {
    breaks = breaks + 1
    break
}
visits + breaks
`)
	testIntegerObject(t, result, 3)
}

func TestForLoopValidatesCollectionShape(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"for key, value in [1] {}", "array for loop requires one binding"},
		{"let entries = {1: 2}\nfor item in entries {}", "map for loop requires key and value bindings"},
		{"for item in 1 {}", "not iterable: int"},
	}

	for _, test := range tests {
		result := testEval(test.input)
		err, ok := result.(*object.Error)
		if !ok {
			t.Fatalf("result for %q is %T, want *object.Error", test.input, result)
		}
		if err.MessageText() != test.want {
			t.Fatalf("error for %q is %q, want %q", test.input, err.MessageText(), test.want)
		}
	}
}
