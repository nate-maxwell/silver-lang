package main

import (
	"os"
	"silver/astcache"
	"silver/packages"
	"testing"
)

func TestGeneratePackageCachesInternalOperatorMembers(t *testing.T) {
	t.Chdir(t.TempDir())
	sources := map[string]string{
		"api.slv":    "let value = 20 @@ 22",
		"helper.slv": "let calculate = fn() int { return 20 @@ 22 }",
		"ops.slv":    "operator @@ = fn(left: int, right: int) int { return left + right }",
	}
	for filename, input := range sources {
		if err := os.WriteFile(filename, []byte(input), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile("package.yaml", []byte("package: example\nmembers: [helper.slv, ops.slv]\nexport: [api.slv]\n"), 0600); err != nil {
		t.Fatal(err)
	}
	manifest, err := packages.ReadManifestFS(os.DirFS("."), "package.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := generatePackage(manifest); err != nil {
		t.Fatal(err)
	}
	for filename, input := range sources {
		if _, ok := astcache.Load(filename, []byte(input)); !ok {
			t.Errorf("missing cache for package member %q", filename)
		}
	}
}
