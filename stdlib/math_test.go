package stdlib_test

import (
	"math"
	"silver/object"
	"testing"
)

const mathImport = "let math = import(\"math\")\n"

func TestAbs(t *testing.T) {
	tests := []struct {
		input string
		want  object.Object
	}{
		{input: `math.abs(-5)`, want: &object.Integer{Value: 5}},
		{input: `math.abs(0)`, want: &object.Integer{Value: 0}},
		{input: `math.abs(-1.25)`, want: &object.Float{Value: 1.25}},
		{input: `math.abs(1.25)`, want: &object.Float{Value: 1.25}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(mathImport + tt.input)
			switch want := tt.want.(type) {
			case *object.Integer:
				testIntegerObject(t, result, want.Value)
			case *object.Float:
				testFloatObject(t, result, want.Value)
			}
		})
	}
}

func TestMinAndMax(t *testing.T) {
	tests := []struct {
		input string
		want  object.Object
	}{
		{input: `math.min(2, 1)`, want: &object.Integer{Value: 1}},
		{input: `math.max(2, 1)`, want: &object.Integer{Value: 2}},
		{input: `math.min(1.5, 2)`, want: &object.Float{Value: 1.5}},
		{input: `math.max(1.5, 2)`, want: &object.Integer{Value: 2}},
		{input: `math.min(-2, -1.5)`, want: &object.Integer{Value: -2}},
		{input: `math.max(-2, -1.5)`, want: &object.Float{Value: -1.5}},
		{input: `math.min(9007199254740993, 9007199254740992.0)`, want: &object.Float{Value: 9007199254740992.0}},
		{input: `math.max(9007199254740993, 9007199254740992.0)`, want: &object.Integer{Value: 9007199254740993}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(mathImport + tt.input)
			switch want := tt.want.(type) {
			case *object.Integer:
				testIntegerObject(t, result, want.Value)
			case *object.Float:
				testFloatObject(t, result, want.Value)
			}
		})
	}
}

func TestMathIntegerFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  int64
	}{
		{input: `math.factorial(0)`, want: 1},
		{input: `math.factorial(20)`, want: 2432902008176640000},
		{input: `math.gcd([48, -18, 0])`, want: 6},
		{input: `math.gcd([(-9223372036854775807) - 1, 2])`, want: 2},
		{input: `math.gcd([])`, want: 0},
		{input: `math.lcd([4, 6, 10])`, want: 60},
		{input: `math.lcd([9223372036854775807, 2, 0])`, want: 0},
		{input: `math.lcd([])`, want: 1},
		{input: `math.lcm([9, 6])`, want: 18},
		{input: `math.isqrt(9223372036854775807)`, want: 3037000499},
		{input: `math.ceil(-1.25)`, want: -1},
		{input: `math.floor(-1.25)`, want: -2},
		{input: `math.round(1.4)`, want: 1},
		{input: `math.round(1.5)`, want: 2},
		{input: `math.round(-1.5)`, want: -2},
		{input: `math.truc(-1.75)`, want: -1},
		{input: `math.trunc(1.75)`, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testIntegerObject(t, testEval(mathImport+tt.input), tt.want)
		})
	}
}

func TestMathFloatingPointFunctions(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{input: `math.fmod(5.5, 2)`, want: 1.5},
		{input: `math.remainer(5.5, 2)`, want: -0.5},
		{input: `math.remainder(7, 2)`, want: -1},
		{input: `math.acos(1)`, want: 0},
		{input: `math.asin(1)`, want: math.Pi / 2},
		{input: `math.atan(1)`, want: math.Pi / 4},
		{input: `math.cos(0)`, want: 1},
		{input: `math.sin(0)`, want: 0},
		{input: `math.tan(0)`, want: 0},
		{input: `math.cbrt(-8)`, want: -2},
		{input: `math.exp(1)`, want: math.E},
		{input: `math.exp2(3)`, want: 8},
		{input: `math.expm1(1)`, want: math.E - 1},
		{input: `math.log(81, 3)`, want: 4},
		{input: `math.log1p(1)`, want: math.Log(2)},
		{input: `math.log2(8)`, want: 3},
		{input: `math.log10(1000)`, want: 3},
		{input: `math.sqrt(9)`, want: 3},
		{input: `math.degrees(math.pi)`, want: 180},
		{input: `math.radians(180)`, want: math.Pi},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := testEval(mathImport + tt.input)
			value, ok := result.(*object.Float)
			if !ok {
				t.Fatalf("result is %T (%v), want *object.Float", result, result)
			}
			if math.Abs(value.Value-tt.want) > 1e-12 {
				t.Fatalf("result is %.16g, want %.16g", value.Value, tt.want)
			}
		})
	}
}

func TestMathModf(t *testing.T) {
	result := testEval(mathImport + `math.modf(-3.25)`)
	array, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("result is %T (%v), want *object.Array", result, result)
	}
	if len(array.Elements) != 2 {
		t.Fatalf("result has %d elements, want 2", len(array.Elements))
	}
	testFloatObject(t, array.Elements[0], -0.25)
	testFloatObject(t, array.Elements[1], -3)
}

func TestMathConstants(t *testing.T) {
	tests := []struct {
		name string
		want float64
	}{
		{name: "pi", want: math.Pi},
		{name: "e", want: math.E},
		{name: "tau", want: 2 * math.Pi},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFloatObject(t, testEval(mathImport+"math."+tt.name), tt.want)
		})
	}

	nan, ok := testEval(mathImport + `math.nan`).(*object.Float)
	if !ok || !math.IsNaN(nan.Value) {
		t.Fatalf("math.nan is %T (%v), want NaN float", nan, nan)
	}
}

func TestMathFunctionErrors(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `math.factorial(-1)`, message: "argument to `factorial` must be nonnegative"},
		{input: `math.abs("one")`, message: "argument to `abs` must be INTEGER or FLOAT, got STRING"},
		{input: `math.abs((-9223372036854775807) - 1)`, message: "absolute value out of range for INTEGER"},
		{input: `math.min("one", 2)`, message: "argument 1 to `min` must be INTEGER or FLOAT, got STRING"},
		{input: `math.max(1, False)`, message: "argument 2 to `max` must be INTEGER or FLOAT, got BOOLEAN"},
		{input: `math.factorial(21)`, message: "result of `factorial` is out of range for INTEGER"},
		{input: `math.factorial(1.0)`, message: "argument to `factorial` must be INTEGER, got FLOAT"},
		{input: `math.gcd(12)`, message: "argument to `gcd` must be ARRAY, got INTEGER"},
		{input: `math.gcd([12, 3.0])`, message: "element 2 of argument to `gcd` must be INTEGER, got FLOAT"},
		{input: `math.lcd([9223372036854775807, 2])`, message: "result of `lcd` is out of range for INTEGER"},
		{input: `math.isqrt(-1)`, message: "argument to `isqrt` must be nonnegative"},
		{input: `math.ceil(math.nan)`, message: "result of `ceil` is out of range for INTEGER"},
		{input: `math.round(1)`, message: "argument 1 to `round` must be FLOAT, got INTEGER"},
		{input: `math.round(math.nan)`, message: "result of `round` is out of range for INTEGER"},
		{input: `math.sin("zero")`, message: "argument 1 to `sin` must be INTEGER or FLOAT, got STRING"},
		{input: `math.log(8, False)`, message: "argument 2 to `log` must be INTEGER or FLOAT, got BOOLEAN"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, ok := testEval(mathImport + tt.input).(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if result.MessageText() != tt.message {
				t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
			}
		})
	}
}
