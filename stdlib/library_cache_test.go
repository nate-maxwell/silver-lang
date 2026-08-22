package stdlib

import (
	"io"
	"silver/astcache"
	"silver/object"
	"testing"
)

func TestEmbeddedSilverModuleCacheIsCurrent(t *testing.T) {
	library := New(io.Discard, &object.Null{}, &object.Boolean{Value: true}, &object.Boolean{Value: false})
	for name, module := range library.sourceModules {
		if _, ok := astcache.LoadBytes(module.sourceName, []byte(module.source), module.cache); !ok {
			t.Errorf("embedded AST cache for Silver module %q is missing or stale; run go generate ./stdlib", name)
		}
	}
}

func TestJSONIsSilverAuthoredModule(t *testing.T) {
	library := New(io.Discard, &object.Null{}, &object.Boolean{Value: true}, &object.Boolean{Value: false})
	if _, native := library.modules["json"]; native {
		t.Fatal("json is registered as a native Go module")
	}
	module, silver := library.sourceModules["json"]
	if !silver || module.sourceName != "stdlib/silver/json/json.slv" {
		t.Fatalf("json Silver source module is %#v, present=%t", module, silver)
	}
}
