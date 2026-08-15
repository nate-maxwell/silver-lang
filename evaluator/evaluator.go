package evaluator

import (
	"fmt"
	"io"
	"os"
	"silver/ast"
	"silver/object"
	stdlibpkg "silver/stdlib"
	"sync"
	"sync/atomic"
)

// NULL is the canonical null singleton used by identity-based truthiness.
var NULL = &object.Null{}

// Evaluator owns the state shared across one execution session: the standard
// library, imported-module caches, circular-import state, and traceback
// contexts. Reuse one evaluator for a REPL or a group of related evaluations.
type Evaluator struct {
	standardLibrary *stdlibpkg.Library
	constants       *constantPool
	modules         map[string]*object.Module // filepath or standard-library name to module
	loading         map[string]bool           // module load state | circular import detection
	contexts        []string                  // active Silver function/module names
	warnings        io.Writer                 // scope-exit task diagnostics
	// nextEnumValueID gives every evaluated enum member a session-unique hash
	// identity, even when separate modules declare enums with the same names.
	nextEnumValueID *atomic.Uint64
}

// synchronizedWriter makes builtin output and warnings safe when tasks write
// from multiple goroutines.
type synchronizedWriter struct {
	mu *sync.Mutex
	w  io.Writer
}

func (w *synchronizedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}

// New constructs an evaluator whose print builtin writes to standard output.
func New() *Evaluator {
	return NewWithStreams(os.Stdin, os.Stdout, os.Stderr)
}

// NewWithOutput constructs an evaluator with an explicit destination for
// language-level output. A nil writer discards output safely.
func NewWithOutput(out io.Writer) *Evaluator {
	if out == nil {
		out = io.Discard
	}
	safe := &synchronizedWriter{mu: &sync.Mutex{}, w: out}
	return newEvaluator(os.Stdin, safe, safe, safe)
}

// NewWithWriters constructs an evaluator with separate program-output and
// warning destinations. The CLI uses stderr for warnings while the REPL keeps
// all diagnostics in its single output stream.
func NewWithWriters(out, warnings io.Writer) *Evaluator {
	if out == nil {
		out = io.Discard
	}
	if warnings == nil {
		warnings = io.Discard
	}
	streamLock := &sync.Mutex{}
	safeOut := &synchronizedWriter{mu: streamLock, w: out}
	safeWarnings := &synchronizedWriter{mu: streamLock, w: warnings}
	return newEvaluator(os.Stdin, safeOut, safeWarnings, safeWarnings)
}

// NewWithStreams constructs an evaluator with explicit language-level stdin,
// stdout, and stderr. Runtime warnings share stderr.
func NewWithStreams(in io.Reader, out, errOut io.Writer) *Evaluator {
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = io.Discard
	}
	streamLock := &sync.Mutex{}
	safeOut := &synchronizedWriter{mu: streamLock, w: out}
	safeErrOut := &synchronizedWriter{mu: streamLock, w: errOut}
	return newEvaluator(in, safeOut, safeErrOut, safeErrOut)
}

func newEvaluator(in io.Reader, out, errOut, warnings io.Writer) *Evaluator {
	return &Evaluator{
		standardLibrary: stdlibpkg.NewWithStreams(in, out, errOut, NULL, TRUE, FALSE),
		constants:       newConstantPool(),
		modules:         make(map[string]*object.Module),
		loading:         make(map[string]bool),
		contexts:        make([]string, 0),
		warnings:        warnings,
		nextEnumValueID: &atomic.Uint64{},
	}
}

// fork gives a task independent mutable evaluator state while sharing the
// immutable standard library, synchronized output, and enum identity
// source.
func (e *Evaluator) fork() *Evaluator {
	modules := make(map[string]*object.Module, len(e.modules))
	for path, module := range e.modules {
		modules[path] = module
	}
	return &Evaluator{
		standardLibrary: e.standardLibrary,
		constants:       newConstantPool(),
		modules:         modules,
		loading:         make(map[string]bool),
		contexts:        append([]string(nil), e.contexts...),
		warnings:        e.warnings,
		nextEnumValueID: e.nextEnumValueID,
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
		val := e.Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.BreakStatement:
		return &object.Break{}

	case *ast.ContinueStatement:
		return &object.Continue{}

	case *ast.AssertStatement:
		condition := e.Eval(node.Condition, env)
		if isError(condition) {
			return condition
		}
		if isTruthy(condition) {
			return NULL
		}
		message := ""
		if node.Message != nil {
			value := e.Eval(node.Message, env)
			if isError(value) {
				return value
			}
			message = value.Inspect()
		}
		return newError(object.RuntimeErrorKindAssertion, "%s", message)

	case *ast.DeferStatement:
		function := e.Eval(node.Call.Function, env)
		if isError(function) {
			return function
		}
		arguments := e.evalExpressions(node.Call.Arguments, env)
		if len(arguments) == 1 && isError(arguments[0]) {
			return arguments[0]
		}
		env.RegisterDefer(object.DeferredCall{Function: function, Arguments: arguments, Call: node.Call})
		return NULL

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
		if task, ok := val.(*object.Task); ok {
			task.SetName(node.Name.Value)
		}
		env.SetTyped(node.Name.Value, val, node.Name.Type)

	case *ast.AssignmentStatement:
		return e.evalAssignment(node, env)

	case *ast.MemberAssignmentStatement:
		return e.evalMemberAssignment(node, env)

	case *ast.IndexAssignmentStatement:
		return e.evalIndexAssignment(node, env)

	case *ast.EnumStatement:
		return e.evalEnumStatement(node, env)

	case *ast.StructStatement:
		return e.evalStructStatement(node, env)

	case *ast.ForStatement:
		return e.evalForStatement(node, env)

	case *ast.WhileStatement:
		return e.evalWhileStatement(node, env)

	case *ast.Identifier:
		return e.evalIdentifier(node, env)

	case *ast.ImportExpression:
		pathValue := e.Eval(node.Path, env)
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
		value := e.Eval(node.Object, env)
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
		if node.Operator == "&&" && !isTruthy(left) {
			return FALSE
		}
		if node.Operator == "||" && isTruthy(left) {
			return TRUE
		}

		right := e.Eval(node.Right, env)
		if isError(right) {
			return right
		}
		if node.Operator == "&&" || node.Operator == "||" {
			return nativeBoolToBooleanObject(isTruthy(right))
		}

		return e.evalInfixExpression(node, left, right)

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
		for _, errorType := range node.ErrorTypes {
			if err := e.validateErrorTypeAnnotation(errorType, env); err != nil {
				return err
			}
		}
		params := node.Parameters
		body := node.Body
		return &object.Function{
			Parameters: params,
			ReturnType: node.ReturnType,
			ErrorTypes: node.ErrorTypes,
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

	case *ast.StructLiteral:
		structType := e.Eval(node.StructType, env)
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
		left := e.Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := e.Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return e.evalIndexExpression(node, left, index)

	case *ast.MapLiteral:
		return e.evalMapLiteral(node, env)

	case *ast.TaskExpression:
		return e.evalTaskExpression(node, env)

	case *ast.CollectExpression:
		return e.evalCollectExpression(node, env)
	}

	return nil
}
