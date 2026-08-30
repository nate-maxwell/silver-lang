package evaluator

import (
	"os"
	"path/filepath"
	"silver/ast"
	"silver/astcache"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"strings"
	"testing"
)

func TestEvalFileWithNestedRelativeImport(t *testing.T) {
	dir := t.TempDir()
	libDir := filepath.Join(dir, "lib")
	if err := os.Mkdir(libDir, 0755); err != nil {
		t.Fatal(err)
	}
	writeSilverFile(t, filepath.Join(libDir, "base.slv"), `let factor = 2`)
	writeSilverFile(t, filepath.Join(libDir, "math.slv"), `
let base = import("./base.slv")
let double = fn(x) int { return x * base.factor }
`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let math = import("./lib/math.slv")
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

func TestEvalFileCreatesAndRefreshesASTCache(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.slv")
	source := []byte("let answer = 41\nanswer")
	writeSilverFile(t, path, string(source))

	result := New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 41)
	if _, ok := astcache.Load(path, source); !ok {
		t.Fatal("EvalFile did not create a usable AST cache")
	}

	changedSource := []byte("let answer = 42\nanswer")
	writeSilverFile(t, path, string(changedSource))
	result = New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 42)
	if _, ok := astcache.Load(path, changedSource); !ok {
		t.Fatal("EvalFile did not refresh the AST cache after a source change")
	}
	if _, ok := astcache.Load(path, source); ok {
		t.Fatal("refreshed cache still matches the old source")
	}
}

func TestEvalFileCachesFoldedAST(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	source := []byte("1 + 2 * 3")
	writeSilverFile(t, path, string(source))

	result := New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 7)
	program, ok := astcache.Load(path, source)
	if !ok {
		t.Fatal("could not load EvalFile's AST cache")
	}
	expression := program.Statements[0].(*ast.ExpressionStatement).Expression
	integer, ok := expression.(*ast.IntegerLiteral)
	if !ok || integer.Value != 7 {
		t.Fatalf("cached expression is %T (%v), want folded integer 7", expression, expression)
	}
}

func TestEvalFileRepairsDamagedASTCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	source := []byte("42")
	writeSilverFile(t, path, string(source))
	if err := os.WriteFile(astcache.Path(path), []byte("damaged"), 0600); err != nil {
		t.Fatal(err)
	}

	result := New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 42)
	if _, ok := astcache.Load(path, source); !ok {
		t.Fatal("EvalFile did not replace the damaged AST cache")
	}
}

func TestImportsAreCached(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "module.slv"), `let value = 1`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	engine := New()

	first := evalInput(t, engine, env, `import("./module.slv")`)
	second := evalInput(t, engine, env, `import("./module.slv")`)
	if first != second {
		t.Fatal("the same module was evaluated more than once")
	}
}

func TestImportSearchesSilverPath(t *testing.T) {
	sourceDir := t.TempDir()
	firstLibraryDir := t.TempDir()
	secondLibraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(secondLibraryDir, "library.slv"), `let value = 42`)
	t.Setenv(importPathEnvironment, strings.Join([]string{firstLibraryDir, secondLibraryDir}, string(os.PathListSeparator)))

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("library.slv").value`)
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
	if err := os.WriteFile(filepath.Join(packageDir, "example.yaml"), []byte(`
package: example
export:
  - ./nested/library.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(importPathEnvironment, packageDir)

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	engine := New()
	assertInteger(t, evalInput(t, engine, env, `import("nested/library.slv").value`), 42)
	assertInteger(t, evalInput(t, engine, env, `import("library.slv").value`), 42)

	result := evalInput(t, engine, env, `import("hidden.slv")`)
	if failure, ok := result.(*object.Error); !ok || !strings.Contains(failure.MessageText(), "could not read") {
		t.Fatalf("hidden import is %#v, want package-interface import error", result)
	}
}

func TestSilverPathPackageUsesAndCreatesASTCaches(t *testing.T) {
	sourceDir := t.TempDir()
	packageDir := t.TempDir()
	cachedPath := filepath.Join(packageDir, "cached.slv")
	uncachedPath := filepath.Join(packageDir, "uncached.slv")
	cachedSource := []byte("let value = 1")
	uncachedSource := []byte("let value = 42")
	writeSilverFile(t, cachedPath, string(cachedSource))
	writeSilverFile(t, uncachedPath, string(uncachedSource))

	// Store an AST whose source hash matches cached.slv but whose value makes
	// cache use observable. A package import should load this program instead
	// of reparsing the source.
	cachedProgram, parseError := ParseSource(cachedPath, []byte("let value = 41"))
	if parseError != nil {
		t.Fatal(parseError)
	}
	if err := astcache.Store(cachedPath, cachedSource, cachedProgram); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(packageDir, "package.yaml"), []byte(`
package: example
export:
  - cached.slv
  - uncached.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(importPathEnvironment, packageDir)

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	assertInteger(t, evalInput(t, New(), env, `import("cached.slv").value`), 41)
	if _, ok := astcache.Load(uncachedPath, uncachedSource); !ok {
		t.Fatal("package import did not create an AST cache for an uncached export")
	}
}

func TestSilverPathAcceptsExplicitSourceFile(t *testing.T) {
	sourceDir := t.TempDir()
	fileDir := t.TempDir()
	path := filepath.Join(fileDir, "single.slv")
	writeSilverFile(t, path, `let value = 42`)
	t.Setenv(importPathEnvironment, path)

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("single.slv").value`)
	assertInteger(t, result, 42)
}

func TestSilverPathRefreshesAfterEnvironmentChange(t *testing.T) {
	sourceDir := t.TempDir()
	packageDir := t.TempDir()
	writeSilverFile(t, filepath.Join(packageDir, "library.slv"), `let value = 42`)
	if err := os.WriteFile(filepath.Join(packageDir, "example.yaml"), []byte(`
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
	if result := evalInput(t, engine, env, `import("missing.slv")`); !isError(result) {
		t.Fatalf("initial import is %#v, want an error", result)
	}
	if err := os.Setenv(importPathEnvironment, packageDir); err != nil {
		t.Fatal(err)
	}
	assertInteger(t, evalInput(t, engine, env, `import("library.slv").value`), 42)
}

func TestPackageOperatorsAreScopedByManifest(t *testing.T) {
	packageDir := t.TempDir()
	writePackageOperatorFixture(t, packageDir)
	t.Setenv(importPathEnvironment, packageDir)

	mainPath := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, mainPath, `
let foo = import("foo.slv")
let bar = import("bar.slv")
foo.apply(foo.make(4), 2) * 10 + bar.apply(foo.make(4), 2)
`)
	result := New().EvalFile(mainPath, object.NewEnvironment())
	assertInteger(t, result, 427)
	for _, name := range []string{"foo.slv", "foo_operators.slv", "bar.slv", "bar_operators.slv"} {
		if _, err := os.Stat(astcache.Path(filepath.Join(packageDir, name))); err != nil {
			t.Fatalf("package operator cache for %s was not created: %v", name, err)
		}
	}

	// A fresh evaluator prepares both packages again and consumes their
	// context-tagged caches without leaking either operator grammar.
	result = New().EvalFile(mainPath, object.NewEnvironment())
	assertInteger(t, result, 427)
}

func TestImportedPackageOperatorIsNotVisibleToImporter(t *testing.T) {
	packageDir := t.TempDir()
	writePackageOperatorFixture(t, packageDir)
	t.Setenv(importPathEnvironment, packageDir)

	mainPath := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, mainPath, `
let foo = import("foo.slv")
foo.make(4) @@ 2
`)
	result := New().EvalFile(mainPath, object.NewEnvironment())
	failure, ok := result.(*object.Error)
	if !ok || !strings.Contains(failure.MessageText(), "could not parse") {
		t.Fatalf("result is %#v, want syntax error for package-local operator", result)
	}
}

func writePackageOperatorFixture(t *testing.T, directory string) {
	t.Helper()
	writeSilverFile(t, filepath.Join(directory, "foo_operators.slv"), `operator @@ = fn(left, right) int { return 999 }`)
	writeSilverFile(t, filepath.Join(directory, "foo.slv"), `
export { Foo, make, apply }
let operators = import("./foo_operators.slv")
struct Foo {
    value: int
    @@: call(self: Foo, other: int) int
}
let overload = fn(self: Foo, other: int) int { return self.value * 10 + other }
let make = fn(value: int) Foo { return Foo{value, overload} }
let apply = fn(left: Foo, right: int) int { return left @@ right }
`)
	if err := os.WriteFile(filepath.Join(directory, "foo.yaml"), []byte(`
package: package_foo
export:
  - ./foo.slv
  - ./foo_operators.slv
`), 0644); err != nil {
		t.Fatal(err)
	}

	writeSilverFile(t, filepath.Join(directory, "bar_operators.slv"), `operator @@ = fn(left, right) int { return 7 }`)
	writeSilverFile(t, filepath.Join(directory, "bar.slv"), `
export { apply }
let operators = import("./bar_operators.slv")
let foo = import("./foo.slv")
let apply = fn(left: foo.Foo, right: int) int { return left @@ right }
`)
	if err := os.WriteFile(filepath.Join(directory, "bar.yaml"), []byte(`
package: package_bar
export:
  - ./bar.slv
  - ./bar_operators.slv
`), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestImportPrefersImporterDirectoryOverSilverPath(t *testing.T) {
	sourceDir := t.TempDir()
	libraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(sourceDir, "library.slv"), `let value = 1`)
	writeSilverFile(t, filepath.Join(libraryDir, "library.slv"), `let value = 2`)
	t.Setenv(importPathEnvironment, libraryDir)

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("library.slv").value`)
	assertInteger(t, result, 1)
}

func TestSilverPathModuleResolvesRelativeImportsBesideItself(t *testing.T) {
	sourceDir := t.TempDir()
	libraryDir := t.TempDir()
	writeSilverFile(t, filepath.Join(libraryDir, "dependency.slv"), `let value = 21`)
	writeSilverFile(t, filepath.Join(libraryDir, "library.slv"), `
let dependency = import("./dependency.slv")
let value = dependency.value * 2
`)
	t.Setenv(importPathEnvironment, libraryDir)

	env := object.NewEnvironment()
	env.SetSourceDir(sourceDir)
	result := evalInput(t, New(), env, `import("library.slv").value`)
	assertInteger(t, result, 42)
}

func TestSilverStandardLibraryImportsAreCached(t *testing.T) {
	engine := New()
	env := object.NewEnvironment()

	first := evalInput(t, engine, env, `import("testing")`)
	second := evalInput(t, engine, env, `import("testing")`)
	if first != second {
		t.Fatal("the same Silver standard-library module was evaluated more than once")
	}
}

func TestSilverStandardLibraryDoesNotReplacePathImports(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "testing.slv"), `let origin = "user"`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `
let standard = import("testing")
let user = import("./testing.slv")
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
	writeSilverFile(t, filepath.Join(dir, "module.slv"), `let value = 42`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `
let module_path = "./module.slv"
import(module_path).value
`)
	assertInteger(t, result, 42)
}

func TestModuleExportDeclarationLimitsPublicBindings(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `
export { public_value }
let public_value = 42
let private_value = 99
`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	engine := New()

	assertInteger(t, evalInput(t, engine, env, `import("./library.slv").public_value`), 42)
	result := evalInput(t, engine, env, `import("./library.slv").private_value`)
	if err, ok := result.(*object.Error); !ok || !strings.Contains(err.MessageText(), `has no exported member "private_value"`) {
		t.Fatalf("private member result is %T (%v), want AttributeError", result, result)
	}
}

func TestModuleWithoutExportDeclarationExportsEveryBinding(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `let first = 1
let second = 2`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `let library = import("./library.slv")
library.first + library.second`)
	assertInteger(t, result, 3)
}

func TestEmptyModuleExportDeclarationExportsNothing(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "library.slv"), "export {}\nlet hidden = 42")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("./library.slv").hidden`)
	if _, ok := result.(*object.Error); !ok {
		t.Fatalf("result is %T (%v), want *object.Error", result, result)
	}
}

func TestModuleCanExportImportedBinding(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "dependency.slv"), "let value = 42")
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `export { dependency }
let dependency = import("./dependency.slv")`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("./library.slv").dependency.value`)
	assertInteger(t, result, 42)
}

func TestUndefinedModuleExportFailsImport(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "library.slv"), "export { missing }\nlet present = 42")
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("./library.slv")`)
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
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `
let message = "loaded"
let double = fn(value: int) int { return value * 2 }
`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("./library.slv")
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
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `let value = 42`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("./library.slv")
let read = fn(library: module) int { return library.value }
read(library)
`)

	result := New().EvalFile(mainPath, object.NewEnvironment())
	assertInteger(t, result, 42)
}

func TestDestructuredModuleExportMustMatchParameterType(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "library.slv"), `let value = "wrong"`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("./library.slv")
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
	writeSilverFile(t, filepath.Join(dir, "module.slv"), `let present = 1`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("./module.slv").missing`)
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
	writeSilverFile(t, filepath.Join(dir, "a.slv"), `let b = import("./b.slv")`)
	writeSilverFile(t, filepath.Join(dir, "b.slv"), `let a = import("./a.slv")`)

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
