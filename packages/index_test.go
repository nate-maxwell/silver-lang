package packages

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexResolvesOnlyManifestExports(t *testing.T) {
	directory := t.TempDir()
	if err := os.Mkdir(filepath.Join(directory, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	exposed := filepath.Join(directory, "nested", "library.slv")
	hidden := filepath.Join(directory, "hidden.slv")
	writeTestFile(t, exposed, "let value = 42")
	writeTestFile(t, hidden, "let value = 99")
	writeTestFile(t, filepath.Join(directory, "example.yaml"), `
package: example
export:
  - nested/library.slv
`)

	index := NewIndex()
	if err := index.Refresh(directory); err != nil {
		t.Fatal(err)
	}
	for _, request := range []string{"nested/library.slv", "library.slv"} {
		path, manifest, found, err := index.Resolve(request)
		if err != nil {
			t.Fatal(err)
		}
		if !found || path != exposed || manifest == nil || manifest.Name() != "example" {
			t.Fatalf("Resolve(%q) = %q, %#v, %v; want exposed package file", request, path, manifest, found)
		}
	}
	if manifest := index.ManifestFor(exposed); manifest == nil || manifest.Name() != "example" {
		t.Fatalf("ManifestFor(exposed) = %#v, want example manifest", manifest)
	}
	if _, _, found, err := index.Resolve("hidden.slv"); err != nil || found {
		t.Fatalf("hidden file resolved: found=%v, err=%v", found, err)
	}
}

func TestIndexAcceptsYAMLManifestSuffixes(t *testing.T) {
	for _, suffix := range []string{".yaml", ".yml"} {
		t.Run(suffix, func(t *testing.T) {
			directory := t.TempDir()
			source := filepath.Join(directory, "library.slv")
			writeTestFile(t, source, "let value = 42")
			manifestPath := filepath.Join(directory, "example"+suffix)
			writeTestFile(t, manifestPath, "package: example\nexport:\n  - library.slv\n")

			index := NewIndex()
			if err := index.Refresh(manifestPath); err != nil {
				t.Fatal(err)
			}
			path, manifest, found, err := index.Resolve("library.slv")
			if err != nil || !found || path != source || manifest == nil {
				t.Fatalf("Resolve(library.slv) = %q, %#v, %v, %v; want %q", path, manifest, found, err, source)
			}
		})
	}
}

func TestIndexSupportsLegacyDirectoriesAndExplicitFiles(t *testing.T) {
	legacyDirectory := t.TempDir()
	legacyFile := filepath.Join(legacyDirectory, "legacy.slv")
	writeTestFile(t, legacyFile, "let value = 1")
	explicitDirectory := t.TempDir()
	explicitFile := filepath.Join(explicitDirectory, "explicit.slv")
	writeTestFile(t, explicitFile, "let value = 2")

	index := NewIndex()
	searchPath := legacyDirectory + string(os.PathListSeparator) + explicitFile
	if err := index.Refresh(searchPath); err != nil {
		t.Fatal(err)
	}
	for request, want := range map[string]string{
		"legacy.slv":   legacyFile,
		"explicit.slv": explicitFile,
	} {
		path, manifest, found, err := index.Resolve(request)
		if err != nil || !found || path != want || manifest != nil {
			t.Fatalf("Resolve(%q) = %q, %#v, %v, %v; want %q", request, path, manifest, found, err, want)
		}
	}
}

func TestIndexRefreshesWhenSearchPathChanges(t *testing.T) {
	firstDirectory := t.TempDir()
	secondDirectory := t.TempDir()
	first := filepath.Join(firstDirectory, "first.slv")
	second := filepath.Join(secondDirectory, "second.slv")
	writeTestFile(t, first, "let value = 1")
	writeTestFile(t, second, "let value = 2")

	index := NewIndex()
	if err := index.Refresh(firstDirectory); err != nil {
		t.Fatal(err)
	}
	if _, _, found, _ := index.Resolve("first.slv"); !found {
		t.Fatal("first search path did not resolve")
	}
	if err := index.Refresh(secondDirectory); err != nil {
		t.Fatal(err)
	}
	if _, _, found, _ := index.Resolve("first.slv"); found {
		t.Fatal("stale first search path still resolves")
	}
	if path, _, found, err := index.Resolve("second.slv"); err != nil || !found || path != second {
		t.Fatalf("second search path resolved as %q, %v, %v", path, found, err)
	}
}
