package parser

import (
	"silver/ast"
	"silver/lexer"
	"strings"
	"testing"
)

func TestTaskAndCollectExpressions(t *testing.T) {
	p := New(lexer.New(`
let run = fn(value: int) int { value }
let task_a = fn() int { run(1) }
let task_b = fn() int { run(2) }
let a = task task_a
let b = task task_b
collect a, b
`))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	first := program.Statements[3].(*ast.LetStatement).Value.(*ast.TaskExpression)
	if target, ok := first.Work.(*ast.Identifier); !ok || target.Value != "task_a" {
		t.Fatalf("first task target parsed incorrectly: %#v", first)
	}
	second := program.Statements[4].(*ast.LetStatement).Value.(*ast.TaskExpression)
	if target, ok := second.Work.(*ast.Identifier); !ok || target.Value != "task_b" {
		t.Fatalf("second task target parsed incorrectly: %#v", second)
	}
	collect := program.Statements[5].(*ast.ExpressionStatement).Expression.(*ast.CollectExpression)
	if len(collect.Handles) != 2 || collect.Handles[0].Value != "a" || collect.Handles[1].Value != "b" {
		t.Fatalf("collect handles are %#v", collect.Handles)
	}
}

func TestCollectRejectsExpressions(t *testing.T) {
	p := New(lexer.New(`collect make_task()`))
	p.ParseProgram()
	if len(p.Errors()) == 0 || !strings.Contains(p.Errors()[0], "CannotCollectExpressionError") {
		t.Fatalf("errors are %v, want CannotCollectExpressionError", p.Errors())
	}
}

func TestTaskRejectsCalls(t *testing.T) {
	p := New(lexer.New(`let a = task run()`))
	p.ParseProgram()
	if len(p.Errors()) == 0 || !strings.Contains(p.Errors()[0], "TaskTargetError") {
		t.Fatalf("errors are %v, want TaskTargetError", p.Errors())
	}
}

func TestTaskAcceptsZeroArgumentAnonymousFunction(t *testing.T) {
	p := New(lexer.New(`let handle = task fn() int { 42 }`))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	task := program.Statements[0].(*ast.LetStatement).Value.(*ast.TaskExpression)
	function, ok := task.Work.(*ast.FunctionLiteral)
	if !ok || len(function.Parameters) != 0 {
		t.Fatalf("task target is %#v, want zero-argument function literal", task.Work)
	}
}

func TestTaskRejectsParameterizedAnonymousFunction(t *testing.T) {
	p := New(lexer.New(`let handle = task fn(value: int) int { value }`))
	p.ParseProgram()
	if len(p.Errors()) == 0 || !strings.Contains(p.Errors()[0], "must have no parameters") {
		t.Fatalf("errors are %v, want anonymous-task arity error", p.Errors())
	}
}

func TestTaskAndCollectRejectParenthesizedSyntax(t *testing.T) {
	for _, input := range []string{`let a = task(run())`, `collect(a)`} {
		p := New(lexer.New(input))
		p.ParseProgram()
		if len(p.Errors()) == 0 {
			t.Fatalf("parser accepted obsolete syntax %q", input)
		}
	}
}

func TestTaskHandleMayOnlyBeCollectedOnce(t *testing.T) {
	p := New(lexer.New(`
let run = fn() int { 1 }
let a = task run
let alias = a
collect a
collect alias
`))
	p.ParseProgram()
	if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], "TaskAlreadyCollectedError") {
		t.Fatalf("errors are %v, want one TaskAlreadyCollectedError", p.Errors())
	}
}
