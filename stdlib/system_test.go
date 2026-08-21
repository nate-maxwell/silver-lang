package stdlib_test

import (
	"os"
	"runtime"
	"silver/internal/version"
	"silver/object"
	"strings"
	"testing"
)

const systemImport = `let system = import("system")
`

func TestSystemVersion(t *testing.T) {
	components := []struct {
		name string
		want int64
	}{
		{name: "MAJOR", want: version.Major},
		{name: "MINOR", want: version.Minor},
		{name: "PATCH", want: version.Patch},
	}
	for _, component := range components {
		value, ok := testEval(systemImport + "system." + component.name).(*object.Integer)
		if !ok || value.Value != component.want {
			t.Fatalf("system.%s is %T (%v), want %d", component.name, value, value, component.want)
		}
	}

	value, ok := testEval(systemImport + `system.VERSION`).(*object.String)
	if !ok || value.Value != version.String() {
		t.Fatalf("system.VERSION is %T (%v), want %q", value, value, version.String())
	}
}

func TestSystemHostInformation(t *testing.T) {
	tests := []string{
		`system.machine()`,
		`system.node()`,
		`system.processor()`,
		`system.release()`,
		`system.system()`,
	}
	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			if _, ok := testEval(systemImport + input).(*object.String); !ok {
				t.Fatalf("%s did not return a string", input)
			}
		})
	}

	machine := testEval(systemImport + `system.machine()`).(*object.String)
	if machine.Value == "" {
		t.Fatal("machine is empty on a supported Go architecture")
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}
	node := testEval(systemImport + `system.node()`).(*object.String)
	if node.Value != hostname {
		t.Fatalf("node is %q, want %q", node.Value, hostname)
	}

	wantSystems := map[string]string{
		"aix": "AIX", "android": "Android", "darwin": "Darwin", "dragonfly": "DragonFlyBSD",
		"freebsd": "FreeBSD", "illumos": "SunOS", "ios": "iOS", "js": "JavaScript",
		"linux": "Linux", "netbsd": "NetBSD", "openbsd": "OpenBSD", "plan9": "Plan9",
		"solaris": "SunOS", "wasip1": "WASI", "windows": "Windows",
	}
	system := testEval(systemImport + `system.system()`).(*object.String)
	if system.Value != wantSystems[runtime.GOOS] {
		t.Fatalf("system is %q, want %q", system.Value, wantSystems[runtime.GOOS])
	}
}

func TestSystemEnvironmentFunctions(t *testing.T) {
	const key = "SILVER_STDLIB_SYSTEM_TEST"
	t.Setenv(key, "before")

	testNullObject(t, testEval(systemImport+`system.setenv("`+key+`", "after")`))
	if value := os.Getenv(key); value != "after" {
		t.Fatalf("process environment is %q, want after", value)
	}

	value, ok := testEval(systemImport + `system.getenv("` + key + `")`).(*object.String)
	if !ok || value.Value != "after" {
		t.Fatalf("getenv returned %T (%v), want after", value, value)
	}

	input := `let maps = import("map")
maps.get(system.environment(), "` + key + `")`
	value, ok = testEval(systemImport + input).(*object.String)
	if !ok || value.Value != "after" {
		t.Fatalf("environment returned %T (%v), want after", value, value)
	}

	missing := testEval(systemImport + `system.getenv("SILVER_STDLIB_DEFINITELY_MISSING")`).(*object.String)
	if missing.Value != "" {
		t.Fatalf("missing environment variable is %q, want empty", missing.Value)
	}
}

func TestSystemSilverPathHelpers(t *testing.T) {
	const silverPathName = "SILVER_PATH"

	name, ok := testEval(systemImport + `system.ENV_SILVER_PATH`).(*object.String)
	if !ok || name.Value != silverPathName {
		t.Fatalf("ENV_SILVER_PATH is %T (%v), want SILVER_PATH", name, name)
	}

	separator, ok := testEval(systemImport + `system.get_path_sep()`).(*object.String)
	if !ok || separator.Value != string(os.PathListSeparator) {
		t.Fatalf("get_path_sep returned %T (%v), want %q", separator, separator, os.PathListSeparator)
	}

	t.Setenv(silverPathName, "")
	testNullObject(t, testEval(systemImport+`system.append_path("first")`))
	if value := os.Getenv(silverPathName); value != "first" {
		t.Fatalf("append_path on an empty value produced %q, want first", value)
	}

	separatorText := string(os.PathListSeparator)
	testNullObject(t, testEval(systemImport+`system.append_path("second")`))
	if value, want := os.Getenv(silverPathName), "first"+separatorText+"second"; value != want {
		t.Fatalf("append_path produced %q, want %q", value, want)
	}
}

func TestSystemFunctionErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `system.machine(1)`, message: "wrong number of arguments. got=1, want=0"},
		{input: `system.get_path_sep(1)`, message: "wrong number of arguments. got=1, want=0"},
		{input: `system.append_path(1)`, message: "argument 1 to `append_path` must be STRING, got INTEGER"},
		{input: `system.getenv(1)`, message: "argument 1 to `getenv` must be STRING, got INTEGER"},
		{input: `system.setenv("key", 1)`, message: "argument 2 to `setenv` must be STRING, got INTEGER"},
	}
	for _, tt := range tests {
		result, ok := testEval(systemImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if result.MessageText() != tt.message {
			t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
		}
	}

	result, ok := testEval(systemImport + `system.setenv("invalid=key", "value")`).(*object.Error)
	if !ok || !strings.Contains(result.MessageText(), "could not set environment variable") {
		t.Fatalf("invalid key returned %T (%v), want setenv error", result, result)
	}
}
