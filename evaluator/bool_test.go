package evaluator

import "testing"

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"True", true},
		{"False", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 <= 1", true},
		{"2 <= 1", false},
		{"1 >= 1", true},
		{"1 >= 2", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"True == True", true},
		{"False == False", true},
		{"True == False", false},
		{"True != False", true},
		{"False != True", true},
		{"(1 < 2) == True", true},
		{"(1 < 2) == False", false},
		{"(1 > 2) == True", false},
		{"(1 > 2) == False", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"True && True", true},
		{"True && False", false},
		{"False && True", false},
		{"False || False", false},
		{"False || True", true},
		{"True || False", true},
		{"True || False && False", true},
		{"False || True && False", false},
		{"42 && True", true},
		{"False || \"value\"", true},
		{"struct Value {}\nValue{} && True", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			testBooleanObject(t, testEval(test.input), test.expected)
		})
	}
}

func TestLogicalOperatorsShortCircuit(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "and",
			input: `let called = False
let mark = fn() bool {
    called = True
    return True
}
False && mark()
called`,
		},
		{
			name: "or",
			input: `let called = False
let mark = fn() bool {
    called = True
    return False
}
True || mark()
called`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testBooleanObject(t, testEval(test.input), false)
		})
	}

	testBooleanObject(t, testEval("False && missing"), false)
	testBooleanObject(t, testEval("True || missing"), true)
}
