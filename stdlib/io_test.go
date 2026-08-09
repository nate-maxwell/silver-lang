package stdlib_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"silver/object"
	"strconv"
	"strings"
	"testing"
)

const ioImport = "let io = import(\"io\")\n"

func TestOpenReturnsReadableFileStruct(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("hello, Silver"), 0600); err != nil {
		t.Fatal(err)
	}

	input := `let file: File = io.open(` + silverString(path) + `)
let reader: call() str | IOError = file.read
let contents = reader()
file.close()
contents`
	result, ok := testEval(ioImport + input).(*object.String)
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

	input := `let file = io.open(` + silverString(path) + `)
let writer: call(contents: str) | IOError = file.write
writer("new contents")
let contents = file.read()
file.close()
contents`
	result, ok := testEval(ioImport + input).(*object.String)
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

	input := `let file = io.open(` + silverString(path) + `)
let matches = core.type(file) == File
let path_matches = file.path == ` + silverString(path) + `
file.close()
if matches { path_matches } else { False }`
	testBooleanObject(t, testEval(ioImport+coreImport+input), true)
}

func TestFileOperationsAfterCloseReturnIOError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "message.txt")
	if err := os.WriteFile(path, []byte("contents"), 0600); err != nil {
		t.Fatal(err)
	}

	tests := []string{
		`file.read()`,
		`file.write("replacement")`,
		`file.close()`,
	}
	for _, operation := range tests {
		input := `let file = io.open(` + silverString(path) + `)
file.close()
try {
` + operation + `
False
} catch IOError err {
err.message != ""
}`
		testBooleanObject(t, testEval(ioImport+input), true)
	}
}

func TestOpenReturnsStructuredErrors(t *testing.T) {
	directory := t.TempDir()
	missing := filepath.Join(directory, "missing.txt")
	testBooleanObject(t, testEval(ioImport+`try {
io.open(`+silverString(missing)+`)
} catch FileNotFound err {
err.message != ""
}`), true)
	testBooleanObject(t, testEval(ioImport+`try {
io.open(`+silverString(directory)+`)
} catch PermissionDenied err {
err.message != ""
}`), true)

	result := testEval(ioImport + `try {
io.open(` + silverString(missing) + `)
} catch FileNotFound err {
err.message
}`)
	message, ok := result.(*object.String)
	if !ok || message.Value == "" {
		t.Fatalf("missing-file message is %#v, want a non-empty string", result)
	}
}

func TestOpenHasDeclaredCallSignature(t *testing.T) {
	input := `let opener: call(path: str) File | FileNotFound | PermissionDenied = io.open
core.type(opener) == call`
	testBooleanObject(t, testEval(ioImport+coreImport+input), true)
}

func TestOpenRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `io.open()`, message: "wrong number of arguments. got=0, want=1"},
		{input: `io.open(1)`, message: "argument to `open` must be STRING, got INTEGER"},
	}
	for _, tt := range tests {
		result, ok := testEval(ioImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if result.MessageText() != tt.message {
			t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
		}
	}
}

func TestStandardStreamsUseInjectedIO(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := ioImport + `let contents = io.stdin.read()
io.stdout.write("out:" + contents)
io.stderr.write("err:" + contents)
contents`
	result, ok := testEvalWithStreams(input, strings.NewReader("hello"), &stdout, &stderr).(*object.String)
	if !ok || result.Value != "hello" {
		t.Fatalf("stdin read returned %#v, want hello", result)
	}
	if got, want := stdout.String(), "out:hello"; got != want {
		t.Fatalf("stdout is %q, want %q", got, want)
	}
	if got, want := stderr.String(), "err:hello"; got != want {
		t.Fatalf("stderr is %q, want %q", got, want)
	}
}

func TestStandardStreamsExposeTypesAndSignatures(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := ioImport + coreImport + `let input: IOStream = io.stdin
let output: IOStream = io.stdout
let errors: IOStream = io.stderr
let read: call() str | IOError = input.read
let write_out: call(data: str) | IOError = output.write
let write_err: call(data: str) | IOError = errors.write
core.type(input) == IOStream`
	testBooleanObject(t, testEvalWithStreams(input, strings.NewReader(""), &stdout, &stderr), true)
}

func TestStandardStreamsRejectUnsupportedDirections(t *testing.T) {
	var stdout, stderr bytes.Buffer
	tests := []string{
		`try { io.stdin.write("data") } catch IOError err { err.message != "" }`,
		`try { io.stdout.read() } catch IOError err { err.message != "" }`,
		`try { io.stderr.read() } catch IOError err { err.message != "" }`,
	}
	for _, input := range tests {
		testBooleanObject(t, testEvalWithStreams(ioImport+input, strings.NewReader(""), &stdout, &stderr), true)
	}
}

func TestStandardStreamsReturnIOErrors(t *testing.T) {
	result := testEvalWithStreams(ioImport+`try {
io.stdin.read()
} catch IOError err {
err.message
}`, failingReader{}, &bytes.Buffer{}, &bytes.Buffer{})
	message, ok := result.(*object.String)
	if !ok || !strings.Contains(message.Value, "read failed") {
		t.Fatalf("read failure returned %#v", result)
	}

	result = testEvalWithStreams(ioImport+`try {
io.stdout.write("data")
} catch IOError err {
err.message
}`, strings.NewReader(""), failingWriter{}, &bytes.Buffer{})
	message, ok = result.(*object.String)
	if !ok || !strings.Contains(message.Value, "write failed") {
		t.Fatalf("write failure returned %#v", result)
	}
}

func TestStandardStreamsRejectInvalidArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	tests := []struct {
		input   string
		message string
	}{
		{input: `io.stdin.read(1)`, message: "wrong number of arguments. got=1, want=0"},
		{input: `io.stdout.write()`, message: "wrong number of arguments. got=0, want=1"},
		{input: `io.stderr.write(1)`, message: "argument to `IOStream.write` must be STRING, got INTEGER"},
	}
	for _, tt := range tests {
		result, ok := testEvalWithStreams(ioImport+tt.input, strings.NewReader(""), &stdout, &stderr).(*object.Error)
		if !ok || result.MessageText() != tt.message {
			t.Fatalf("%s returned %#v, want %q", tt.input, result, tt.message)
		}
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func silverString(path string) string {
	return strconv.Quote(filepath.ToSlash(path))
}
