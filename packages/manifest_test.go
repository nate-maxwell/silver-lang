package packages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestReadManifest(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	first := filepath.Join(directory, "first.slv")
	second := filepath.Join(directory, "nested", "second.slv")
	writeTestFile(t, first, "let first = 1")
	writeTestFile(t, second, "let second = 2")
	manifestPath := filepath.Join(directory, "example.yaml")
	writeTestFile(t, manifestPath, `
# Package comments are ignored.
package: example_2
authors: [Ada, "Grace Hopper"]
export:
  - ./first.slv
  - nested/second.slv # Export comments are ignored too.
`)

	manifest, err := ReadManifest(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := manifest.Name(), "example_2"; got != want {
		t.Fatalf("name is %q, want %q", got, want)
	}
	authors := manifest.Authors()
	if len(authors) != 2 || authors[0] != "Ada" || authors[1] != "Grace Hopper" {
		t.Fatalf("unexpected authors: %v", authors)
	}
	authors[0] = "changed"
	if manifest.Authors()[0] != "Ada" {
		t.Fatal("Authors exposed mutable manifest state")
	}
	if got, want := manifest.Path(), filepath.Clean(manifestPath); got != want {
		t.Fatalf("path is %q, want %q", got, want)
	}
	if got, want := manifest.Root(), directory; got != want {
		t.Fatalf("root is %q, want %q", got, want)
	}
	if !strings.HasPrefix(manifest.ID(), "package:") {
		t.Fatalf("ID is %q, want package identity", manifest.ID())
	}

	exports := manifest.Exports()
	if len(exports) != 2 {
		t.Fatalf("exports are %#v, want two entries", exports)
	}
	if got, want := exports[0].Declared(), "first.slv"; got != want {
		t.Fatalf("first declared path is %q, want %q", got, want)
	}
	if got, want := exports[0].Path(), first; got != want {
		t.Fatalf("first resolved path is %q, want %q", got, want)
	}
	if got, want := exports[1].Declared(), "nested/second.slv"; got != want {
		t.Fatalf("second declared path is %q, want %q", got, want)
	}
	if got, want := exports[1].Path(), second; got != want {
		t.Fatalf("second resolved path is %q, want %q", got, want)
	}

	// Callers cannot mutate the manifest through the returned slice.
	exports[0] = Export{}
	if got := manifest.Exports()[0].Path(); got != first {
		t.Fatalf("manifest export changed through returned slice: %q", got)
	}
}

func TestDecodeManifestRejectsInvalidYAML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		message string
	}{
		{name: "empty", input: "", message: "manifest is empty"},
		{name: "missing package", input: "export: []", message: `missing required field "package"`},
		{name: "invalid package name", input: "package: 2example\nexport: []", message: "package name"},
		{name: "missing export", input: "package: example", message: `missing required field "export"`},
		{name: "unknown field", input: "package: example\nexport: []\nunexpected: true", message: "field unexpected not found"},
		{name: "package is not a string", input: "package: true\nexport: []", message: "must be a string"},
		{name: "export is not a list", input: "package: example\nexport: source.slv", message: "cannot unmarshal"},
		{name: "export path is not a string", input: "package: example\nexport: [42]", message: "must be a string"},
		{name: "members is not a list", input: "package: example\nexport: []\nmembers: helper.slv", message: "cannot unmarshal"},
		{name: "authors is not a list", input: "package: example\nexport: []\nauthors: Ada", message: "cannot unmarshal"},
		{name: "author is not a string", input: "package: example\nexport: []\nauthors: [42]", message: "must be a string"},
		{name: "member path is not a string", input: "package: example\nexport: []\nmembers: [42]", message: "must be a string"},
		{name: "multiple documents", input: "package: example\nexport: []\n---\npackage: other\nexport: []", message: "multiple YAML documents"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := decodeManifest("example.yaml", []byte(test.input))
			if err == nil {
				t.Fatal("decode succeeded, want an error")
			}
			if !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error is %q, want %q", err, test.message)
			}
		})
	}
}

func TestManifestSeparatesMembersFromExports(t *testing.T) {
	files := fstest.MapFS{
		"library/package.yaml": {Data: []byte("package: library\nmembers: [api.slv, helper.slv]\nexport: [api.slv]\n")},
		"library/api.slv":      {Data: []byte("let value = 42")},
		"library/helper.slv":   {Data: []byte("let value = 21")},
	}
	manifest, err := ReadManifestFS(files, "library/package.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if got := manifest.Exports(); len(got) != 1 || got[0].Path() != "library/api.slv" {
		t.Fatalf("exports = %#v, want only api.slv", got)
	}
	members := manifest.Members()
	if len(members) != 2 || members[1].Path() != "library/helper.slv" {
		t.Fatalf("members = %#v, want api.slv and helper.slv once each", members)
	}
	members[0] = File{}
	if manifest.Members()[0].Path() != "library/api.slv" {
		t.Fatal("Members exposed mutable manifest state")
	}
}

func TestReadManifestRejectsInvalidMembers(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "source.slv"), "let value = 1")
	writeTestFile(t, filepath.Join(dir, "wrong.txt"), "not Silver")
	for _, test := range []struct{ name, members, message string }{
		{"outside root", "[../outside.slv]", "leaves the package root"},
		{"absolute", "[" + filepath.ToSlash(filepath.Join(dir, "source.slv")) + "]", "must be relative"},
		{"missing", "[missing.slv]", "could not access member"},
		{"wrong extension", "[wrong.txt]", "is not a .slv file"},
		{"directory", "[.]", "not a regular file"},
		{"duplicate", "[source.slv, ./source.slv]", "duplicate member"},
	} {
		t.Run(test.name, func(t *testing.T) {
			filename := filepath.Join(dir, "package.yaml")
			writeTestFile(t, filename, "package: example\nexport: []\nmembers: "+test.members+"\n")
			if _, err := ReadManifest(filename); err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("ReadManifest error = %v, want %q", err, test.message)
			}
		})
	}
}

func TestReadManifestRejectsInvalidExports(t *testing.T) {
	directory := t.TempDir()
	writeTestFile(t, filepath.Join(directory, "source.slv"), "let value = 1")
	writeTestFile(t, filepath.Join(directory, "wrong.txt"), "not Silver")

	tests := []struct {
		name       string
		exportYAML string
		message    string
	}{
		{name: "outside root", exportYAML: "  - ../outside.slv", message: "leaves the package root"},
		{name: "missing", exportYAML: "  - missing.slv", message: "could not access export"},
		{name: "wrong extension", exportYAML: "  - wrong.txt", message: "is not a .slv file"},
		{name: "duplicate", exportYAML: "  - source.slv\n  - ./source.slv", message: "duplicate export"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			manifestPath := filepath.Join(directory, strings.ReplaceAll(test.name, " ", "_")+".yaml")
			writeTestFile(t, manifestPath, "package: example\nexport:\n"+test.exportYAML)
			_, err := ReadManifest(manifestPath)
			if err == nil || !strings.Contains(err.Error(), test.message) {
				t.Fatalf("error is %v, want %q", err, test.message)
			}
		})
	}
}

func writeTestFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}
