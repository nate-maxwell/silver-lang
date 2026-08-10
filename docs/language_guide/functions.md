# Functions

Functions are first-class values introduced by `fn`. They may be assigned to bindings, passed to other functions,
returned as results, stored in arrays and maps, or placed in callable struct fields.

## Declaring and calling functions

A function lists its parameters between parentheses and places an optional return annotation before its body:

```silver
let add = fn(left: int, right: int) int {
    left + right
}

add(20, 22) # 42
```

Functions are expressions, so an anonymous function can be passed or called directly:

```silver
let apply = fn(operation: call(int) int, value: int) int {
    operation(value)
}

apply(fn(value: int) int { value * 2 }, 21) # 42
fn(value: int) int { value + 1 }(41)        # 42
```

Arguments are evaluated from left to right and normally bind by position. Every parameter must be filled and extra
arguments are rejected. Structs and modules can also fill several parameters through
[object destructuring](objects.md#object-destructuring).

## Parameters and runtime contracts

Parameter annotations are optional and checked when arguments bind:

```silver
let choose = fn(value: int, enabled) int {
    if enabled { value } else { 0 }
}
```

Here `value` must be an integer while `enabled` accepts any value. Silver has no `any` annotation; leave a parameter
unannotated when it should accept every value.

An annotation can name a primitive type, struct, enum, built-in nominal type, or a qualified type from a
[module](modules.md#module-values-and-qualified-types):

```silver
let paths = import("path")

let describe = fn(value: paths.Path) str {
    value.path
}
```

## Return behavior

An annotated function returns its final body expression implicitly. `return expression` leaves the function early, and
`return` by itself produces null:

```silver
let first_positive = fn(values: array) int {
    for value in values {
        if value > 0 {
            return value
        }
    }
    0
}
```

Silver checks the result against the return annotation. Use `fn() null` to declare an explicit null result.

An unannotated function always returns null. Its body still runs, and an expression supplied to `return` is evaluated,
but that value is discarded at the function boundary:

```silver
let announce = fn(message: str) {
    import("io").print(message)
    42
}

announce("ready") # null
```

Functions that may propagate typed application errors place their alternatives after the success type. See
[Errors](errors.md#declaring-error-contracts).

## Closures

A function captures its lexical environment. The captured bindings remain available after the declaring function has
returned:

```silver
let make_adder = fn(amount: int) call(int) int {
    fn(value: int) int { value + amount }
}

let add_two = make_adder(2)
add_two(40) # 42
```

Closures capture bindings rather than frozen copies. They observe later assignments and can update bindings in an
enclosing scope:

```silver
let count = 0
let next = fn() int {
    count = count + 1
    count
}

next() # 1
next() # 2
```

## Callable annotations

The primitive `call` type accepts any Silver function or native callable:

```silver
let callback: call = fn(value: int) int { value }
```

A detailed callable annotation constrains parameter types, its result, and any declared errors:

```text
call(int, str) bool
call(path: str) str | NotFound
```

Parameter names are optional. When a callable annotation includes names, both the names and types must match the
supplied callable. This matters for callable struct fields because named parameters define
[method receivers](objects.md#methods-are-callable-fields):

```silver
struct Counter {
    value: int
    increment: call(self: Counter, amount: int) int
}
```

Callable error contracts are compatible by possibility: a function that produces fewer error types can satisfy a
contract that permits more, but not the reverse. See [Callable error contracts](errors.md#callable-error-contracts).

## Functions and other language features

- [`return`](control_flow.md#returns-and-errors) can leave loops and nested conditional blocks inside a function.
- [`defer`](control_flow.md#deferred-calls) schedules cleanup for function exit, including error propagation.
- [Tasks](concurrency.md) accept zero-argument functions and retain their results or errors.
- [Template strings](template_strings.md) capture scope and reevaluate their interpolations when `.eval()` is called.

[Language guide](language_guide.md) | [Documentation index](../table_of_contents.md)
