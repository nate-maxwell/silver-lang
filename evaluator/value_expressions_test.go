package evaluator

import (
	"silver/object"
	"strings"
	"testing"
)

func TestSuccessfulBlocksWithoutValuesProduceNull(t *testing.T) {
	expressions := []string{
		`if True {}`,
		`if False { 1 } else {}`,
		`if False { 1 }`,
		"if True { 1\nlet local = 2 }",
		`if True { type Local = int }`,
		`if True { type Local = struct {} }`,
		`if True { type Local = enum { One } }`,
		`switch 1 { case 1: }`,
		`switch 1 { default: }`,
		`try {} catch RuntimeError err { 1 }`,
		`try { 1 / 0 } catch ZeroDivisionError err {}`,
	}
	for _, expression := range expressions {
		t.Run(expression, func(t *testing.T) {
			testNullObject(t, evalInput(t, New(), object.NewEnvironment(), expression))
			for _, annotation := range []string{"", ": null", ": any"} {
				input := "let value" + annotation + " = " + expression + "\nimport(\"core\").type(value) == null && !value"
				testBooleanObject(t, evalInput(t, New(), object.NewEnvironment(), input), true)
			}
		})
	}
}

func TestValueExpressionRejectsEscapingControlFlow(t *testing.T) {
	// Each position consumes a value, even when the enclosing construct is a
	// statement. The loop and function make all three control statements legal
	// to parse, so these checks exercise runtime evaluation.
	positions := []struct {
		name  string
		input string
	}{
		{"binding", `let value = VALUE`},
		{"typed binding", `let value: int = VALUE`},
		{"assignment", "let value = 0\nvalue = VALUE"},
		{"member assignment", "type Box = struct { value: any }\nlet box = Box{0}\nbox.value = VALUE"},
		{"member assignment target", `(VALUE).value = 1`},
		{"index assignment", "let values = [0]\nvalues[0] = VALUE"},
		{"index assignment target", `(VALUE)[0] = 1`},
		{"index assignment key", "let values = [0]\nvalues[VALUE] = 1"},
		{"return operand", `return VALUE`},
		{"assert condition", `assert VALUE`},
		{"assert message", `assert False, VALUE`},
		{"call argument", `import("core").type(VALUE)`},
		{"call target", `(VALUE)()`},
		{"defer argument", `defer import("core").type(VALUE)`},
		{"defer target", `defer (VALUE)()`},
		{"import path", `import(VALUE)`},
		{"member receiver", `(VALUE).value`},
		{"prefix operand", `!(VALUE)`},
		{"left operand", `(VALUE) + 1`},
		{"right operand", `1 + (VALUE)`},
		{"logical left operand", `(VALUE) && True`},
		{"logical right operand", `True && (VALUE)`},
		{"struct field", "type Box = struct { value: any }\nBox{VALUE}"},
		{"struct type", `(VALUE){1}`},
		{"array element", `[VALUE]`},
		{"index receiver", `(VALUE)[0]`},
		{"index operand", `[1][VALUE]`},
		{"map key", `{(VALUE): 1}`},
		{"map value", `{"key": VALUE}`},
		{"if condition", `if (VALUE) { 1 }`},
		{"switch subject", `switch (VALUE) { case 1: 1 }`},
		{"switch case", `switch 1 { case VALUE: 1 }`},
		{"for iterable", `for item in (VALUE) {}`},
		{"while condition", `while (VALUE) { break }`},
	}
	for _, control := range []string{"return", "break", "continue"} {
		for _, position := range positions {
			t.Run(control+"/"+position.name, func(t *testing.T) {
				statement := control
				if control == "return" {
					statement += " 7"
				}
				body := strings.ReplaceAll(position.input, "VALUE", "if True { "+statement+" }")
				input := "let f = fn() int {\nfor outer in [0] {\n" + body + "\n}\nreturn 9\n}\nf()"
				failure := assertErrorStruct(t, evalInput(t, New(), object.NewEnvironment(), input), "RuntimeError")
				if want := control + " cannot escape a value expression"; failure.MessageText() != want {
					t.Fatalf("error is %q, want %q", failure.MessageText(), want)
				}
				if len(failure.Frames) == 0 {
					t.Fatal("control-flow error has no source location")
				}
			})
		}
	}
}

func TestValueExpressionRejectsControlFlowFromSelectedBlock(t *testing.T) {
	for _, expression := range []string{
		`if True { return 7 }`,
		`if False {} else { return 7 }`,
		`switch 1 { case 1: return 7 }`,
		`switch 1 { default: return 7 }`,
		`try { return 7 } catch RuntimeError err { 0 }`,
		`try { 1 / 0 } catch ZeroDivisionError err { return 7 }`,
		`if True { if True { return 7 } }`,
		`if True { for item in [1] { return 7 } }`,
	} {
		t.Run(expression, func(t *testing.T) {
			input := "let f = fn() int {\nlet x = " + expression + "\nreturn 9\n}\nf()"
			assertErrorStruct(t, evalInput(t, New(), object.NewEnvironment(), input), "RuntimeError")
		})
	}
}

func TestRejectedValueExpressionDoesNotModifyBinding(t *testing.T) {
	for _, statement := range []string{"let value =", "value ="} {
		t.Run(statement, func(t *testing.T) {
			input := "let value = 3\nlet f = fn() int {\n" + statement + ` if True { return 7 }
return 9
}
let caught = try { f() } catch RuntimeError err { err.message }
value == 3 && caught == "return cannot escape a value expression"`
			testBooleanObject(t, evalInput(t, New(), object.NewEnvironment(), input), true)
		})
	}
	env := object.NewEnvironment()
	assertErrorStruct(t, evalInput(t, New(), env, `let value = if True { return 7 }`), "RuntimeError")
	if _, exists := env.Get("value"); exists {
		t.Fatal("failed initializer created a binding")
	}
}

func TestValueExpressionPreservesLocalControlFlowAndErrors(t *testing.T) {
	for _, input := range []string{
		`let value = if True { fn() int { return 7 }() }
value == 7`,
		`let value = if True {
    let count = 0
    for item in [1, 2, 3] {
        if item == 1 { continue }
        if item == 3 { break }
        count = count + item
    }
    count
}
value == 2`,
		`let value = if False { return 7 } else { 9 }
value == 9 && (False && if True { return 7 }) == False`,
		`try { let value = if True { 1 / 0 } } catch ZeroDivisionError err { True }`,
		`type Missing = struct { message: str }
let fail = fn() int | Missing { return Missing{"missing"} }
try { let value = if True { fail() } } catch Missing err { err.message == "missing" }`,
		`let f = fn() int {
    switch 1 { case 1: return 7 }
    return 9
}
f() == 7`,
		`let f = fn() int {
    try { return 7 } catch RuntimeError err { return 0 }
    return 9
}
f() == 7`,
	} {
		t.Run(input, func(t *testing.T) {
			testBooleanObject(t, evalInput(t, New(), object.NewEnvironment(), input), true)
		})
	}
}

func TestTemplateInterpolationUsesValueSemantics(t *testing.T) {
	input := "let template = " + templateLiteral(`{if True {}}`) + "\ntemplate.eval()"
	testStringObject(t, evalInput(t, New(), object.NewEnvironment(), input), "null")
	input = "let template = " + templateLiteral(`{if True { return 7 }}`) + "\ntemplate.eval()"
	assertErrorStruct(t, evalInput(t, New(), object.NewEnvironment(), input), "RuntimeError")
}
