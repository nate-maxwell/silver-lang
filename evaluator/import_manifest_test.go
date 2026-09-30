package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestImportsRejectLegacyFormsEvenWithManifest(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "module.slv"), "let value = 42")
	writePackageManifest(t, dir, "module.slv")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	for _, request := range []string{"io", "http/cookiejar", "_networking", "module.slv", "./module.slv", "../module.slv", filepath.Join(dir, "module.slv")} {
		assertManifestImportError(t, evalInput(t, New(), env, fmt.Sprintf("import(%q)", request)), "invalid package import")
	}
	assertInteger(t, evalInput(t, New(), env, `import("fixture:module").value`), 42)
}

func TestQualifiedImportsRequireExplicitRegistration(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "module.slv"), "let value = 42")
	writePackageManifest(t, dir, "module.slv")
	t.Setenv(importPathEnvironment, "")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	assertManifestImportError(t, evalInput(t, New(), env, `import("fixture:module")`), "not found")
	main := filepath.Join(dir, "main.slv")
	writeSilverFile(t, main, `import("fixture:module")`)
	assertManifestImportError(t, New().EvalFile(main, object.NewEnvironment()), "not found")
}

func TestPackageImportRejectsUnlistedFile(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "member.slv"), "let value = 1")
	writeSilverFile(t, filepath.Join(dir, "unlisted.slv"), "let value = 42")
	writePackageManifest(t, dir, "member.slv")
	assertManifestImportError(t, evalInput(t, New(), object.NewEnvironment(), `import("fixture:unlisted")`), "not found")
}

func TestMovingManifestInvalidatesCachedImports(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "module.slv"), "let value = 42")
	writePackageManifest(t, dir, "module.slv")
	engine := New()
	env := object.NewEnvironment()
	assertInteger(t, evalInput(t, engine, env, `import("fixture:module").value`), 42)
	manifest := filepath.Join(dir, "package.yaml")
	if err := os.Rename(manifest, filepath.Join(dir, "manifest.backup")); err != nil {
		t.Fatal(err)
	}
	for _, evaluator := range []*Evaluator{engine, New()} {
		assertManifestImportError(t, evalInput(t, evaluator, env, `import("fixture:module")`), "required package YAML manifest")
	}
	if err := os.Mkdir(manifest, 0755); err != nil {
		t.Fatal(err)
	}
	assertManifestImportError(t, evalInput(t, engine, env, `import("fixture:module")`), "not a regular file")
}

func TestPackageImportRejectsInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "module.slv"), "let value = 42")
	manifest := filepath.Join(dir, "package.yaml")
	writeSilverFile(t, manifest, "package: example\nexport: not-a-list\n")
	t.Setenv(importPathEnvironment, manifest)
	assertManifestImportError(t, evalInput(t, New(), object.NewEnvironment(), `import("example:module")`), "invalid package file")
}

func assertManifestImportError(t *testing.T, result object.Object, message string) {
	t.Helper()
	failure, ok := result.(*object.Error)
	definition, _ := object.BuiltinStructDefinitionByName("ImportError")
	if !ok || failure.Value.Struct != definition || !strings.Contains(failure.MessageText(), message) {
		t.Fatalf("result is %s, want ImportError containing %q", result.Inspect(), message)
	}
}
