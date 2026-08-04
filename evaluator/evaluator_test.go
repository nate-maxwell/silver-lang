package evaluator

import (
	"bytes"
	"silver/object"
	"testing"
)

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`len("")`, 0},
		{`len("four")`, 4},
		{`len("hello world")`, 11},
		{`len([1, 2, 3])`, 3},
		{`len(1)`, "argument to `len` not supported, got INTEGER"},
		{`len("one", "two")`, "wrong number of arguments. got=2, want=1"},
		{`first([1, 2, 3])`, 1},
		{`first([])`, nil},
		{`first(1)`, "argument to `first` must be ARRAY, got INTEGER"},
		{`last([1, 2, 3])`, 3},
		{`last([])`, nil},
		{`rest([1, 2, 3])`, []int64{2, 3}},
		{`rest([1])`, []int64{}},
		{`rest([])`, nil},
		{`append([1, 2], 3)`, []int64{1, 2, 3}},
		{`append(1, 2)`, "argument to `append` must be ARRAY, got INTEGER"},
		{`append([1])`, "wrong number of arguments. got=1, want=2"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		switch expected := tt.expected.(type) {
		case int:
			testIntegerObject(t, evaluated, int64(expected))
		case []int64:
			array, ok := evaluated.(*object.Array)
			if !ok {
				t.Errorf("object is not Array. got=%T (%+v)", evaluated, evaluated)
				continue
			}
			if len(array.Elements) != len(expected) {
				t.Errorf("array has wrong length. got=%d, want=%d", len(array.Elements), len(expected))
				continue
			}
			for i, value := range expected {
				testIntegerObject(t, array.Elements[i], value)
			}
		case string:
			errObj, ok := evaluated.(*object.Error)
			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)",
					evaluated, evaluated)
				continue
			}
			if errObj.MessageText() != expected {
				t.Errorf("wrong error message. expected=%q, got=%q",
					expected, errObj.MessageText())
			}
		case nil:
			testNullObject(t, evaluated)
		}
	}
}

func TestPrintBuiltinUsesEvaluatorOutput(t *testing.T) {
	var out bytes.Buffer
	result := evalInput(t, NewWithOutput(&out), object.NewEnvironment(), `print("hello", 42)`)

	testNullObject(t, result)
	if got, want := out.String(), "hello\n42\n"; got != want {
		t.Fatalf("print output is %q, want %q", got, want)
	}
}

func TestClosures(t *testing.T) {
	input := `
		let newAdder = fn(x) call {
			fn(y) int { x + y }
		}

		let addTwo = newAdder(2)
		addTwo(2)
	`

	testIntegerObject(t, testEval(input), 4)
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let identity = fn(x) int { x }\nidentity(5)", 5},
		{"let identity = fn(x) int { return x }\nidentity(5)", 5},
		{"let double = fn(x) int { x * 2 }\ndouble(5)", 10},
		{"let add = fn(x, y) int { x + y }\nadd(5, 5)", 10},
		{"let add = fn(x, y) int { x + y }\nadd(5 + 5, add(5, 5))", 20},
		{"fn(x) int { x }(5)", 5},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestEnclosingEnvironments(t *testing.T) {
	input := `
			let first = 10
			let second = 10
			let third = 10

			let ourFunction = fn(first) int {
			let second = 20

			first + second + third
			}

			ourFunction(20) + first + second`

	testIntegerObject(t, testEval(input), 70)
}

func TestFunctionObject(t *testing.T) {
	input := "fn(x) int { x + 2 }"

	evaluated := testEval(input)
	fn, ok := evaluated.(*object.Function)
	if !ok {
		t.Fatalf("object is not Fucntion. got=%T (%+v)", evaluated, evaluated)
	}

	if len(fn.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v",
			fn.Parameters)
	}

	if fn.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", fn.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if fn.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got%q", expectedBody, fn.Body.String())
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"let a = 5\na", 5},
		{"let a = 5 * 5\na", 25},
		{"let a = 5\nlet b = a\nb", 5},
		{"let a = 5\nlet b = a\nlet c = a + b + 5\nc", 15},
	}

	for _, tt := range tests {
		testIntegerObject(t, testEval(tt.input), tt.expected)
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{
			"5 + True",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"5 + True\n5",
			"type mismatch: INTEGER + BOOLEAN",
		},
		{
			"-True",
			"unknown operator: -BOOLEAN",
		},
		{
			"True + False",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"5\nTrue + False\n5",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"if (10 > 1) { True + False }",
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return True + False
				}

				return 1
			}
			`,
			"unknown operator: BOOLEAN + BOOLEAN",
		},
		{
			"foobar",
			"identifier not found: foobar",
		},
		{
			`"hello" - "world"`,
			"unknown operator: STRING - STRING",
		},
		{
			`{"name": "Silver"}[fn(x) { x }]`,
			"unusable as hash key: FUNCTION",
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		errObj, ok := evaluated.(*object.Error)
		if !ok {
			t.Errorf("no error object returned. got=%T(%+v)",
				evaluated, evaluated)
			continue
		}

		if errObj.MessageText() != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got =%q",
				tt.expectedMessage, errObj.MessageText())
		}
	}
}

func TestReturnStatement(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10", 10},
		{"return 10\n9", 10},
		{"return 2 * 5\n9", 10},
		{"9\nreturn 2 * 5\n9", 10},
		{
			`
			if (10 > 1) {
				if (10 > 1) {
					return 10
				}
			
				return 1
			}
			`,
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testIntegerObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if True { 10 }", 10},
		{"if False { 10 }", nil},
		{"if 1 { 10 }", 10},
		{"if 1 < 2 { 10 }", 10},
		{"if 1 > 2 { 10 }", nil},
		{"if 1 > 2 { 10 } else { 20 }", 20},
		{"if 1 < 2 { 10 } else { 20 }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)
		if ok {
			testIntegerObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}
