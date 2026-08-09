# Silver Language Guide

This guide describes the language implemented by the current interpreter. Silver is dynamically executed: type
annotations are optional runtime contracts, not a separate compile-time phase.

## Table of contents

- [Types and values](types.md) - primitives, collections, annotations, callable signatures, typed errors, and
  first-class types.
- [Control flow](control_flow.md) - truthiness, conditionals, loops, returns, assertions, error handling, and deferred
  calls.
- [Objects](objects.md) - structs, destructuring, methods, operator and indexing protocols, enums, and modules.
- [Concurrency](concurrency.md) - starting tasks, collecting results, error propagation, handle ownership, and scope
  exit.

## Source files and statements

Silver source files conventionally use the `.slv` extension. Source is UTF-8, and `#` begins a line comment:

```silver
# Comments continue to the end of the line.
let answer = 6 * 7
```

Physical newlines separate statements. Silver rejects semicolons and adjacent statements on one line. Expressions can
span lines inside calls, arrays, maps, struct literals, and other delimited forms.

Identifiers begin with a letter or underscore and may contain letters, digits, and underscores. Keywords and names are
case-sensitive; the boolean values are spelled `True` and `False`.

## Functions and closures

Functions are expressions introduced by `fn`. They capture their lexical environment:

```silver
let make_adder = fn(amount: int) call(int) int {
    fn(value: int) int { value + amount }
}

let add_two = make_adder(2)
add_two(40) # 42
```

Parameters and return values may be annotated:

```text
fn(name: Type, ...) ReturnType { ... }
```

For an annotated function, the final body expression is its result unless `return` runs first. A function without a
return annotation always returns null, even if its body evaluates a value or executes `return value`. Use `return`
without a value for an explicit null result.

Callable types can state the expected parameter and return types:

```silver
let apply = fn(operation: call(value: int) int, value: int) int {
    operation(value)
}
```

Named parameters inside `call(...)` are part of the callable contract. Bare `call` accepts any callable without checking
a detailed signature. See [Types and values](types.md#callable-types) for the full rules and
[Objects](objects.md#methods-are-callable-fields) for method binding.

## Modules and imports

Every source file is a module. `import(expression)` evaluates a string and returns the corresponding module value:

```silver
let io = import("io")
let helpers = import("./helpers.slv")

io.print(helpers.answer)
```

Bare standard-library names resolve to modules embedded in the interpreter. Relative paths resolve beside the importing
file. Other source paths check the importing directory and then each directory in the platform-separated `SILVERPATH`
environment variable.

Imports are evaluated once per interpreter session and cached. Nested relative imports resolve beside their own source
module, and circular imports fail with `ImportError`. Module bindings do not leak into the importing scope; access them
through member syntax.

The import path can be computed, but must evaluate to a string:

```silver
let module_path = "./helpers.slv"
let helpers = import(module_path)
```

Modules also participate in argument destructuring. When a module value does not satisfy the parameter at its position,
its exports can fill parameters with matching names. See [Object destructuring](objects.md#object-destructuring).

## Template strings

Triple-backtick literals create a `TemplateString`, not a `str`. Interpolations are ordinary Silver expressions, but
evaluation is delayed until `.eval()`:

````silver
let table = "users"
let minimum = 21
let query = ```SELECT * FROM {table} WHERE age >= {minimum}```

minimum = 25
let rendered: str = query.eval()
````

A template captures its lexical scope, observes the current values of captured bindings, and reevaluates its expressions
on every `.eval()`. Interpolation failures are therefore delayed too. Use `{{` and `}}` for literal braces.

````silver
let count = 0
let next = fn() int {
    count = count + 1
    count
}

let value = ```call {next()}```
value.eval() # "call 1"
value.eval() # "call 2"
````

`TemplateString` is a nominal built-in struct type whose `eval` field has the signature `call() str`.

## Runtime diagnostics

Parser and evaluator errors include source positions. Uncaught errors from nested functions and imported modules include
traceback frames with function names and locations. A file invocation exits unsuccessfully after writing the diagnostic
to standard error.

Runtime failures use nominal error structs, so programs may handle them with the same typed error mechanism as
application errors. Continue with [typed errors](types.md#typed-error-contracts) and
[`try`/`catch`](control_flow.md#try-and-catch).

## Where next?

Read the chapters in any order:

- [Types and values](types.md)
- [Control flow](control_flow.md)
- [Objects](objects.md)
- [Concurrency](concurrency.md)

Then browse the [standard-library reference](../table_of_contents.md#standard-library) or return to
[Getting Started](../getting_started.md).

[Documentation index](../table_of_contents.md) | [Project README](../../README.md)
