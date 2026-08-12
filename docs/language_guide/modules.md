# Modules and Imports

Every Silver source file is a module. A module is an isolated namespace whose members are the top-level bindings
created while that file is evaluated.

## Importing modules

`import(expression)` evaluates its argument, requires a string, loads the corresponding module, and returns a value of
type `module`:

```silver
let io = import("io")
let helpers = import("./helpers.slv")

io.print(helpers.answer)
```

The path can be computed at runtime:

```silver
let module_path = "./helpers.slv"
let helpers = import(module_path)
```

A non-string path raises `TypeError`. A module that cannot be read or resolved raises `ImportError`.

## Path resolution

Silver resolves module names in this order:

| Import form | Resolution |
| ----------- | ---------- |
| Bundled standard-library name, such as `"io"` | Load the module embedded in the interpreter. |
| Absolute filesystem path | Load that exact path. |
| Relative or bare source path | Check beside the importing file first. |
| Unresolved source path | Check each directory in `SILVER_PATH` using the platform's path-list separator. |

Relative imports inside an imported module resolve beside that module, not beside the original entry file:

```text
project/
    main.slv
    lib/
        math.slv
        constants.slv
```

```silver
# lib/math.slv
let constants = import("./constants.slv")
```

Use an explicit relative path such as `./testing.slv` when a user file has the same name as a bundled module.

## Module members

Silver has no separate visibility declaration or modifier. Every top-level binding becomes a module member, including
`let` bindings and struct or enum definitions:

```silver
# geometry.slv
struct Point { x: int, y: int }

let origin = Point{0, 0}
let translate = fn(point: Point, amount: int) Point {
    return Point{point.x + amount, point.y + amount}
}
```

Importers access those bindings through member syntax:

```silver
let geometry = import("./geometry.slv")
let point: geometry.Point = geometry.translate(geometry.origin, 5)
```

Top-level bindings stay isolated from the importing scope. Importing `geometry.slv` does not introduce `Point`,
`origin`, or `translate` as unqualified names. Accessing an absent member raises `AttributeError`.

Module members cannot be replaced through member assignment. Member assignment is reserved for mutable struct fields:

```silver
# geometry.origin = geometry.Point{1, 1} # TypeError
```

## Evaluation, caching, and state

A successful module is evaluated once per interpreter session and cached by its bundled name or canonical absolute
path. Repeated imports return the same module object:

```silver
let first = import("testing")
let second = import("testing")
first == second # True
```

Consequently, module functions share the top-level bindings they captured during evaluation. This is useful for modules
such as the stateful [`testing`](../stdlib/testing.md) runner, but libraries should make shared mutation deliberate:

```silver
# counter.slv
let count = 0
let next = fn() int {
    count = count + 1
    return count
}
```

Every importer of `counter.slv` observes the same sequence through `next()`. Reassigning `count` updates the binding seen
by the module's closures; it does not replace the value already exposed as `counter.count`. Mutating an exposed array,
map, or struct is visible through every import because those importers hold the same value.

A module is not cached when its evaluation fails. An import cycle detected while modules are loading raises
`ImportError` rather than exposing a partially initialized namespace.

## Module values and qualified types

Use the primitive `module` annotation when a function expects an intact module value:

```silver
let announce = fn(library: module) {
    library.print("ready")
}

announce(import("io"))
```

Nominal definitions belonging to a module are named through the binding that holds the module:

```silver
let paths = import("path")
let current: paths.Path = paths.cwd()
```

The qualification matters: nominal types are identified by their exact definitions, not merely by their names or
fields.

## Module destructuring

Modules implement the same named-value protocol as structs. If a module argument does not satisfy the parameter at its
position, its members can fill remaining parameters with matching names:

```silver
let announce = fn(print: call) {
    print("ready")
}

announce(import("io"))
```

An argument that satisfies a `module` parameter remains intact and is not destructured. The complete binding algorithm
is described under [Object destructuring](objects.md#object-destructuring).

## Failures and diagnostics

Failures raised while evaluating a module propagate through `import`. Tracebacks retain the imported file's path and
include module and function frames. See [Errors and diagnostics](errors.md#runtime-diagnostics).

[Language guide](language_guide.md) | [Documentation index](../table_of_contents.md)
