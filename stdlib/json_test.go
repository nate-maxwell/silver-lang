package stdlib_test

import (
	"os"
	"path/filepath"
	"silver/object"
	"testing"
)

const jsonImport = "let json = import(\"json\")\n"

func TestJSONLoadsBuildsNestedSilverValues(t *testing.T) {
	input := `let maps = import("map")
let value = json.loads("{\"project\":{\"name\":\"Silver\",\"active\":true},\"versions\":[1,2.5,null]}")
let project = maps.get(value, "project")
let versions = maps.get(value, "versions")
if maps.get(project, "name") == "Silver" {
    if maps.get(project, "active") {
        if versions[0] == 1 {
            if versions[1] == 2.5 { versions[2] } else { False }
        } else { False }
    } else { False }
} else {
    False
}`
	result := testEval(jsonImport + input)
	if err, ok := result.(*object.Error); ok {
		t.Fatal(err.Inspect())
	}
	testNullObject(t, result)
}

func TestJSONLoadsAcceptsEveryJSONRootType(t *testing.T) {
	tests := []struct {
		input string
		want  object.Object
	}{
		{input: `json.loads("null")`, want: &object.Null{}},
		{input: `json.loads("true")`, want: &object.Boolean{Value: true}},
		{input: `json.loads("42")`, want: &object.Integer{Value: 42}},
		{input: `json.loads("2.5")`, want: &object.Float{Value: 2.5}},
		{input: `json.loads("\"text\"")`, want: &object.String{Value: "text"}},
	}
	for _, tt := range tests {
		result := testEval(jsonImport + tt.input)
		if result.Inspect() != tt.want.Inspect() || result.Type() != tt.want.Type() {
			t.Errorf("%s returned %s (%s), want %s (%s)", tt.input, result.Inspect(), result.Type(), tt.want.Inspect(), tt.want.Type())
		}
	}
}

func TestJSONDecodeErrorHasPythonFields(t *testing.T) {
	input := `try {
    json.loads("{\n  \"name\": }")
} catch json.JSONDecodeError err {
    if err.doc == "{\n  \"name\": }" {
        if err.lineno == 2 {
            if err.colno == 11 {
                if err.pos == 12 { err.msg } else { "wrong position" }
            } else { "wrong column" }
        } else { "wrong line" }
    } else { "wrong document" }
}`
	evaluated := testEval(jsonImport + input)
	if err, ok := evaluated.(*object.Error); ok {
		t.Fatal(err.Inspect())
	}
	result, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("result is %T, want *object.String", result)
	}
	if got, want := result.Value, "invalid character '}' looking for beginning of value"; got != want {
		t.Fatalf("message is %q, want %q", got, want)
	}
}

func TestJSONLoadsRejectsEmptyAndTrailingData(t *testing.T) {
	tests := []struct {
		document string
		message  string
		position int64
	}{
		{document: "", message: "Expecting value", position: 0},
		{document: "   ", message: "Expecting value", position: 3},
		{document: "{\"x\"", message: "Unexpected end of JSON input", position: 4},
		{document: "{} []", message: "Extra data", position: 3},
	}
	for _, tt := range tests {
		input := `try {
    json.loads(` + silverString(tt.document) + `)
} catch json.JSONDecodeError err {
    if err.msg == ` + silverString(tt.message) + ` { err.pos } else { -1 }
}`
		testIntegerObject(t, testEval(jsonImport+input), tt.position)
	}
}

func TestJSONDumpsRoundTripsNestedValues(t *testing.T) {
	input := `let maps = import("map")
let source = json.loads("{\"nested\":{\"values\":[1,true,null]}}")
let encoded = json.dumps(source)
let decoded = json.loads(encoded)
maps.get(maps.get(decoded, "nested"), "values")[1]`
	result := testEval(jsonImport + input)
	if err, ok := result.(*object.Error); ok {
		t.Fatal(err.Inspect())
	}
	testBooleanObject(t, result, true)
}

func TestJSONDumpsSupportsIntegerAndStringIndent(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `json.dumps({"value": [1, 2]}, 2)`, want: "{\n  \"value\": [\n    1,\n    2\n  ]\n}"},
		{input: `json.dumps({"value": 1}, "--")`, want: "{\n--\"value\": 1\n}"},
		{input: `json.dumps({"value": 1}, 0)`, want: "{\n\"value\": 1\n}"},
	}
	for _, tt := range tests {
		result, ok := testEval(jsonImport + tt.input).(*object.String)
		if !ok {
			t.Fatalf("%s returned %T, want *object.String", tt.input, result)
		}
		if result.Value != tt.want {
			t.Errorf("%s returned %q, want %q", tt.input, result.Value, tt.want)
		}
	}
}

func TestJSONLoadAndDumpUseFiles(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")
	if err := os.WriteFile(path, []byte(`{"old":true}`), 0600); err != nil {
		t.Fatal(err)
	}

	input := `let io = import("io")
let maps = import("map")
let file = io.open(` + silverString(path) + `)
json.dump({"nested": {"answer": 42}}, file, 2)
let value = json.load(file)
file.close()
maps.get(maps.get(value, "nested"), "answer")`
	testIntegerObject(t, testEval(jsonImport+input), 42)
}

func TestJSONRejectsUnsupportedEncodeValues(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `json.dumps({1: "value"})`, message: "JSON map keys must be STRING, got INTEGER"},
		{input: `json.dumps(json)`, message: "object of type MODULE is not JSON serializable"},
	}
	for _, tt := range tests {
		result, ok := testEval(jsonImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if got := result.MessageText(); got != tt.message {
			t.Errorf("%s error is %q, want %q", tt.input, got, tt.message)
		}
	}
}

func TestJSONFunctionsRejectInvalidArguments(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `json.loads()`, message: "wrong number of arguments. got=0, want=1"},
		{input: `json.loads(1)`, message: "argument to `loads` must be STRING, got INTEGER"},
		{input: `json.dumps({}, True)`, message: "argument 2 to `dumps` must be INTEGER or STRING, got BOOLEAN"},
		{input: `json.load(1)`, message: "argument to `load` must be file-like, got INTEGER"},
		{input: `json.dump({}, 1)`, message: "argument to `dump` must be file-like, got INTEGER"},
	}
	for _, tt := range tests {
		result, ok := testEval(jsonImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if got := result.MessageText(); got != tt.message {
			t.Errorf("%s error is %q, want %q", tt.input, got, tt.message)
		}
	}
}
