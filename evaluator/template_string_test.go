package evaluator

import (
	"silver/object"
	"testing"
)

func templateLiteral(contents string) string {
	return "```" + contents + "```"
}

func TestTemplateStringSQLExample(t *testing.T) {
	input := `let min_age = 21
let status = "active"
let query: TemplateString = ` + templateLiteral(`
SELECT name, age, status FROM users
WHERE age >= {min_age} AND status = '{status}'
`) + `
query.eval()`

	result, ok := testEval(input).(*object.String)
	if !ok {
		t.Fatalf("result is %T, want string", result)
	}
	want := "\nSELECT name, age, status FROM users\nWHERE age >= 21 AND status = 'active'\n"
	if result.Value != want {
		t.Fatalf("result is %q, want %q", result.Value, want)
	}
}

func TestTemplateStringDoesNotEvaluateOnDeclaration(t *testing.T) {
	input := `let query = ` + templateLiteral(`{missing}`) + `
"declared"`
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "declared" {
		t.Fatalf("declaration returned %#v, want declared", result)
	}
}

func TestTemplateStringEvalUsesCurrentCapturedBindings(t *testing.T) {
	input := `let value = 1
let template = ` + templateLiteral(`value={value}`) + `
value = 2
template.eval()`
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "value=2" {
		t.Fatalf("evaluation returned %#v, want value=2", result)
	}
}

func TestTemplateStringEvalReevaluatesExpressions(t *testing.T) {
	input := `let count = 0
let next = fn() int {
    count = count + 1
    count
}
let template = ` + templateLiteral(`{next()}`) + `
let first = template.eval()
let second = template.eval()
first + second`
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "12" {
		t.Fatalf("evaluations returned %#v, want 12", result)
	}
}

func TestTemplateStringEvalUsesNormalSilverExpressions(t *testing.T) {
	input := `let maps = import("map")
let word = "silver"
let template = ` + templateLiteral(`sum={1 + 2}; value={maps.get({"answer": 42}, "answer")}; text={word}`) + `
template.eval()`
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "sum=3; value=42; text=silver" {
		t.Fatalf("evaluation returned %#v", result)
	}
}

func TestTemplateStringLiteralBraces(t *testing.T) {
	input := `let value = "inside"
let template = ` + templateLiteral(`{{literal}} {value}`) + `
template.eval()`
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "{literal} inside" {
		t.Fatalf("evaluation returned %#v, want literal braces", result)
	}
}

func TestTemplateStringMayContainAnotherTemplateExpression(t *testing.T) {
	input := "let template = ```outer {```inner```.eval()}```\ntemplate.eval()"
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "outer inner" {
		t.Fatalf("evaluation returned %#v, want nested template result", result)
	}
}

func TestTemplateStringCapturesFunctionScope(t *testing.T) {
	input := `let make = fn(value: str) TemplateString {
    ` + templateLiteral(`captured {value}`) + `
}
let template = make("scope")
template.eval()`
	result, ok := testEval(input).(*object.String)
	if !ok || result.Value != "captured scope" {
		t.Fatalf("evaluation returned %#v, want captured scope", result)
	}
}

func TestTemplateStringHasNominalTypeAndEvalSignature(t *testing.T) {
	input := coreImport + `let template: TemplateString = ` + templateLiteral(`hello`) + `
let evaluate: call() str = template.eval
core.type(template) == TemplateString`
	testBooleanObject(t, testEval(input), true)
}

func TestTemplateStringEvalReportsDelayedErrors(t *testing.T) {
	input := `let template = ` + templateLiteral(`{missing}`) + `
template.eval()`
	result, ok := testEval(input).(*object.Error)
	if !ok {
		t.Fatalf("evaluation returned %T, want NameError", result)
	}
	if result.Value.Struct.Name != "NameError" {
		t.Fatalf("error is %s, want NameError", result.Value.Struct.Name)
	}
}

func TestTemplateStringEvalRejectsArguments(t *testing.T) {
	input := `let template = ` + templateLiteral(`hello`) + `
template.eval(1)`
	result, ok := testEval(input).(*object.Error)
	if !ok || result.MessageText() != "wrong number of arguments. got=1, want=0" {
		t.Fatalf("evaluation returned %#v, want arity error", result)
	}
}
