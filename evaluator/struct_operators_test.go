package evaluator

import (
	"silver/object"
	"strings"
	"testing"
)

const operatorStructPrelude = `
struct Number {
    value: int
    add: call(self: Number, other: int) int
    sub: call(self: Number, other: int) int
	mul: call(self: Number, other: int) int
	div: call(self: Number, other: int) float
	mod: call(self: Number, other: int) int
	int_div: call(self: Number, other: int) int
    pow: call(self: Number, other: int) int
    eq: call(self: Number, other: int) bool
    not_eq: call(self: Number, other: int) bool
    lt: call(self: Number, other: int) bool
	gt: call(self: Number, other: int) bool
	lte: call(self: Number, other: int) bool
	gte: call(self: Number, other: int) bool
}
let add = fn(self: Number, other: int) int { self.value + other }
let sub = fn(self: Number, other: int) int { self.value - other }
let mul = fn(self: Number, other: int) int { self.value * other }
let div = fn(self: Number, other: int) float { self.value / other }
let mod = fn(self: Number, other: int) int { self.value % other }
let int_div = fn(self: Number, other: int) int { self.value // other }
let pow = fn(self: Number, other: int) int { self.value ** other }
let eq = fn(self: Number, other: int) bool { self.value == other }
let not_eq = fn(self: Number, other: int) bool { self.value != other }
let lt = fn(self: Number, other: int) bool { self.value < other }
let gt = fn(self: Number, other: int) bool { self.value > other }
let lte = fn(self: Number, other: int) bool { self.value <= other }
let gte = fn(self: Number, other: int) bool { self.value >= other }
let number = Number{10, add, sub, mul, div, mod, int_div, pow, eq, not_eq, lt, gt, lte, gte}
`

func TestStructArithmeticOperatorMethods(t *testing.T) {
	tests := []struct {
		expression string
		want       int64
	}{
		{"number + 2", 12},
		{"number - 2", 8},
		{"number * 2", 20},
		{"number % 3", 1},
		{"number // 3", 3},
		{"number ** 2", 100},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			testIntegerObject(t, testEval(operatorStructPrelude+test.expression), test.want)
		})
	}
}

func TestStructDivisionMethodMayReturnFloat(t *testing.T) {
	testFloatObject(t, testEval(operatorStructPrelude+"number / 4"), 2.5)
}

func TestStructComparisonOperatorMethods(t *testing.T) {
	tests := []struct {
		expression string
		want       bool
	}{
		{"number == 10", true},
		{"number != 10", false},
		{"number < 11", true},
		{"number > 9", true},
		{"number <= 10", true},
		{"number <= 9", false},
		{"number >= 10", true},
		{"number >= 11", false},
	}
	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			testBooleanObject(t, testEval(operatorStructPrelude+test.expression), test.want)
		})
	}
}

func TestStructOperatorMayReturnStruct(t *testing.T) {
	evaluated := testEval(`
struct Vector {
    x: int
    y: int
    add: call(self: Vector, other: Vector) Vector
}
let vector_add = fn(self: Vector, other: Vector) Vector {
    Vector{self.x + other.x, self.y + other.y, vector_add}
}
let left = Vector{2, 3, vector_add}
let right = Vector{5, 7, vector_add}
let sum = left + right
sum.x * 10 + sum.y
`)
	testIntegerObject(t, evaluated, 80)
}

func TestStructOperatorRequiresMappedMethod(t *testing.T) {
	evaluated := testEval(`
struct Vector { x: int }
Vector{1} + Vector{2}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	for _, part := range []string{`operator "+"`, `struct "Vector"`, `method "add"`} {
		if !strings.Contains(err.MessageText(), part) {
			t.Fatalf("error %q does not contain %q", err.MessageText(), part)
		}
	}
}

func TestStructOperatorMethodMustBeCallable(t *testing.T) {
	evaluated := testEval(`
struct Invalid { add: int }
Invalid{1} + 2
`)
	err, ok := evaluated.(*object.Error)
	if !ok || !strings.Contains(err.MessageText(), `operator method "add" on struct "Invalid" is not callable`) {
		t.Fatalf("result is %#v, want non-callable operator error", evaluated)
	}
}

func TestStructTruthinessDoesNotLookUpOperatorMethod(t *testing.T) {
	evaluated := testEval(`
struct Empty {}
if (Empty{}) { 1 } else { 0 }
`)
	testIntegerObject(t, evaluated, 1)
}

func TestStructOperatorRegistryCoversEveryBinaryOperator(t *testing.T) {
	want := map[string]string{
		"+": "add", "-": "sub", "*": "mul", "/": "div", "%": "mod", "//": "int_div", "**": "pow",
		"==": "eq", "!=": "not_eq", "<": "lt", ">": "gt",
		"<=": "lte", ">=": "gte",
	}
	if len(structInfixOperatorMethods) != len(want) {
		t.Fatalf("registry has %d entries, want %d", len(structInfixOperatorMethods), len(want))
	}
	for operator, method := range want {
		if structInfixOperatorMethods[operator] != method {
			t.Fatalf("operator %q maps to %q, want %q", operator, structInfixOperatorMethods[operator], method)
		}
	}
}
