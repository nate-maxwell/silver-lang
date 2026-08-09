# `testing`

`testing` is a small stateful test runner written in Silver. It records assertion failures, continues with later tests,
and prints progress through `io.print`.

```silver
let testing = import("testing")

testing.suite("math", fn() {
    testing.run("addition", fn() {
        testing.equal(2 + 2, 4, "two plus two")
    })

    testing.run("truth", fn() {
        testing.is_true(10 > 3, "comparison should hold")
    })
})

testing.report()
```

## Assertions

All assertion helpers require a message and raise the built-in `AssertionError` on failure:

| Export                                 | Description                   |
| -------------------------------------- | ----------------------------- |
| `fail(message)`                        | Fail unconditionally.         |
| `check(condition, message)`            | Require a truthy condition.   |
| `is_true(value, message)`              | Require exact `True`.         |
| `is_false(value, message)`             | Require exact `False`.        |
| `equal(actual, expected, message)`     | Require `actual == expected`. |
| `not_equal(actual, expected, message)` | Require `actual != expected`. |

## Running and reporting

- `run(name, test: call())`: Run one zero-argument test, record `AssertionError`, and continue. Unexpected errors
  propagate.
- `suite(name, tests: call())`: Prefix contained test names; suites may nest.
- `results() array`: Return recorded `Result` values in execution order.
- `summary() Summary`: Return current total, passed, and failed counts.
- `report() bool`: Print totals and failure details; return whether all tests passed.
- `reset()`: Clear results, counters, and suite state.

`Result` has `name`, `passed`, and `message`. `Summary` has `total`, `passed`, and `failed`. Module state is shared
because standard-library imports are cached within an interpreter session; call `reset` when reusing the runner for an
independent batch.

[Standard library index](../table_of_contents.md#standard-library)
