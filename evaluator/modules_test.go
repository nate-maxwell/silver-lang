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
let double = fn(x) int { x * base.factor }
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
let double = fn(value: int) int { value * 2 }
`)
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, mainPath, `
let library = import("./library.slv")
let process = fn(double: call(int) int, message: str) int {
	double(21)
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
let read = fn(library: module) int { library.value }
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
let read = fn(value: int) int { value }
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
	if !strings.Contains(err.MessageText(), `has no member "missing"`) {
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
