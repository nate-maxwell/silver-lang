package stdlib_test

import (
	"silver/object"
	"testing"
)

const collectionsImport = "let collections = import(\"collections\")\n"

func TestDequeOperations(t *testing.T) {
	result := testEval(collectionsImport + `let values = collections.deque([2, 3])
values.appendleft(1)
values.append(4)
let right = values.pop()
let left = collections.popleft(values)
[values.values, left, right]`)

	if got, want := result.Inspect(), `[[2, 3], 1, 4]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestDequeAppendOperationsAreMethods(t *testing.T) {
	for _, input := range []string{`collections.append(collections.deque(), 1)`, `collections.appendleft(collections.deque(), 1)`} {
		if _, ok := testEval(collectionsImport + input).(*object.Error); !ok {
			t.Fatalf("%s did not require method syntax", input)
		}
	}

	if _, ok := testEval(collectionsImport + `[1].append(2)`).(*object.Error); !ok {
		t.Fatal("ordinary array unexpectedly has deque append method")
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
[values.values, collections.count(values, 2), collections.index(values, 3)]`)

	if got, want := result.Inspect(), `[[3, 2, 1, -1, 0, 5, 4], 1, 0]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestStackOperations(t *testing.T) {
	result := testEval(collectionsImport + `let values = collections.stack()
values.push("first")
values.push("second")
let top = values.peek()
let popped = values.pop()
[values.values, top, popped]`)

	if got, want := result.Inspect(), `[[first], second, second]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestSequenceIndexMethods(t *testing.T) {
	result := testEval(collectionsImport + `let values = collections.deque([1, 2])
values[1] = 9
[values[0], values[1]]`)

	if got, want := result.Inspect(), `[1, 9]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestStackOperationsAreMethods(t *testing.T) {
	for _, input := range []string{
		`collections.push(collections.stack(), 1)`,
		`collections.peek(collections.stack())`,
		`collections.pop(collections.stack())`,
	} {
		if _, ok := testEval(collectionsImport + input).(*object.Error); !ok {
			t.Fatalf("%s did not require method syntax", input)
		}
	}

	if _, ok := testEval(collectionsImport + `[1].pop()`).(*object.Error); !ok {
		t.Fatal("ordinary array unexpectedly has stack pop method")
	}
}

func TestDefaultMapCreatesAndStoresMissingValues(t *testing.T) {
	result := testEval(collectionsImport + `let calls = {"count": 0}
let make_list = fn() array {
    calls["count"] = calls["count"] + 1
    return []
}
let grouped = collections.defaultmap(make_list, {"known": [1]})
let first = grouped["missing"]
let second = grouped["missing"]
[grouped["known"], first == second, calls["count"], grouped.values]`)

	if got, want := result.Inspect(), `[[1], true, 1, {known: [1], missing: []}]`; got != want && got != `[[1], true, 1, {missing: [], known: [1]}]` {
		t.Fatalf("result is %q, want map with known and missing entries", got)
	}
}

func TestDefaultMapSetAndNominalType(t *testing.T) {
	result := testEval(collectionsImport + `let core = import("core")
let make_count = fn() int { 0 }
let counts = collections.defaultmap(make_count)
counts["silver"] = counts["silver"] + 1
[counts["silver"], core.type(counts) == collections.DefaultMap]`)

	if got, want := result.Inspect(), `[1, true]`; got != want {
		t.Fatalf("result is %q, want %q", got, want)
	}
}

func TestCollectionErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `collections.stack().pop()`, message: "pop from an empty collection"},
		{input: `collections.stack().peek()`, message: "peek from an empty stack"},
		{input: `collections.defaultmap(0)`, message: "default factory must be a Silver function, got INTEGER"},
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
