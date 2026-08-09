package evaluator

import (
	"os"
	"path/filepath"
	"silver/object"
	"testing"
)

func TestRuntimeFaultsProduceSpecificErrorStructs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "type", input: `1 + True`, want: "TypeError"},
		{name: "value", input: coreImport + `core.range(0, 1000001)`, want: "ValueError"},
		{name: "zero division", input: `1 / 0`, want: "ZeroDivisionError"},
		{name: "name", input: `missing`, want: "NameError"},
		{name: "attribute", input: `struct Point { x: int }
Point{1}.missing`, want: "AttributeError"},
		{name: "key", input: `{"known": 1}["missing"]`, want: "KeyError"},
		{name: "index", input: `[1][2]`, want: "IndexError"},
		{name: "runtime", input: `struct Missing { message: str }
let read = fn() str | Missing { Missing{"not found"} }
let caller = fn() str { read() }
caller()`, want: "RuntimeError"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assertErrorStruct(t, testEval(test.input), test.want)
		})
	}
}

func TestFileLoadingProducesImportAndSyntaxErrors(t *testing.T) {
	directory := t.TempDir()
	assertErrorStruct(
		t,
		New().EvalFile(filepath.Join(directory, "missing.slv"), object.NewEnvironment()),
		"ImportError",
	)

	invalid := filepath.Join(directory, "invalid.slv")
	if err := os.WriteFile(invalid, []byte("let ="), 0600); err != nil {
		t.Fatal(err)
	}
	assertErrorStruct(t, New().EvalFile(invalid, object.NewEnvironment()), "SyntaxError")
}

func assertErrorStruct(t *testing.T, value object.Object, want string) *object.Error {
	t.Helper()
	err, ok := value.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want %s", value, want)
	}
	if err.Value == nil || err.Value.Struct == nil || err.Value.Struct.Name != want {
		t.Fatalf("error carries %#v, want %s", err.Value, want)
	}
	return err
}
