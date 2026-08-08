package stdlib_test

import (
	"silver/object"
	"testing"
)

const collectionsImport = "let collections = import(\"collections\")\n"

func TestDequeOperations(t *testing.T) {
	result := testEval(collectionsImport + `let values = collections.deque([2, 3])
collections.appendleft(values, 1)
collections.append(values, 4)
let right = collections.pop(values)
let left = collections.popleft(values)
[values, left, right]`)

	if got, want := result.Inspect(), `[[2, 3], 1, 4]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestDequeEditingOperations(t *testing.T) {
	result := testEval(collectionsImport + `let values = collections.deque([1, 2, 3])
collections.extend(values, [4, 5])
collections.extendleft(values, [-1, 0])
collections.insert(values, 2, 9)
collections.remove(values, 9)
collections.rotate(values, 2)
collections.reverse(values)
[values, collections.count(values, 2), collections.index(values, 3)]`)

	if got, want := result.Inspect(), `[[3, 2, 1, -1, 0, 5, 4], 1, 0]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestStackOperations(t *testing.T) {
	result := testEval(collectionsImport + `let values = collections.stack()
collections.push(values, "first")
collections.push(values, "second")
let top = collections.peek(values)
let popped = collections.pop(values)
[values, top, popped]`)

	if got, want := result.Inspect(), `[[first], second, second]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestDefaultDictCreatesAndStoresMissingValues(t *testing.T) {
	result := testEval(collectionsImport + `let calls = {"count": 0}
let make_list = fn() array {
    calls["count"] = calls["count"] + 1
    return []
}
let grouped = collections.defaultdict(make_list, {"known": [1]})
let first = grouped["missing"]
let second = grouped["missing"]
[grouped["known"], first == second, calls["count"]]`)

	if got, want := result.Inspect(), `[[1], true, 1]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestDefaultDictGetDoesNotInvokeFactory(t *testing.T) {
	result := testEval(collectionsImport + `let maps = import("map")
let calls = {"count": 0}
let make_value = fn() int {
    calls["count"] = calls["count"] + 1
    return 0
}
let values = collections.defaultdict(make_value)
let missing = maps.get(values, "missing")
[missing, calls["count"]]`)

	if got, want := result.Inspect(), `[null, 0]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestCollectionErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `collections.pop([])`, message: "pop from an empty collection"},
		{input: `collections.peek([])`, message: "peek from an empty stack"},
		{input: `collections.defaultdict(0)`, message: "default factory must be callable, got INTEGER"},
		{input: `collections.rotate([], "one")`, message: "rotation argument to `rotate` must be INTEGER, got STRING"},
	}

	for _, tt := range tests {
		result, ok := testEval(collectionsImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if result.MessageText() != tt.message {
			t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
		}
	}
}
