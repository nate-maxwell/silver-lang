// Package packages discovers Silver packages, validates their YAML manifests,
// and resolves the source files they expose through SILVER_PATH.
//
// A package manifest separates package members from public entry files.
// ReadManifest turns that document into an immutable Manifest. An Index is a
// concurrency-safe snapshot of a search path and resolves import requests to
// manifest exports.
package packages

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"silver/source"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Manifest describes one validated YAML package manifest.
//
// Paths are absolute for ReadManifest and filesystem-relative for ReadManifestFS.
// Callers cannot modify a Manifest after it has been read.
type Manifest struct {
	name    string
	authors []string
	id      string
	path    string
	root    string
	exports []Export
	members []Export
}

// File describes one source file belonging to a Manifest.
type File struct {
	declared string
	path     string
}

// Export is a public entry file.
type Export = File

// Name returns the package name declared in the manifest's package field.
func (manifest *Manifest) Name() string { return manifest.name }

// Authors returns a copy of the optional author metadata.
func (manifest *Manifest) Authors() []string {
	return append([]string(nil), manifest.authors...)
}

// ID returns the package's stable runtime identity. The identity is derived
// from the manifest's canonical path, so packages with the same declared name
// but different manifests remain distinct.
func (manifest *Manifest) ID() string { return manifest.id }

// Path returns the manifest path.
func (manifest *Manifest) Path() string { return manifest.path }

// Root returns the directory containing the manifest.
func (manifest *Manifest) Root() string { return manifest.root }

// Exports returns a copy of the manifest's exports in declaration order.
func (manifest *Manifest) Exports() []Export {
	return append([]Export(nil), manifest.exports...)
}

// Members returns all files sharing this package's operator grammar. Exports
// are automatic members, followed by any additional files in members.
func (manifest *Manifest) Members() []File {
	return append([]File(nil), manifest.members...)
}

// Declared returns the slash-separated, cleaned relative path from the
// manifest's members or export list.
func (file File) Declared() string { return file.declared }

// Path returns the resolved source path.
func (file File) Path() string { return file.path }

// ReadManifest reads path as a Silver package manifest.
//
// The manifest must contain exactly one YAML document with a valid package
// name and an export list. Every member or export must name an existing regular .slv
// file within the manifest's directory. Returned paths are absolute and
// cleaned.
func ReadManifest(path string) (*Manifest, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	absolute = filepath.Clean(absolute)
	root := filepath.Dir(absolute)
	manifest, err := readManifestFS(os.DirFS(root), filepath.Base(absolute), source.PathKey)
	if err != nil {
		return nil, fmt.Errorf("package file %q: %w", absolute, err)
	}
	manifest.id = "package:" + source.PathKey(absolute)
	manifest.path = absolute
	manifest.root = root
	for index := range manifest.exports {
		manifest.exports[index].path = filepath.Join(root, filepath.FromSlash(manifest.exports[index].path))
	}
	for index := range manifest.members {
		manifest.members[index].path = filepath.Join(root, filepath.FromSlash(manifest.members[index].path))
	}
	return manifest, nil
}

// ReadManifestFS validates a manifest and its sources in an embedded or virtual
// filesystem. Returned paths are slash-separated names relative to filesystem.
func ReadManifestFS(filesystem fs.FS, filename string) (*Manifest, error) {
	return readManifestFS(filesystem, filename, path.Clean)
}

func readManifestFS(filesystem fs.FS, filename string, pathKey func(string) string) (*Manifest, error) {
	input, err := fs.ReadFile(filesystem, filename)
	if err != nil {
		return nil, fmt.Errorf("could not read package file %q: %w", filename, err)
	}
	document, err := decodeManifest(filename, input)
	if err != nil {
		return nil, err
	}
	root := path.Dir(filename)
	manifest := &Manifest{
		name:    document.packageName,
		authors: document.authors,
		id:      "package:" + filename,
		path:    filename,
		root:    root,
	}
	all := make(map[string]bool)
	for _, list := range []struct {
		kind  string
		files []string
	}{{"export", document.exports}, {"member", document.members}} {
		seen := make(map[string]bool)
		for _, declared := range list.files {
			resolved, err := validateSource(filesystem, filename, root, declared, list.kind)
			if err != nil {
				return nil, err
			}
			key := pathKey(resolved)
			if seen[key] {
				return nil, fmt.Errorf("invalid package file %q: duplicate %s %q", filename, list.kind, declared)
			}
			seen[key] = true
			file := Export{declared: cleanImportName(declared), path: resolved}
			if list.kind == "export" {
				manifest.exports = append(manifest.exports, file)
			}
			if !all[key] {
				manifest.members = append(manifest.members, file)
				all[key] = true
			}
		}
	}
	return manifest, nil
}

// manifestDocument mirrors the accepted YAML fields while retaining whether a
// required field was omitted.
type manifestDocument struct {
	Package *manifestString   `yaml:"package"`
	Authors []manifestString  `yaml:"authors"`
	Export  *[]manifestString `yaml:"export"`
	Members []manifestString  `yaml:"members"`
}

// decodedManifest is the filesystem-independent result of schema validation.
type decodedManifest struct {
	packageName string
	authors     []string
	exports     []string
	members     []string
}

// manifestString is a YAML string that does not accept implicit conversions
// from booleans, numbers, or other scalar types.
type manifestString string

// UnmarshalYAML rejects YAML's implicit scalar conversions so manifest names
// and export paths must be written as strings.
func (value *manifestString) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.ShortTag() != "!!str" {
		return fmt.Errorf("must be a string")
	}
	*value = manifestString(node.Value)
	return nil
}

// decodeManifest decodes and validates the document-level manifest schema.
// Filesystem validation of individual exports is deferred to ReadManifest.
func decodeManifest(path string, input []byte) (decodedManifest, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(input))
	decoder.KnownFields(true)

	var document manifestDocument
	if err := decoder.Decode(&document); err != nil {
		if err == io.EOF {
			return decodedManifest{}, fmt.Errorf("invalid package file %q: manifest is empty", path)
		}
		return decodedManifest{}, fmt.Errorf("invalid package file %q: %w", path, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return decodedManifest{}, fmt.Errorf("invalid package file %q: %w", path, err)
		}
		return decodedManifest{}, fmt.Errorf("invalid package file %q: multiple YAML documents are not allowed", path)
	}
	if document.Package == nil {
		return decodedManifest{}, fmt.Errorf("invalid package file %q: missing required field %q", path, "package")
	}
	packageName := string(*document.Package)
	if !validPackageName(packageName) {
		return decodedManifest{}, fmt.Errorf("invalid package file %q: package name %q is invalid", path, packageName)
	}
	if document.Export == nil {
		return decodedManifest{}, fmt.Errorf("invalid package file %q: missing required field %q", path, "export")
	}
	exports := make([]string, len(*document.Export))
	for index, declared := range *document.Export {
		exports[index] = string(declared)
	}
	members := make([]string, len(document.Members))
	for index, declared := range document.Members {
		members[index] = string(declared)
	}
	authors := make([]string, len(document.Authors))
	for index, author := range document.Authors {
		authors[index] = string(author)
	}
	return decodedManifest{packageName: packageName, authors: authors, exports: exports, members: members}, nil
}

// validateSource resolves one declared file relative to root and ensures it
// is an existing regular Silver source file contained by that root.
func validateSource(filesystem fs.FS, manifestPath, root, declared, kind string) (string, error) {
	if filepath.IsAbs(declared) || path.IsAbs(declared) || filepath.VolumeName(declared) != "" {
		return "", fmt.Errorf("invalid package file %q: %s %q must be relative", manifestPath, kind, declared)
	}
	relative := cleanImportName(declared)
	if relative == ".." || strings.HasPrefix(relative, "../") {
		return "", fmt.Errorf("invalid package file %q: %s %q leaves the package root", manifestPath, kind, declared)
	}
	exported := path.Join(root, relative)
	info, err := fs.Stat(filesystem, exported)
	if err != nil {
		return "", fmt.Errorf("invalid package file %q: could not access %s %q: %w", manifestPath, kind, declared, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("invalid package file %q: %s %q is not a regular file", manifestPath, kind, declared)
	}
	if !strings.EqualFold(filepath.Ext(exported), ".slv") {
		return "", fmt.Errorf("invalid package file %q: %s %q is not a .slv file", manifestPath, kind, declared)
	}
	return exported, nil
}

// validPackageName reports whether name is a nonempty ASCII identifier.
func validPackageName(name string) bool {
	for index, ch := range name {
		if ch == '_' || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' {
			continue
		}
		if index > 0 && ch >= '0' && ch <= '9' {
			continue
		}
		return false
	}
	return name != ""
}
