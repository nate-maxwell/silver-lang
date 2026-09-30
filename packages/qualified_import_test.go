package packages

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQualifiedExportsUseDeclaredPaths(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(dir, "nested", "module.slv"), "let value = 1")
	writeTestFile(t, filepath.Join(dir, "hidden.slv"), "let value = 2")
	manifestPath := filepath.Join(dir, "package.yaml")
	writeTestFile(t, manifestPath, "package: example\nauthors: []\nmembers: [hidden.slv]\nexport: [./nested/module.slv]\n")
	index := NewIndex()
	if err := index.Refresh(manifestPath); err != nil {
		t.Fatal(err)
	}
	resolved, manifest, found, err := index.Resolve("example:nested/module")
	if err != nil || !found || resolved != filepath.Join(dir, "nested", "module.slv") || manifest.Name() != "example" {
		t.Fatalf("qualified export = %q, %v, %v, %v", resolved, manifest, found, err)
	}
	for _, request := range []string{"example:module", "example:hidden", "other:nested/module", "Example:nested/module", "example:Nested/module"} {
		if _, _, found, err := index.Resolve(request); err != nil || found {
			t.Fatalf("Resolve(%q) = found %v, error %v", request, found, err)
		}
	}
}

func TestFirstRegisteredPackageOwnsItsNamespace(t *testing.T) {
	var manifests []string
	for _, module := range []string{"first", "second"} {
		dir := t.TempDir()
		writeTestFile(t, filepath.Join(dir, module+".slv"), "let value = 1")
		manifestPath := filepath.Join(dir, "package.yaml")
		writeTestFile(t, manifestPath, "package: example\nexport: ["+module+".slv]\n")
		manifests = append(manifests, manifestPath)
	}
	index := NewIndex()
	if err := index.Refresh(strings.Join(manifests, string(os.PathListSeparator))); err != nil {
		t.Fatal(err)
	}
	if _, _, found, err := index.Resolve("example:first"); err != nil || !found {
		t.Fatalf("first package did not resolve: %v, %v", found, err)
	}
	if _, _, found, err := index.Resolve("example:second"); err != nil || found {
		t.Fatalf("namespace merged with later package: %v, %v", found, err)
	}
}
