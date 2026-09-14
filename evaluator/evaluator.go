package evaluator

import (
	"fmt"
	"io"
	"os"
	"silver/ast"
	"silver/object"
	"silver/packages"
	"silver/parser"
	stdlibpkg "silver/stdlib"
	"sync"
	"sync/atomic"
)

// NULL is the canonical null singleton used by identity-based truthiness.
var NULL = &object.Null{}

// Evaluator combines a shared interpreter session with its execution context.
// Reuse one evaluator for a REPL or a group of related evaluations.
type Evaluator struct {
	*evaluatorSession
	constants *constantPool
	contexts  []string // active Silver function/module names
}

// evaluatorSession is retained by every evaluator fork. Module identity and
// loading state must outlive any individual template invocation.
type evaluatorSession struct {
	standardLibrary *stdlibpkg.Library
	modules         *moduleStore
	// nextEnumValueID gives every evaluated enum member a session-unique hash
	// identity, even when separate modules declare enums with the same names.
	nextEnumValueID *atomic.Uint64
	operatorScopes  *operatorScopeSet
	packages        *packages.Index
	packageStates   *packageStateSet
}

type operatorScopeSet struct {
	mu     sync.Mutex
	values map[string]*operatorScope
}

type operatorScope struct {
	registry *parser.InfixRegistry
	infix    *infixDefinitions
}

type infixDefinitions struct {
	mu     sync.Mutex
	values map[string]*infixDefinition
}

type infixDefinition struct {
	node     *ast.OperatorStatement
	env      *object.Environment
	callable *object.Function
}

// New constructs an evaluator whose print builtin writes to standard output.
func New() *Evaluator {
	return NewWithStreams(os.Stdin, os.Stdout, os.Stderr)
}

// InfixRegistry returns the parser registry for this interpreter session.
// Interactive and embedding frontends should use it for every parsed source.
func (e *Evaluator) InfixRegistry() *parser.InfixRegistry {
	return e.operatorScope("").registry
}

func (e *Evaluator) operatorScope(packageID string) *operatorScope {
	e.operatorScopes.mu.Lock()
	defer e.operatorScopes.mu.Unlock()
	if scope := e.operatorScopes.values[packageID]; scope != nil {
		return scope
	}
	scope := &operatorScope{
		registry: parser.NewInfixRegistry(),
		infix:    &infixDefinitions{values: make(map[string]*infixDefinition)},
	}
	e.operatorScopes.values[packageID] = scope
	return scope
}

// NewWithOutput constructs an evaluator with an explicit destination for
// language-level output. A nil writer discards output safely.
func NewWithOutput(out io.Writer) *Evaluator {
	if out == nil {
		out = io.Discard
	}
	return newEvaluator(os.Stdin, out, out)
}

// NewWithStreams constructs an evaluator with explicit language-level stdin,
// stdout, and stderr.
func NewWithStreams(in io.Reader, out, errOut io.Writer) *Evaluator {
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}
	return newEvaluator(in, out, errOut)
}

func newEvaluator(in io.Reader, out, errOut io.Writer) *Evaluator {
	return &Evaluator{
		evaluatorSession: &evaluatorSession{
			standardLibrary: stdlibpkg.NewWithStreams(in, out, errOut, NULL, TRUE, FALSE),
			modules:         newModuleStore(),
			nextEnumValueID: &atomic.Uint64{},
			operatorScopes:  &operatorScopeSet{values: make(map[string]*operatorScope)},
			packages:        packages.NewIndex(),
			packageStates:   newPackageStateSet(),
		},
		constants: newConstantPool(),
	}
}

// fork copies execution context for lazy templates and retains the session.
func (e *Evaluator) fork() *Evaluator {
	return &Evaluator{
		evaluatorSession: e.evaluatorSession,
		constants:        newConstantPool(),
		contexts:         append([]string(nil), e.contexts...),
	}
}

// Eval annotates a newly-created runtime error with the current AST location.
// Errors that are merely propagating already have an origin, so SetOrigin
// leaves their traceback unchanged.
func (e *Evaluator) Eval(node ast.Node, env *object.Environment) object.Object {
	result := e.eval(node, env)
	if failure, ok := result.(*object.Error); ok {
		failure.SetOrigin(e.traceFrame(node))
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
		val := e.evalValue(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &object.Break{}

	case *ast.ContinueStatement:
		return &object.Continue{}

	case *ast.AssertStatement:
		condition := e.evalValue(node.Condition, env)
		if isError(condition) {
			return condition
		}
		if isTruthy(condition) {
			return NULL
		}
		message := ""
		if node.Message != nil {
			value := e.evalValue(node.Message, env)
			if isError(value) {
				return value
			}
			message = value.Inspect()
		}
		return newError(object.RuntimeErrorKindAssertion, "%s", message)

	case *ast.DeferStatement:
		function := e.evalValue(node.Call.Function, env)
		if isError(function) {
			return function
		}
		arguments := e.evalCallArguments(node.Call.Arguments, env)
		if len(arguments) == 1 && isError(arguments[0]) {
			return arguments[0]
		}
		env.RegisterDefer(object.DeferredCall{Function: function, Arguments: arguments, Call: node.Call})
		return NULL

	case *ast.ExportStatement:
		// Export declarations affect the module object assembled after the
		// source finishes evaluating; they do not alter lexical bindings.
		return NULL

	case *ast.OperatorStatement:
		return e.registerInfix(node, env)

	case *ast.TypeStatement:
		return e.evalTypeStatement(node, env)

	case *ast.LetStatement:
		contract, err := object.ResolveContract(node.Name.Type, env)
		if err != nil {
			return err
		}
		val := e.evalValue(node.Value, env)
		if isError(val) {
			return val
		}
		if err := e.requireType(contract, val, fmt.Sprintf("binding %q", node.Name.Value)); err != nil {
			return err
		}
		if function, ok := val.(*object.Function); ok {
			if function.Name == "" {
				function.Name = node.Name.Value
			}
		}
		env.SetTyped(node.Name.Value, val, contract)

	case *ast.AssignmentStatement:
		return e.evalAssignment(node, env)

	case *ast.MemberAssignmentStatement:
		return e.evalMemberAssignment(node, env)

	case *ast.IndexAssignmentStatement:
		return e.evalIndexAssignment(node, env)

	case *ast.ForStatement:
		return e.evalForStatement(node, env)

	case *ast.WhileStatement:
		return e.evalWhileStatement(node, env)

	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	case *ast.ImportExpression:
		pathValue := e.evalValue(node.Path, env)
		if isError(pathValue) {
			return pathValue
		}
		path, ok := pathValue.(*object.String)
		if !ok {
			return newError(object.RuntimeErrorKindType, "import path must be str, got %s", runtimeTypeName(pathValue))
		}
		result := e.importModule(path.Value, env)
		e.prependCallerFrame(result, node)
		return result

	case *ast.MemberExpression:
		value := e.evalValue(node.Object, env)
		if isError(value) {
			return value
		}
		return e.evalMember(value, node.Member.Value)

	// Expressions
	case *ast.IfExpression:
		return e.evalIfExpression(node, env)

	case *ast.SwitchExpression:
		return e.evalSwitchExpression(node, env)

	case *ast.TryExpression:
		return e.evalTryExpression(node, env)

	case *ast.ExpressionStatement:
		return e.Eval(node.Expression, env)

	case *ast.PrefixExpression:
		right := e.evalValue(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)

	case *ast.InfixExpression:
		left := e.evalValue(node.Left, env)
		if isError(left) {
			return left
		}
		if node.Operator == "&&" && !isTruthy(left) {
			return FALSE
		}
		if node.Operator == "||" && isTruthy(left) {
			return TRUE
		}

		right := e.evalValue(node.Right, env)
		if isError(right) {
			return right
		}
		if node.Operator == "&&" || node.Operator == "||" {
			return nativeBoolToBooleanObject(isTruthy(right))
		}
		if instance, ok := left.(*object.StructInstance); ok && e.structOperatorVisible(node.Operator, instance, env) {
			if _, exists := instance.Get(node.Operator); exists {
				return e.evalStructInfixExpression(node, instance, right)
			}
		}
		if callable, failure := e.infixCallable(node.Operator, env.PackageID()); failure != nil {
			return failure
		} else if callable != nil {
			result := e.applyFunction(callable, []object.Object{left, right})
			e.prependCallerFrame(result, node)
			return result
		}

		return e.evalInfixExpression(node, left, right, env)

	case *ast.IntegerLiteral:
		return e.constants.integer(node.Value)

	case *ast.FloatLiteral:
		return e.constants.float(node.Value)

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.FunctionLiteral:
		parameterTypes := make([]*object.Contract, len(node.Parameters))
		for index, parameter := range node.Parameters {
			contract, err := object.ResolveContract(parameter.Type, env)
			if err != nil {
				return err
			}
			parameterTypes[index] = contract
		}
		returnType, err := object.ResolveContract(node.ReturnType, env)
		if err != nil {
			return err
		}
		errorTypes := make([]*object.Contract, len(node.ErrorTypes))
		for index, errorType := range node.ErrorTypes {
			contract, err := object.ResolveErrorContract(errorType, env)
			if err != nil {
				return err
			}
			errorTypes[index] = contract
		}
		return &object.Function{
			Parameters:     node.Parameters,
			ParameterTypes: parameterTypes,
			ReturnType:     returnType,
			ErrorTypes:     errorTypes,
			Env:            env,
			Body:           node.Body,
		}

	case *ast.CallExpression:
		function := e.evalValue(node.Function, env)
		if isError(function) {
			return function
		}
		args := e.evalCallArguments(node.Arguments, env)
		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		result := e.applyFunction(function, args)
		e.prependCallerFrame(result, node)
		return result

	case *ast.StructLiteral:
		structType := e.evalValue(node.StructType, env)
		if isError(structType) {
			return structType
		}
		definition, ok := structType.(*object.Struct)
		if !ok {
			return newError(object.RuntimeErrorKindType, "not a struct: %s", runtimeTypeName(structType))
		}
		values := e.evalExpressions(node.Values, env)
		if len(values) == 1 && isError(values[0]) {
			return values[0]
		}
		return e.applyStruct(definition, values)

	case *ast.StringLiteral:
		return e.constants.string(node.Value)

	case *ast.TemplateStringLiteral:
		return e.evalTemplateStringLiteral(node, env)

	case *ast.ArrayLiteral:
		elements := e.evalExpressions(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}

	case *ast.IndexExpression:
		left := e.evalValue(node.Left, env)
		if isError(left) {
			return left
		}
		index := e.evalValue(node.Index, env)
		if isError(index) {
			return index
		}
		return e.evalIndexExpression(node, left, index)

	case *ast.MapLiteral:
		return e.evalMapLiteral(node, env)
	}

	return nil
}

func (e *Evaluator) registerInfix(node *ast.OperatorStatement, env *object.Environment) object.Object {
	infix := e.operatorScope(env.PackageID()).infix
	infix.mu.Lock()
	defer infix.mu.Unlock()
	if _, exists := infix.values[node.Symbol]; exists {
		return newError(object.RuntimeErrorKindName, "operator %q is already defined", node.Symbol)
	}
	infix.values[node.Symbol] = &infixDefinition{node: node, env: env}
	return NULL
}

func (e *Evaluator) infixCallable(symbol, packageID string) (*object.Function, *object.Error) {
	infix := e.operatorScope(packageID).infix
	infix.mu.Lock()
	defer infix.mu.Unlock()
	definition := infix.values[symbol]
	if definition == nil {
		return nil, nil
	}
	if definition.callable != nil {
		return definition.callable, nil
	}
	callable := e.evalValue(definition.node.Function, definition.env)
	if failure, ok := callable.(*object.Error); ok {
		return nil, failure
	}
	function, ok := callable.(*object.Function)
	if !ok {
		return nil, newError(object.RuntimeErrorKindType, "operator %q must be defined by a function", symbol)
	}
	function.Operator = true
	function.Name = fmt.Sprintf("operator %q", symbol)
	definition.callable = function
	return function, nil
}
