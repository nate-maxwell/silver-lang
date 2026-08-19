package main

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"silver/astcache"
	"silver/evaluator"
)

func runASTGen(args []string, _ io.Reader, out, errOut io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(errOut, "usage: silver astgen <path>")
		return 2
	}
	if err := generateASTCaches(args[0], out); err != nil {
		fmt.Fprintln(errOut, err)
		return 1
	}
	return 0
}

// generateASTCaches parses path, or every Silver source below path when it is
// a directory, and writes each AST cache beside its source file.
func generateASTCaches(path string, out io.Writer) error {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("could not resolve %q: %w", path, err)
	}
	absolutePath = filepath.Clean(absolutePath)

	info, err := os.Stat(absolutePath)
	if err != nil {
		return fmt.Errorf("could not access %q: %w", absolutePath, err)
	}
	if !info.IsDir() {
		return generateASTCache(absolutePath, out)
	}

	return filepath.WalkDir(absolutePath, func(sourcePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(sourcePath) != ".slv" {
			return nil
		}
		return generateASTCache(sourcePath, out)
	})
}

func generateASTCache(sourcePath string, out io.Writer) error {
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("could not read %q: %w", sourcePath, err)
	}
	program, parseError := evaluator.ParseSource(sourcePath, source)
	if parseError != nil {
		return errors.New(parseError.Inspect())
	}
	if err := astcache.Store(sourcePath, source, program); err != nil {
		return fmt.Errorf("could not generate %q: %w", astcache.Path(sourcePath), err)
	}
	fmt.Fprintln(out, astcache.Path(sourcePath))
	return nil
}
