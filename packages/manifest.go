// Package packages loads Silver package manifests and resolves their exports.
package packages

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Manifest describes one YAML package file.
type Manifest struct {
	name    string
	id      string
	path    string
	root    string
	exports []Export
}

// Export describes one source file exposed by a manifest.
type Export struct {
	declared string
	path     string
}

// Name returns the package name declared by the manifest.
func (manifest *Manifest) Name() string { return manifest.name }

// ID returns the stable runtime identity derived from the manifest path.
func (manifest *Manifest) ID() string { return manifest.id }

// Path returns the absolute manifest path.
func (manifest *Manifest) Path() string { return manifest.path }

// Root returns the directory containing the manifest.
func (manifest *Manifest) Root() string { return manifest.root }

// Exports returns the manifest's exports in declaration order.
func (manifest *Manifest) Exports() []Export {
	return append([]Export(nil), manifest.exports...)
}

// Declared returns the normalized path written in the export block.
func (export Export) Declared() string { return export.declared }

// Path returns the export's absolute source path.
func (export Export) Path() string { return export.path }

// ReadManifest parses path as a YAML manifest and validates every export.
func ReadManifest(path string) (*Manifest, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read package file %q: %w", path, err)
	}

	document, err := decodeManifest(path, input)
	if err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	absolute = filepath.Clean(absolute)
	root := filepath.Dir(absolute)
	manifest := &Manifest{
		name: document.packageName,
		id:   "package:" + pathKey(absolute),
		path: absolute,
		root: root,
	}

	seen := make(map[string]bool, len(document.exports))
	for _, declared := range document.exports {
		exported, err := validateExport(path, root, declared)
		if err != nil {
			return nil, err
		}
		key := pathKey(exported)
		if seen[key] {
			return nil, fmt.Errorf("invalid package file %q: duplicate export %q", path, declared)
		}
		seen[key] = true
		manifest.exports = append(manifest.exports, Export{
			declared: cleanImportName(declared),
			path:     exported,
		})
	}
	return manifest, nil
}

type manifestDocument struct {
	Package *manifestString   `yaml:"package"`
	Export  *[]manifestString `yaml:"export"`
}

type decodedManifest struct {
	packageName string
	exports     []string
}

type manifestString string

func (value *manifestString) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode || node.ShortTag() != "!!str" {
		return fmt.Errorf("must be a string")
	}
	*value = manifestString(node.Value)
	return nil
}

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
	return decodedManifest{packageName: packageName, exports: exports}, nil
}

func validateExport(manifestPath, root, declared string) (string, error) {
	if filepath.IsAbs(declared) {
		return "", fmt.Errorf("invalid package file %q: export %q must be relative", manifestPath, declared)
	}
	exported := filepath.Clean(filepath.Join(root, filepath.FromSlash(declared)))
	relative, err := filepath.Rel(root, exported)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid package file %q: export %q leaves the package root", manifestPath, declared)
	}
	info, err := os.Stat(exported)
	if err != nil {
		return "", fmt.Errorf("invalid package file %q: could not access export %q: %w", manifestPath, declared, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("invalid package file %q: export %q is not a regular file", manifestPath, declared)
	}
	if !strings.EqualFold(filepath.Ext(exported), ".slv") {
		return "", fmt.Errorf("invalid package file %q: export %q is not a .slv file", manifestPath, declared)
	}
	return exported, nil
}

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
