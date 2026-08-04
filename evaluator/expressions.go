package evaluator

import (
	"silver/ast"
	"silver/object"
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

// evalIndexExpression dispatches indexing according to the left operand type.
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
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
