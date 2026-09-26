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
	entries     []indexEntry
	byFile      map[source.ModuleID]*Manifest
	err         error
}

// indexEntry contains the manifests found in one search-path component.
type indexEntry struct {
	manifests []*Manifest
}

// NewIndex returns an empty package index. The caller must call Refresh before
// the first call to Resolve.
func NewIndex() *Index {
	return &Index{byFile: make(map[source.ModuleID]*Manifest)}
}

// Refresh loads the entries in the platform-separated searchPath.
//
// Directory entries expose only manifest exports. Manifest files may also be
// listed directly; individual source-file entries are rejected. Refresh is a
// no-op when searchPath has not changed and returns the result of the previous
// load in that case.
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
		entry, err := loadIndexEntry(rawEntry)
		if err != nil {
			index.err = err
			return err
		}
		index.entries = append(index.entries, entry)
		for _, manifest := range entry.manifests {
			if err := index.addManifest(manifest); err != nil {
				index.err = err
				return err
			}
		}
	}
	return nil
}

// loadIndexEntry classifies and loads one search-path entry.
func loadIndexEntry(path string) (indexEntry, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return indexEntry{}, err
	}
	absolute = filepath.Clean(absolute)
	info, err := os.Stat(absolute)
	if err != nil {
		if os.IsNotExist(err) {
			return indexEntry{}, nil
		}
		return indexEntry{}, err
	}

	if info.IsDir() {
		return loadDirectoryEntry(absolute)
	}
	if isManifestPath(absolute) {
		manifest, err := ReadManifest(absolute)
		if err != nil {
			return indexEntry{}, err
		}
		return indexEntry{manifests: []*Manifest{manifest}}, nil
	}
	return indexEntry{}, fmt.Errorf("search-path entry %q must be a package YAML manifest or a directory containing manifests", absolute)
}

// loadDirectoryEntry loads every manifest immediately inside directory.
// A directory without manifests exposes no files.
func loadDirectoryEntry(directory string) (indexEntry, error) {
	entry := indexEntry{}
	candidates, err := os.ReadDir(directory)
	if err != nil {
		return indexEntry{}, err
	}
	for _, candidate := range candidates {
		if candidate.IsDir() || !isManifestPath(candidate.Name()) {
			continue
		}
		manifest, err := ReadManifest(filepath.Join(directory, candidate.Name()))
		if err != nil {
			return indexEntry{}, err
		}
		entry.manifests = append(entry.manifests, manifest)
	}
	return entry, nil
}

// isManifestPath reports whether path has a supported YAML manifest suffix.
func isManifestPath(path string) bool {
	extension := filepath.Ext(path)
	return strings.EqualFold(extension, ".yaml") || strings.EqualFold(extension, ".yml")
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

// DiscoverFor finds a file's package even when it is run or imported directly
// without SILVER_PATH. The nearest package.yaml is the local package boundary;
// only its declared members acquire package identity. Other manifest filenames
// are supported through explicit search-path registration.
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
