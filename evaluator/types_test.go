package evaluator

import (
	"silver/object"
	"testing"
)

func TestTypedLetBinding(t *testing.T) {
	testIntegerObject(t, testEval("let age: int = 36\nage"), 36)
}

func TestTypedLetBindingRejectsMismatch(t *testing.T) {
	evaluated := testEval(`let age: int = "thirty-six"`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for binding "age": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestTypedFunctionParameter(t *testing.T) {
	evaluated := testEval("let double = fn(value: int) int { value * 2 }\ndouble(\"two\")")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "value": expected int, got str`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestTypedFunctionReturnValue(t *testing.T) {
	evaluated := testEval("let label = fn() str { 42 }\nlabel()")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for return value of "label": expected str, got int`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestUnknownType(t *testing.T) {
	evaluated := testEval(`let value: Missing = 1`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `unknown type "Missing"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestStringTypeNameIsRejected(t *testing.T) {
	evaluated := testEval(`let value: string = "old spelling"`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `unknown type "string"`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestEnumTypeAnnotation(t *testing.T) {
	evaluated := testEval("enum Direction { North, South }\nlet direction: Direction = Direction.North\ndirection")
	value, ok := evaluated.(*object.EnumValue)
	if !ok || value.Member != "North" {
		t.Fatalf("result is %#v, want Direction.North", evaluated)
	}
}

func TestBareReturnUsesNull(t *testing.T) {
	evaluated := testEval("let noValue = fn() null {\nreturn\n}\nnoValue()")
	testNullObject(t, evaluated)
}

func TestFunctionWithoutReturnTypeReturnsNull(t *testing.T) {
	evaluated := testEval("let noValue = fn() {\nreturn 42\n}\nnoValue()")
	testNullObject(t, evaluated)
}

func TestTypedHigherOrderFunction(t *testing.T) {
	evaluated := testEval(`
let apply = fn(operation: call(int) int, value: int) int {
	return operation(value)
}
apply(fn(value: int) int { value * 2 }, 21)
`)
	testIntegerObject(t, evaluated, 42)
}

func TestCallSignatureRejectsWrongArgumentSignature(t *testing.T) {
	evaluated := testEval(`
let apply = fn(operation: call(int) int, value: int) int {
	return operation(value)
}
apply(fn(value: str) int { 1 }, 21)
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "operation": expected call(int) int, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestCallSignatureRejectsWrongArity(t *testing.T) {
	evaluated := testEval(`
let apply = fn(operation: call(int) int, value: int) int {
	return operation(value)
}
apply(fn(left: int, right: int) int { left + right }, 21)
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for parameter "operation": expected call(int) int, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestCallSignatureReturnTypeAndClosure(t *testing.T) {
	evaluated := testEval(`
let makeAdder = fn(amount: int) call(int) int {
	return fn(value: int) int { value + amount }
}
let addTwo: call(int) int = makeAdder(2)
addTwo(40)
`)
	testIntegerObject(t, evaluated, 42)
}

func TestCallSignatureRejectsWrongReturnedFunction(t *testing.T) {
	evaluated := testEval(`
let makeIdentity = fn() call(int) int {
	return fn(value: str) int { 1 }
}
makeIdentity()
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for return value of "makeIdentity": expected call(int) int, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestFunctionErrorStructUnwindsIntoDestructuringCatch(t *testing.T) {
	evaluated := testEval(`
struct FileNotFound {
	message: str
}
let open = fn(found: bool) str | FileNotFound {
	if (found) { "contents" } else { FileNotFound{"missing"} }
}
let describe = fn(message: str) str { message }
try {
	describe(open(False))
} catch FileNotFound err {
	describe(err)
}
`)
	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "missing" {
		t.Fatalf("result is %#v, want destructured error message", evaluated)
	}
}

func TestFunctionMayReturnSuccessFromReturnUnion(t *testing.T) {
	evaluated := testEval(`
struct FileNotFound { message: str }
let open = fn() str | FileNotFound { "contents" }
open()
`)
	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "contents" {
		t.Fatalf("result is %#v, want contents", evaluated)
	}
}

func TestLeadingPipeDeclaresNullSuccess(t *testing.T) {
	evaluated := testEval(`
struct PermissionDenied { message: str }
let writeFile = fn(allowed: bool) | PermissionDenied {
	if (allowed) { return }
	return PermissionDenied{"denied"}
}
writeFile(True)
`)
	testNullObject(t, evaluated)
}

func TestLeadingPipeTreatsEmptyBodyAsNullSuccess(t *testing.T) {
	evaluated := testEval(`
struct PermissionDenied { message: str }
let writeFile = fn() | PermissionDenied {}
writeFile()
`)
	testNullObject(t, evaluated)
}

func TestLeadingPipeRaisesErrorStruct(t *testing.T) {
	evaluated := testEval(`
struct PermissionDenied { message: str }
let writeFile = fn() | PermissionDenied {
	return PermissionDenied{"denied"}
}
try {
	writeFile()
} catch PermissionDenied err {
	err.message
}
`)
	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "denied" {
		t.Fatalf("result is %#v, want destructured error message", evaluated)
	}
}

func TestReturnUnionRejectsUndeclaredValue(t *testing.T) {
	evaluated := testEval(`
struct FileNotFound { message: str }
let open = fn() str | FileNotFound { 42 }
open()
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for return value of "open": expected str | FileNotFound, got int`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestErrorReturnAlternativeMustBeStruct(t *testing.T) {
	evaluated := testEval(`let invalid = fn() int | str { 1 }`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `error return type "str" must be a struct`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestCallableReturnUnionAcceptsFunctionWithFewerErrors(t *testing.T) {
	evaluated := testEval(`
struct FileNotFound { message: str }
let opener: call() str | FileNotFound = fn() str { "contents" }
opener()
`)
	value, ok := evaluated.(*object.String)
	if !ok || value.Value != "contents" {
		t.Fatalf("result is %#v, want contents", evaluated)
	}
}

func TestCallableReturnUnionRejectsUndeclaredFunctionError(t *testing.T) {
	evaluated := testEval(`
struct FileNotFound { message: str }
let opener: call() str = fn() str | FileNotFound { "contents" }
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for binding "opener": expected call() str, got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}
