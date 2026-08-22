package evaluator

import (
	"silver/ast"
	"silver/object"
	"silver/token"
)

// evalExpressions evaluates expressions from left to right. On failure it
// returns a one-element slice containing the error so callers can propagate it.
func (e *Evaluator) evalExpressions(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, expression := range exps {
		evaluated := e.Eval(expression, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}

	return result
}

// evalCallArguments evaluates call operands from left to right and expands
// variadic parameter references into their individual positional arguments.
func (e *Evaluator) evalCallArguments(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object

	for _, expression := range exps {
		evaluated := e.Eval(expression, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		if variadic, ok := evaluated.(*object.VariadicArguments); ok {
			result = append(result, variadic.Elements...)
			continue
		}
		result = append(result, evaluated)
	}

	return result
}

// evalPrefixExpression dispatches unary operators by their source spelling.
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError(object.RuntimeErrorKindType, "unknown operator: %s%s", operator, right.Type())
	}
}

// evalInfixExpression selects type-specific binary semantics and performs the
// common type mismatch checks.
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case operator == "&&":
		return nativeBoolToBooleanObject(isTruthy(left) && isTruthy(right))
	case operator == "||":
		return nativeBoolToBooleanObject(isTruthy(left) || isTruthy(right))
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case isNumeric(left) && isNumeric(right):
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError(object.RuntimeErrorKindType, "type mismatch: %s %s %s", left.Type(), operator, right.Type())
	default:
		return newError(object.RuntimeErrorKindType, "unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalIndexExpression dispatches primitive indexing directly and lets struct
// values implement bracket reads through get_item.
func (e *Evaluator) evalIndexExpression(node *ast.IndexExpression, left, index object.Object) object.Object {
	switch left := left.(type) {
	case *object.Array:
		return evalArrayIndexExpression(left, index)
	case *object.Map:
		return evalMapIndexExpression(left, index)
	case *object.StructInstance:
		return e.callStructIndexMethod(node, left, "get_item", []object.Object{index})
	default:
		return newError(object.RuntimeErrorKindType, "index operator not supported: %s", left.Type())
	}
}

// evalIfExpression evaluates only the selected branch.
func (e *Evaluator) evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	condition := e.Eval(ie.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return e.Eval(ie.Consequence, env)
	} else if ie.Alternative != nil {
		return e.Eval(ie.Alternative, env)
	} else {
		return NULL
	}
}

// evalSwitchExpression evaluates the switch value once and each case value
// only when reached. Comparisons use the same evaluator path as source-level
// == expressions, including struct operator methods.
func (e *Evaluator) evalSwitchExpression(se *ast.SwitchExpression, env *object.Environment) object.Object {
	value := e.Eval(se.Value, env)
	if isError(value) {
		return value
	}

	for _, switchCase := range se.Cases {
		caseValue := e.Eval(switchCase.Value, env)
		if isError(caseValue) {
			return caseValue
		}
		comparison := &ast.InfixExpression{
			Token:    token.Token{Type: token.EQ, Literal: "==", Position: switchCase.Token.Position},
			Left:     se.Value,
			Operator: "==",
			Right:    switchCase.Value,
		}
		matched := e.evalInfixExpression(comparison, value, caseValue)
		if isError(matched) {
			return matched
		}
		if isTruthy(matched) {
			return e.Eval(switchCase.Body, env)
		}
	}

	if se.Default != nil {
		return e.Eval(se.Default, env)
	}
	return NULL
}

// evalMinusPrefixOperatorExpression negates an integer or float and reports a
// type error for other operands.
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch right := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -right.Value}
	case *object.Float:
		return &object.Float{Value: -right.Value}
	default:
		return newError(object.RuntimeErrorKindType, "unknown operator: -%s", right.Type())
	}
}
