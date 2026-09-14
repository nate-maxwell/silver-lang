package source

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestFileIdentity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Identity.slv")
	for _, alias := range []string{dir + "/./Identity.slv", dir + "/nested/../Identity.slv", filepath.ToSlash(path)} {
		if FileID(alias) != FileID(path) {
			t.Errorf("%q and %q have different identities", path, alias)
		}
	}
	lower := filepath.Join(dir, "identity.slv")
	if equal, want := FileID(path) == FileID(lower), runtime.GOOS == "windows"; equal != want {
		t.Errorf("case variants have equal identities: %t, want %t", equal, want)
	}
	if FileID(path) == FileID(filepath.Join(dir, "elsewhere", "Identity.slv")) {
		t.Error("different directories have the same module identity")
	}
}

func TestBundledIdentityIsCaseSensitiveAndSeparateFromFiles(t *testing.T) {
	if BundledID("io") == BundledID("IO") {
		t.Error("bundled names must be case-sensitive")
	}
	path := filepath.Join(t.TempDir(), "io")
	if FileID(path) == BundledID(path) {
		t.Error("file and bundled modules have the same identity")
	}
}
