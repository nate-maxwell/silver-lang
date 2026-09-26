package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestEveryFilesystemImportRequiresManifest(t *testing.T) {
	for _, form := range []string{"relative", "absolute", "local name", "parent relative"} {
		t.Run(form, func(t *testing.T) {
			t.Setenv(importPathEnvironment, "")
			dir := t.TempDir()
			modulePath := filepath.Join(dir, "module.slv")
			writeSilverFile(t, modulePath, "let value = 42")
			env := object.NewEnvironment()
			env.SetSourceDir(dir)
			request := "./module.slv"
			switch form {
			case "absolute":
				request = modulePath
			case "local name":
				request = "module.slv"
			case "parent relative":
				child := filepath.Join(dir, "app")
				if err := os.Mkdir(child, 0755); err != nil {
					t.Fatal(err)
				}
				env.SetSourceDir(child)
				request = "../module.slv"
			}
			engine := New()
			expression := fmt.Sprintf("import(%q).value", request)
			assertManifestImportError(t, evalInput(t, engine, env, expression), "package YAML manifest is required")
			// Declaring it as an internal member is sufficient for explicit imports.
			writePackageManifest(t, dir, "module.slv")
			assertInteger(t, evalInput(t, engine, env, expression), 42)
		})
	}
}

func TestFilesystemImportRejectsUnlistedFile(t *testing.T) {
	t.Setenv(importPathEnvironment, "")
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "member.slv"), "let value = 1")
	writeSilverFile(t, filepath.Join(dir, "unlisted.slv"), "let value = 42")
	writePackageManifest(t, dir, "member.slv")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	assertManifestImportError(t, evalInput(t, New(), env, `import("./unlisted.slv")`), "list the file in members or export")
}

func TestMovingManifestInvalidatesCachedImports(t *testing.T) {
	for _, searchPath := range []bool{false, true} {
		t.Run(fmt.Sprintf("search-path=%t", searchPath), func(t *testing.T) {
			dir := t.TempDir()
			writeSilverFile(t, filepath.Join(dir, "module.slv"), "let value = 42")
			writeSilverFile(t, filepath.Join(dir, "package.yaml"), "package: example\nexport: [module.slv]\n")
			env := object.NewEnvironment()
			env.SetSourceDir(dir)
			t.Setenv(importPathEnvironment, "")
			request := `import("./module.slv").value`
			if searchPath {
				t.Setenv(importPathEnvironment, dir)
				env.SetSourceDir(t.TempDir())
				request = `import("module.slv").value`
			}
			engine := New()
			assertInteger(t, evalInput(t, engine, env, request), 42)
			backup := filepath.Join(dir, "bck")
			if err := os.Mkdir(backup, 0755); err != nil {
				t.Fatal(err)
			}
			manifest := filepath.Join(dir, "package.yaml")
			if err := os.Rename(manifest, filepath.Join(backup, "package.yaml")); err != nil {
				t.Fatal(err)
			}
			assertManifestImportError(t, evalInput(t, engine, env, request), "required package YAML manifest")
			// Fresh sessions must also reject the missing manifest.
			freshEnv := object.NewEnvironment()
			freshEnv.SetSourceDir(dir)
			assertManifestImportError(t, evalInput(t, New(), freshEnv, `import("./module.slv")`), "package YAML manifest is required")
			// A directory at the old filename cannot satisfy the requirement either.
			if err := os.Mkdir(manifest, 0755); err != nil {
				t.Fatal(err)
			}
			assertManifestImportError(t, evalInput(t, engine, env, request), "not a regular file")
		})
	}
}

func TestFilesystemImportRejectsInvalidManifest(t *testing.T) {
	t.Setenv(importPathEnvironment, "")
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "module.slv"), "let value = 42")
	writeSilverFile(t, filepath.Join(dir, "package.yaml"), "package: example\nexport: not-a-list\n")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	assertManifestImportError(t, evalInput(t, New(), env, `import("./module.slv")`), "invalid package file")
}

func assertManifestImportError(t *testing.T, result object.Object, message string) {
	t.Helper()
	failure, ok := result.(*object.Error)
	definition, _ := object.BuiltinStructDefinitionByName("ImportError")
	if !ok || failure.Value.Struct != definition || !strings.Contains(failure.MessageText(), message) {
		t.Fatalf("result is %s, want ImportError containing %q", result.Inspect(), message)
	}
}
