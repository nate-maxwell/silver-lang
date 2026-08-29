// Path normalization helpers for handling packages.
// They keep things like ./nested/file.slv and nested\file.slv comparable.

package packages

import (
	"path/filepath"
	"runtime"
	"strings"
)

// cleanImportName returns the platform-independent form used to compare import
// requests, basenames, and manifest-declared paths.
func cleanImportName(path string) string {
	cleaned := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	return strings.TrimPrefix(cleaned, "./")
}

// pathKey returns the normalized form used for filesystem identity. Paths are
// case-insensitive on Windows, matching the host filesystem convention.
func pathKey(path string) string {
	key := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return key
}
