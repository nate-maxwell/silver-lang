package packages

import (
	"fmt"
	"path/filepath"
	"silver/source"
)

// Resolve searches for package exports in search-path
// order. An exported file may be addressed by its declared path or basename.
//
// A successful lookup always has a manifest. If no entry matches, Resolve returns an
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
		for _, manifest := range entry.manifests {
			for _, exported := range manifest.exports {
				if matchesExposedFile(wanted, filepath.Base(exported.path), exported.declared) {
					return exported.path, index.byFile[source.FileID(exported.path)], true, nil
				}
			}
		}
	}
	return "", nil, false, nil
}

// matchesExposedFile reports whether an import request matches an exposed
// file's basename or its manifest-declared relative path.
func matchesExposedFile(wanted, base, declared string) bool {
	wantedKey := source.PathKey(wanted)
	return wantedKey == source.PathKey(base) || declared != "" && wantedKey == source.PathKey(declared)
}
