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
enum State { Ready, Waiting }
struct Person { name: string, age: int }
struct Handler { callback: call(value: int) }
let values = [1, 2.5, True, "four", {"five": 5}]
assert values[0] == 1, "cached assertion"
let person = Person{"Ada", 36}
person.age = 37
let choose = fn(value: int) int { if (value > 0) { return -value } else { return 0 } }
let apply = fn(operation: call(int) int, value: int) int { operation(value) }
struct Missing { message: str }
let read = fn() str | Missing { Missing{"missing"} }
try { read() } catch Missing err { err.message }
for value in values { continue }
for key, value in ({"answer": 42}) { print(key, value) }
while False { break }
let module = import("./library.slv")
module.member(choose(values[0]))
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
