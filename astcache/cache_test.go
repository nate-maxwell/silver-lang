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
	path := filepath.Join(t.TempDir(), "all.slvr")
	source := []byte(`
enum State { Ready, Waiting }
struct Person { name: string, age: int }
let values = [1, 2.5, True, "four", {"five": 5}]
let person = Person{"Ada", 36}
person.age = 37
let choose = fn(value: int) int { if (value > 0) { return -value } else { return 0 } }
let apply = fn(operation: call(int) int, value: int) int { operation(value) }
let module = import("./library.lib")
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
	path := filepath.Join(t.TempDir(), "main.slvr")
	source := []byte("let answer = 42")
	if err := astcache.Store(path, source, parse(t, path, source)); err != nil {
		t.Fatal(err)
	}

	if _, ok := astcache.Load(path, []byte("let answer = 43")); ok {
		t.Fatal("cache matched changed source")
	}
}

func TestDamagedCacheIsAMiss(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slvr")
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
