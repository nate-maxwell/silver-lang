package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	if err := os.WriteFile(path, []byte(`let answer = 42`), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{path}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run returned %d; stderr: %s", code, stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("file mode printed REPL output: %q", stdout.String())
	}
}

func TestRunFileInjectsStandardStreams(t *testing.T) {
	path := filepath.Join(t.TempDir(), "streams.slv")
	source := `let io = import("io")
let contents = io.stdin.read()
io.stdout.write("stdout:" + contents)
io.stderr.write("stderr:" + contents)`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{path}, strings.NewReader("input"), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run returned %d; stderr: %s", code, stderr.String())
	}
	if got, want := stdout.String(), "stdout:input"; got != want {
		t.Fatalf("stdout is %q, want %q", got, want)
	}
	if got, want := stderr.String(), "stderr:input"; got != want {
		t.Fatalf("stderr is %q, want %q", got, want)
	}
}

func TestRunFileReportsErrors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{filepath.Join(t.TempDir(), "missing.slv")}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "could not read") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}

func TestRunFileReportsTraceback(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	if err := os.WriteFile(path, []byte(`let fail = fn() {
    1 + True
}
fail()
`), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{path}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run returned %d, want 1", code)
	}
	traceback := stderr.String()
	for _, part := range []string{
		"Traceback (most recent call last):",
		fmt.Sprintf("File \"%s\", line 4, column 1, in <module>", path),
		fmt.Sprintf("File \"%s\", line 2, column 7, in fail", path),
		"TypeError: type mismatch: INTEGER + BOOLEAN",
	} {
		if !strings.Contains(traceback, part) {
			t.Fatalf("traceback does not contain %q:\n%s", part, traceback)
		}
	}
}

func TestRunFileReportsUnhandledStructError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	source := `struct Missing { message: str }
let read = fn() str | Missing { Missing{"not found"} }
read()
`
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{path}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run returned %d, want 1", code)
	}
	for _, part := range []string{"Traceback (most recent call last):", "Missing: not found"} {
		if !strings.Contains(stderr.String(), part) {
			t.Fatalf("unhandled error does not contain %q:\n%s", part, stderr.String())
		}
	}
}
