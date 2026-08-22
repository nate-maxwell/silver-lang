# Errors and Diagnostics

Silver uses nominal struct values for both application errors and interpreter failures. Recoverable errors propagate as
values with declared types, while uncaught failures produce source-aware diagnostics and tracebacks.

## Error values

Any struct can be used as an application error type. A `message: str` field is conventional and is required by the
built-in error types, but application errors may carry any additional fields they need:

```silver
struct NotFound {
    message: str
    path: str
}

struct PermissionProblem {
    message: str
}
```

Errors are nominal. Two structs with identical fields remain different error types, and `catch` matches the exact
definition rather than the shape of its fields.

## Declaring error contracts

List possible application errors after a function's successful return type:

```silver
let read = fn(path: str) str | NotFound | PermissionProblem {
    return NotFound{"file does not exist", path}
}
```

When a function produces one of its declared error structs, that value unwinds the call instead of becoming an ordinary
successful result. Silver has no separate `throw` or `raise` statement; producing the declared struct is sufficient.
The same struct constructed outside such a return contract is an ordinary value.

If success is null, omit the success type and begin the contract with `|`:

```silver
let save = fn(allowed: bool) | PermissionProblem {
    if !allowed {
        return PermissionProblem{"permission denied"}
    }
    return
}
```

Error alternatives must resolve to struct definitions. A non-error success value must satisfy the successful return
type, and an application error escaping through a function must appear in that function's own contract:

```silver
let load = fn(path: str) str | NotFound | PermissionProblem {
    return read(path)
}
```

Because `read` can also produce `PermissionProblem`, `load` must catch it or include it as shown. Otherwise Silver
raises `RuntimeError` at the function boundary. Built-in runtime failures such as `TypeError` may propagate without
being listed in a return union.

## No `raise` or `throw`

Unlike other `try`/`catch` languages, Silver does not implement a `raise` or `throw` keyword.
Errors are instead returned as values and will be displayed as errors if not handled by a `try/catch` block.

## Handling errors with `try` and `catch`

`try` is an expression followed by one or more typed handlers:

```silver
let contents = try {
    read("settings.json")
} catch NotFound err {
    "{}"
} catch PermissionProblem err {
    "access denied: " + err.message
}
```

Catch clauses are tested in source order. The first exact nominal match runs, and its binding receives the complete
error struct with all of its fields. The selected block becomes the value of the whole `try` expression.

An unmatched error continues to an enclosing `try` or escapes the current function. There is no untyped catch-all
clause.

Built-in failures use the same mechanism:

```silver
let message = try {
    1 + True
} catch TypeError err {
    err.message
}
```

## Assertions

`assert` raises `AssertionError` when its condition is falsey:

```silver
assert amount >= 0
assert amount <= balance, "amount exceeds balance"
```

The optional expression after the comma supplies the error's `message`. Assertions are catchable, which lets the
[`testing`](../stdlib/testing.md) module record a failed check and continue running later tests.

## Built-in error types

Interpreter failures are instances of built-in nominal structs with a `message: str` property:

| Type | Typical cause |
| ---- | ------------- |
| `RuntimeError` | A runtime contract failure that does not fit a more specific category, including an undeclared application error escaping a function. |
| `AssertionError` | A falsey `assert` condition. |
| `TypeError` | An invalid operand, argument, assignment, call, or return type. |
| `ValueError` | A value of the right type that is invalid for an operation. |
| `ZeroDivisionError` | Division by numeric zero. |
| `NameError` | An unknown binding or type name. |
| `AttributeError` | A missing or invalid module/struct member. |
| `ImportError` | A missing, unreadable, or circular module import. |
| `SyntaxError` | Source that cannot be parsed. |
| `KeyError` | A missing map key read through bracket syntax. |
| `IndexError` | An out-of-range sequence index or empty sequence operation. |
| `TaskError` | Invalid task-handle use or collection state. |

Standard-library APIs add nominal errors such as `IOError`, `FileNotFound`, `PermissionDenied`, `ConnectionError`,
`ListenError`, `ReadError`, and `WriteError`. Individual modules may define qualified errors too, including
`json.JSONDecodeError` and `args.ArgumentError`. Their reference pages document the operations that produce them and any
properties beyond `message`.

An entry file's parse failure occurs before program evaluation and therefore cannot be caught inside that file. A
`SyntaxError` encountered while importing another file propagates from the `import` expression and can be handled by
surrounding code.

## Callable error contracts

Detailed `call` annotations may include error alternatives:

```silver
let opener: call(path: str) str | NotFound = read
```

Compatibility is based on what the callable may produce. A function with no application errors can satisfy a callable
contract that permits `NotFound`; a function that may produce `NotFound` cannot satisfy a contract that promises only a
string.

Parameter names in a detailed callable annotation are also part of the contract. See
[Callable annotations](functions.md#callable-annotations).

## Propagation through scopes and tasks

An error immediately leaves the current expression and unwinds calls until it reaches a matching `catch`. Deferred
calls still run while a scope exits; see [`defer`](control_flow.md#deferred-calls).

Tasks retain errors instead of raising them immediately. The error propagates when the handle is collected, so a
`try` expression should surround `collect` when recovery is required. See
[Errors are delayed until collection](concurrency.md#errors-are-delayed-until-collection).

## Runtime diagnostics

Parser and evaluator errors include source positions. An uncaught error from nested functions or imported modules
includes traceback frames with function names and locations. Running a file writes the diagnostic to standard error and
exits unsuccessfully.

Catch errors when the program can recover meaningfully. Let unexpected contract failures escape so their tracebacks
retain the complete path to the defect.

[Language guide](language_guide.md) | [Documentation index](../table_of_contents.md)
