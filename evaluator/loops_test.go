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
