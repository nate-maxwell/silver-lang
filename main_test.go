package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.monkey")
	if err := os.WriteFile(path, []byte(`let answer = 42;`), 0644); err != nil {
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

func TestRunFileReportsErrors(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{filepath.Join(t.TempDir(), "missing.monkey")}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "could not read") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
}
