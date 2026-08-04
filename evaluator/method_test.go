package evaluator

import (
	"silver/object"
	"testing"
)

func TestCallableStructFieldBindsReceiver(t *testing.T) {
	evaluated := testEval(`
struct Scale {
	x: float
	y: float
	z: float
}
let grow = fn(scale: Scale, amount: float) {
	scale.x = scale.x + amount
	scale.y = scale.y + amount
	scale.z = scale.z + amount
}
struct Transform {
	scale: Scale
	grow: call(scale: Scale, amount: float)
}
let actor = Transform{Scale{1.0, 1.0, 1.0}, grow}
actor.grow(5.0)
actor.scale.x + actor.scale.y + actor.scale.z
`)
	testFloatObject(t, evaluated, 18.0)
}

func TestBoundCallableFieldUsesItsOwnReceiver(t *testing.T) {
	evaluated := testEval(`
struct Scale {
	x: int
}
let grow = fn(scale: Scale, amount: int) {
	scale.x = scale.x + amount
}
struct Transform {
	scale: Scale
	grow: call(scale: Scale, amount: int)
}
let first = Transform{Scale{1}, grow}
let second = Transform{Scale{10}, grow}
first.grow(2)
first.scale.x * 10 + second.scale.x
`)
	testIntegerObject(t, evaluated, 40)
}

func TestCallableStructFieldSignatureChecksParameterNames(t *testing.T) {
	evaluated := testEval(`
struct Scale {
	x: int
}
let wrong = fn(target: Scale, amount: int) {}
struct Transform {
	scale: Scale
	grow: call(scale: Scale, amount: int)
}
Transform{Scale{1}, wrong}
`)
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if got, want := err.MessageText(), `type mismatch for field "Transform.grow": expected call(scale: Scale, amount: int), got call`; got != want {
		t.Fatalf("error message is %q, want %q", got, want)
	}
}

func TestBareCallStructFieldDoesNotBindReceiver(t *testing.T) {
	evaluated := testEval(`
let identity = fn(value: int) int { value }
struct Box {
	callback: call
}
let box = Box{identity}
box.callback(42)
`)
	testIntegerObject(t, evaluated, 42)
}
