package packages

import (
	"fmt"
	"path"
	"strings"
)

// Resolve finds an exported <package>:<module> in registered package manifests.
func (index *Index) Resolve(request string) (string, *Manifest, bool, error) {
	return index.ResolveFrom(request, "")
}

// ResolveFrom also permits a package to import its own declared internal members.
// The first registered manifest with a matching name owns the entire namespace.
func (index *Index) ResolveFrom(request, importerPackageID string) (string, *Manifest, bool, error) {
	packageName, moduleName, err := ParseImport(request)
	if err != nil {
		return "", nil, false, err
	}
	index.mu.RLock()
	defer index.mu.RUnlock()
	if !index.initialized {
		return "", nil, false, fmt.Errorf("package index is not initialized")
	}
	if index.err != nil {
		return "", nil, false, index.err
	}
	for _, manifest := range index.entries {
		if manifest.Name() != packageName {
			continue
		}
		files := manifest.exports
		if manifest.ID() == importerPackageID {
			files = manifest.members
		}
		for _, file := range files {
			if strings.TrimSuffix(file.declared, path.Ext(file.declared)) == moduleName {
				return file.path, manifest, true, nil
			}
		}
		// A matching package claims the whole name even when this module is
		// absent or private. Falling through to a later same-named package
		// would mix two packages' public APIs and ownership rules.
		return "", nil, false, nil
	}
	return "", nil, false, nil
}
