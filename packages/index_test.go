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
	writeTestFile(t, filepath.Join(directory, "example.yaml"), `
package: example
members:
  - hidden.slv
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
	if index.ManifestFor(hidden) != index.ManifestFor(exposed) {
		t.Fatal("internal member has a different package owner")
	}
}

func TestIndexRejectsConflictingMembership(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "shared.slv"), "let value = 42")
	writeTestFile(t, filepath.Join(dir, "first.yaml"), "package: first\nexport: [shared.slv]\n")
	writeTestFile(t, filepath.Join(dir, "second.yaml"), "package: second\nexport: []\nmembers: [shared.slv]\n")
	index := NewIndex()
	if err := index.Refresh(dir); err == nil || !strings.Contains(err.Error(), "belongs to both") {
		t.Fatalf("conflicting package membership accepted: %v", err)
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
	if _, _, found, err := index.Resolve("helper.slv"); err != nil || found {
		t.Fatalf("internal helper exposed: %v, %v", found, err)
	}
	writeTestFile(t, filepath.Join(dir, "nested", "package.yaml"), "package: nested\nexport: []\n")
	unlisted := filepath.Join(dir, "nested", "unlisted.slv")
	writeTestFile(t, unlisted, "let value = 1")
	if owner, err := index.DiscoverFor(unlisted); err != nil || owner != nil {
		t.Fatalf("unlisted source acquired package: %v, %v", owner, err)
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

func TestIndexRequiresManifests(t *testing.T) {
	legacyDirectory := t.TempDir()
	legacyFile := filepath.Join(legacyDirectory, "legacy.slv")
	writeTestFile(t, legacyFile, "let value = 1")
	explicitDirectory := t.TempDir()
	explicitFile := filepath.Join(explicitDirectory, "explicit.slv")
	writeTestFile(t, explicitFile, "let value = 2")

	index := NewIndex()
	if err := index.Refresh(legacyDirectory); err != nil {
		t.Fatal(err)
	}
	if _, _, found, err := index.Resolve("legacy.slv"); err != nil || found {
		t.Fatalf("manifest-free directory exposed a file: found=%v, err=%v", found, err)
	}
	if err := index.Refresh(explicitFile); err == nil || !strings.Contains(err.Error(), "package YAML manifest") {
		t.Fatalf("source-file search entry was accepted: %v", err)
	}
	if _, _, _, err := index.Resolve("explicit.slv"); err == nil {
		t.Fatal("Resolve ignored failed refresh")
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
	for _, request := range []string{"identity.slv", "NESTED/IDENTITY.SLV", `nested\identity.slv`} {
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
