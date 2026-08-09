# Control Flow

Silver uses expressions for branching and typed propagation for recoverable errors. Blocks are brace-delimited, while
physical newlines separate their statements.

## Truthiness and logical operators

Only `False` and the null value are falsey. Everything else is truthy, including `0`, `0.0`, `""`, `[]`, and `{}`.

`!value` returns the logical negation of a value's truthiness. `&&` and `||` short-circuit and return a boolean:

```silver
let ready = True
let fallback = False

ready && !fallback # True
```

With `left && right`, Silver skips `right` when `left` is falsey. With `left || right`, it skips `right` when `left` is
truthy.

## Conditional expressions

`if` is an expression and may produce a value. Parentheses around its condition are optional:

```silver
let score = 92
let label = if score >= 90 {
    "excellent"
} else {
    "keep going"
}
```

Only the selected branch runs. When an `if` expression has no `else` and its condition is falsey, its value is null.

```silver
if ready {
    import("io").print("starting")
}
```

## `for` loops

Iterate an array with one binding:

```silver
let total = 0
for value in [10, 20, 30] {
    total = total + value
}
```

Iterate a map with key and value bindings:

```silver
let total = 0
let entries = {"first": 10, "second": 20}
for key, value in entries {
    total = total + value
}
```

Map iteration order is unspecified. The iterable is evaluated before iteration, and its shape is checked at runtime.
Array iteration requires one loop variable; map iteration requires two.

Loop variables live in the loop's execution scope. Assignment inside a loop can update bindings from an enclosing scope,
as the `total` examples do.

## `while` loops

`while` evaluates its condition before every iteration:

```silver
let count = 3
while count > 0 {
    import("io").print(count)
    count = count - 1
}
```

Silver currently has no `break` or `continue` statement. A `return` inside a function can leave a loop and the
surrounding function.

## Return behavior

`return expression` immediately leaves the current function. `return` by itself produces null:

```silver
let find_positive = fn(values: array) int {
    for value in values {
        if value > 0 {
            return value
        }
    }
    return 0
}
```

An annotated function also returns its final expression implicitly. An unannotated function always produces null; any
explicit return expression is evaluated but its value is discarded at the function boundary.

## Assertions

`assert` raises the built-in `AssertionError` when its condition is falsey:

```silver
assert amount >= 0
assert amount <= balance, "amount exceeds balance"
```

The optional expression after the comma supplies the error's `message`. Assertions are ordinary catchable runtime
errors, which lets the [`testing`](../stdlib/testing.md) module record failed checks without terminating the entire test
run.

## `try` and `catch`

`try` is a value-producing expression followed by one or more typed handlers:

```silver
struct NotFound { message: str }
struct Denied { message: str }

let read = fn(path: str) str | NotFound | Denied {
    NotFound{"missing: " + path}
}

let contents = try {
    read("settings.json")
} catch NotFound err {
    "{}"
} catch Denied err {
    "access denied: " + err.message
}
```

Catch clauses are tested in source order and match by exact nominal struct type. The declared binding receives the
caught struct, including all its fields. An unmatched error continues to an enclosing `try` or escapes the current
function.

Built-in runtime failures use this same mechanism:

```silver
let message = try {
    1 + True
} catch TypeError err {
    err.message
}
```

See [Typed error contracts](types.md#typed-error-contracts) for declaring the errors a function may propagate.

## Deferred calls

`defer` schedules a call for the end of the surrounding function, module, or script:

```silver
let file = import("io").open("data.txt")
defer file.close()

let contents = file.read()
```

Deferred calls run in last-in-first-out order. Silver captures the callable and evaluates its arguments immediately when
it reaches `defer`; it does not reevaluate the argument expressions during scope exit.

```silver
let io = import("io")

let show = fn(message: str) {
    io.print(message)
}

let example = fn() {
    let message = "first value"
    defer show(message)
    message = "changed later"
}

example() # prints "first value"
```

Only a call expression may follow `defer`. Deferred calls still run when a return or propagated error leaves the scope.

[Documentation index](../table_of_contents.md)
