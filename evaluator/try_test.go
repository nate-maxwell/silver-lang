package evaluator

import (
	"silver/object"
	"strings"
	"testing"
)

func TestCatchBindsCaughtStructToDeclaredVariable(t *testing.T) {
	evaluated := testEval(`
struct Missing { message: str, path: str }
let read = fn() str | Missing { Missing{"not found", "/tmp/data"} }
try {
	read()
} catch Missing err {
	err.message + ": " + err.path
}
`)
	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "not found: /tmp/data" {
		t.Fatalf("result is %#v, want fields on explicitly bound error", evaluated)
	}
}

func TestCatchClausesMatchByNominalStructType(t *testing.T) {
	testBooleanObject(t, testEval(`
struct Missing { message: str }
struct Denied { message: str }
let read = fn() str | Denied { Denied{"no"} }
try {
	read()
} catch Missing err {
	False
} catch Denied err {
	err.message == "no"
}
`), true)
}

func TestUnmatchedErrorPropagatesToOuterCatch(t *testing.T) {
	testBooleanObject(t, testEval(`
struct Missing { message: str }
struct Denied { message: str }
let read = fn() str | Denied { Denied{"no"} }
try {
	try {
		read()
	} catch Missing err {
		False
	}
} catch Denied err {
	err.message == "no"
}
`), true)
}

func TestCaughtErrorDoesNotEscapeFunction(t *testing.T) {
	value := testEval(`
struct Missing { message: str }
let read = fn() str | Missing { Missing{"not found"} }
let recover = fn() str {
	try {
		read()
	} catch Missing err {
		err.message
	}
}
recover()
`)
	text, ok := value.(*object.String)
	if !ok || text.Value != "not found" {
		t.Fatalf("result is %#v, want recovered string", value)
	}
}

func TestFunctionMustDeclarePropagatedError(t *testing.T) {
	result := testEval(`
struct Missing { message: str }
let read = fn() str | Missing { Missing{"not found"} }
let caller = fn() str { read() }
caller()
`)
	err, ok := result.(*object.Error)
	if !ok || !strings.Contains(err.MessageText(), `error Missing escaped "caller" but is not declared`) {
		t.Fatalf("result is %#v, want undeclared error error", result)
	}
}

func TestRuntimeErrorsUseTheirBuiltinErrorStruct(t *testing.T) {
	result := testEval(`
try {
	1 + True
} catch TypeError err {
	err.message
}
`)
	message, ok := result.(*object.String)
	if !ok || message.Value != "type mismatch: INTEGER + BOOLEAN" {
		t.Fatalf("result is %#v, want the caught TypeError message", result)
	}
}
