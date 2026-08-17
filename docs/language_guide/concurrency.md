# Concurrency

Silver's `task` and `collect` expressions provide structured concurrency for zero-argument callables. Tasks start
immediately, retain their result or error, and are joined explicitly with `collect` or automatically when their
scope exits.

## Starting a task

Place `task` before an unparenthesized callable name or member:

```silver
let fetch_user = fn() str { return "Ada" }
let user = task fetch_user
```

The target is invoked with no arguments. `task` accepts:

- an identifier, such as `task fetch_user`;
- a member expression, such as `task service.fetch`;
- a zero-argument anonymous function, such as `task fn() int { 6 * 7 }`.

It does not accept a call expression or parenthesized call syntax:

```silver
# Invalid: the work would be called before task could start it.
task fetch_user()

# Invalid: task syntax is intentionally unparenthesized.
task(fetch_user)
```

Store the returned handle in a binding so it can be collected:

```silver
let calculation = task fn() int { return 6 * 7 }
```

## Collecting results

`collect` takes an unparenthesized, comma-separated list of task-handle identifiers:

```silver
let fetch_user = fn() str { return "Ada" }
let fetch_score = fn() int { return 100 }

let user = task fetch_user
let score = task fetch_score
let result = collect user, score

result.user  # "Ada"
result.score # 100
```

Collection waits until every named task completes. It returns an anonymous struct with one field per non-null result.
Field names come from the handle identifiers supplied to `collect`.

A task whose callable returns null is joined but omitted from the result struct:

```silver
let announce = fn() {
    import("io").print("ready")
}

let notification = task announce
let result = collect notification
# result has no notification field because announce returned null.
```

Like `task`, `collect` does not use parentheses:

```silver
let result = collect user, score # valid
# collect(user, score)           # invalid
```

## Errors are delayed until collection

A task retains both declared error values and ordinary runtime failures. The starting scope continues to run until the
handle is collected:

```silver
struct Missing { message: str }

let read = fn() str | Missing {
    return Missing{"not found"}
}

let pending = task read
import("io").print("task has started")

let outcome = try {
    collect pending
} catch Missing err {
    err.message
}
```

The error propagates from `collect`, so normal [`try`/`catch`](errors.md#handling-errors-with-try-and-catch) rules apply
at that point.
A runtime error such as `TypeError` is retained in the same way.

When several handles are collected together, collection joins all of them. If a collected task failed, its error is
returned rather than a partial result struct.

## Handles may be collected once

Task handles are affine values: one task result may be consumed only once. Aliasing a handle does not create another
result:

```silver
let work = fn() int { return 42 }
let original = task work
let alias = original

let first = collect original
# collect alias # parser error: the same task was already collected
```

The parser tracks direct task-handle bindings and aliases through branches, loops, functions, and `try` expressions. It
rejects a program when a known handle could be collected more than once.

## Scope exit and uncollected tasks

Every task belongs to the scope that starts it. When that function, module, or script exits, Silver joins any task that
was not explicitly collected. This prevents concurrent work from being silently abandoned.

An uncollected task produces a warning on the evaluator's warning stream:

```silver
let work = fn() {
    import("io").print("finished")
}

let abandoned = task work
# Scope exit still waits for "finished" and warns about abandoned.
```

Explicit collection is preferred because it gives the program a clear point for observing results and failures.

## Data shared between tasks

Tasks execute closures over Silver values. Arrays, maps, and struct instances are mutable reference values, and native
I/O/network objects may wrap shared host resources. Concurrent mutation is therefore observable. Prefer independent
values, immutable-by-convention data, or native objects whose API documents synchronization.

The task API currently provides joining, not cancellation, timeouts, priorities, or channels. Use standard-library
operations such as [`time.sleep`](../stdlib/time.md) and blocking networking calls with care: scope exit must still wait
for unfinished work.

## Complete example

```silver
let io = import("io")
let times = import("time")

let slow_answer = fn() int {
    times.sleep(times.duration(25, "ms"))
    return 40
}

let quick_answer = fn() int { return 2 }

let slow = task slow_answer
let quick = task quick_answer
let answers = collect slow, quick

io.println(answers.slow + answers.quick) # 42
```

[Documentation index](../table_of_contents.md)
