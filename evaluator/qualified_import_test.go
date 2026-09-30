package evaluator

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"silver/object"
	"testing"
)

func TestQualifiedStandardLibraryImports(t *testing.T) {
	t.Setenv(importPathEnvironment, "")
	engine := New()
	env := object.NewEnvironment()
	for _, name := range []string{
		"core", "io", "system", "arrays", "collections", "maps", "math", "random",
		"regex", "string", "terminal", "time", "args", "json", "logging", "networking",
		"path", "testing", "http", "http/client", "http/server", "http/cookies", "http/cookiejar",
	} {
		t.Run(name, func(t *testing.T) {
			qualified := evalInput(t, engine, env, fmt.Sprintf("import(%q)", "core:"+name))
			if _, ok := qualified.(*object.Module); !ok {
				t.Fatalf("core:%s returned %s", name, qualified.Inspect())
			}
			again := evalInput(t, engine, env, fmt.Sprintf("import(%q)", "core:"+name))
			assertManifestImportError(t, evalInput(t, engine, env, fmt.Sprintf("import(%q)", name)), "invalid package import")
			if qualified != again {
				t.Fatal("repeated qualified imports have different module identities")
			}
		})
	}
	var output bytes.Buffer
	result := evalInput(t, NewWithOutput(&output), object.NewEnvironment(), `
let println = import("core:io").println
println("hello")`)
	if isError(result) || output.String() != "hello\n" {
		t.Fatalf("member alias returned %s, output %q", result.Inspect(), output.String())
	}
}

func TestQualifiedPackageImportsAfterAppendPath(t *testing.T) {
	t.Setenv(importPathEnvironment, "")
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "my_library"), 0755); err != nil {
		t.Fatal(err)
	}
	writeSilverFile(t, filepath.Join(dir, "that_module.slv"), `let foo = fn() int { return 42 }`)
	writeSilverFile(t, filepath.Join(dir, "my_library", "my_sub_pkg.slv"), `
let api = import("my_package:that_module")
let answer = api.foo()`)
	writeSilverFile(t, filepath.Join(dir, "hidden.slv"), "let value = 99")
	manifestPath := filepath.Join(dir, "package.yaml")
	writeSilverFile(t, manifestPath, `package: my_package
authors: ["Ada"]
members: [./hidden.slv]
export:
  - ./that_module.slv
  - ./my_library/my_sub_pkg.slv
`)
	engine := New()
	env := object.NewEnvironment()
	env.SetSourceDir(t.TempDir())
	assertManifestImportError(t, evalInput(t, engine, env, `import("my_package:that_module")`), "not found")
	assertInteger(t, evalInput(t, engine, env, fmt.Sprintf(`
let system = import("core:system")
system.append_path(%q)
let my_func = import("my_package:that_module").foo
my_func()`, manifestPath)), 42)
	assertInteger(t, evalInput(t, engine, env, `import("my_package:my_library/my_sub_pkg").answer`), 42)
	for _, request := range []string{"my_package:hidden", "my_package:my_sub_pkg", "wrong:that_module"} {
		assertManifestImportError(t, evalInput(t, engine, env, fmt.Sprintf("import(%q)", request)), "not found")
	}
	qualified := evalInput(t, engine, env, `import("my_package:that_module")`)
	again := evalInput(t, engine, env, `import("my_package:that_module")`)
	assertManifestImportError(t, evalInput(t, engine, env, fmt.Sprintf("import(%q)", filepath.Join(dir, "that_module.slv"))), "invalid package import")
	if qualified != again {
		t.Fatal("repeated imports have different module identities")
	}
	if err := os.Rename(manifestPath, filepath.Join(dir, "manifest.backup")); err != nil {
		t.Fatal(err)
	}
	assertManifestImportError(t, evalInput(t, engine, env, `import("my_package:that_module")`), "required package YAML manifest")
}

func TestQualifiedImportsIsolatePackagesAndReserveCore(t *testing.T) {
	var manifests []string
	for i, name := range []string{"first", "second", "core"} {
		dir := t.TempDir()
		writeSilverFile(t, filepath.Join(dir, "io.slv"), fmt.Sprintf("let value = %d", i+1))
		writeSilverFile(t, filepath.Join(dir, "missing.slv"), "let value = 99")
		manifestPath := filepath.Join(dir, "package.yaml")
		writeSilverFile(t, manifestPath, fmt.Sprintf("package: %s\nexport: [io.slv, missing.slv]\n", name))
		manifests = append(manifests, manifestPath)
	}
	t.Setenv(importPathEnvironment, manifests[0]+string(os.PathListSeparator)+manifests[1]+string(os.PathListSeparator)+manifests[2])
	engine := New()
	env := object.NewEnvironment()
	assertInteger(t, evalInput(t, engine, env, `import("first:io").value`), 1)
	assertInteger(t, evalInput(t, engine, env, `import("second:io").value`), 2)
	module := evalInput(t, engine, env, `import("core:io")`).(*object.Module)
	if module.Exports["println"] == nil {
		t.Fatal("registered package shadowed the standard library")
	}
	assertManifestImportError(t, evalInput(t, engine, env, `import("core:missing")`), "does not exist")
	assertManifestImportError(t, evalInput(t, engine, env, `import("core:_networking")`), "does not exist")
}

func TestQualifiedImportsRejectMalformedNames(t *testing.T) {
	t.Setenv(importPathEnvironment, "")
	for _, request := range []string{
		":io", "core:", "bad-name:module", "core:io:extra", "core:../io", "core:./io",
		"core:io/", "core:http//cookies", `core:http\cookies`, "core:io.slv", "core: io",
	} {
		t.Run(request, func(t *testing.T) {
			assertManifestImportError(t, evalInput(t, New(), object.NewEnvironment(), fmt.Sprintf("import(%q)", request)), "invalid package import")
		})
	}
}

func TestQualifiedImportCycle(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "a.slv"), `let b = import("cycle:b")`)
	writeSilverFile(t, filepath.Join(dir, "b.slv"), `let a = import("cycle:a")`)
	manifestPath := filepath.Join(dir, "package.yaml")
	writeSilverFile(t, manifestPath, "package: cycle\nexport: [a.slv, b.slv]\n")
	t.Setenv(importPathEnvironment, manifestPath)
	assertManifestImportError(t, evalInput(t, New(), object.NewEnvironment(), `import("cycle:a")`), "circular import")
}
