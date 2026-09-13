package evaluator

import (
	"bytes"
	"silver/object"
	"testing"
)

func TestDeferredCallsRunLIFOOnReturn(t *testing.T) {
	var out bytes.Buffer
	input := `let io = import("io")
let record = fn(value: str) { io.println(value) }
let run = fn() int {
    defer record("first")
    defer record("second")
    return 7
}
run()`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	testIntegerObject(t, result, 7)
	if got, want := out.String(), "second\nfirst\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestDeferCapturesArgumentsAndCallableImmediately(t *testing.T) {
	var out bytes.Buffer
	input := `let io = import("io")
let value = "before"
let action = fn(text: str) { io.println("old " + text) }
defer action(value)
value = "after"
action = fn(text: str) { io.println("new " + text) }`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	testNullObject(t, result)
	if got, want := out.String(), "old before\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestDeferredCallsRunWhileUnwindingError(t *testing.T) {
	var out bytes.Buffer
	input := `let io = import("io")
let run = fn() {
    defer io.println("cleanup")
    missing_name
}
run()`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	if _, ok := result.(*object.Error); !ok {
		t.Fatalf("result is %T, want *object.Error", result)
	}
	if got, want := out.String(), "cleanup\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestAllDeferredCallsRunAndLaterFailureWins(t *testing.T) {
	var out bytes.Buffer
	input := `let io = import("io")
let fail_first = fn() {
    io.println("first")
    missing_first
}
let fail_second = fn() {
    io.println("second")
    missing_second
}
let run = fn() {
    defer fail_first()
    defer fail_second()
    missing_body
}
run()`

	result, ok := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input).(*object.Error)
	if !ok {
		t.Fatalf("result is not *object.Error")
	}
	if got, want := result.MessageText(), "identifier not found: missing_first"; got != want {
		t.Fatalf("error is %q, want %q", got, want)
	}
	if got, want := out.String(), "second\nfirst\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestDeferredCallsUseFunctionScope(t *testing.T) {
	var out bytes.Buffer
	input := `let io = import("io")
let run = fn() {
    if True { defer io.println("cleanup") }
    io.println("body")
}
run()`

	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
	testNullObject(t, result)
	if got, want := out.String(), "body\ncleanup\n"; got != want {
		t.Fatalf("output is %q, want %q", got, want)
	}
}

func TestForLoopDefersUseEnclosingScope(t *testing.T) {
	for _, test := range []struct {
		name, loop, want string
	}{
		{
			name: "array closures",
			loop: `for value in [1, 2] {
    let record = fn() { io.println(value) }
    defer record()
    value = value + 10
}`,
			want: "body\n12\n11\nbefore\n",
		},
		{
			name: "map",
			loop: `let entries = {"key": "value"}
for key, value in entries {
    defer io.println(key)
    defer io.println(value)
}`,
			want: "body\nvalue\nkey\nbefore\n",
		},
		{
			name: "nested loops",
			loop: `for outer in [1, 2] {
    defer io.println(outer)
    for inner in [3, 4] {
        defer io.println(inner)
    }
}`,
			want: "body\n4\n3\n2\n4\n3\n1\nbefore\n",
		},
	} {
		for _, scope := range []string{"script", "function"} {
			t.Run(test.name+"/"+scope, func(t *testing.T) {
				var out bytes.Buffer
				body := "defer io.println(\"before\")\n" + test.loop + "\nio.println(\"body\")"
				if scope == "function" {
					body = "let run = fn() {\n" + body + "\n}\nrun()"
				}
				input := "let io = import(\"io\")\n" + body
				result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
				testNullObject(t, result)
				if got := out.String(); got != test.want {
					t.Fatalf("output is %q, want %q", got, test.want)
				}
			})
		}
	}
}

func TestForLoopDefersSurviveControlFlow(t *testing.T) {
	for _, test := range []struct {
		control, want string
	}{
		{"break", "body\n1\n"},
		{"continue", "body\n2\n1\n"},
		{"return", "1\n"},
		{"missing_name", "1\n"},
	} {
		t.Run(test.control, func(t *testing.T) {
			var out bytes.Buffer
			input := `let io = import("io")
let run = fn() {
    for value in [1, 2] {
        defer io.println(value)
        ` + test.control + `
    }
    io.println("body")
}
run()`
			result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), input)
			if test.control == "missing_name" {
				err, ok := result.(*object.Error)
				if !ok || err.MessageText() != "identifier not found: missing_name" {
					t.Fatalf("result is %v, want missing_name error", result)
				}
			} else {
				testNullObject(t, result)
			}
			if got := out.String(); got != test.want {
				t.Fatalf("output is %q, want %q", got, test.want)
			}
		})
	}
}
