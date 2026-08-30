package packages

import (
	"fmt"
	"os"
	"path/filepath"
)

// Resolve searches for explicit source entries and package exports in search-path
// order. An exported file may be addressed by its declared path or basename.
//
// The returned Manifest is nil for an individual source-file entry or a file
// found through a legacy directory. If no entry matches, Resolve returns an
// empty path, a nil Manifest, false, and a nil error. Resolve returns an error
// when the index has not been refreshed or its latest refresh failed.
func (index *Index) Resolve(request string) (string, *Manifest, bool, error) {
	index.mu.RLock()
	defer index.mu.RUnlock()
	if !index.initialized {
		return "", nil, false, fmt.Errorf("package index is not initialized")
	}
	if index.err != nil {
		return "", nil, false, index.err
	}

	wanted := cleanImportName(request)
	for _, entry := range index.entries {
		if entry.legacy {
			candidate := filepath.Clean(filepath.Join(entry.directory, request))
			if importCandidateExists(candidate) {
				return candidate, index.byFile[pathKey(candidate)], true, nil
			}
		}
		if entry.file != "" && matchesExposedFile(wanted, filepath.Base(entry.file), "") {
			return entry.file, nil, true, nil
		}
		for _, manifest := range entry.manifests {
			for _, exported := range manifest.exports {
				if matchesExposedFile(wanted, filepath.Base(exported.path), exported.declared) {
					return exported.path, manifest, true, nil
				}
			}
		}
	}
	return "", nil, false, nil
}

// matchesExposedFile reports whether an import request matches an exposed
// file's basename or its manifest-declared relative path.
func matchesExposedFile(wanted, base, declared string) bool {
	return wanted == cleanImportName(base) || declared != "" && wanted == cleanImportName(declared)
}

// importCandidateExists reports whether a legacy import candidate exists. It
// preserves non-not-found filesystem errors for the later module read.
func importCandidateExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}
