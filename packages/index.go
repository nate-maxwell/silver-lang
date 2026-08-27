package packages

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Index is a concurrency-safe snapshot of package and source entries from a
// platform-separated search path such as SILVER_PATH.
//
// The zero value is ready for use, though Refresh must be called before the
// first call to Resolve.
type Index struct {
	mu          sync.RWMutex
	initialized bool
	searchPath  string
	entries     []indexEntry
	byFile      map[string]*Manifest
	err         error
}

// indexEntry represents one search-path component: a directory, an individual
// source file, a manifest file, or a directory's collected manifests. legacy
// selects directory-wide lookup when no manifest governs the directory.
type indexEntry struct {
	directory string
	legacy    bool
	file      string
	manifests []*Manifest
}

// NewIndex returns an empty package index. The caller must call Refresh before
// the first call to Resolve.
func NewIndex() *Index {
	return &Index{byFile: make(map[string]*Manifest)}
}

// Refresh loads the entries in the platform-separated searchPath.
//
// Directory entries containing manifests expose only their declared exports;
// directories without manifests retain legacy directory-wide lookup. Manifest
// files and individual source files may also be listed directly. Refresh is a
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
	index.byFile = make(map[string]*Manifest)
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
			index.addManifest(manifest)
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
			return indexEntry{directory: absolute, legacy: true}, nil
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
	return indexEntry{file: absolute}, nil
}

// loadDirectoryEntry loads every manifest immediately inside directory. A
// directory without manifests is marked for legacy directory-wide lookup.
func loadDirectoryEntry(directory string) (indexEntry, error) {
	entry := indexEntry{directory: directory}
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
	entry.legacy = len(entry.manifests) == 0
	return entry, nil
}

// isManifestPath reports whether path has a supported YAML manifest suffix.
func isManifestPath(path string) bool {
	extension := filepath.Ext(path)
	return strings.EqualFold(extension, ".yaml") || strings.EqualFold(extension, ".yml")
}

// addManifest records ownership of every file exported by manifest.
func (index *Index) addManifest(manifest *Manifest) {
	for _, exported := range manifest.exports {
		index.byFile[pathKey(exported.path)] = manifest
	}
}

// ManifestFor returns the manifest that exports path. It returns nil when path
// is not a declared export in the current snapshot.
func (index *Index) ManifestFor(path string) *Manifest {
	index.mu.RLock()
	defer index.mu.RUnlock()
	return index.byFile[pathKey(path)]
}
