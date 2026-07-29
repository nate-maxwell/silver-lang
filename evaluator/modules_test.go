package evaluator

import (
	"os"
	"path/filepath"
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
	writeMonkeyFile(t, filepath.Join(libDir, "base.monkey"), `let factor = 2;`)
	writeMonkeyFile(t, filepath.Join(libDir, "math.monkey"), `
let base = import("./base.monkey");
let double = fn(x) { x * base.factor; };
`)
	mainPath := filepath.Join(dir, "main.monkey")
	writeMonkeyFile(t, mainPath, `
let math = import("./lib/math.monkey");
math.double(21);
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
	path := filepath.Join(dir, "main.slvr")
	source := []byte("let answer = 41\nanswer")
	writeMonkeyFile(t, path, string(source))

	result := New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 41)
	if _, ok := astcache.Load(path, source); !ok {
		t.Fatal("EvalFile did not create a usable AST cache")
	}

	changedSource := []byte("let answer = 42\nanswer")
	writeMonkeyFile(t, path, string(changedSource))
	result = New().EvalFile(path, object.NewEnvironment())
	assertInteger(t, result, 42)
	if _, ok := astcache.Load(path, changedSource); !ok {
		t.Fatal("EvalFile did not refresh the AST cache after a source change")
	}
	if _, ok := astcache.Load(path, source); ok {
		t.Fatal("refreshed cache still matches the old source")
	}
}

func TestEvalFileRepairsDamagedASTCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slvr")
	source := []byte("42")
	writeMonkeyFile(t, path, string(source))
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
	writeMonkeyFile(t, filepath.Join(dir, "module.monkey"), `let value = 1;`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	engine := New()

	first := evalInput(t, engine, env, `import("./module.monkey")`)
	second := evalInput(t, engine, env, `import("./module.monkey")`)
	if first != second {
		t.Fatal("the same module was evaluated more than once")
	}
}

func TestMissingModuleMember(t *testing.T) {
	dir := t.TempDir()
	writeMonkeyFile(t, filepath.Join(dir, "module.monkey"), `let present = 1;`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)

	result := evalInput(t, New(), env, `import("./module.monkey").missing`)
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if !strings.Contains(err.Message, `has no member "missing"`) {
		t.Fatalf("unexpected error: %s", err.Message)
	}
}

func TestCircularImport(t *testing.T) {
	dir := t.TempDir()
	writeMonkeyFile(t, filepath.Join(dir, "a.monkey"), `let b = import("./b.monkey");`)
	writeMonkeyFile(t, filepath.Join(dir, "b.monkey"), `let a = import("./a.monkey");`)

	result := New().EvalFile(filepath.Join(dir, "a.monkey"), object.NewEnvironment())
	err, ok := result.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if !strings.Contains(err.Message, "circular import detected") {
		t.Fatalf("unexpected error: %s", err.Message)
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

func writeMonkeyFile(t *testing.T, path, contents string) {
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
