# Silver Language Guide

This guide describes the language implemented by the current interpreter. Silver is dynamically executed: type
annotations are optional runtime contracts, not a separate compile-time phase.

## Table of contents

- [Types and values](types.md) - primitives, collections, annotations, nominal types, and first-class types.
- [Functions](functions.md) - declarations, calls, parameters, returns, closures, and callable annotations.
- [Control flow](control_flow.md) - truthiness, conditionals, loops, and deferred calls.
- [Errors and diagnostics](errors.md) - typed contracts, `try`/`catch`, assertions, built-in errors, and tracebacks.
- [Objects](objects.md) - structs, destructuring, methods, operator and indexing protocols, and enums.
- [Modules and imports](modules.md) - resolution, members, caching, shared state, and qualified types.
- [Concurrency](concurrency.md) - starting tasks, collecting results, error propagation, handle ownership, and scope
  exit.
- [Template strings](template_strings.md) - delayed interpolation, captured scope, rendering, and literal braces.

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

## Where next?

Browse the [standard-library reference](../table_of_contents.md#standard-library) or return to
[Getting Started](../getting_started.md).

[Documentation index](../table_of_contents.md) | [Project README](../../README.md)
