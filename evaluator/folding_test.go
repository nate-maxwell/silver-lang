package evaluator

import (
	"silver/ast"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"testing"
)

func TestFoldConstants(t *testing.T) {
	tests := []struct {
		input string
		check func(*testing.T, ast.Expression)
	}{
		{"1 + 2 * 3", expectFoldedInteger(7)},
		{"(1 + 2) * 3.0", expectFoldedFloat(9)},
		{"1 < 2 == True", expectFoldedBoolean(true)},
		{`"silver" + "lang"`, expectFoldedString("silverlang")},
		{"!!5", expectFoldedBoolean(true)},
		{"5 / 2", expectFoldedInteger(2)},
		{"2 ** 3", expectFoldedInteger(8)},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			program := parseForFolding(t, test.input)
			foldConstants(program)
			test.check(t, program.Statements[0].(*ast.ExpressionStatement).Expression)
		})
	}
}

func TestFoldConstantsThroughoutTree(t *testing.T) {
	program := parseForFolding(t, `
let calculate = fn(value) hash {
    let numbers = [1 + 2, value + (3 * 4)]
    return {1 + 1: numbers[6 / 2]}
}
calculate(10 - 5)
`)
	foldConstants(program)

	declaration := program.Statements[0].(*ast.LetStatement)
	function := declaration.Value.(*ast.FunctionLiteral)
	arrayDeclaration := function.Body.Statements[0].(*ast.LetStatement)
	array := arrayDeclaration.Value.(*ast.ArrayLiteral)
	expectFoldedInteger(3)(t, array.Elements[0])
	partial := array.Elements[1].(*ast.InfixExpression)
	expectFoldedInteger(12)(t, partial.Right)

	returnStatement := function.Body.Statements[1].(*ast.ReturnStatement)
	for key, value := range returnStatement.ReturnValue.(*ast.HashLiteral).Pairs {
		expectFoldedInteger(2)(t, key)
		expectFoldedInteger(3)(t, value.(*ast.IndexExpression).Index)
	}

	call := program.Statements[1].(*ast.ExpressionStatement).Expression.(*ast.CallExpression)
	expectFoldedInteger(5)(t, call.Arguments[0])
}

func TestFoldConstantsLeavesRuntimeErrors(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1 / 0", "division by zero"},
		{"1 + True", "type mismatch: INTEGER + BOOLEAN"},
		{`"left" - "right"`, "unknown operator: STRING - STRING"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			program := parseForFolding(t, test.input)
			foldConstants(program)
			expression := program.Statements[0].(*ast.ExpressionStatement).Expression
			if _, ok := expression.(*ast.InfixExpression); !ok {
				t.Fatalf("unsafe expression folded to %T", expression)
			}

			result := New().Eval(program, object.NewEnvironment())
			err, ok := result.(*object.Error)
			if !ok {
				t.Fatalf("result is %T, want *object.Error", result)
			}
			if err.Message != test.want {
				t.Fatalf("error is %q, want %q", err.Message, test.want)
			}
		})
	}
}

func TestFoldedLiteralRetainsExpressionPosition(t *testing.T) {
	program := parseForFolding(t, "1 + 2")
	original := program.Statements[0].(*ast.ExpressionStatement).Expression
	want := original.Position()
	foldConstants(program)
	got := program.Statements[0].(*ast.ExpressionStatement).Expression.Position()
	if got != want {
		t.Fatalf("folded position is %+v, want %+v", got, want)
	}
}

func parseForFolding(t *testing.T, input string) *ast.Program {
	t.Helper()
	p := parser.New(lexer.NewWithSource(input, "folding.slvr"))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	return program
}

func expectFoldedInteger(want int64) func(*testing.T, ast.Expression) {
	return func(t *testing.T, expression ast.Expression) {
		t.Helper()
		value, ok := expression.(*ast.IntegerLiteral)
		if !ok || value.Value != want {
			t.Fatalf("expression is %T (%v), want integer %d", expression, expression, want)
		}
	}
}

func expectFoldedFloat(want float64) func(*testing.T, ast.Expression) {
	return func(t *testing.T, expression ast.Expression) {
		t.Helper()
		value, ok := expression.(*ast.FloatLiteral)
		if !ok || value.Value != want {
			t.Fatalf("expression is %T (%v), want float %g", expression, expression, want)
		}
	}
}

func expectFoldedBoolean(want bool) func(*testing.T, ast.Expression) {
	return func(t *testing.T, expression ast.Expression) {
		t.Helper()
		value, ok := expression.(*ast.Boolean)
		if !ok || value.Value != want {
			t.Fatalf("expression is %T (%v), want boolean %t", expression, expression, want)
		}
	}
}

func expectFoldedString(want string) func(*testing.T, ast.Expression) {
	return func(t *testing.T, expression ast.Expression) {
		t.Helper()
		value, ok := expression.(*ast.StringLiteral)
		if !ok || value.Value != want {
			t.Fatalf("expression is %T (%v), want string %q", expression, expression, want)
		}
	}
}
