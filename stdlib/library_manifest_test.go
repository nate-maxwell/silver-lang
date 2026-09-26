package stdlib

import (
	"io"
	"silver/object"
	"strings"
	"testing"
	"testing/fstest"
)

func TestBundledManifestControlsMembershipAndExports(t *testing.T) {
	files := fstest.MapFS{
		"silver/example/package.yaml":      {Data: []byte("package: example\nmembers: [helper.slv]\nexport: [example.slv, nested/entry.slv]\n")},
		"silver/example/example.slv":       {Data: []byte("let value = 42")},
		"silver/example/nested/entry.slv":  {Data: []byte("let value = 21")},
		"silver/example/helper.slv":        {Data: []byte("let value = 1")},
		"silver/example/unlisted.slv":      {Data: []byte("let value = 0")},
		"silver/unpackaged/unpackaged.slv": {Data: []byte("let value = 99")},
	}
	library, err := loadPackageManifests(files, nil)
	if err != nil {
		t.Fatal(err)
	}
	main, ok := library.LookupSource("core:example")
	if !ok {
		t.Fatal("missing package entry")
	}
	entry, ok := library.LookupSource("core:example/nested/entry")
	if !ok || entry.PackageID != main.PackageID {
		t.Fatal("nested entry lost package identity")
	}
	for _, name := range []string{"example/helper", "example/unlisted", "unpackaged"} {
		if _, ok := library.LookupSource("core:" + name); ok {
			t.Fatalf("undeclared export %q is importable", name)
		}
	}
	helper, ok := library.LookupSourceFrom("core:example/helper", entry.PackageID)
	if !ok || helper.PackageID != main.PackageID {
		t.Fatal("qualified internal import lost package identity")
	}
	if members := library.SourceMembers(main.PackageID); len(members) != 3 {
		t.Fatalf("got %d members, want 3", len(members))
	}
	delete(files, "silver/example/package.yaml")
	library, err = loadPackageManifests(files, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := library.LookupSource("core:example"); ok {
		t.Fatal("module remained importable without its manifest")
	}
}

func TestNativeModulesRequireManifests(t *testing.T) {
	native := map[string]*object.Module{"example": {Path: "example"}}
	files := fstest.MapFS{}
	if _, err := loadPackageManifests(files, native); err == nil || !strings.Contains(err.Error(), "requires a package.yaml") {
		t.Fatalf("native module without manifest was accepted: %v", err)
	}
	files["native/example/package.yaml"] = &fstest.MapFile{Data: []byte("package: example\nexport: []\n")}
	library, err := loadPackageManifests(files, native)
	if err != nil {
		t.Fatal(err)
	}
	if module, ok := library.LookupModule("core:example"); !ok || module != native["example"] {
		t.Fatal("native manifest did not expose its implementation")
	}
}

func TestPackageCombinesNativeBindingsWithSilverEntry(t *testing.T) {
	native := map[string]*object.Module{
		"example": {Path: "example", Exports: map[string]object.Object{"native_value": &object.Integer{Value: 42}}},
	}
	files := fstest.MapFS{
		"silver/example/package.yaml": {Data: []byte("package: example\nmembers: [helper.slv]\nexport: [example.slv]\n")},
		"silver/example/example.slv":  {Data: []byte("export { native_value }")},
		"silver/example/helper.slv":   {Data: []byte("let value = 1")},
	}
	library, err := loadPackageManifests(files, native)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := library.LookupModule("core:example"); ok {
		t.Fatal("native bindings bypass the Silver entry")
	}
	entry, ok := library.LookupSource("core:example")
	if !ok || entry.NativeBindings["native_value"] != native["example"].Exports["native_value"] {
		t.Fatal("native bindings missing from the public entry")
	}
	helper, ok := library.LookupSourceFrom("core:example/helper", entry.PackageID)
	if !ok || len(helper.NativeBindings) != 0 {
		t.Fatal("entry bindings leaked into a separate member")
	}
	files["silver/example/package.yaml"].Data = []byte("package: example\nmembers: [example.slv]\nexport: [helper.slv]\n")
	if _, err := loadPackageManifests(files, native); err == nil || !strings.Contains(err.Error(), "requires an exported example.slv") {
		t.Fatalf("native bindings without a public entry were accepted: %v", err)
	}
}

func TestStandardLibraryManifestEntryPoints(t *testing.T) {
	library := New(io.Discard, &object.Null{}, &object.Boolean{Value: true}, &object.Boolean{Value: false})
	for _, name := range []string{"args", "http", "http/client", "http/cookiejar", "http/cookies", "http/server", "json", "logging", "networking", "path", "testing"} {
		if _, ok := library.LookupSource("core:" + name); !ok {
			t.Errorf("manifest migration lost %q", name)
		}
	}
	for _, name := range []string{"arrays", "collections", "core", "io", "maps", "math", "random", "regex", "string", "system", "terminal", "time"} {
		if _, ok := library.LookupModule("core:" + name); !ok {
			t.Errorf("manifest migration lost %q", name)
		}
	}
}
