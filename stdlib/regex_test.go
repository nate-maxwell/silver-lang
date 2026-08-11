package stdlib_test

import (
	"silver/object"
	"strings"
	"testing"
)

const regexImport = `let regex = import("regex")
`

func TestRegexMatchSearchAndFullMatch(t *testing.T) {
	result := testEval(regexImport + `regex.match("[a-z]+", "silver 2")`)
	match, ok := result.(*object.StructInstance)
	if !ok || match.Struct.Name != "MatchObject" {
		t.Fatalf("match result is %T (%v), want MatchObject", result, result)
	}

	testNullObject(t, testEval(regexImport+`regex.match("[0-9]+", "silver 2")`))
	testIntegerObject(t, testEval(regexImport+`regex.search("[0-9]+", "silver 22").start()`), 7)

	// fullmatch must allow a later alternative to consume the whole input.
	full := testEval(regexImport + `regex.fullmatch("a|ab", "ab")`)
	if _, ok := full.(*object.StructInstance); !ok {
		t.Fatalf("fullmatch result is %T (%v), want MatchObject", full, full)
	}
	testNullObject(t, testEval(regexImport+`regex.fullmatch("[a-z]+", "silver 2")`))
}

func TestRegexMatchObject(t *testing.T) {
	input := `let match = regex.search("(?P<label>[a-z]+)=([0-9]+)", "xx id=42 yy")
`
	tests := []struct {
		expression string
		want       string
	}{
		{expression: `match.group()`, want: "id=42"},
		{expression: `match.group(1)`, want: "id"},
		{expression: `match.group("label")`, want: "id"},
		{expression: `match.groups()`, want: "[id, 42]"},
		{expression: `match.start()`, want: "3"},
		{expression: `match.end()`, want: "8"},
		{expression: `match.span()`, want: "[3, 8]"},
		{expression: `match.string`, want: "xx id=42 yy"},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			if got := testEval(regexImport + input + tt.expression).Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}

	groupMap := `let maps = import("map")
` + input + `maps.get(match.groupmap(), "label")`
	if got := testEval(regexImport + groupMap).Inspect(); got != "id" {
		t.Fatalf("named group is %q, want id", got)
	}

	optional := `regex.match("(a)?b", "b").group(1)`
	testNullObject(t, testEval(regexImport+optional))
}

func TestRegexCollectionAndReplacementFunctions(t *testing.T) {
	tests := []struct {
		expression string
		want       string
	}{
		{expression: `regex.findall("[0-9]+", "a12 b345")`, want: "[12, 345]"},
		{expression: `regex.findlist("[0-9]+", "a12 b345")[1].group()`, want: "345"},
		{expression: `regex.sub("[0-9]+", "#", "a12 b345")`, want: "a# b#"},
		{expression: `regex.subn("[0-9]+", "#", "a12 b345")`, want: "[a# b#, 2]"},
		{expression: `regex.split("[ ,]+", "one, two three")`, want: "[one, two, three]"},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			if got := testEval(regexImport + tt.expression).Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}

	escaped, ok := testEval(regexImport + `regex.escape("a+b?")`).(*object.String)
	if !ok || escaped.Value != `a\+b\?` {
		t.Fatalf("escape result is %T (%v), want %q", escaped, escaped, `a\+b\?`)
	}
}

func TestCompiledRegexMethods(t *testing.T) {
	prefix := `let expression = regex.compile("[0-9]+")
`
	tests := []struct {
		expression string
		want       string
	}{
		{expression: `expression.match("12 silver").group()`, want: "12"},
		{expression: `expression.search("silver 12").group()`, want: "12"},
		{expression: `expression.findall("1 22")`, want: "[1, 22]"},
		{expression: `expression.findlist("1 22")[1].span()`, want: "[2, 4]"},
		{expression: `expression.sub("#", "1 22")`, want: "# #"},
		{expression: `expression.subn("#", "1 22")`, want: "[# #, 2]"},
		{expression: `expression.split("a1b22c")`, want: "[a, b, c]"},
		{expression: `expression.fullmatch("123") .group()`, want: "123"},
		{expression: `expression.escape("a+b")`, want: `a\+b`},
	}
	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			if got := testEval(regexImport + prefix + tt.expression).Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}

	compiled := testEval(regexImport + `regex.compile("[a-z]+")`)
	expression, ok := compiled.(*object.StructInstance)
	if !ok || expression.Struct.Name != "Expression" {
		t.Fatalf("compile result is %T (%v), want Expression", compiled, compiled)
	}
	for _, field := range expression.Struct.Fields {
		if _, exists := expression.Values[field]; !exists {
			t.Errorf("Expression is missing method %q", field)
		}
	}
}

func TestRegexErrors(t *testing.T) {
	invalid, ok := testEval(regexImport + `regex.compile("[")`).(*object.Error)
	if !ok || !strings.Contains(invalid.MessageText(), "invalid pattern passed to `compile`") {
		t.Fatalf("invalid pattern result is %T (%v)", invalid, invalid)
	}

	wrongType, ok := testEval(regexImport + `regex.search("x", 1)`).(*object.Error)
	if !ok || wrongType.MessageText() != "argument 2 to `search` must be STRING, got INTEGER" {
		t.Fatalf("wrong type result is %T (%v)", wrongType, wrongType)
	}
}
