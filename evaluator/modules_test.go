package evaluator

import (
	"fmt"
	"os"
	"path/filepath"
	"silver/ast"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"strings"
	"testing"
)

func TestEvalFileWithNestedQualifiedImport(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "lib/base.slv", "lib/math.slv")
	libDir := filepath.Join(dir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeSilverFile(t, filepath.Join(libDir, "base.slv"), `let factor = 2`)
	writeSilverFile(t, filepath.Join(libDir, "math.slv"), `
let base = import("fixture:lib/base")
let double = fn(x) int { return x * base.factor }
`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let math = import("fixture:lib/math")
math.double(21)
`)

	env := object.NewEnvironment()
	result := New().EvalFile(mainPath, env)
	integer, ok := result.(*object.Integer)
	if !ok {
		t.Fatalf("result is %T (%v), want *object.Integer", result, result)
	}
	if integer.Value != 42 {
		t.Fatalf("result is %d, want 42", integer.Value)
	}
	if _, ok := env.Get("double"); ok {
		t.Fatal("module binding leaked into the importing environment")
	}
}

func TestEvalFileReadsCurrentSourceWithoutWritingFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.slv")
	writeSilverFile(t, path, "let answer = 41\nanswer")
	engine := New()

	result := engine.EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 41)

	writeSilverFile(t, path, "let answer = 42\nanswer")
	result = engine.EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 42)
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "main.slv" {
		t.Fatalf("EvalFile created files beside the source: %v", entries)
	}
}

func TestParseFileFoldsConstants(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, path, "1 + 2 * 3")

	program, parseError := New().parseFile(path, nil)
	if parseError != nil {
		t.Fatal(parseError.Inspect())
	}
	expression := program.Statements[0].(*ast.ExpressionStatement).Expression
	integer, ok := expression.(*ast.IntegerLiteral)
	if !ok || integer.Value != 7 {
		t.Fatalf("parsed expression is %T (%v), want folded integer 7", expression, expression)
	}
}

func TestEvalFileLeavesLegacyASTCacheUntouched(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, path, "42")
	cachePath := path + ".astc"
	if err := os.WriteFile(cachePath, []byte("legacy cache"), 0600); err != nil {
		t.Fatal(err)
	}

	result := New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 42)
	contents, err := os.ReadFile(cachePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "legacy cache" {
		t.Fatalf("EvalFile modified a legacy cache: %q", contents)
	}
}

func TestImportsAreCached(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "module.slv")
	writeSilverFile(t, filepath.Join(dir, "module.slv"), `let value = 1`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	engine := New()

	first := evalInput(t, engine, env, `import("fixture:module")`)
	if _, ok := first.(*object.Module); !ok {
		t.Fatalf("initial import failed: %s", first.Inspect())
	}
	second := evalInput(t, engine, env, `import("fixture:module")`)
	if first != second {
		t.Fatal("the same module was evaluated more than once")
	}
}

func TestImportSearchesSilverPath(t *testing.T) {
	sourceDir := t.TempDir()
	firstLibraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(firstLibraryDir, "package.yaml"), "package: empty\nexport: []\n")
	secondLibraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(secondLibraryDir, "library.slv"), `let value = 42`)
	writeSilverFile(t, filepath.Join(secondLibraryDir, "package.yaml"), "package: library\nexport: [library.slv]\n")
	t.Setenv(importPathEnvironment, strings.Join([]string{filepath.Join(firstLibraryDir, "package.yaml"), filepath.Join(secondLibraryDir, "package.yaml")}, string(os.PathListSeparator)))

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("library:library").value`)
	assertInteger(t, result, 42)
}

func TestSilverPathPackageExposesManifestFiles(t *testing.T) {
	sourceDir := t.TempDir()
	packageDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(packageDir, "nested"), 0755); err != nil {
		t.Fatal(err)
	}
	writeSilverFile(t, filepath.Join(packageDir, "nested", "library.slv"), `let value = 42`)
	writeSilverFile(t, filepath.Join(packageDir, "hidden.slv"), `let value = 99`)
	if err := os.WriteFile(filepath.Join(packageDir, "package.yaml"), []byte(`
package: example
export:
  - ./nested/library.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(importPathEnvironment, filepath.Join(packageDir, "package.yaml"))

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	engine := New()
	assertInteger(t, evalInput(t, engine, env, `import("example:nested/library").value`), 42)
	assertManifestImportError(t, evalInput(t, engine, env, `import("example:library")`), "not found")

	result := evalInput(t, engine, env, `import("example:hidden")`)
	if failure, ok := result.(*object.Error); !ok || !strings.Contains(failure.MessageText(), "not found") {
		t.Fatalf("hidden import is %#v, want package-interface import error", result)
	}
}

func TestSilverPathPackageReadsSourcesWithoutWritingFiles(t *testing.T) {
	sourceDir := t.TempDir()
	packageDir := t.TempDir()
	modulePath := filepath.Join(packageDir, "module.slv")
	writeSilverFile(t, modulePath, "let value = 41")
	writeSilverFile(t, filepath.Join(packageDir, "helper.slv"), "let value = 1")
	if err := os.WriteFile(filepath.Join(packageDir, "package.yaml"), []byte(`
package: example
export:
  - module.slv
members:
  - helper.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(importPathEnvironment, filepath.Join(packageDir, "package.yaml"))

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	assertInteger(t, evalInput(t, New(), env, `import("example:module").value`), 41)
	writeSilverFile(t, modulePath, "let value = 42")
	assertInteger(t, evalInput(t, New(), env, `import("example:module").value`), 42)
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("package import created files beside the sources: %v", entries)
	}
}

func TestSilverPathRejectsExplicitSourceFile(t *testing.T) {
	sourceDir := t.TempDir()
	fileDir := t.TempDir()
	path := filepath.Join(fileDir, "single.slv")
	writeSilverFile(t, path, `let value = 42`)
	t.Setenv(importPathEnvironment, path)

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("example:single").value`)
	if failure, ok := result.(*object.Error); !ok || !strings.Contains(failure.MessageText(), "package YAML manifest") {
		t.Fatalf("result is %#v, want manifest requirement error", result)
	}
}

func TestSilverPathRefreshesAfterEnvironmentChange(t *testing.T) {
	sourceDir := t.TempDir()
	packageDir := t.TempDir()
	writeSilverFile(t, filepath.Join(packageDir, "library.slv"), `let value = 42`)
	if err := os.WriteFile(filepath.Join(packageDir, "package.yaml"), []byte(`
package: example
export:
  - ./library.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(importPathEnvironment, "")

	engine := New()
	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	if result := evalInput(t, engine, env, `import("example:missing")`); !isError(result) {
		t.Fatalf("initial import is %#v, want an error", result)
	}
	if err := os.Setenv(importPathEnvironment, filepath.Join(packageDir, "package.yaml")); err != nil {
		t.Fatal(err)
	}
	assertInteger(t, evalInput(t, engine, env, `import("example:library").value`), 42)
}

func TestPackageOperatorsAreScopedByManifest(t *testing.T) {
	packageDir := t.TempDir()
	writePackageOperatorFixture(t, packageDir)

	mainPath := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, mainPath, `
let foo = import("package_foo:foo")
let bar = import("package_bar:bar")
foo.apply(foo.make(4), 2) * 10 + bar.apply(foo.make(4), 2)
`)
	result := New().EvalFile(mainPath, object.NewEnvironment())
	assertInteger(t, result, 427)
}

func TestImportedPackageOperatorIsNotVisibleToImporter(t *testing.T) {
	packageDir := t.TempDir()
	writePackageOperatorFixture(t, packageDir)

	mainPath := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, mainPath, `
let foo = import("package_foo:foo")
foo.make(4) @@ 2
`)
	result := New().EvalFile(mainPath, object.NewEnvironment())
	failure, ok := result.(*object.Error)
	if !ok || !strings.Contains(failure.MessageText(), "could not parse") {
		t.Fatalf("result is %#v, want syntax error for package-local operator", result)
	}
}

func TestInternalPackageMembersShareOperators(t *testing.T) {
	for _, publicOperators := range []bool{true, false} {
		name := "internal operator definition"
		if publicOperators {
			name = "public operator definition"
		}
		t.Run(name, func(t *testing.T) {
			packageDir := t.TempDir()
			helperPath := filepath.Join(packageDir, "helper.slv")
			apiPath := filepath.Join(packageDir, "api.slv")
			writeSilverFile(t, filepath.Join(packageDir, "ops.slv"), `operator @@ = fn(left: int, right: int) int { return left + right }`)
			writeSilverFile(t, helperPath, `let calculate = fn() int { return 20 @@ 22 }`)
			writeSilverFile(t, apiPath, `
let ops = import("library:ops")
let helper = import("library:helper")
let value = helper.calculate()
value
`)
			exports := "[api.slv]"
			if publicOperators {
				exports = "[api.slv, ops.slv]"
			}
			writeSilverFile(t, filepath.Join(packageDir, "package.yaml"), "package: library\nmembers: [helper.slv, ops.slv]\nexport: "+exports+"\n")
			t.Setenv(importPathEnvironment, filepath.Join(packageDir, "package.yaml"))
			for run := 0; run < 2; run++ {
				// Each evaluator discovers declarations in internal members.
				engine := New()
				env := object.NewEnvironment()
				env.SetSourceDir(t.TempDir())
				assertInteger(t, evalInput(t, engine, env, `import("library:api").value`), 42)
				if result := evalInput(t, engine, env, `import("library:helper")`); !isError(result) {
					t.Fatalf("internal member was exposed through SILVER_PATH: %v", result)
				}
				// Internal members cannot be imported from another package.
				assertManifestImportError(t, evalInput(t, engine, env, fmt.Sprintf("import(%q)", helperPath)), "invalid package import")
			}
			// Running the entry point uses the registered package and its operator scope.
			assertInteger(t, New().EvalFile(apiPath, object.NewEnvironment()), 42)
		})
	}
}

func TestManifestFreeSilverPathDoesNotExposeSources(t *testing.T) {
	directory := t.TempDir()
	writeSilverFile(t, filepath.Join(directory, "library.slv"), "let value = 42")
	t.Setenv(importPathEnvironment, directory)
	env := object.NewEnvironment()
	env.SetSourceDir(t.TempDir())
	if result := evalInput(t, New(), env, `import("example:library")`); !isError(result) {
		t.Fatalf("manifest-free search directory exposed source: %v", result)
	}
}

func TestBundledSourcesUseSeparatePackageScopes(t *testing.T) {
	engine := New()
	for _, name := range []string{"json", "http", "http/client"} {
		module, ok := engine.standardLibrary.LookupSource("core:" + name)
		if !ok {
			t.Fatalf("missing bundled module %q", name)
		}
		if module.PackageID == "stdlib" || !strings.HasSuffix(module.PackageID, "package.yaml") {
			t.Fatalf("%s has no manifest identity: %q", name, module.PackageID)
		}
		if _, parseError := engine.prepareSourcePackage(module); parseError != nil {
			t.Fatal(parseError)
		}
	}
	json, _ := engine.standardLibrary.LookupSource("core:json")
	http, _ := engine.standardLibrary.LookupSource("core:http")
	client, _ := engine.standardLibrary.LookupSource("core:http/client")
	if json.PackageID == http.PackageID || http.PackageID != client.PackageID {
		t.Fatal("bundled operator scopes do not follow package manifests")
	}
	// An operator installed in one bundled package must not enter another's grammar.
	jsonEnv := object.NewEnvironment()
	jsonEnv.SetPackageID(json.PackageID)
	program, parseError := ParseSourceWithRegistry("json-ops", []byte(`operator @@ = fn(left: int, right: int) int { return left + right }`), engine.operatorScope(json.PackageID).registry)
	if parseError != nil {
		t.Fatal(parseError.Inspect())
	}
	result := engine.Eval(program, jsonEnv)
	if isError(result) {
		t.Fatal(result.Inspect())
	}
	if _, err := ParseSourceWithRegistry("json-helper", []byte("20 @@ 22"), engine.operatorScope(json.PackageID).registry); err != nil {
		t.Fatal(err.Inspect())
	}
	if _, err := ParseSourceWithRegistry("http-helper", []byte("20 @@ 22"), engine.operatorScope(http.PackageID).registry); err == nil {
		t.Fatal("json's operator grammar leaked into http")
	}
}

func writePackageOperatorFixture(t *testing.T, directory string) {
	t.Helper()
	fooDir := filepath.Join(directory, "foo")
	barDir := filepath.Join(directory, "bar")
	for _, dir := range []string{fooDir, barDir} {
		if err := os.Mkdir(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	writeSilverFile(t, filepath.Join(fooDir, "foo_operators.slv"), `operator @@ = fn(left, right) int { return 999 }`)
	writeSilverFile(t, filepath.Join(fooDir, "foo.slv"), `
export { Foo, make, apply }
let operators = import("package_foo:foo_operators")
type Foo = struct {
    value: int
    @@: call(self: Foo, other: int) int
}
let overload = fn(self: Foo, other: int) int { return self.value * 10 + other }
let make = fn(value: int) Foo { return Foo{value, overload} }
let apply = fn(left: Foo, right: int) int { return left @@ right }
`)
	if err := os.WriteFile(filepath.Join(fooDir, "package.yaml"), []byte(`
package: package_foo
export:
  - ./foo.slv
  - ./foo_operators.slv
`), 0644); err != nil {
		t.Fatal(err)
	}

	writeSilverFile(t, filepath.Join(barDir, "bar_operators.slv"), `operator @@ = fn(left, right) int { return 7 }`)
	writeSilverFile(t, filepath.Join(barDir, "bar.slv"), `
export { apply }
let operators = import("package_bar:bar_operators")
let foo = import("package_foo:foo")
let apply = fn(left: foo.Foo, right: int) int { return left @@ right }
`)
	if err := os.WriteFile(filepath.Join(barDir, "package.yaml"), []byte(`
package: package_bar
export:
  - ./bar.slv
  - ./bar_operators.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(importPathEnvironment, filepath.Join(fooDir, "package.yaml")+string(os.PathListSeparator)+filepath.Join(barDir, "package.yaml"))
}

func TestQualifiedImportDoesNotUseNeighboringFiles(t *testing.T) {
	sourceDir := t.TempDir()
	writePackageManifest(t, sourceDir, "library.slv")
	libraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(sourceDir, "library.slv"), `let value = 1`)
	writeSilverFile(t, filepath.Join(libraryDir, "library.slv"), `let value = 2`)
	writeSilverFile(t, filepath.Join(libraryDir, "package.yaml"), "package: library\nexport: [library.slv]\n")
	t.Setenv(importPathEnvironment, filepath.Join(libraryDir, "package.yaml"))

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("library:library").value`)
	assertInteger(t, result, 2)
}

func TestSilverPathModuleResolvesQualifiedInternalMembers(t *testing.T) {
	sourceDir := t.TempDir()
	libraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(libraryDir, "dependency.slv"), `let value = 21`)
	writeSilverFile(t, filepath.Join(libraryDir, "package.yaml"), "package: library\nmembers: [dependency.slv]\nexport: [library.slv]\n")
	writeSilverFile(t, filepath.Join(libraryDir, "library.slv"), `
let dependency = import("library:dependency")
let value = dependency.value * 2
`)
	t.Setenv(importPathEnvironment, filepath.Join(libraryDir, "package.yaml"))

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("library:library").value`)
	assertInteger(t, result, 42)
}

func TestSilverStandardLibraryImportsAreCached(t *testing.T) {
	engine := New()
	env := object.NewEnvironment()

	first := evalInput(t, engine, env, `import("core:testing")`)
	second := evalInput(t, engine, env, `import("core:testing")`)
	if first != second {
		t.Fatal("the same Silver standard-library module was evaluated more than once")
	}
}

func TestStandardLibraryAndUserPackageNamesAreSeparate(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "testing.slv")
	writeSilverFile(t, filepath.Join(dir, "testing.slv"), `let origin = "user"`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `
let standard = import("core:testing")
let user = import("fixture:testing")
standard.check(user.origin == "user", "path import should load the user module")
user.origin
`)
	text, ok := result.(*object.String)
	if !ok || text.Value != "user" {
		t.Fatalf("result is %T (%v), want user module value", result, result)
	}
}

func TestImportAcceptsPathExpression(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "module.slv")
	writeSilverFile(t, filepath.Join(dir, "module.slv"), `let value = 42`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `
let module_path = "fixture:module"
import(module_path).value
`)
	assertInteger(t, result, 42)
}

func TestModuleExportDeclarationLimitsPublicBindings(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `
export { public_value }
let public_value = 42
let private_value = 99
`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	engine := New()

	assertInteger(t, evalInput(t, engine, env, `import("fixture:library").public_value`), 42)
	result := evalInput(t, engine, env, `import("fixture:library").private_value`)
	if err, ok := result.(*object.Error); !ok || !strings.Contains(err.MessageText(), `has no exported member "private_value"`) {
		t.Fatalf("private member result is %T (%v), want AttributeError", result, result)
	}
}

func TestModuleWithoutExportDeclarationExportsEveryBinding(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `let first = 1
let second = 2`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `let library = import("fixture:library")
library.first + library.second`)
	assertInteger(t, result, 3)
}

func TestEmptyModuleExportDeclarationExportsNothing(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), "export {}\nlet hidden = 42")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("fixture:library").hidden`)
	if failure, ok := result.(*object.Error); !ok || !strings.Contains(failure.MessageText(), `has no exported member "hidden"`) {
		t.Fatalf("result is %s, want missing member error", result.Inspect())
	}
}

func TestModuleCanExportImportedBinding(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv", "dependency.slv")
	writeSilverFile(t, filepath.Join(dir, "dependency.slv"), "let value = 42")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `export { dependency }
let dependency = import("fixture:dependency")`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("fixture:library").dependency.value`)
	assertInteger(t, result, 42)
}

func TestUndefinedModuleExportFailsImport(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), "export { missing }\nlet present = 42")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("fixture:library")`)
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T (%v), want *object.Error", result, result)
	}
	if got, want := err.MessageText(), `exported symbol "missing" is not defined`; got != want {
		t.Fatalf("error is %q, want %q", got, want)
	}
}

func TestImportRejectsNonStringPath(t *testing.T) {
	result := evalInput(t, New(), object.NewEnvironment(), `import(42)`)
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if got, want := err.MessageText(), "import path must be str, got int"; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestFunctionDestructuresModuleExports(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `
let message = "loaded"
let double = fn(value: int) int { return value * 2 }
`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("fixture:library")
let process = fn(double: call(int) int, message: str) int {
	return double(21)
}
process(library)
`)

	result := New().EvalFile(mainPath, object.NewEnvironment())
	assertInteger(t, result, 42)
}

func TestMatchingModuleParameterIsNotDestructured(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `let value = 42`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("fixture:library")
let read = fn(library: module) int { return library.value }
read(library)
`)

	result := New().EvalFile(mainPath, object.NewEnvironment())
	assertInteger(t, result, 42)
}

func TestDestructuredModuleExportMustMatchParameterType(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `let value = "wrong"`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("fixture:library")
let read = fn(value: int) int { return value }
read(library)
`)

	result := New().EvalFile(mainPath, object.NewEnvironment())
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "value": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestMissingModuleMember(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "module.slv")
	writeSilverFile(t, filepath.Join(dir, "module.slv"), `let present = 1`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("fixture:module").missing`)
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if !strings.Contains(err.MessageText(), `has no exported member "missing"`) {
		t.Fatalf("unexpected error: %s", err.MessageText())
	}
}

func TestCircularImport(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "a.slv", "b.slv")
	writeSilverFile(t, filepath.Join(dir, "a.slv"), `let b = import("fixture:b")`)
	writeSilverFile(t, filepath.Join(dir, "b.slv"), `let a = import("fixture:a")`)

	result := New().EvalFile(filepath.Join(dir, "a.slv"), object.NewEnvironment())
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if !strings.Contains(err.MessageText(), "circular import detected") {
		t.Fatalf("unexpected error: %s", err.MessageText())
	}
}

func evalInput(t *testing.T, engine *Evaluator, env *object.Environment, input string) object.Object {
	t.Helper()
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return engine.Eval(program, env)
}

func writeSilverFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func writePackageManifest(t *testing.T, directory string, members ...string) {
	t.Helper()
	var document strings.Builder
	document.WriteString("package: fixture\nauthors: []\nexport:\n")
	for _, member := range members {
		fmt.Fprintf(&document, "  - %q\n", filepath.ToSlash(member))
	}
	writeSilverFile(t, filepath.Join(directory, "package.yaml"), document.String())
	t.Setenv(importPathEnvironment, filepath.Join(directory, "package.yaml"))
}

func assertInteger(t *testing.T, result object.Object, want int64) {
	t.Helper()
	integer, ok := result.(*object.Integer)
	if !ok {
		t.Fatalf("result is %T (%v), want *object.Integer", result, result)
	}
	if integer.Value != want {
		t.Fatalf("result is %d, want %d", integer.Value, want)
	}
}
