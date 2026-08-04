package evaluator

import (
	"math"
	"path/filepath"
	"silver/ast"
	"silver/object"
	"silver/token"
	"testing"
)

func TestScalarConstantsArePooled(t *testing.T) {
	engine := New()
	environment := object.NewEnvironment()

	tests := []struct {
		name  string
		input string
	}{
		{"integer", "42"},
		{"float", "4.25"},
		{"string", `"silver"`},
		{"true", "True"},
		{"false", "False"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			first := evalInput(t, engine, environment, test.input)
			second := evalInput(t, engine, environment, test.input)
			if first != second {
				t.Fatalf("equal %s constants were allocated separately", test.name)
			}
		})
	}
}

func TestConstantPoolsAreEvaluatorLocal(t *testing.T) {
	first := evalInput(t, New(), object.NewEnvironment(), `"silver"`)
	second := evalInput(t, New(), object.NewEnvironment(), `"silver"`)
	if first == second {
		t.Fatal("separate evaluators unexpectedly share their string pool")
	}
}

func TestFloatPoolUsesExactBits(t *testing.T) {
	engine := New()
	positiveZero := &ast.FloatLiteral{
		Token: token.Token{Type: token.FLOAT, Literal: "0.0"},
		Value: 0,
	}
	negativeZero := &ast.FloatLiteral{
		Token: token.Token{Type: token.FLOAT, Literal: "-0.0"},
		Value: math.Copysign(0, -1),
	}

	positive := engine.Eval(positiveZero, object.NewEnvironment())
	negative := engine.Eval(negativeZero, object.NewEnvironment())
	if positive == negative {
		t.Fatal("positive and negative zero used the same pooled object")
	}
	if repeated := engine.Eval(negativeZero, object.NewEnvironment()); repeated != negative {
		t.Fatal("identical negative-zero constants were not pooled")
	}
}

func TestFoldedCachedConstantsArePooled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	writeSilverFile(t, path, "40 + 2")
	engine := New()

	first := engine.EvalFile(path, object.NewEnvironment())
	second := engine.EvalFile(path, object.NewEnvironment())
	if first != second {
		t.Fatal("folded constant loaded from the AST cache was not pooled")
	}
	assertInteger(t, first, 42)
}

func TestStringEqualityIsIndependentOfPooling(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`"same" == "same"`, true},
		{`"left" != "right"`, true},
		{`("sil" + "ver") == "silver"`, true},
		{`"left" == "right"`, false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			testBooleanObject(t, testEval(test.input), test.want)
		})
	}
}
