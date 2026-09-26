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
	for _, preload := range []string{"", `import("fixture:counter").next()`} {
		t.Run(fmt.Sprintf("preload=%t", preload != ""), func(t *testing.T) {
			dir := t.TempDir()
			writePackageManifest(t, dir, "counter.slv")
			writeSilverFile(t, filepath.Join(dir, "counter.slv"), counterModule)
			env := object.NewEnvironment()
			env.SetSourceDir(dir)
			engine := New()
			input := `let first = ` + templateLiteral(`{import("fixture:counter").next()}`) + `
let second = ` + templateLiteral(`{import("fixture:counter").next()}`) + "\n" + preload + `
[first.eval(), first.eval(), second.eval(), import("fixture:counter").next()]
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
			assertInteger(t, evalInput(t, New(), env, `import("fixture:counter").next()`), 1)
		})
	}
}

func TestTemplateImportsPreserveNominalDefinitions(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "identity.slv")
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
let template = `+templateLiteral(`{capture(import("fixture:identity"))}`)+`
template.eval()
let first = captured
template.eval()
let second = import("fixture:identity")
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
let template = `+templateLiteral(`{import("core:testing").run("pass", fn() {})}`)+`
template.eval()
template.eval()
import("core:testing").summary().total
`)
	assertInteger(t, result, 2)
}

func TestTemplateReturnedByModuleCanImportItsOwner(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "owner.slv")
	writeSilverFile(t, filepath.Join(dir, "owner.slv"), counterModule+`
let template = `+templateLiteral(`{import("fixture:owner").next()}`))
	env := object.NewEnvironment()
	env.SetSourceDir(dir)
	result := evalInput(t, New(), env, `
let owner = import("fixture:owner")
[owner.template.eval(), owner.template.eval(), owner.next()]
`)
	if want := "[1, 2, 3]"; result.Inspect() != want {
		t.Fatalf("result is %s, want %s", result.Inspect(), want)
	}
}

func TestRegisteredManifestPathCasingPreservesModuleIdentity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path casing policy")
	}
	dir := t.TempDir()
	writePackageManifest(t, dir, "Identity.slv")
	filename := filepath.Join(dir, "Identity.slv")
	writeSilverFile(t, filename, `type Item = struct { value: int }
 type Choice = enum { First }`)
	env := object.NewEnvironment()
	engine := New()
	first := evalInput(t, engine, env, `import("fixture:Identity")`)
	module, ok := first.(*object.Module)
	if !ok {
		t.Fatalf("import returned %s", first.Inspect())
	}
	if module.ID != source.FileID(filename) {
		t.Fatal("lost canonical file identity")
	}
	t.Setenv(importPathEnvironment, strings.ToUpper(filepath.Join(dir, "package.yaml")))
	second := evalInput(t, engine, env, `import("fixture:Identity")`)
	if first != second {
		t.Fatal("manifest path casing changed module identity")
	}
	assertManifestImportError(t, evalInput(t, engine, env, `import("fixture:identity")`), "not found")
}

func TestWindowsPackageImportsShareModuleAndOperatorIdentity(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path casing policy")
	}
	dir := t.TempDir()
	writePackageOperatorFixture(t, dir)
	env := object.NewEnvironment()
	engine := New()
	evalInput(t, engine, env, `let first = import("package_foo:foo")`)
	t.Setenv(importPathEnvironment, strings.ToUpper(filepath.Join(dir, "foo", "package.yaml")))
	result := evalInput(t, engine, env, `let second = import("package_foo:foo")
 [first == second, second.apply(first.make(4), 2)]`)
	if want := "[true, 42]"; result.Inspect() != want {
		t.Fatalf("result is %s, want %s", result.Inspect(), want)
	}
}

func TestImportCyclesShareIdentityAndTemplateLoadingState(t *testing.T) {
	for _, alias := range []string{"fixture:Cycle"} {
		for _, template := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/template=%t", alias, template), func(t *testing.T) {

				dir := t.TempDir()
				writePackageManifest(t, dir, "guard.slv", "Cycle.slv")
				writeSilverFile(t, filepath.Join(dir, "guard.slv"), counterModule)
				expression := fmt.Sprintf("import(%q)", alias)
				if template {
					expression = templateLiteral("{"+expression+"}") + ".eval()"
				}
				// Bound recursion even if loading state is lost, and expose a
				// duplicate evaluation before it can obscure the original cycle.
				writeSilverFile(t, filepath.Join(dir, "Cycle.slv"), `
let guard = import("fixture:guard")
let attempt = guard.next()
assert attempt <= 2, "module evaluated twice"
let value = if attempt == 1 { `+expression+` } else { 42 }`)
				env := object.NewEnvironment()
				env.SetSourceDir(dir)
				engine := New()
				result := evalInput(t, engine, env, `import("fixture:Cycle")`)
				failure, ok := result.(*object.Error)
				if !ok || !strings.Contains(failure.MessageText(), "circular import detected") {
					t.Fatalf("result is %s, want circular import error", result.Inspect())
				}

				// Failed loads must leave no stale loading marker or cached module.
				// The already-loaded guard makes the next evaluation succeed.
				assertInteger(t, evalInput(t, engine, env, `import("fixture:Cycle").value`), 42)
			})
		}
	}
}
