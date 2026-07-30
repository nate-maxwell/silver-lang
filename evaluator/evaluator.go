package evaluator

import (
	"fmt"
	"io"
	"os"
	"silver/ast"
	builtinpkg "silver/evaluator/builtins"
	"silver/object"
)

// NULL is the canonical null singleton used by identity-based truthiness.
var NULL = &object.Null{}

// Evaluator owns the state shared across one execution session: native
// builtins, imported-module caches, circular-import state, and traceback
// contexts. Reuse one evaluator for a REPL or a group of related evaluations.
type Evaluator struct {
	builtins  *builtinpkg.Registry
	constants *constantPool
	modules   map[string]*object.Module // filepath to module
	loading   map[string]bool           // module load state | circular import detection
	contexts  []string                  // active Silver function/module names
	// nextEnumValueID gives every evaluated enum member a session-unique hash
	// identity, even when separate modules declare enums with the same names.
	nextEnumValueID uint64
}

// New constructs an evaluator whose print builtin writes to standard output.
func New() *Evaluator {
	return NewWithOutput(os.Stdout)
}

// NewWithOutput constructs an evaluator with an explicit destination for
// language-level output. A nil writer discards output safely.
func NewWithOutput(out io.Writer) *Evaluator {
	if out == nil {
		out = io.Discard
	}
	return &Evaluator{
		builtins:  builtinpkg.New(out, NULL),
		constants: newConstantPool(),
		modules:   make(map[string]*object.Module),
		loading:   make(map[string]bool),
		contexts:  make([]string, 0),
	}
}

// Eval preserves the original package API. Callers that evaluate more than
// one program, such as the REPL, should reuse an Evaluator returned by New so
// imported modules remain cached.
func Eval(node ast.Node, env *object.Environment) object.Object {
	return New().Eval(node, env)
}

// Eval annotates a newly-created runtime error with the current AST location.
// Errors that are merely propagating already have an origin, so SetOrigin
// leaves their traceback unchanged.
func (e *Evaluator) Eval(node ast.Node, env *object.Environment) object.Object {
	result := e.eval(node, env)
	if err, ok := result.(*object.Error); ok {
		err.SetOrigin(e.traceFrame(node))
	}
	return result
}

// eval dispatches AST node semantics. Eval wraps this method to attach source
// information to newly-created errors in one central place.
func (e *Evaluator) eval(node ast.Node, env *object.Environment) object.Object {
	switch node := node.(type) {

	//Statements
	case *ast.Program:
		return e.evalProgram(node, env)

	case *ast.BlockStatement:
		return e.evalBlockStatement(node, env)

	case *ast.ReturnStatement:
		if node.ReturnValue == nil {
			return &object.ReturnValue{Value: NULL}
		}
		val := e.Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		if err := e.validateTypeAnnotation(node.Name.Type, env); err != nil {
			return err
		}
		val := e.Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if err := e.requireType(node.Name.Type, val, env, fmt.Sprintf("binding %q", node.Name.Value)); err != nil {
			return err
		}
		if function, ok := val.(*object.Function); ok {
			if function.Name == "" {
				function.Name = node.Name.Value
			}
		}
		env.Set(node.Name.Value, val)

	case *ast.EnumStatement:
		return e.evalEnumStatement(node, env)

	case *ast.StructStatement:
		return e.evalStructStatement(node, env)

	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	case *ast.ImportExpression:
		result := e.importModule(node.Path.Value, env)
		e.prependCallerFrame(result, node)
		return result

	case *ast.MemberExpression:
		value := e.Eval(node.Object, env)
		if isError(value) {
			return value
		}
		return evalMember(value, node.Member.Value)

	// Expressions
	case *ast.IfExpression:
		return e.evalIfExpression(node, env)

	case *ast.ExpressionStatement:
		return e.Eval(node.Expression, env)

	case *ast.PrefixExpression:
		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := e.Eval(node.Left, env)
		if isError(left) {
			return left
		}

		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}

		return evalInfixExpression(node.Operator, left, right)

	case *ast.IntegerLiteral:
		return e.constants.integer(node.Value)

	case *ast.FloatLiteral:
		return e.constants.float(node.Value)

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.FunctionLiteral:
		for _, parameter := range node.Parameters {
			if err := e.validateTypeAnnotation(parameter.Type, env); err != nil {
				return err
			}
		}
		if err := e.validateTypeAnnotation(node.ReturnType, env); err != nil {
			return err
		}
		params := node.Parameters
		body := node.Body
		return &object.Function{
			Parameters: params,
			ReturnType: node.ReturnType,
			Env:        env,
			Body:       body,
		}

	case *ast.CallExpression:
		function := e.Eval(node.Function, env)
		if isError(function) {
			return function
		}
		args := e.evalExpressions(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		result := e.applyFunction(function, args)
		e.prependCallerFrame(result, node)
		return result

	case *ast.StringLiteral:
		return e.constants.string(node.Value)

	case *ast.ArrayLiteral:
		elements := e.evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := e.Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := e.Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index)

	case *ast.HashLiteral:
		return e.evalHashLiteral(node, env)
	}

	return nil
}
