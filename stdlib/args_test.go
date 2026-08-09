package stdlib_test

import (
	"silver/object"
	"strings"
	"testing"
)

func TestArgsParsesOptionsFlagsCountsAndPositionals(t *testing.T) {
	input := `let args = import("args")
let maps = import("map")
let parser = args.new("build", "Build a project")
parser.add(args.positional("source", "Source file"))
parser.add(args.optional_positional("output", "Output file", "app.exe"))
parser.add(args.required_option("target", "t", "Target platform"))
parser.add(args.option_with_default("mode", "m", "Build mode", "debug"))
parser.add(args.option("color", "", "Color output"))
parser.add(args.flag("force", "f", "Overwrite output"))
parser.add(args.count("verbose", "v", "Increase verbosity"))

let parsed = parser.parse([
    "main.slv", "--target", "linux-amd64", "-m", "release",
    "-f", "-v", "-v"
])
assert parsed.help_requested == False
assert parsed.values["source"] == "main.slv"
assert parsed.values["output"] == "app.exe"
assert parsed.values["target"] == "linux-amd64"
assert parsed.values["mode"] == "release"
assert parsed.values["force"] == True
assert parsed.values["verbose"] == 2
assert maps.contains(parsed.values, "color") == False
True`
	testBooleanObject(t, testEval(input), true)
}

func TestArgsSupportsOptionTerminatorAndHelpRequest(t *testing.T) {
	input := `let args = import("args")
let parser = args.new("show", "")
parser.add(args.positional("value", "Value to show"))
let parsed = parser.parse(["--", "--literal"])
assert parsed.values["value"] == "--literal"

let required = args.new("required", "")
required.add(args.required_option("target", "t", "Target"))
let help = required.parse(["--help"])
assert help.help_requested
True`
	testBooleanObject(t, testEval(input), true)
}

func TestArgsReturnsTypedParseErrors(t *testing.T) {
	for _, test := range []struct {
		name    string
		source  string
		message string
	}{
		{
			name: "missing required option",
			source: `let args = import("args")
let parser = args.new("build", "")
parser.add(args.required_option("target", "t", "Target"))
parser.parse([])`,
			message: "missing required option: --target",
		},
		{
			name: "unknown option",
			source: `let args = import("args")
args.new("build", "").parse(["--unknown"])`,
			message: "unknown option: --unknown",
		},
		{
			name: "missing option value",
			source: `let args = import("args")
let parser = args.new("build", "")
parser.add(args.option("output", "o", "Output"))
parser.parse(["-o"])`,
			message: "option requires a value: -o",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			failure, ok := testEval(test.source).(*object.Error)
			if !ok {
				t.Fatalf("result is not an error")
			}
			if failure.Value == nil || failure.Value.Struct.Name != "ArgumentError" {
				t.Fatalf("error is %v, want ArgumentError", failure)
			}
			if got := failure.MessageText(); got != test.message {
				t.Fatalf("message is %q, want %q", got, test.message)
			}
		})
	}
}

func TestArgsValidatesDefinitionsAndBuildsHelp(t *testing.T) {
	failure, ok := testEval(`let args = import("args")
let parser = args.new("build", "")
parser.add(args.flag("help", "x", "Conflict"))`).(*object.Error)
	if !ok || failure.MessageText() != "-h and --help are reserved for parser help" {
		t.Fatalf("result is %v, want reserved help ArgumentError", failure)
	}

	result := testEval(`let args = import("args")
let parser = args.new("build", "Build a Silver project.")
parser.add(args.positional("source", "Source file"))
parser.add(args.option_with_default("mode", "m", "Build mode", "debug"))
parser.help()`)
	help, ok := result.(*object.String)
	if !ok {
		t.Fatalf("result is %T, want string", result)
	}
	for _, text := range []string{
		"Usage: build <source> [options]",
		"Build a Silver project.",
		"Arguments:",
		"source\tSource file",
		"-h, --help",
		"-m, --mode <value>\tBuild mode (default: debug)",
	} {
		if !strings.Contains(help.Value, text) {
			t.Fatalf("help does not contain %q:\n%s", text, help.Value)
		}
	}
}
