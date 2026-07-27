package evaluator

import (
	"math"
	"silver/object"
	"testing"
)

func TestFloatEvaluation(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"5.5", 5.5},
		{"-5.5", -5.5},
		{"1.5 + 2.25", 3.75},
		{"5.0 - 2", 3.0},
		{"2 * 1.25", 2.5},
		{"5 / 2.0", 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testFloatObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestIntegerDivisionRemainsIntegerDivision(t *testing.T) {
	testIntegerObject(t, testEval("5 / 2"), 2)
}

func TestMixedNumericComparisons(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1 == 1.0", true},
		{"1 != 1.0", false},
		{"1 < 1.5", true},
		{"2.5 > 2", true},
		{"9007199254740992 == 9007199254740992.0", true},
		{"9007199254740993 == 9007199254740992.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestFloatDivisionByZero(t *testing.T) {
	evaluated := testEval("1.0 / 0.0")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if err.Message != "division by zero" {
		t.Fatalf("error message is %q, want division by zero", err.Message)
	}
}

func TestFloatTypeMismatch(t *testing.T) {
	evaluated := testEval("1.5 + True")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if err.Message != "type mismatch: FLOAT + BOOLEAN" {
		t.Fatalf("error message is %q", err.Message)
	}
}

func TestFloatHashKeysRespectNumericEquality(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`let values = {1: "integer"}; values[1.0];`, "integer"},
		{`let values = {1.5: "float"}; values[1.5];`, "float"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		value, ok := evaluated.(*object.String)
		if !ok {
			t.Fatalf("result is %T, want *object.String", evaluated)
		}
		if value.Value != tt.want {
			t.Fatalf("result is %q, want %q", value.Value, tt.want)
		}
	}
}

func testFloatObject(t *testing.T, value object.Object, want float64) bool {
	t.Helper()
	result, ok := value.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", value, value)
		return false
	}
	if math.Abs(result.Value-want) > 1e-12 {
		t.Errorf("float has value %g, want %g", result.Value, want)
		return false
	}
	return true
}
