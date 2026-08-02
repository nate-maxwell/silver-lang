package evaluator

import "testing"

func TestEvalIntegerExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"2 ** 10", 1024},
		{"2 ** 3 ** 2", 512},
		{"-2 ** 2", -4},
		{"(-2) ** 2", 4},
		{"2 * 3 ** 2", 18},
		{"0 ** 0", 1},
		{"let pow = fn(x: int, y: int) int { return x ** y }\npow(2, 8)", 256},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 // 2 * 2 + 10", 60},
		{"50 // 4", 12},
		{"50 // 4 * 2 + 1", 25},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 // 3) * 2 + -10", 50},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}
