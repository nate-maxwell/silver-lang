package packages

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	writeTestFile(t, filepath.Join(directory, "package.yaml"), `
package: example
members:
  - hidden.slv
export:
  - nested/library.slv
`)

	index := NewIndex()
	if err := index.Refresh(filepath.Join(directory, "package.yaml")); err != nil {
		t.Fatal(err)
	}
	for _, request := range []string{"example:nested/library"} {
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
	if _, _, found, err := index.Resolve("example:hidden"); err != nil || found {
		t.Fatalf("hidden file resolved: found=%v, err=%v", found, err)
	}
	if index.ManifestFor(hidden) != index.ManifestFor(exposed) {
		t.Fatal("internal member has a different package owner")
	}
}

func TestIndexRejectsConflictingMembership(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(nested, "shared.slv"), "let value = 42")
	first := filepath.Join(dir, "package.yaml")
	second := filepath.Join(nested, "package.yaml")
	writeTestFile(t, first, "package: first\nexport: [nested/shared.slv]\n")
	writeTestFile(t, second, "package: second\nexport: []\nmembers: [shared.slv]\n")
	if err := NewIndex().Refresh(first + string(os.PathListSeparator) + second); err == nil || !strings.Contains(err.Error(), "belongs to both") {
		t.Fatalf("conflicting ownership: %v", err)
	}
}

func TestDiscoverForFindsLocalMembersWithoutExposingThem(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(dir, "nested", "helper.slv")
	writeTestFile(t, helper, "let value = 42")
	writeTestFile(t, filepath.Join(dir, "package.yaml"), "package: library\nexport: []\nmembers: [nested/helper.slv]\n")
	index := NewIndex()
	if err := index.Refresh(""); err != nil {
		t.Fatal(err)
	}
	owner, err := index.DiscoverFor(helper)
	if err != nil || owner == nil || owner.Name() != "library" {
		t.Fatalf("DiscoverFor = %v, %v", owner, err)
	}
	if _, _, found, err := index.Resolve("library:nested/helper"); err != nil || found {
		t.Fatalf("internal helper exposed: %v, %v", found, err)
	}
	writeTestFile(t, filepath.Join(dir, "nested", "package.yaml"), "package: nested\nexport: []\n")
	unlisted := filepath.Join(dir, "nested", "unlisted.slv")
	writeTestFile(t, unlisted, "let value = 1")
	if owner, err := index.DiscoverFor(unlisted); err != nil || owner != nil {
		t.Fatalf("unlisted source acquired package: %v, %v", owner, err)
	}
}

func TestIndexRequiresPackageYAMLFilename(t *testing.T) {
	for _, name := range []string{"other.yaml", "package.yml", "package.slv"} {
		t.Run(name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), name)
			writeTestFile(t, filename, "package: example\nexport: []\n")
			if err := NewIndex().Refresh(filename); err == nil {
				t.Fatalf("accepted %s", name)
			}
		})
	}
}

func TestIndexRequiresManifests(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "library.slv"), "let value = 42")
	// A directory is invalid even when it contains a valid manifest.
	writeTestFile(t, filepath.Join(dir, "package.yaml"), "package: example\nexport: [library.slv]\n")
	for _, entry := range []string{dir, t.TempDir(), filepath.Join(dir, "library.slv"), filepath.Join(t.TempDir(), "package.yaml")} {
		index := NewIndex()
		if err := index.Refresh(entry); err == nil {
			t.Fatalf("accepted entry %q", entry)
		}
		if _, _, _, err := index.Resolve("example:library"); err == nil {
			t.Fatal("Resolve ignored failed refresh")
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
	writeTestFile(t, filepath.Join(firstDirectory, "package.yaml"), "package: first\nexport: [first.slv]\n")
	writeTestFile(t, filepath.Join(secondDirectory, "package.yaml"), "package: second\nexport: [second.slv]\n")

	index := NewIndex()
	if err := index.Refresh(filepath.Join(firstDirectory, "package.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, _, found, _ := index.Resolve("first:first"); !found {
		t.Fatal("first search path did not resolve")
	}
	if err := index.Refresh(filepath.Join(secondDirectory, "package.yaml")); err != nil {
		t.Fatal(err)
	}
	if _, _, found, _ := index.Resolve("first:first"); found {
		t.Fatal("stale first search path still resolves")
	}
	if path, _, found, err := index.Resolve("second:second"); err != nil || !found || path != second {
		t.Fatalf("second search path resolved as %q, %v, %v", path, found, err)
	}
}

func TestWindowsPackagePathIdentity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path casing policy")
	}
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "Nested"), 0755); err != nil {
		t.Fatal(err)
	}
	exposed := filepath.Join(dir, "Nested", "Identity.slv")
	manifestPath := filepath.Join(dir, "Package.yaml")
	writeTestFile(t, exposed, "let value = 42")
	writeTestFile(t, manifestPath, "package: example\nexport:\n  - Nested/Identity.slv\n")
	index := NewIndex()
	if err := index.Refresh(manifestPath); err != nil {
		t.Fatal(err)
	}
	owner := index.ManifestFor(exposed)
	if owner == nil || index.ManifestFor(strings.ToUpper(exposed)) != owner {
		t.Fatal("case variant lost its package owner")
	}
	for _, request := range []string{"example:Nested/Identity"} {
		path, manifest, found, err := index.Resolve(request)
		if err != nil || !found || path != exposed || manifest != owner {
			t.Fatalf("Resolve(%q) = %q, %v, %v, %v; want canonical package export", request, path, manifest, found, err)
		}
	}
	alias, err := ReadManifest(strings.ToUpper(manifestPath))
	if err != nil || alias.ID() != owner.ID() {
		t.Fatalf("manifest alias has a different identity: %v, %v", alias, err)
	}
	writeTestFile(t, manifestPath, "package: example\nexport:\n  - Nested/Identity.slv\n  - nested/identity.slv\n")
	if _, err := ReadManifest(manifestPath); err == nil || !strings.Contains(err.Error(), "duplicate export") {
		t.Fatalf("duplicate case variant export: %v", err)
	}
}
