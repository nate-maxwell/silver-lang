package astcache_test

import (
	"os"
	"path/filepath"
	"silver/ast"
	"silver/astcache"
	"silver/lexer"
	"silver/parser"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "all.slv")
	source := []byte(`
export { State, Person, Details, values, choose, module }
operator @ = fn(left: int, right: int) int { left + right }
type Point = struct { x: float, y: float }
type Direction = enum { North, South }
type Names = array[str]
type Transform = call(Point) Point
type Read = call() array[Point]
type State = enum { Ready, Waiting }
type Person = struct { name: string, age: int }
type Details = struct { person :: Person }
type Handler = struct { callback: call(value: int) }
let values = [1, 2.5, True, "four", {"five": 5}]
assert values[0] == 1, "cached assertion"
let person = Person{"Ada", 36}
person.age = 37
let choose = fn(value: int) int { if (value > 0) { return -value } else { return 0 } }
let variadic = fn(prefix: str, parts: str...) { parts }
let label = switch values[0] {
case 1:
    "one"
default:
    "other"
}
let apply = fn(operation: call(int) int, value: int) int { operation(value) }
type Missing = struct { message: str }
let read = fn() str | Missing { Missing{"missing"} }
try { read() } catch Missing err { err.message }
for value in values { continue }
for key, value in ({"answer": 42}) { print(key, value) }
while False { break }
let module = import("./library.slv")
module.member(choose(values[0]))
1 @ 2
`)
	program := parse(t, path, source)

	if err := astcache.Store(path, source, program); err != nil {
		t.Fatal(err)
	}
	loaded, ok := astcache.Load(path, source)
	if !ok {
		t.Fatal("cache was not loaded")
	}
	if got, want := loaded.String(), program.String(); got != want {
		t.Fatalf("loaded AST differs:\ngot:  %s\nwant: %s", got, want)
	}
	if got := loaded.Position().Source; got != path {
		t.Fatalf("loaded source is %q, want %q", got, path)
	}
}

func TestSourceChangeInvalidatesCache(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	source := []byte("let answer = 42")
	if err := astcache.Store(path, source, parse(t, path, source)); err != nil {
		t.Fatal(err)
	}

	if _, ok := astcache.Load(path, []byte("let answer = 43")); ok {
		t.Fatal("cache matched changed source")
	}
}

func TestTypeAliasRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "aliases.slv")
	source := []byte(`type Read = call() array[array[int]]`)
	program := parse(t, path, source)
	if err := astcache.Store(path, source, program); err != nil {
		t.Fatal(err)
	}
	loaded, ok := astcache.Load(path, source)
	if !ok {
		t.Fatal("type alias cache was not loaded")
	}
	alias := loaded.Statements[0].(*ast.TypeStatement).Value.(*ast.TypeAnnotation)
	if !alias.IsCallSignature() || len(alias.ParameterTypes) != 0 {
		t.Fatal("cache lost zero-argument call signature")
	}
	if got, want := alias.String(), "call() array[array[int]]"; got != want {
		t.Fatalf("loaded alias = %q, want %q", got, want)
	}
}

func TestTypeDeclarationRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "types.slv")
	source := []byte("type Empty = struct {}\ntype Nothing = enum {}\ntype Read = call() array[int]")
	program := parse(t, path, source)
	if err := astcache.Store(path, source, program); err != nil {
		t.Fatal(err)
	}
	loaded, ok := astcache.Load(path, source)
	if !ok {
		t.Fatal("type declaration cache was not loaded")
	}
	if got, want := loaded.String(), program.String(); got != want {
		t.Fatalf("loaded declarations = %q, want %q", got, want)
	}
	alias := loaded.Statements[2].(*ast.TypeStatement).Value.(*ast.TypeAnnotation)
	if !alias.IsCallSignature() || len(alias.ParameterTypes) != 0 {
		t.Fatal("cache lost zero-argument call signature")
	}
}

func TestParserContextSeparatesCaches(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	source := []byte("let answer = 42")
	context := []byte("package operators: @@=60")
	if err := astcache.StoreWithContext(path, source, context, parse(t, path, source)); err != nil {
		t.Fatal(err)
	}

	if _, ok := astcache.LoadWithContext(path, source, context); !ok {
		t.Fatal("cache did not match its parser context")
	}
	if _, ok := astcache.LoadWithContext(path, source, []byte("package operators: @@=70")); ok {
		t.Fatal("cache matched a different parser context")
	}
	if _, ok := astcache.Load(path, source); ok {
		t.Fatal("contextual cache matched a context-free load")
	}
}

func TestLoadBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "embedded.slv")
	source := []byte("let answer = 42")
	if err := astcache.Store(path, source, parse(t, path, source)); err != nil {
		t.Fatal(err)
	}
	cache, err := os.ReadFile(astcache.Path(path))
	if err != nil {
		t.Fatal(err)
	}

	loaded, ok := astcache.LoadBytes(path, source, cache)
	if !ok || loaded == nil {
		t.Fatal("in-memory cache was not loaded")
	}
	if _, ok := astcache.LoadBytes(path, []byte("let answer = 43"), cache); ok {
		t.Fatal("in-memory cache matched changed source")
	}
}

func TestTemplateStringRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "template.slv")
	source := []byte("let value = 42\nlet template = ```answer: {value}```\ntemplate.eval()")
	program := parse(t, path, source)

	if err := astcache.Store(path, source, program); err != nil {
		t.Fatal(err)
	}
	loaded, ok := astcache.Load(path, source)
	if !ok {
		t.Fatal("template string AST cache was not loaded")
	}
	if got, want := loaded.String(), program.String(); got != want {
		t.Fatalf("loaded template AST differs:\ngot:  %s\nwant: %s", got, want)
	}
}

func TestDamagedCacheIsAMiss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	if err := os.WriteFile(astcache.Path(path), []byte("not an AST cache"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, ok := astcache.Load(path, []byte("42")); ok {
		t.Fatal("damaged cache was loaded")
	}
}

func parse(t *testing.T, path string, source []byte) *ast.Program {
	t.Helper()
	p := parser.New(lexer.NewWithSource(string(source), path))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return program
}
