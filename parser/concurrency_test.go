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
let a = task(run(1))
let b = task { run(2) }
collect(a, b)
`))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}
	first := program.Statements[1].(*ast.LetStatement).Value.(*ast.TaskExpression)
	if first.Call == nil || first.Body != nil {
		t.Fatalf("parenthesized task parsed incorrectly: %#v", first)
	}
	second := program.Statements[2].(*ast.LetStatement).Value.(*ast.TaskExpression)
	if second.Body == nil || second.Call != nil {
		t.Fatalf("block task parsed incorrectly: %#v", second)
	}
	collect := program.Statements[3].(*ast.ExpressionStatement).Expression.(*ast.CollectExpression)
	if len(collect.Handles) != 2 || collect.Handles[0].Value != "a" || collect.Handles[1].Value != "b" {
		t.Fatalf("collect handles are %#v", collect.Handles)
	}
}

func TestCollectRejectsExpressions(t *testing.T) {
	p := New(lexer.New(`collect(make_task())`))
	p.ParseProgram()
	if len(p.Errors()) == 0 || !strings.Contains(p.Errors()[0], "CannotCollectExpressionError") {
		t.Fatalf("errors are %v, want CannotCollectExpressionError", p.Errors())
	}
}

func TestTaskRequiresCallOrBlock(t *testing.T) {
	p := New(lexer.New(`let a = task(1 + 2)`))
	p.ParseProgram()
	if len(p.Errors()) == 0 || !strings.Contains(p.Errors()[0], "TaskArgumentError") {
		t.Fatalf("errors are %v, want TaskArgumentError", p.Errors())
	}
}

func TestTaskHandleMayOnlyBeCollectedOnce(t *testing.T) {
	p := New(lexer.New(`
let run = fn() int { 1 }
let a = task(run())
let alias = a
collect(a)
collect(alias)
`))
	p.ParseProgram()
	if len(p.Errors()) != 1 || !strings.Contains(p.Errors()[0], "TaskAlreadyCollectedError") {
		t.Fatalf("errors are %v, want one TaskAlreadyCollectedError", p.Errors())
	}
}
