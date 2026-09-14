package evaluator

import (
	"fmt"
	"path/filepath"
	"runtime"
	"silver/object"
	"silver/source"
	"strings"
	"testing"
)

const counterModule = `
let count = 0
let next = fn() int {
    count = count + 1
    return count
}
`

func TestTemplateImportsShareSessionModules(t *testing.T) {
	for _, preload := range []string{"", `import("./counter.slv").next()`} {
		t.Run(fmt.Sprintf("preload=%t", preload != ""), func(t *testing.T) {
			dir := t.TempDir()
			writeSilverFile(t, filepath.Join(dir, "counter.slv"), counterModule)
			env := object.NewEnvironment()
			env.SetSourceDir(dir)
			engine := New()
			input := `let first = ` + templateLiteral(`{import("./counter.slv").next()}`) + `
let second = ` + templateLiteral(`{import("./counter.slv").next()}`) + "\n" + preload + `
[first.eval(), first.eval(), second.eval(), import("./counter.slv").next()]
`
			result := evalInput(t, engine, env, input)
			want := "[1, 2, 3, 4]"
			if preload != "" {
				want = "[2, 3, 4, 5]"
			}
			if result.Inspect() != want {
				t.Fatalf("result is %s, want %s", result.Inspect(), want)
			}

			// A new interpreter session must still get its own module instance.
			assertInteger(t, evalInput(t, New(), env, `import("./counter.slv").next()`), 1)
		})
	}
}

func TestTemplateImportsPreserveNominalDefinitions(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "identity.slv"), `
type Item = struct { value: int }
type Choice = enum { First }
let accept = fn(item: Item) int { return item.value }
`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	result := evalInput(t, New(), env, `
let captured = null
let capture = fn(library: module) str {
    captured = library
    return ""
}
let template = `+templateLiteral(`{capture(import("./identity.slv"))}`)+`
template.eval()
let first = captured
template.eval()
let second = import("./identity.slv")
[first == captured, first == second, first.Item == second.Item,
 first.Choice.First == second.Choice.First, first.accept(second.Item{42})]
`)
	if want := "[true, true, true, true, 42]"; result.Inspect() != want {
		t.Fatalf("result is %s, want %s", result.Inspect(), want)
	}
}

func TestTemplateImportsShareBundledModules(t *testing.T) {
	env := object.NewEnvironment()
	engine := NewWithOutput(nil)
	result := evalInput(t, engine, env, `
let template = `+templateLiteral(`{import("testing").run("pass", fn() {})}`)+`
template.eval()
template.eval()
import("testing").summary().total
`)
	assertInteger(t, result, 2)
}

func TestTemplateReturnedByModuleCanImportItsOwner(t *testing.T) {
	dir := t.TempDir()
	writeSilverFile(t, filepath.Join(dir, "owner.slv"), counterModule+`
let template = `+templateLiteral(`{import("./owner.slv").next()}`))
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	result := evalInput(t, New(), env, `
let owner = import("./owner.slv")
[owner.template.eval(), owner.template.eval(), owner.next()]
`)
	if want := "[1, 2, 3]"; result.Inspect() != want {
		t.Fatalf("result is %s, want %s", result.Inspect(), want)
	}
}

func TestImportsCanonicalizeWindowsPaths(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path casing policy")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "Identity.slv")
	writeSilverFile(t, path, `type Item = struct { value: int }
type Choice = enum { First }`)
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	engine := New()
	first := evalInput(t, engine, env, `import("./Identity.slv")`)
	module, ok := first.(*object.Module)
	if !ok {
		t.Fatalf("import returned %s", first.Inspect())
	}
	if module.ID != source.FileID(path) || module.Path != path {
		t.Fatalf("module identity/path is %q/%q, want canonical ID and original diagnostic path", module.ID, module.Path)
	}
	for _, alias := range []string{"./identity.slv", "./IDENTITY.slv", strings.ToUpper(filepath.ToSlash(path)), "./sub/../Identity.slv"} {
		second := evalInput(t, engine, env, fmt.Sprintf("import(%q)", alias))
		if second != first {
			t.Fatalf("import(%q) returned a different module: %s", alias, second.Inspect())
		}
	}
	result := evalInput(t, engine, env, `
let first = import("./Identity.slv")
let second = import("./identity.slv")
let accept = fn(item: first.Item) int { return item.value }
[first.Item == second.Item, first.Choice.First == second.Choice.First, accept(second.Item{42})]
`)
	if want := "[true, true, 42]"; result.Inspect() != want {
		t.Fatalf("result is %s, want %s", result.Inspect(), want)
	}
}

func TestWindowsPackageImportsShareModuleAndOperatorIdentity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path casing policy")
	}
	dir := t.TempDir()
	writePackageOperatorFixture(t, dir)
	t.Setenv(importPathEnvironment, dir)
	env := object.NewEnvironment()
	env.SetSourceDir(t.TempDir())
	input := fmt.Sprintf(`
let first = import(%q)
let second = import("FOO.SLV")
[first == second, second.apply(first.make(4), 2)]
`, strings.ToUpper(filepath.ToSlash(filepath.Join(dir, "foo.slv"))))
	result := evalInput(t, New(), env, input)
	if want := "[true, 42]"; result.Inspect() != want {
		t.Fatalf("result is %s, want %s", result.Inspect(), want)
	}
}

func TestImportCyclesShareIdentityAndTemplateLoadingState(t *testing.T) {
	for _, alias := range []string{"./Cycle.slv", "./cycle.slv"} {
		for _, template := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/template=%t", alias, template), func(t *testing.T) {
				if alias == "./cycle.slv" && runtime.GOOS != "windows" {
					t.Skip("Windows path casing policy")
				}
				dir := t.TempDir()
				writeSilverFile(t, filepath.Join(dir, "guard.slv"), counterModule)
				expression := fmt.Sprintf("import(%q)", alias)
				if template {
					expression = templateLiteral("{"+expression+"}") + ".eval()"
				}
				// Bound recursion even if loading state is lost, and expose a
				// duplicate evaluation before it can obscure the original cycle.
				writeSilverFile(t, filepath.Join(dir, "Cycle.slv"), `
let guard = import("./guard.slv")
assert guard.next() == 1, "module evaluated twice"
let value = `+expression)
				env := object.NewEnvironment()
				env.SetSourceDir(dir)
				engine := New()
				result := evalInput(t, engine, env, `import("./Cycle.slv")`)
				failure, ok := result.(*object.Error)
				if !ok || !strings.Contains(failure.MessageText(), "circular import detected") {
					t.Fatalf("result is %s, want circular import error", result.Inspect())
				}

				// Failed loads must leave no stale loading marker or cached module.
				writeSilverFile(t, filepath.Join(dir, "Cycle.slv"), `let value = 42`)
				assertInteger(t, evalInput(t, engine, env, `import("./Cycle.slv").value`), 42)
			})
		}
	}
}
