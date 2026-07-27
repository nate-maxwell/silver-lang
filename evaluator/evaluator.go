package evaluator

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"silver/ast"
	"silver/lexer"
	"silver/object"
	"silver/parser"
	"strings"
)

// Canonical singleton values make identity-based boolean and null comparisons
// deterministic throughout evaluation.
var (
	NULL  = &object.Null{}
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
)

// Evaluator owns the state shared across one execution session: native
// builtins, imported-module caches, circular-import state, and traceback
// contexts. Reuse one evaluator for a REPL or a group of related evaluations.
type Evaluator struct {
	builtins builtinRegistry
	modules  map[string]*object.Module // filepath to module
	loading  map[string]bool           // module load state | circular import detection
	contexts []string                  // active Silver function/module names
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
		builtins: newDefaultBuiltinRegistry(out),
		modules:  make(map[string]*object.Module),
		loading:  make(map[string]bool),
		contexts: make([]string, 0),
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
		val := e.Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}

	case *ast.LetStatement:
		val := e.Eval(node.Value, env)
		if isError(val) {
			return val
		}
		if function, ok := val.(*object.Function); ok && function.Name == "" {
			function.Name = node.Name.Value
		}
		env.Set(node.Name.Value, val)

	case *ast.EnumStatement:
		return e.evalEnumStatement(node, env)

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
		return &object.Integer{Value: node.Value}

	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)

	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		return &object.Function{Parameters: params, Env: env, Body: body}

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
		return &object.String{Value: node.Value}

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

// EvalFile parses and evaluates path in env. It also sets env's source
// directory so relative imports resolve beside the entry file.
func (e *Evaluator) EvalFile(path string, env *object.Environment) object.Object {
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return newError("could not resolve file %q: %s", path, err)
	}
	absolutePath = filepath.Clean(absolutePath)

	program, parseError := parseFile(absolutePath)
	if parseError != nil {
		return parseError
	}

	env.SetSourceDir(filepath.Dir(absolutePath))
	return e.Eval(program, env)
}

// importModule resolves, loads, and evaluates a module in an isolated top-level
// environment. Successful modules are cached by canonical absolute path.
func (e *Evaluator) importModule(path string, env *object.Environment) object.Object {
	absolutePath, err := resolveImportPath(path, env.SourceDir())
	if err != nil {
		return newError("could not resolve import %q: %s", path, err)
	}

	if module, ok := e.modules[absolutePath]; ok {
		return module
	}
	if e.loading[absolutePath] {
		return newError("circular import detected while loading %q", absolutePath)
	}

	e.loading[absolutePath] = true
	defer delete(e.loading, absolutePath)

	program, parseError := parseFile(absolutePath)
	if parseError != nil {
		return parseError
	}

	moduleEnv := object.NewEnvironment()
	moduleEnv.SetSourceDir(filepath.Dir(absolutePath))
	e.pushContext("<module>")
	defer e.popContext()
	result := e.Eval(program, moduleEnv)
	if isError(result) {
		return result
	}

	module := &object.Module{Path: absolutePath, Exports: moduleEnv.Bindings()}
	e.modules[absolutePath] = module
	return module
}

// resolveImportPath resolves path relative to sourceDir, falling back to the
// process working directory for in-memory evaluation.
func resolveImportPath(path, sourceDir string) (string, error) {
	if sourceDir == "" {
		var err error
		sourceDir, err = os.Getwd()
		if err != nil {
			return "", err
		}
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(sourceDir, path)
	}

	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(absolutePath), nil
}

// parseFile reads a source file and parses it with its absolute path attached
// to every token for diagnostics and tracebacks.
func parseFile(path string) (*ast.Program, *object.Error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, newError("could not read %q: %s", path, err)
	}

	p := parser.New(lexer.NewWithSource(string(input), path))
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return nil, newError("could not parse %q:\n%s", path, strings.Join(p.Errors(), "\n"))
	}
	return program, nil
}

// evalMember resolves members on module and enum namespace objects.
func evalMember(value object.Object, member string) object.Object {
	switch value := value.(type) {
	case *object.Module:
		export, ok := value.Exports[member]
		if !ok {
			return newError("module %q has no member %q", value.Path, member)
		}
		return export
	case *object.Enum:
		enumValue, ok := value.Members[member]
		if !ok {
			return newError("enum %q has no member %q", value.Name, member)
		}
		return enumValue
	default:
		return newError("member access not supported on %s", value.Type())
	}
}

// evalIdentifier resolves lexical bindings before falling back to the native
// builtin registry.
func (e *Evaluator) evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := e.builtins.get(node.Value); ok {
		return builtin
	}
	return newError("identifier not found: %s", node.Value)
}

/* ----------------------------------------------------------------------------------------------------------
Functions and callables
---------------------------------------------------------------------------------------------------------- */

// applyFunction invokes either a Silver closure or native builtin. Silver calls
// create a lexical child environment and a named traceback context.
func (e *Evaluator) applyFunction(fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		if len(args) != len(fn.Parameters) {
			return newError("wrong number of arguments. got=%d, want=%d", len(args), len(fn.Parameters))
		}
		extendedEnv := extendFunctionEnv(fn, args)
		name := fn.Name
		if name == "" {
			name = "<anonymous>"
		}
		e.pushContext(name)
		defer e.popContext()
		evaluated := e.Eval(fn.Body, extendedEnv)
		return unwrapReturnValue(evaluated)

	case *object.Builtin:
		return fn.Fn(args...)

	default:
		return newError("not a function: %s", fn.Type())
	}
}

// extendFunctionEnv binds evaluated arguments to parameters in a child of the
// function's captured lexical environment. Arity is validated by applyFunction.
func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env)

	for i, param := range fn.Parameters {
		env.Set(param.Value, args[i])
	}

	return env
}

// unwrapReturnValue removes the evaluator's internal function-return wrapper.
func unwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}

	return obj
}

/* ----------------------------------------------------------------------------------------------------------
General expressions and statements
---------------------------------------------------------------------------------------------------------- */

// evalProgram evaluates top-level statements in order, stopping immediately on
// a return or error object.
func (e *Evaluator) evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range program.Statements {
		result = e.Eval(statement, env)

		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}

	return result
}

// evalBlockStatement evaluates a block until completion or until return/error
// control flow must propagate to an enclosing evaluator.
func (e *Evaluator) evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object

	for _, statement := range block.Statements {
		result = e.Eval(statement, env)

		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}

	return result
}

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
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

// evalInfixExpression selects type-specific binary semantics and performs the
// common type mismatch checks.
func evalInfixExpression(operator string, left, right object.Object) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case operator == "==":
		return nativeBoolToBooleanObject(left == right)
	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s", left.Type(), operator, right.Type())
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

// evalStringInfixExpression implements the operators supported by strings.
func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	if operator != "+" {
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}

	leftVal := left.(*object.String).Value
	rightVal := right.(*object.String).Value
	return &object.String{Value: leftVal + rightVal}
}

// evalIndexExpression dispatches indexing according to the left operand type.
func evalIndexExpression(left, index object.Object) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ && index.Type() == object.INTEGER_OBJ:
		return evalArrayIndexExpression(left, index)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index)
	default:
		return newError("index operator not supported: %s", left.Type())
	}
}

// evalArrayIndexExpression returns null for indexes outside the array bounds.
func evalArrayIndexExpression(array, index object.Object) object.Object {
	arrayObject := array.(*object.Array)
	idx := index.(*object.Integer).Value
	max := int64(len(arrayObject.Elements) - 1)

	if idx < 0 || idx > max {
		return NULL
	}

	return arrayObject.Elements[idx]
}

// evalHashIndexExpression normalizes a hashable key and returns null when the
// key is absent.
func evalHashIndexExpression(hash, index object.Object) object.Object {
	hashObject := hash.(*object.Hash)

	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}

	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		return NULL
	}

	return pair.Value
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

// isTruthy implements Silver truthiness. Only null and False are falsey.
func isTruthy(obj object.Object) bool {
	switch obj {
	case NULL:
		return false
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

// evalHashLiteral evaluates key/value pairs and rejects keys that do not
// implement object.Hashable.
func (e *Evaluator) evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)

	for keyNode, valueNode := range node.Pairs {
		key := e.Eval(keyNode, env)
		if isError(key) {
			return key
		}

		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}

		value := e.Eval(valueNode, env)
		if isError(value) {
			return value
		}

		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}
	}

	return &object.Hash{Pairs: pairs}
}

/* ----------------------------------------------------------------------------------------------------------
Unary operators
---------------------------------------------------------------------------------------------------------- */

// evalBangOperatorExpression implements logical negation using canonical
// singleton values.
func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	case NULL:
		return TRUE
	default:
		return FALSE
	}
}

// evalMinusPrefixOperatorExpression negates an integer or reports a type error.
func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	if right.Type() != object.INTEGER_OBJ {
		return newError("unknown operator: -%s", right.Type())
	}

	value := right.(*object.Integer).Value
	return &object.Integer{Value: -value}
}

/* ----------------------------------------------------------------------------------------------------------
Binary operators
---------------------------------------------------------------------------------------------------------- */

// evalIntegerInfixExpression implements integer arithmetic and comparisons.
// Division by zero becomes a Silver error rather than a Go panic.
func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value

	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		if rightVal == 0 {
			return newError("division by zero")
		}
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s", left.Type(), operator, right.Type())
	}
}

/* ----------------------------------------------------------------------------------------------------------
Booleans
---------------------------------------------------------------------------------------------------------- */

// nativeBoolToBooleanObject returns canonical boolean instances because
// equality for these values relies on object identity.
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

/* ----------------------------------------------------------------------------------------------------------
Error handling
---------------------------------------------------------------------------------------------------------- */

// newError formats a Silver runtime error without a source frame. Eval attaches
// the origin as the error propagates out of the AST node that created it.
func newError(format string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(format, a...)}
}

// isError safely identifies runtime error objects, including a nil result.
func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}
