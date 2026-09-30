package evaluator

import (
	"fmt"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"testing"
)

func TestBindingContractForms(t *testing.T) {
	for _, tt := range []struct {
		declaration string
		strict      bool
	}{
		{"let value = 32", false},
		{"let value: any = 32", false},
		{"let value: int = 32", true},
		{"let value := 32", true},
	} {
		t.Run(tt.declaration, func(t *testing.T) {
			result := testEval(tt.declaration + "\nvalue = 14\nvalue = \"changed\"\nvalue")
			if !tt.strict {
				testStringObject(t, result, "changed")
				return
			}
			err, ok := result.(*object.Error)
			if !ok || err.MessageText() != `type mismatch for binding "value": expected int, got str` {
				t.Fatalf("result = %v, want integer binding type error", result)
			}
		})
	}
}

func TestInferredBindingRuntimeTypes(t *testing.T) {
	for _, tt := range []struct {
		name, setup, initial, same, different, typeName, otherType string
	}{
		{"integer", "", "32", "14", `"wrong"`, "int", "str"},
		{"float", "", "3.2", "1.4", "14", "float", "int"},
		{"boolean", "", "True", "False", "14", "bool", "int"},
		{"string", "", `"first"`, `"second"`, "14", "str", "int"},
		{"null", "", "if False { 1 }", "if False { 2 }", "14", "null", "int"},
		{"array", "", "[32]", `["different element type"]`, "{}", "array", "map"},
		{"map", "", `{"x": 32}`, `{14: "different contents"}`, "[]", "map", "array"},
		{"call", coreImport, "fn() int { return 32 }", "core.len", "14", "call", "int"},
		{"native call", coreImport, "core.len", "fn() str { return \"ok\" }", "14", "call", "int"},
		{"struct", "type A = struct { x: int }\ntype B = struct { x: int }", "A{32}", "A{14}", "B{14}", "A", "B"},
		{"enum", "type A = enum { First, Second }\ntype B = enum { First }", "A.First", "A.Second", "B.First", "A", "B"},
		{"type value", "", "int", "str", "14", "type", "int"},
		{"type alias", "type Names = array[str]", "Names", "int", "14", "type", "int"},
		{"struct definition", "type A = struct {}\ntype B = struct {}", "A", "B", "B{}", "struct", "B"},
		{"enum definition", "type A = enum { First }\ntype B = enum { First }", "A", "B", "B.First", "enum", "B"},
		{"module", coreImport, "core", `import("core:arrays")`, "14", "module", "int"},
		{"template", "", "```first```", "```second```", `"text"`, "TemplateString", "str"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := fmt.Sprintf(`%s
let value := %s
value = %s
let saved = value
let message = try {
    value = %s
    "assignment unexpectedly succeeded"
} catch TypeError err {
    err.message
}
message`, tt.setup, tt.initial, tt.same, tt.different)
			want := fmt.Sprintf(`type mismatch for binding "value": expected %s, got %s`, tt.typeName, tt.otherType)
			p := parser.New(lexer.New(input))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			env := object.NewEnvironment()
			testStringObject(t, New().Eval(program, env), want)
			value, _ := env.Get("value")
			saved, _ := env.Get("saved")
			if value != saved {
				t.Fatal("rejected assignment changed the binding's value")
			}
		})
	}
}

func TestInferredTryBindingUsesSelectedResultPerInvocation(t *testing.T) {
	result := testEval(`
type Missing = struct { message: str }
let read = fn(fail: bool) int | Missing {
    if fail { return Missing{"missing"} }
    return 32
}
let make = fn(fail: bool) call {
    let value := try { read(fail) } catch Missing err { "fallback" }
    return fn(replacement) any {
        value = replacement
        return value
    }
}
let good = make(False)
let recovered = make(True)
assert good(14) == 14
assert recovered("updated") == "updated"
let first = try { good("wrong") } catch TypeError err { err.message }
let second = try { recovered(14) } catch TypeError err { err.message }
assert first == "type mismatch for binding \"value\": expected int, got str"
assert second == "type mismatch for binding \"value\": expected str, got int"
True
`)
	testBooleanObject(t, result, true)
}

func TestInferredBindingUsesActualAnyReturnAndEvaluatesOnce(t *testing.T) {
	result := testEval(`
let calls = 0
let produce = fn() any {
    calls = calls + 1
    return 32
}
let value := produce()
assert calls == 1
try { value = "wrong" } catch TypeError err { err.message }
`)
	testStringObject(t, result, `type mismatch for binding "value": expected int, got str`)
}

func TestInferredBindingRetainsNominalIdentity(t *testing.T) {
	for name, declarations := range map[string]string{
		"struct": "type A = struct { x: int }\nlet value := A{32}\nlet saved = A{14}\ntype A = struct { x: int }\nlet other = A{14}",
		"enum":   "type A = enum { First, Second }\nlet value := A.First\nlet saved = A.Second\ntype A = enum { First, Second }\nlet other = A.First",
	} {
		t.Run(name, func(t *testing.T) {
			result := testEval(declarations + `
value = saved
try { value = other } catch TypeError err { err.message }
`)
			testStringObject(t, result, `type mismatch for binding "value": expected A, got A`)
		})
	}
}

func TestFailedInferredInitializerDoesNotInstallBinding(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprint(existing), func(t *testing.T) {
			env := object.NewEnvironment()
			if existing {
				env.Set("value", &object.Integer{Value: 7})
			}
			p := parser.New(lexer.New(`
type Missing = struct { message: str }
let fail = fn() int | Missing { return Missing{"missing"} }
let value := fail()
let reached = True
`))
			program := p.ParseProgram()
			if len(p.Errors()) != 0 {
				t.Fatal(p.Errors())
			}
			result := New().Eval(program, env)
			if err, ok := result.(*object.Error); !ok || err.MessageText() != "missing" {
				t.Fatalf("result = %v, want propagated Missing error", result)
			}
			value, present := env.Get("value")
			if present != existing {
				t.Fatalf("binding exists = %t, want %t", present, existing)
			}
			if existing {
				testIntegerObject(t, value, 7)
			}
			if _, present := env.Get("reached"); present {
				t.Fatal("evaluation continued after the error")
			}
		})
	}
}

func TestTryBindingExplicitAndDynamicContracts(t *testing.T) {
	for _, tt := range []struct {
		declaration, fail, want string
	}{
		{"let value: int =", "False", "32"},
		{"let value: int =", "True", `type mismatch for binding "value": expected int, got str`},
		{"let value =", "True", "fallback"},
	} {
		t.Run(tt.declaration+tt.fail, func(t *testing.T) {
			input := fmt.Sprintf(`
type Missing = struct { message: str }
let read = fn(fail: bool) int | Missing {
    if fail { return Missing{"missing"} }
    return 32
}
try {
    %s try { read(%s) } catch Missing err { "fallback" }
    value
} catch TypeError err { err.message }
`, tt.declaration, tt.fail)
			result := testEval(input)
			if result == nil || isError(result) || result.Inspect() != tt.want {
				t.Fatalf("result = %v, want %q", result, tt.want)
			}
		})
	}
}

func TestInferredBindingAcceptsCyclicCollection(t *testing.T) {
	testBooleanObject(t, testEval(`
let values = [0]
values[0] = values
let fixed := values
fixed = ["still an array"]
fixed[0] == "still an array"
`), true)
}
