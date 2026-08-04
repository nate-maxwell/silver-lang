package evaluator

import (
	"silver/object"
	"testing"
)

func TestAbsBuiltin(t *testing.T) {
	tests := []struct {
		input string
		want  object.Object
	}{
		{input: `abs(-5)`, want: &object.Integer{Value: 5}},
		{input: `abs(0)`, want: &object.Integer{Value: 0}},
		{input: `abs(-1.25)`, want: &object.Float{Value: 1.25}},
		{input: `abs(1.25)`, want: &object.Float{Value: 1.25}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			switch want := tt.want.(type) {
			case *object.Integer:
				testIntegerObject(t, result, want.Value)
			case *object.Float:
				testFloatObject(t, result, want.Value)
			}
		})
	}
}

func TestMinAndMaxBuiltins(t *testing.T) {
	tests := []struct {
		input string
		want  object.Object
	}{
		{input: `min(2, 1)`, want: &object.Integer{Value: 1}},
		{input: `max(2, 1)`, want: &object.Integer{Value: 2}},
		{input: `min(1.5, 2)`, want: &object.Float{Value: 1.5}},
		{input: `max(1.5, 2)`, want: &object.Integer{Value: 2}},
		{input: `min(-2, -1.5)`, want: &object.Integer{Value: -2}},
		{input: `max(-2, -1.5)`, want: &object.Float{Value: -1.5}},
		{input: `min(9007199254740993, 9007199254740992.0)`, want: &object.Float{Value: 9007199254740992.0}},
		{input: `max(9007199254740993, 9007199254740992.0)`, want: &object.Integer{Value: 9007199254740993}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			switch want := tt.want.(type) {
			case *object.Integer:
				testIntegerObject(t, result, want.Value)
			case *object.Float:
				testFloatObject(t, result, want.Value)
			}
		})
	}
}

func TestRangeBuiltin(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: `range(1, 5)`, want: `[1, 2, 3, 4]`},
		{input: `range(-2, 2)`, want: `[-2, -1, 0, 1]`},
		{input: `range(3, 3)`, want: `[]`},
		{input: `range(3, 1)`, want: `[]`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(tt.input)
			if got := result.Inspect(); got != tt.want {
				t.Fatalf("result is %q, want %q", got, tt.want)
			}
		})
	}
}

func TestLenBuiltinSupportsMaps(t *testing.T) {
	testIntegerObject(t, testEval(`len({"one": 1, "two": 2})`), 2)
	testIntegerObject(t, testEval(`len({})`), 0)
}

func TestNewCoreBuiltinErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `abs("one")`, message: "argument to `abs` must be INTEGER or FLOAT, got STRING"},
		{input: `abs((-9223372036854775807) - 1)`, message: "absolute value out of range for INTEGER"},
		{input: `min("one", 2)`, message: "argument 1 to `min` must be INTEGER or FLOAT, got STRING"},
		{input: `max(1, False)`, message: "argument 2 to `max` must be INTEGER or FLOAT, got BOOLEAN"},
		{input: `range(1.0, 2)`, message: "argument 1 to `range` must be INTEGER, got FLOAT"},
		{input: `range(1, 2.0)`, message: "argument 2 to `range` must be INTEGER, got FLOAT"},
		{input: `range(1)`, message: "wrong number of arguments. got=1, want=2"},
		{input: `range(0, 1000001)`, message: "range contains too many elements: 1000001 (maximum 1000000)"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}
}
