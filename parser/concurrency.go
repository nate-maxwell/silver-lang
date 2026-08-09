package parser

import (
	"fmt"
	"silver/ast"
)

type taskUsage struct {
	collected bool
}

// validateTaskUsage performs the small affine check required by task handles.
// It runs after syntax parsing but before the program can be evaluated.
func (p *Parser) validateTaskUsage(program *ast.Program) {
	bindings := make(map[string]*taskUsage)
	p.validateTaskStatements(program.Statements, bindings)
}

func (p *Parser) validateTaskStatements(statements []ast.Statement, bindings map[string]*taskUsage) {
	for _, statement := range statements {
		if statement == nil {
			continue
		}
		switch node := statement.(type) {
		case *ast.ExpressionStatement:
			p.validateTaskExpression(node.Expression, bindings)
		case *ast.LetStatement:
			p.validateTaskExpression(node.Value, bindings)
			p.bindTaskUsage(node.Name.Value, node.Value, bindings)
		case *ast.AssignmentStatement:
			p.validateTaskExpression(node.Value, bindings)
			p.bindTaskUsage(node.Name.Value, node.Value, bindings)
		case *ast.MemberAssignmentStatement:
			p.validateTaskExpression(node.Target, bindings)
			p.validateTaskExpression(node.Value, bindings)
		case *ast.IndexAssignmentStatement:
			p.validateTaskExpression(node.Target, bindings)
			p.validateTaskExpression(node.Value, bindings)
		case *ast.ReturnStatement:
			p.validateTaskExpression(node.ReturnValue, bindings)
		case *ast.AssertStatement:
			p.validateTaskExpression(node.Condition, bindings)
			p.validateTaskExpression(node.Message, bindings)
		case *ast.DeferStatement:
			p.validateTaskExpression(node.Call, bindings)
		case *ast.ForStatement:
			p.validateTaskExpression(node.Iterable, bindings)
			before := cloneTaskBindings(bindings)
			body := cloneTaskBindings(bindings)
			delete(body, node.Key.Value)
			if node.Value != nil {
				delete(body, node.Value.Value)
			}
			p.validateTaskStatements(node.Body.Statements, body)
			mergeCollectedTaskBindings(bindings, before, body)
		case *ast.WhileStatement:
			p.validateTaskExpression(node.Condition, bindings)
			before := cloneTaskBindings(bindings)
			body := cloneTaskBindings(bindings)
			p.validateTaskStatements(node.Body.Statements, body)
			mergeCollectedTaskBindings(bindings, before, body)
		}
	}
}

func (p *Parser) bindTaskUsage(name string, value ast.Expression, bindings map[string]*taskUsage) {
	switch value := value.(type) {
	case *ast.TaskExpression:
		bindings[name] = &taskUsage{}
	case *ast.Identifier:
		if handle, ok := bindings[value.Value]; ok {
			bindings[name] = handle
		} else {
			delete(bindings, name)
		}
	default:
		delete(bindings, name)
	}
}

func (p *Parser) validateTaskExpression(expression ast.Expression, bindings map[string]*taskUsage) {
	if expression == nil {
		return
	}
	switch node := expression.(type) {
	case *ast.TaskExpression:
		p.validateTaskExpression(node.Work, cloneTaskBindings(bindings))
	case *ast.CollectExpression:
		for _, identifier := range node.Handles {
			handle, ok := bindings[identifier.Value]
			if !ok {
				continue
			}
			if handle.collected {
				p.addError(identifier.Position(), fmt.Sprintf("TaskAlreadyCollectedError: task handle %q may only be collected once", identifier.Value))
				continue
			}
			handle.collected = true
		}
	case *ast.PrefixExpression:
		p.validateTaskExpression(node.Right, bindings)
	case *ast.InfixExpression:
		p.validateTaskExpression(node.Left, bindings)
		p.validateTaskExpression(node.Right, bindings)
	case *ast.MemberExpression:
		p.validateTaskExpression(node.Object, bindings)
	case *ast.IfExpression:
		p.validateTaskExpression(node.Condition, bindings)
		consequence := cloneTaskBindings(bindings)
		alternative := cloneTaskBindings(bindings)
		p.validateTaskStatements(node.Consequence.Statements, consequence)
		if node.Alternative != nil {
			p.validateTaskStatements(node.Alternative.Statements, alternative)
		}
		mergeCollectedTaskBindings(bindings, consequence, alternative)
	case *ast.TryExpression:
		// The try body runs in the surrounding scope. Catch bindings are local,
		// but collecting an outer task in any handler must still count as a
		// possible consumption for the affine task check.
		p.validateTaskStatements(node.Body.Statements, bindings)
		for _, clause := range node.Catches {
			handler := cloneTaskBindings(bindings)
			delete(handler, clause.Binding.Value)
			p.validateTaskStatements(clause.Body.Statements, handler)
			for name, usage := range bindings {
				if handled, ok := handler[name]; ok && handled.collected {
					usage.collected = true
				}
			}
		}
	case *ast.FunctionLiteral:
		functionBindings := cloneTaskBindings(bindings)
		for _, parameter := range node.Parameters {
			delete(functionBindings, parameter.Value)
		}
		p.validateTaskStatements(node.Body.Statements, functionBindings)
	case *ast.TemplateStringLiteral:
		templateBindings := cloneTaskBindings(bindings)
		for _, part := range node.Parts {
			p.validateTaskExpression(part.Expression, templateBindings)
		}
	case *ast.CallExpression:
		p.validateTaskExpression(node.Function, bindings)
		for _, argument := range node.Arguments {
			p.validateTaskExpression(argument, bindings)
		}
	case *ast.StructLiteral:
		p.validateTaskExpression(node.StructType, bindings)
		for _, value := range node.Values {
			p.validateTaskExpression(value, bindings)
		}
	case *ast.ArrayLiteral:
		for _, element := range node.Elements {
			p.validateTaskExpression(element, bindings)
		}
	case *ast.IndexExpression:
		p.validateTaskExpression(node.Left, bindings)
		p.validateTaskExpression(node.Index, bindings)
	case *ast.MapLiteral:
		for key, value := range node.Pairs {
			p.validateTaskExpression(key, bindings)
			p.validateTaskExpression(value, bindings)
		}
	}
}

func cloneTaskBindings(source map[string]*taskUsage) map[string]*taskUsage {
	clone := make(map[string]*taskUsage, len(source))
	byUsage := make(map[*taskUsage]*taskUsage)
	for name, usage := range source {
		copy, ok := byUsage[usage]
		if !ok {
			copy = &taskUsage{collected: usage.collected}
			byUsage[usage] = copy
		}
		clone[name] = copy
	}
	return clone
}

func mergeCollectedTaskBindings(target, left, right map[string]*taskUsage) {
	for name, usage := range left {
		if _, ok := target[name]; !ok {
			target[name] = &taskUsage{collected: usage.collected}
		}
	}
	for name, usage := range right {
		if existing, ok := target[name]; !ok {
			target[name] = &taskUsage{collected: usage.collected}
		} else if usage.collected {
			existing.collected = true
		}
	}
	for name, usage := range target {
		leftUsage, leftOK := left[name]
		rightUsage, rightOK := right[name]
		if leftOK && leftUsage.collected || rightOK && rightUsage.collected {
			usage.collected = true
		}
	}
}
