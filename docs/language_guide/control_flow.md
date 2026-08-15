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

Additional conditions may be chained with `else if`:

```silver
let label = if score >= 90 {
    "excellent"
} else if score >= 70 {
    "passing"
} else {
    "keep going"
}
```

```silver
if ready {
    import("io").print("starting")
}
```

## Switch expressions

`switch` compares one value against ordered `case` expressions and may produce a value:

```silver
let label = switch value {
    case 1:
        "one"
    case 2:
        "two"
    case 3:
        "three"
    default:
        "something else"
}
```

Silver evaluates the switch value once. It then evaluates case expressions in source order as they are reached. Each
comparison has exactly the semantics of `switchValue == caseValue`: switch does not define another form of equality,
and a struct switch value therefore uses its existing `eq` operator method. Case expressions are not cached; evaluating
the same switch again reevaluates every case expression reached during that evaluation.

Only the first matching case body runs. There is no fallthrough and no switch-specific `break`; `break` remains loop
control and is not needed to leave a switch. Case values are ordinary expressions and cannot have type annotations.

The final `default` clause is optional. If no case matches and there is no default, the switch evaluates to null. As with
`if`, the selected body's result is the switch expression's result, and a switch can instead be used as a standalone
expression for side effects:

```silver
switch command {
    case "start":
        start()
    case "stop":
        stop()
    default:
        io.print("unknown command")
}
```

Type definitions are first-class values, so switches can dispatch on [`core.type`](../stdlib/core.md):

```silver
switch core.type(value) {
    case int:
        handle_integer(value)
    case str:
        handle_string(value)
    default:
        handle_other(value)
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

## Loop control

`continue` skips the rest of the current iteration. `break` exits the nearest enclosing loop:

```silver
let firstLargeOdd = 0
for value in [2, 7, 10, 13, 18] {
    if value % 2 == 0 { continue }
    if value > 10 {
        firstLargeOdd = value
        break
    }
}
```

Both statements work in `for` and `while` loops and apply only to the innermost loop when loops are nested. They are
only valid inside a loop body. A [`return`](functions.md#return-behavior) inside a function leaves both the loop and the
surrounding function.

## Returns and errors

Function return behavior is documented in [Functions](functions.md#return-behavior). Typed propagation, `try`/`catch`,
and assertions are documented in [Errors and diagnostics](errors.md).

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
