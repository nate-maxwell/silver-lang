package evaluator

import (
	"silver/object"
	"testing"
)

func TestStringConcatenation(t *testing.T) {
	input := `"hello" + " " + "world"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "hello world" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"hello world"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "hello world" {
		t.Errorf("String has wrong value. got=%q", str.Value)
	}
}

func TestStringLiteralEscapeValues(t *testing.T) {
	evaluated := testEval(`"line one\nline two\t\u263A"`)
	str, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}
	if got, want := str.Value, "line one\nline two\t☺"; got != want {
		t.Fatalf("string is %q, want %q", got, want)
	}
}
