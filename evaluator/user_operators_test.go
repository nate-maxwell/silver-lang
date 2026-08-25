package evaluator

import (
	"os"
	"path/filepath"
	"silver/object"
	"strings"
	"testing"
)

func TestUserOperatorCanHostBuiltinFunction(t *testing.T) {
	evaluated := testEval(`
let arrays = import("array")
operator ;; fn(left: array, right) array {
    arrays.append(left, right)
}
[1, 2] ;; 3
`)
	array, ok := evaluated.(*object.Array)
	if !ok {
		t.Fatalf("evaluated is %T (%s), want *object.Array", evaluated, evaluated.Inspect())
	}
	if len(array.Elements) != 3 {
		t.Fatalf("array contains %d elements, want 3", len(array.Elements))
	}
	testIntegerObject(t, array.Elements[2], 3)
}

func TestUserOperatorCanPipeValuesIntoFunctions(t *testing.T) {
	evaluated := testEval(`
operator |> fn(left, right: call) {
    return right(left)
}
let first_function = fn() int { 5 }
let second_function = fn(value: int) int { value * 2 }
let third_function = fn(value: int) int { value + 1 }
first_function() |> second_function |> third_function
`)
	testIntegerObject(t, evaluated, 11)
}

func TestDoubleColonCanBeUsedAsUserOperator(t *testing.T) {
	evaluated := testEval(`
operator :: fn(left: int, right: int) int { left + right }
20 :: 22
`)
	testIntegerObject(t, evaluated, 42)
}

func TestUserOperatorBindingPowerAffectsEvaluation(t *testing.T) {
	evaluated := testEval(`
operator @ 55 fn(left: int, right: int) int { left * 10 + right }
1 + 2 @ 3 * 4
`)
	testIntegerObject(t, evaluated, 42)
}

func TestOperatorDefinedInFunctionIsGloballyAvailableAfterward(t *testing.T) {
	evaluated := testEval(`
let install = fn() {
    operator @ fn(left: int, right: int) int { left + right }
}
install()
20 @ 22
`)
	testIntegerObject(t, evaluated, 42)
}

func TestOperatorIsAvailableToSourcesParsedLater(t *testing.T) {
	directory := t.TempDir()
	operatorPath := filepath.Join(directory, "operators.slv")
	consumerPath := filepath.Join(directory, "consumer.slv")
	if err := os.WriteFile(operatorPath, []byte(`operator %% 55 fn(left: int, right: int) int { left + right }`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(consumerPath, []byte("20 %% 22\n"), 0600); err != nil {
		t.Fatal(err)
	}

	engine := New()
	if result := engine.EvalFile(operatorPath, object.NewEnvironment()); isError(result) {
		t.Fatalf("operator definition failed: %s", result.Inspect())
	}
	evaluated := engine.EvalFile(consumerPath, object.NewEnvironment())
	testIntegerObject(t, evaluated, 42)
}

func TestGlobalOperatorCannotBeRedefinedByLaterSource(t *testing.T) {
	directory := t.TempDir()
	for _, name := range []string{"first.slv", "second.slv"} {
		path := filepath.Join(directory, name)
		if err := os.WriteFile(path, []byte(`operator @ fn(left, right) { return left }`), 0600); err != nil {
			t.Fatal(err)
		}
	}
	engine := New()
	if result := engine.EvalFile(filepath.Join(directory, "first.slv"), object.NewEnvironment()); isError(result) {
		t.Fatalf("first definition failed: %s", result.Inspect())
	}
	evaluated := engine.EvalFile(filepath.Join(directory, "second.slv"), object.NewEnvironment())
	failure, ok := evaluated.(*object.Error)
	if !ok || !strings.Contains(failure.MessageText(), `operator "@" is already defined`) {
		t.Fatalf("evaluated is %#v, want duplicate-operator error", evaluated)
	}
}
