package evaluator

import (
	"os"
	"path/filepath"
	"silver/object"
	"strconv"
	"testing"
)

func TestOpenReturnsReadableFileStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("hello, Silver"), 0600); err != nil {
		t.Fatal(err)
	}

	input := `let file: File = open(` + silverString(path) + `)
let reader: call() str | IOError = file.read
let contents = reader()
file.close()
contents`
	result, ok := testEval(input).(*object.String)
	if !ok {
		t.Fatalf("result is not a string")
	}
	if got, want := result.Value, "hello, Silver"; got != want {
		t.Fatalf("contents are %q, want %q", got, want)
	}
}

func TestFileWriteReplacesContents(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("old contents"), 0600); err != nil {
		t.Fatal(err)
	}

	input := `let file = open(` + silverString(path) + `)
let writer: call(contents: str) | IOError = file.write
writer("new contents")
let contents = file.read()
file.close()
contents`
	result, ok := testEval(input).(*object.String)
	if !ok {
		t.Fatalf("result is not a string")
	}
	if got, want := result.Value, "new contents"; got != want {
		t.Fatalf("contents are %q, want %q", got, want)
	}
	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(onDisk), "new contents"; got != want {
		t.Fatalf("disk contents are %q, want %q", got, want)
	}
}

func TestFilePathAndNominalTypeAreExposed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, nil, 0600); err != nil {
		t.Fatal(err)
	}

	input := `let file = open(` + silverString(path) + `)
let matches = type(file) == File
let path_matches = file.path == ` + silverString(path) + `
file.close()
if matches { path_matches } else { False }`
	testBooleanObject(t, testEval(input), true)
}

func TestFileOperationsAfterCloseReturnIOError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("contents"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []string{
		`type(file.read()) == IOError`,
		`type(file.write("replacement")) == IOError`,
		`type(file.close()) == IOError`,
	}
	for _, assertion := range tests {
		input := `let file = open(` + silverString(path) + `)
file.close()
` + assertion
		testBooleanObject(t, testEval(input), true)
	}
}

func TestOpenReturnsStructuredErrors(t *testing.T) {
	directory := t.TempDir()
	missing := filepath.Join(directory, "missing.txt")
	testBooleanObject(t, testEval(`type(open(`+silverString(missing)+`)) == FileNotFound`), true)
	testBooleanObject(t, testEval(`type(open(`+silverString(directory)+`)) == PermissionDenied`), true)

	result := testEval(`open(` + silverString(missing) + `).message`)
	message, ok := result.(*object.String)
	if !ok || message.Value == "" {
		t.Fatalf("missing-file message is %#v, want a non-empty string", result)
	}
}

func TestOpenHasDeclaredCallSignature(t *testing.T) {
	input := `let opener: call(path: str) File | FileNotFound | PermissionDenied = open
type(opener) == call`
	testBooleanObject(t, testEval(input), true)
}

func TestOpenRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `open()`, message: "wrong number of arguments. got=0, want=1"},
		{input: `open(1)`, message: "argument to `open` must be STRING, got INTEGER"},
	}
	for _, tt := range tests {
		result, ok := testEval(tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if result.Message != tt.message {
			t.Fatalf("error is %q, want %q", result.Message, tt.message)
		}
	}
}

func silverString(path string) string {
	return strconv.Quote(filepath.ToSlash(path))
}
