package packages

import (
	"fmt"
	"os"
	"path/filepath"
	"silver/source"
	"strings"
	"sync"
)

// Index is a concurrency-safe snapshot of packages from a
// platform-separated search path such as SILVER_PATH.
//
// The zero value is ready for use, though Refresh must be called before the
// first call to Resolve.
//
// Typical flow is: package.yaml -> ReadManifest -> Index.Refresh -> Index.Resolve -> Evaluator loads module.
type Index struct {
	mu          sync.RWMutex
	initialized bool
	searchPath  string
	entries     []*Manifest
	byFile      map[source.ModuleID]*Manifest
	err         error
}

// NewIndex returns an empty package index. The caller must call Refresh before
// the first call to Resolve.
func NewIndex() *Index {
	return &Index{byFile: make(map[source.ModuleID]*Manifest)}
}

// Refresh loads the entries in the platform-separated searchPath.
//
// Every entry must point to an existing package.yaml file. Refresh is a
// no-op when searchPath has not changed and returns the result of the previous
// load in that case.
// This is keyed by the search-path string, not file modification times: edits
// to manifests or repairs after a failed load do not invalidate the snapshot.
func (index *Index) Refresh(searchPath string) error {
	index.mu.Lock()
	defer index.mu.Unlock()
	if index.initialized && index.searchPath == searchPath {
		return index.err
	}

	index.initialized = true
	index.searchPath = searchPath
	index.entries = nil
	index.byFile = make(map[source.ModuleID]*Manifest)
	index.err = nil

	for _, rawEntry := range filepath.SplitList(searchPath) {
		if strings.TrimSpace(rawEntry) == "" {
			continue
		}
		manifest, err := ReadRegisteredManifest(rawEntry)
		if err != nil {
			index.err = err
			return err
		}
		index.entries = append(index.entries, manifest)
		if err := index.addManifest(manifest); err != nil {
			index.err = err
			return err
		}
	}
	return nil
}

// ReadRegisteredManifest validates a SILVER_PATH entry. Only an existing
// package.yaml file may be registered; directories and source files are errors.
func ReadRegisteredManifest(filename string) (*Manifest, error) {
	if source.PathKey(filepath.Base(filename)) != source.PathKey(manifestFilename) {
		return nil, fmt.Errorf("search-path entry %q must be a package YAML manifest named package.yaml", filename)
	}
	info, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("required package YAML manifest %q is unavailable: %w", filename, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("search-path entry %q must be a regular package.yaml file", filename)
	}
	return ReadManifest(filename)
}

// addManifest records ownership of every member and rejects ambiguous ownership.
// The caller holds index.mu.
func (index *Index) addManifest(manifest *Manifest) error {
	if index.byFile == nil {
		index.byFile = make(map[source.ModuleID]*Manifest)
	}
	for _, member := range manifest.members {
		if owner := index.byFile[source.FileID(member.path)]; owner != nil && owner.ID() != manifest.ID() {
			return fmt.Errorf("source %q belongs to both package manifests %q and %q", member.path, owner.Path(), manifest.Path())
		}
	}
	// Validate all ownership conflicts before inserting anything from this
	// manifest, so a rejected manifest cannot leave some members registered.
	for _, member := range manifest.members {
		key := source.FileID(member.path)
		if index.byFile[key] == nil {
			index.byFile[key] = manifest
		}
	}
	return nil
}

// ManifestFor returns the manifest that owns path, including internal members.
// It returns nil when path is not a member in the current snapshot.
func (index *Index) ManifestFor(path string) *Manifest {
	index.mu.RLock()
	defer index.mu.RUnlock()
	return index.byFile[source.FileID(path)]
}

// DiscoverFor finds an entry script's package when it is run directly
// without SILVER_PATH. The nearest package.yaml is the local package boundary;
// only its declared members acquire package identity. This does not register
// the package for imports; its package.yaml must still be on SILVER_PATH.
func (index *Index) DiscoverFor(path string) (*Manifest, error) {
	index.mu.Lock()
	defer index.mu.Unlock()
	if index.err != nil {
		return nil, index.err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	key := source.FileID(absolute)
	if owner := index.byFile[key]; owner != nil {
		return owner, nil
	}
	for directory := filepath.Dir(absolute); ; directory = filepath.Dir(directory) {
		filename := filepath.Join(directory, manifestFilename)
		if _, err := os.Stat(filename); err == nil {
			manifest, err := ReadManifest(filename)
			if err != nil {
				return nil, err
			}
			if err := index.addManifest(manifest); err != nil {
				return nil, err
			}
			return index.byFile[key], nil
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		if filepath.Dir(directory) == directory {
			return nil, nil
		}
	}
}
