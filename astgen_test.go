package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunASTGenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	if err := os.WriteFile(path, []byte(`let answer = 42`), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"astgen", path}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run returned %d; stderr: %s", code, stderr.String())
	}
	cachePath := path + ".astc"
	if _, err := os.Stat(cachePath); err != nil {
		t.Fatalf("cache was not generated: %s", err)
	}
	if got, want := strings.TrimSpace(stdout.String()), cachePath; got != want {
		t.Fatalf("stdout is %q, want %q", got, want)
	}
}

func TestRunASTGenDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	sources := []string{
		filepath.Join(root, "one.slv"),
		filepath.Join(nested, "two.slv"),
	}
	for _, path := range sources {
		if err := os.WriteFile(path, []byte(`let answer = 42`), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "ignored.txt"), []byte(`let answer = 42`), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"astgen", root}, strings.NewReader(""), &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run returned %d; stderr: %s", code, stderr.String())
	}
	for _, path := range sources {
		if _, err := os.Stat(path + ".astc"); err != nil {
			t.Errorf("cache for %s was not generated: %s", path, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "ignored.txt.astc")); !os.IsNotExist(err) {
		t.Fatalf("non-Silver file generated a cache; stat error: %v", err)
	}
}

func TestRunASTGenReportsParseErrors(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.slv")
	if err := os.WriteFile(path, []byte(`let =`), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := run([]string{"astgen", path}, strings.NewReader(""), &stdout, &stderr)
	if code != 1 {
		t.Fatalf("run returned %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "could not parse") {
		t.Fatalf("unexpected stderr: %q", stderr.String())
	}
	if _, err := os.Stat(path + ".astc"); !os.IsNotExist(err) {
		t.Fatalf("invalid source generated a cache; stat error: %v", err)
	}
}

func TestRunASTGenRequiresOnePath(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"astgen"}, strings.NewReader(""), &stdout, &stderr)
	if code != 2 {
		t.Fatalf("run returned %d, want 2", code)
	}
	if got, want := strings.TrimSpace(stderr.String()), "usage: silver astgen <path>"; got != want {
		t.Fatalf("stderr is %q, want %q", got, want)
	}
}
