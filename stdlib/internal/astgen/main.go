// Command astgen regenerates the parsed-AST caches embedded with Silver's
// Silver-authored standard-library modules.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"silver/astcache"
	"silver/evaluator"
	"silver/packages"
	"silver/parser"
)

func main() {
	packageDir, err := os.Getwd()
	if err != nil {
		fatal(err)
	}
	repositoryRoot := filepath.Dir(packageDir)
	if err := os.Chdir(repositoryRoot); err != nil {
		fatal(err)
	}

	sourceRoot := filepath.Join("stdlib", "silver")
	err = filepath.WalkDir(sourceRoot, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() != "package.yaml" {
			return nil
		}

		manifest, err := packages.ReadManifestFS(os.DirFS("."), filepath.ToSlash(sourcePath))
		if err != nil {
			return err
		}
		return generatePackage(manifest)
	})
	if err != nil {
		fatal(err)
	}
}

// generatePackage discovers operators in every member before parsing any file,
// matching runtime package preparation even when a definition is internal.
func generatePackage(manifest *packages.Manifest) error {
	registry := parser.NewInfixRegistry()
	inputs := make(map[string][]byte)
	for _, member := range manifest.Members() {
		input, err := os.ReadFile(member.Path())
		if err != nil {
			return err
		}
		inputs[member.Path()] = input
		for _, declaration := range parser.DiscoverOperatorDeclarations(string(input), member.Path()) {
			if message := registry.Predefine(declaration); message != "" {
				return fmt.Errorf("%s:%d:%d: %s", declaration.Position.Source, declaration.Position.Line, declaration.Position.Column, message)
			}
		}
	}
	for _, member := range manifest.Members() {
		logicalPath := member.Path()
		input := inputs[logicalPath]
		program, parseError := evaluator.ParseSourceWithRegistry(logicalPath, input, registry)
		if parseError != nil {
			return fmt.Errorf("parse %s: %s", logicalPath, parseError.MessageText())
		}
		if err := astcache.Store(logicalPath, input, program); err != nil {
			return fmt.Errorf("cache %s: %w", logicalPath, err)
		}
		fmt.Println("generated", astcache.Path(logicalPath))
	}
	return nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
