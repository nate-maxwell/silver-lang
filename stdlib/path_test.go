package stdlib_test

import (
	"os"
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

const pathImport = `let paths = import("path")
`

func TestPathConstructorAndNominalType(t *testing.T) {
	input := `let value: paths.Path = paths.new("projects/silver/main.slv")
core.type(value) == paths.Path`
	testBooleanObject(t, testEval(pathImport+coreImport+input), true)

	path := testEval(pathImport + `paths.new("projects/silver/main.slv")`).(*object.StructInstance)
	if path.Struct.Name != "Path" {
		t.Fatalf("constructed struct is %q, want Path", path.Struct.Name)
	}
	if _, exists := path.Values["joinpath"]; !exists {
		t.Fatal("Path does not expose joinpath")
	}
	for _, field := range path.Struct.Fields {
		if _, exists := path.Values[field]; !exists {
			t.Fatalf("Path value is missing declared field %q", field)
		}
	}
}

func TestPathPropertiesAndTransformations(t *testing.T) {
	path := filepath.Join("projects", "silver", "archive.tar.gz")
	parent := filepath.Join("projects", "silver")
	tests := []struct {
		expression string
		want       string
	}{
		{expression: `value.path`, want: path},
		{expression: `value.name`, want: "archive.tar.gz"},
		{expression: `value.parent().path`, want: parent},
		{expression: `value.parent().name`, want: "silver"},
		{expression: `value.stem`, want: "archive.tar"},
		{expression: `value.suffix`, want: ".gz"},
		{expression: `value.joinpath("release", "main.slv").path`, want: filepath.Join(path, "release", "main.slv")},
		{expression: `value.with_name("release.zip").path`, want: filepath.Join(parent, "release.zip")},
		{expression: `value.with_stem("release").path`, want: filepath.Join(parent, "release.gz")},
		{expression: `value.with_suffix(".zip").path`, want: filepath.Join(parent, "archive.tar.zip")},
		{expression: `value.as_posix()`, want: "projects/silver/archive.tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			input := `let value = paths.new(` + silverString(path) + `)
` + tt.expression
			result, ok := testEval(pathImport + input).(*object.String)
			if !ok {
				t.Fatalf("result is %T (%v), want *object.String", result, result)
			}
			if result.Value != tt.want {
				t.Fatalf("result is %q, want %q", result.Value, tt.want)
			}
		})
	}

	input := `let value = paths.new("archive.tar.gz")
value.suffixes`
	if got := testEval(pathImport + input).Inspect(); got != `[.tar, .gz]` {
		t.Fatalf("suffixes are %q, want %q", got, `[.tar, .gz]`)
	}
	input = `let value = paths.new("projects/silver/main.slv")
value.parts`
	if got := testEval(pathImport + input).Inspect(); got != `[projects, silver, main.slv]` {
		t.Fatalf("parts are %q, want %q", got, `[projects, silver, main.slv]`)
	}
	input = `let value = paths.new("src/lib/main.slv")
value.is_relative_to(paths.new("src"))`
	testBooleanObject(t, testEval(pathImport+input), true)
	input = `let value = paths.new("src/main.slv")
value.match("*.slv")`
	testBooleanObject(t, testEval(pathImport+input), true)
}

func TestPathMethodsReturnPathValues(t *testing.T) {
	input := `let value = paths.new("projects/silver/main.slv")
let derived = value.parent().joinpath("library.slv")
core.type(derived) == paths.Path`
	testBooleanObject(t, testEval(pathImport+coreImport+input), true)

	input = `let value = paths.new("projects/silver/main.slv")
core.type(value.parents()[0]) == paths.Path`
	testBooleanObject(t, testEval(pathImport+coreImport+input), true)

	for _, factory := range []string{`paths.cwd()`, `paths.home()`} {
		input = `core.type(` + factory + `) == paths.Path`
		testBooleanObject(t, testEval(pathImport+coreImport+input), true)
	}
}

func TestPathFilesystemMethods(t *testing.T) {
	directory := t.TempDir()
	nested := filepath.Join(directory, "nested")
	file := filepath.Join(nested, "message.txt")
	renamed := filepath.Join(nested, "renamed.txt")

	testNullObject(t, testEval(pathImport+`paths.new(`+silverString(nested)+`).mkdir()`))
	testIntegerObject(t, testEval(pathImport+`paths.new(`+silverString(file)+`).write_text("hello")`), 5)

	contents, ok := testEval(pathImport + `paths.new(` + silverString(file) + `).read_text()`).(*object.String)
	if !ok || contents.Value != "hello" {
		t.Fatalf("read_text result is %T (%v), want hello", contents, contents)
	}
	testBooleanObject(t, testEval(pathImport+`paths.new(`+silverString(file)+`).exists()`), true)
	testBooleanObject(t, testEval(pathImport+`paths.new(`+silverString(file)+`).is_file()`), true)
	testBooleanObject(t, testEval(pathImport+`paths.new(`+silverString(nested)+`).is_dir()`), true)

	input := `let file = paths.new(` + silverString(file) + `)
file.samefile(file)`
	testBooleanObject(t, testEval(pathImport+input), true)

	input = `let directory = paths.new(` + silverString(nested) + `)
directory.iterdir()[0].path`
	child, ok := testEval(pathImport + input).(*object.String)
	if !ok || child.Value != filepath.Clean(file) {
		t.Fatalf("iterdir child is %T (%v), want %q", child, child, file)
	}
	input = `let directory = paths.new(` + silverString(nested) + `)
core.type(directory.glob("*.txt")[0]) == paths.Path`
	testBooleanObject(t, testEval(pathImport+coreImport+input), true)
	input = `let directory = paths.new(` + silverString(directory) + `)
directory.rglob("*.txt")[0].path`
	match, ok := testEval(pathImport + input).(*object.String)
	if !ok || match.Value != filepath.Clean(file) {
		t.Fatalf("rglob match is %T (%v), want %q", match, match, file)
	}

	input = `let maps = import("map")
let file = paths.new(` + silverString(file) + `)
maps.get(file.stat(), "size")`
	testIntegerObject(t, testEval(pathImport+input), 5)

	input = `let file = paths.new(` + silverString(file) + `)
file.rename(paths.new(` + silverString(renamed) + `)).path`
	result, ok := testEval(pathImport + input).(*object.String)
	if !ok || result.Value != filepath.Clean(renamed) {
		t.Fatalf("rename result is %T (%v), want %q", result, result, renamed)
	}
	testNullObject(t, testEval(pathImport+`paths.new(`+silverString(renamed)+`).unlink()`))
	testNullObject(t, testEval(pathImport+`paths.new(`+silverString(nested)+`).rmdir()`))
}

func TestPathErrorsAndNoChildPathTypes(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `paths.new(1)`, message: "argument 1 to `new` must be STRING, got INTEGER"},
		{input: `paths.new("one").joinpath()`, message: "wrong number of arguments. got=1, want at least=2"},
		{input: `paths.new("main.slv").with_suffix("txt")`, message: `invalid suffix "txt"`},
		{input: `paths.new("one").is_relative_to(2)`, message: "argument 2 to `is_relative_to` must be STRING, got INTEGER"},
	}
	for _, tt := range tests {
		result, ok := testEval(pathImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if result.MessageText() != tt.message {
			t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
		}
	}

	for _, child := range []string{"PosixPath", "WindowsPath", "PurePath"} {
		result, ok := testEval(pathImport + "paths." + child).(*object.Error)
		if !ok || !strings.Contains(result.MessageText(), `has no exported member "`+child+`"`) {
			t.Fatalf("paths.%s returned %T (%v), want missing-member error", child, result, result)
		}
	}
}

func TestPathFactoriesUseCurrentEnvironment(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	for input, want := range map[string]string{
		`paths.cwd().path`:  filepath.Clean(cwd),
		`paths.home().path`: filepath.Clean(home),
	} {
		result, ok := testEval(pathImport + input).(*object.String)
		if !ok || result.Value != want {
			t.Fatalf("%s returned %T (%v), want %q", input, result, result, want)
		}
	}
}
