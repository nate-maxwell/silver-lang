package evaluator

import (
	"silver/ast"
	"silver/object"
	"silver/token"
	"strconv"
)

// foldConstants replaces pure literal expressions with their evaluated value.
// It mutates program in place and returns it for convenient use in the parse
// pipeline. Expressions that would produce an error remain in the tree so the
// evaluator can report them at runtime with their original source location.
func foldConstants(program *ast.Program) *ast.Program {
	if program == nil {
		return nil
	}
	for index, statement := range program.Statements {
		program.Statements[index] = foldStatementConstants(statement)
	}
	return program
}

func foldStatementConstants(statement ast.Statement) ast.Statement {
	switch node := statement.(type) {
	case *ast.BlockStatement:
		for index, child := range node.Statements {
			node.Statements[index] = foldStatementConstants(child)
		}
	case *ast.ExpressionStatement:
		node.Expression = foldExpressionConstants(node.Expression)
	case *ast.LetStatement:
		node.Value = foldExpressionConstants(node.Value)
	case *ast.AssignmentStatement:
		node.Value = foldExpressionConstants(node.Value)
	case *ast.MemberAssignmentStatement:
		node.Target.Object = foldExpressionConstants(node.Target.Object)
		node.Value = foldExpressionConstants(node.Value)
	case *ast.IndexAssignmentStatement:
		node.Target.Left = foldExpressionConstants(node.Target.Left)
		node.Target.Index = foldExpressionConstants(node.Target.Index)
		node.Value = foldExpressionConstants(node.Value)
	case *ast.ReturnStatement:
		node.ReturnValue = foldExpressionConstants(node.ReturnValue)
	case *ast.DeferStatement:
		node.Call.Function = foldExpressionConstants(node.Call.Function)
		foldExpressionSlice(node.Call.Arguments)
	case *ast.ForStatement:
		node.Iterable = foldExpressionConstants(node.Iterable)
		foldStatementConstants(node.Body)
	case *ast.WhileStatement:
		node.Condition = foldExpressionConstants(node.Condition)
		foldStatementConstants(node.Body)
	}
	return statement
}

func foldExpressionConstants(expression ast.Expression) ast.Expression {
	if expression == nil {
		return nil
	}

	switch node := expression.(type) {
	case *ast.PrefixExpression:
		node.Right = foldExpressionConstants(node.Right)
		right, ok := literalObject(node.Right)
		if !ok {
			return node
		}
		return foldedLiteralOrOriginal(evalPrefixExpression(node.Operator, right), node)

	case *ast.InfixExpression:
		node.Left = foldExpressionConstants(node.Left)
		node.Right = foldExpressionConstants(node.Right)
		left, leftOK := literalObject(node.Left)
		right, rightOK := literalObject(node.Right)
		if !leftOK || !rightOK {
			return node
		}
		return foldedLiteralOrOriginal(evalInfixExpression(node.Operator, left, right), node)

	case *ast.IfExpression:
		node.Condition = foldExpressionConstants(node.Condition)
		foldStatementConstants(node.Consequence)
		if node.Alternative != nil {
			foldStatementConstants(node.Alternative)
		}

	case *ast.TryExpression:
		foldStatementConstants(node.Body)
		for _, clause := range node.Catches {
			foldStatementConstants(clause.Body)
		}

	case *ast.FunctionLiteral:
		foldStatementConstants(node.Body)

	case *ast.CallExpression:
		node.Function = foldExpressionConstants(node.Function)
		foldExpressionSlice(node.Arguments)

	case *ast.StructLiteral:
		node.StructType = foldExpressionConstants(node.StructType)
		foldExpressionSlice(node.Values)

	case *ast.ArrayLiteral:
		foldExpressionSlice(node.Elements)

	case *ast.IndexExpression:
		node.Left = foldExpressionConstants(node.Left)
		node.Index = foldExpressionConstants(node.Index)

	case *ast.MapLiteral:
		pairs := make(map[ast.Expression]ast.Expression, len(node.Pairs))
		for key, value := range node.Pairs {
			pairs[foldExpressionConstants(key)] = foldExpressionConstants(value)
		}
		node.Pairs = pairs

	case *ast.MemberExpression:
		node.Object = foldExpressionConstants(node.Object)

	case *ast.ImportExpression:
		node.Path = foldExpressionConstants(node.Path)

	case *ast.TaskExpression:
		node.Work = foldExpressionConstants(node.Work)

	case *ast.CollectExpression:
		// Handles are identifiers, so there are no constant children to fold.
	}

	return expression
}

func foldExpressionSlice(expressions []ast.Expression) {
	for index, expression := range expressions {
		expressions[index] = foldExpressionConstants(expression)
	}
}

// literalObject converts only scalar, side-effect-free AST values. These are
// exactly the inputs the folder may pass to the evaluator's operator helpers.
func literalObject(expression ast.Expression) (object.Object, bool) {
	switch node := expression.(type) {
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}, true
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}, true
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value), true
	case *ast.StringLiteral:
		return &object.String{Value: node.Value}, true
	default:
		return nil, false
	}
}

func foldedLiteralOrOriginal(result object.Object, original ast.Expression) ast.Expression {
	position := original.Position()
	switch value := result.(type) {
	case *object.Integer:
		literal := strconv.FormatInt(value.Value, 10)
		return &ast.IntegerLiteral{
			Token: token.Token{Type: token.INT, Literal: literal, Position: position},
			Value: value.Value,
		}
	case *object.Float:
		return &ast.FloatLiteral{
			Token: token.Token{Type: token.FLOAT, Literal: value.Inspect(), Position: position},
			Value: value.Value,
		}
	case *object.Boolean:
		literal := "False"
		var tokenType token.TokenType = token.FALSE
		if value.Value {
			literal = "True"
			tokenType = token.TRUE
		}
		return &ast.Boolean{
			Token: token.Token{Type: tokenType, Literal: literal, Position: position},
			Value: value.Value,
		}
	case *object.String:
		return &ast.StringLiteral{
			Token: token.Token{Type: token.STRING, Literal: value.Value, Position: position},
			Value: value.Value,
		}
	default:
		return original
	}
}
