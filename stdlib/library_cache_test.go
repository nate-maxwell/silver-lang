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
