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
		if entry.IsDir() || filepath.Ext(sourcePath) != ".slv" {
			return nil
		}

		source, err := os.ReadFile(sourcePath)
		if err != nil {
			return err
		}
		logicalPath := filepath.ToSlash(sourcePath)
		program, parseError := evaluator.ParseSource(logicalPath, source)
		if parseError != nil {
			return fmt.Errorf("parse %s: %s", logicalPath, parseError.MessageText())
		}
		if err := astcache.Store(logicalPath, source, program); err != nil {
			return fmt.Errorf("cache %s: %w", logicalPath, err)
		}
		fmt.Println("generated", astcache.Path(logicalPath))
		return nil
	})
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
