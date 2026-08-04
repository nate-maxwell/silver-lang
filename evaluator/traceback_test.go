package evaluator

import (
	"fmt"
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestTracebackIncludesSourceLocationsAndFunctionNames(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, path, `let fail = fn() {
    1 + True
}
let middle = fn() {
    fail()
}
middle()
`)

	result := New().EvalFile(path, object.NewEnvironment())
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}

	want := []object.TraceFrame{
		{Source: path, Line: 7, Column: 1, Function: "<module>"},
		{Source: path, Line: 5, Column: 5, Function: "middle"},
		{Source: path, Line: 2, Column: 7, Function: "fail"},
	}
	assertTraceFrames(t, err.Frames, want)

	traceback := err.Inspect()
	if !strings.HasPrefix(traceback, "Traceback (most recent call last):\n") {
		t.Fatalf("traceback has unexpected header:\n%s", traceback)
	}
	if !strings.Contains(traceback, fmt.Sprintf("File \"%s\", line 2, column 7, in fail", path)) {
		t.Fatalf("traceback does not include the error origin:\n%s", traceback)
	}
	if !strings.HasSuffix(traceback, "TypeError: type mismatch: INTEGER + BOOLEAN") {
		t.Fatalf("traceback has unexpected error message:\n%s", traceback)
	}
}

func TestTracebackPreservesLocationsAcrossModules(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library.slv")
	mainPath := filepath.Join(dir, "main.slv")

	writeSilverFile(t, libraryPath, `let fail = fn() {
    1 / 0
}
let run = fn() {
    fail()
}
`)
	writeSilverFile(t, mainPath, `let library = import("./library.slv")
library.run()
`)

	result := New().EvalFile(mainPath, object.NewEnvironment())
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}

	want := []object.TraceFrame{
		{Source: mainPath, Line: 2, Column: 1, Function: "<module>"},
		{Source: libraryPath, Line: 5, Column: 5, Function: "run"},
		{Source: libraryPath, Line: 2, Column: 7, Function: "fail"},
	}
	assertTraceFrames(t, err.Frames, want)
	if err.MessageText() != "division by zero" {
		t.Fatalf("error message is %q, want %q", err.MessageText(), "division by zero")
	}
}

func TestFunctionArityErrorHasCallLocation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "arity.slv")
	writeSilverFile(t, path, `let identity = fn(value) { value }
identity()
`)

	result := New().EvalFile(path, object.NewEnvironment())
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}

	want := []object.TraceFrame{
		{Source: path, Line: 2, Column: 1, Function: "<module>"},
	}
	assertTraceFrames(t, err.Frames, want)
	if err.MessageText() != "wrong number of arguments. got=0, want=1" {
		t.Fatalf("error message is %q", err.MessageText())
	}
}

func assertTraceFrames(t *testing.T, got, want []object.TraceFrame) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("trace frame count is %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("frame %d is %+v, want %+v", i, got[i], want[i])
		}
	}
}
