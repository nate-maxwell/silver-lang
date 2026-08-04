package evaluator

import (
	"silver/object"
	"testing"
)

func TestArrayIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"[1, 2, 3][0]", 1},
		{"[1, 2, 3][1]", 2},
		{"[1, 2, 3][2]", 3},
		{"let i = 0\n[1][i]", 1},
		{"[1, 2, 3][1 + 1]", 3},
		{"let myArray = [1, 2, 3]\nmyArray[2]", 3},
		{"let myArray = [1, 2, 3]\nmyArray[0] + myArray[1] + myArray[2]", 6},
		{"let myArray = [1, 2, 3]\nlet i = myArray[0]\nmyArray[i]", 2},
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

func TestArrayIndexOutOfRangeProducesIndexError(t *testing.T) {
	for _, input := range []string{"[1, 2, 3][3]", "[1, 2, 3][-1]"} {
		err, ok := testEval(input).(*object.Error)
		if !ok || err.Value.Struct.Name != "IndexError" {
			t.Fatalf("%s returned %#v, want IndexError", input, err)
		}
	}
}

func TestArrayLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("object is not Array. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("array has wrong num of elements. got=%d", len(result.Elements))
	}

	testIntegerObject(t, result.Elements[0], 1)
	testIntegerObject(t, result.Elements[1], 4)
	testIntegerObject(t, result.Elements[2], 6)
}
