package stdlib_test

import (
	"io"
	"math"
	"silver/evaluator"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"testing"
)

func testEval(input string) object.Object {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	return evaluator.New().Eval(program, object.NewEnvironment())
}

func testEvalWithStreams(input string, in io.Reader, out, errOut io.Writer) object.Object {
	p := parser.New(lexer.New(input))
	program := p.ParseProgram()
	return evaluator.NewWithStreams(in, out, errOut).Eval(program, object.NewEnvironment())
}

func testNullObject(t *testing.T, value object.Object) {
	t.Helper()
	if value != evaluator.NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", value, value)
	}
}

func testBooleanObject(t *testing.T, value object.Object, want bool) {
	t.Helper()
	boolean, ok := value.(*object.Boolean)
	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", value, value)
		return
	}
	if boolean.Value != want {
		t.Errorf("object has wrong value. got=%t, want=%t", boolean.Value, want)
	}
}

func testIntegerObject(t *testing.T, value object.Object, want int64) {
	t.Helper()
	integer, ok := value.(*object.Integer)
	if !ok {
		t.Errorf("object is not Integer. got=%T (%+v)", value, value)
		return
	}
	if integer.Value != want {
		t.Errorf("object has wrong value. got=%d, want=%d", integer.Value, want)
	}
}

func testFloatObject(t *testing.T, value object.Object, want float64) bool {
	t.Helper()
	float, ok := value.(*object.Float)
	if !ok {
		t.Errorf("object is not Float. got=%T (%+v)", value, value)
		return false
	}
	if math.Abs(float.Value-want) > 1e-12 {
		t.Errorf("float has value %g, want %g", float.Value, want)
		return false
	}
	return true
}
