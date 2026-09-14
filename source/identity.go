// Package source defines the identity policy shared by module loading and
// package discovery. Diagnostic paths retain their original spelling.
package source

import (
	"path/filepath"
	"runtime"
	"strings"
)

// ModuleID identifies a resolved source module independently of its import
// spelling. File and bundled modules occupy separate namespaces.
type ModuleID string

// FileID identifies a file whose path has already been resolved to an absolute
// path. Resolution belongs to the importer so relative paths use its directory.
func FileID(absolutePath string) ModuleID {
	return ModuleID("file:" + PathKey(absolutePath))
}

// BundledID identifies a native or embedded standard-library module. Bundled
// names are case-sensitive on every platform.
func BundledID(name string) ModuleID {
	return ModuleID("stdlib:" + name)
}

// PathKey normalizes filesystem paths for identity and name comparisons. It
// cleans separators and dot segments, and folds casing on Windows. Callers
// must resolve paths before using keys as absolute identities. Symlinks and
// hard links are not resolved; identity is lexical, not based on file handles.
func PathKey(path string) string {
	key := filepath.Clean(path)
	if runtime.GOOS == "windows" {
		key = strings.ToLower(key)
	}
	return key
}
