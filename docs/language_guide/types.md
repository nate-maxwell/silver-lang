# Types and Values

Silver evaluates programs dynamically. Type annotations are optional contracts checked when bindings, parameters,
fields, and return values receive values. They document intent and fail close to the boundary without requiring a
separate compilation phase.

## Primitive values

| Type     | Examples                          | Notes                                                        |
| -------- | --------------------------------- | ------------------------------------------------------------ |
| `int`    | `0`, `-42`                        | Signed 64-bit integer.                                       |
| `float`  | `3.14`, `-0.5`                    | IEEE-754 64-bit floating point.                              |
| `bool`   | `True`, `False`                   | Only `False` and null are falsey.                            |
| `str`    | `"silver"`, `"line\n"`            | UTF-8 string.                                                |
| `null`   | produced by value-less operations | No null value literal; `null` is the first-class type value. |
| `array`  | `[1, 2, 3]`                       | Ordered, mutable, zero-indexed sequence.                     |
| `map`    | `{"name": "Ada", 1: True}`        | Mutable hash map.                                            |
| `call`   | `fn(value) { value }`             | Silver function or native callable.                          |
| `module` | `import("io")`                    | Imported module namespace.                                   |

Structs and enums introduce nominal types. TemplateString, task handles, built-in errors, and standard-library objects
add other runtime value kinds.

Silver currently has no `any` annotation. Leave a binding or parameter unannotated when it should accept any value.

## Numbers

Integer literals use base 10. Float literals contain a decimal point. Unary `-` negates either numeric type.

Numeric arithmetic and comparisons may mix integers and floats. `/` always produces a float; `//` produces the integer
quotient. Exponentiation uses `**`.

```silver
5 / 2   # 2.5
5 // 2  # 2
2 ** 10 # 1024
```

Division by numeric zero raises `ZeroDivisionError`. Integer arithmetic uses signed 64-bit values; ordinary arithmetic
overflow follows the implementation's wrapping integer behavior. Standard-library functions may add stricter range
checks where documented.

## Strings

Strings use double quotes and contain UTF-8 data:

```silver
let message = "snowman: \u2603\n"
```

Supported escapes include `\\`, `\"`, `\'`, `\/`, `\0`, `\a`, `\b`, `\f`, `\n`, `\r`, `\t`, `\v`, `\xNN`, `\uNNNN`, and
`\UNNNNNNNN`. Invalid or incomplete escapes are syntax errors.

Strings support concatenation with `+`, equality, and lexical ordering comparisons. They are immutable. The
[`string`](../stdlib/string.md) module provides Unicode-aware operations; note that `core.len(string)` counts UTF-8
bytes rather than code points.

Triple-backtick [template strings](language_guide.md#template-strings) are separate `TemplateString` values and become
ordinary strings only after `.eval()`.

## Arrays

Arrays may contain values of mixed types:

```silver
let values = [1, "two", True]
values[1] = "TWO"
```

Arrays are mutable reference values. Indexes are zero-based integers and must be in range; negative indexes are not
supported by native bracket access. An alias observes indexed mutations:

```silver
let alias = values
alias[0] = 99
values[0] # 99
```

The [`array`](../stdlib/array.md) module provides copy-producing transformations. The
[`collections`](../stdlib/collections.md) module provides intentionally mutable sequence operations and struct-backed
deques/stacks.

## Maps

Map literals contain key/value expressions:

```silver
let record = {
    "name": "Silver",
    "year": 2026
}

record["stable"] = False
```

Integers, floats, booleans, strings, and enum values are hashable keys. Numerically equal integer and integral-float
keys address the same entry. Reading a missing key with brackets raises `KeyError`; [`map.get`](../stdlib/map.md)
returns null instead.

Maps are mutable reference values. Iteration order and the order returned by `map.values` are unspecified.

## Bindings and annotations

`let` introduces a lexical binding:

```silver
let count = 0
let name: str = "Silver"
```

Assignment updates the nearest existing binding and does not declare a new one:

```silver
count = count + 1
```

An annotation remains attached to its binding, so later assignments are checked too:

```silver
let age: int = 36
# age = "unknown" # TypeError
```

Annotations can name primitive types, structs, enums, built-in nominal types, or qualified definitions exported by
modules:

```silver
let paths = import("path")
let working_directory: paths.Path = paths.cwd()
```

An unknown type name raises `NameError` when the contract is resolved.

## Function parameters and returns

Annotate parameters after their names and put the return type after `)`:

```silver
let distance = fn(x: float, y: float) float {
    import("math").sqrt(x * x + y * y)
}
```

Arguments are checked as they bind to parameters. Structs and modules may instead undergo named
[object destructuring](objects.md) when they do not satisfy the current parameter.

An annotated function returns its final expression implicitly or an explicit `return` value. Silver checks the result
against the declared return type. A function without a return annotation always returns null, regardless of a body
value:

```silver
let side_effect = fn() {
    42
}

# side_effect() produces null.
```

Use `fn() null` when an explicit null return contract is useful.

## Callable types

Bare `call` accepts any Silver function or native callable:

```silver
let callback: call = fn(value: int) int { value }
```

A detailed callable annotation constrains its parameters, result, and declared errors:

```silver
let apply = fn(operation: call(int) int, value: int) int {
    operation(value)
}
```

Parameter types may be unnamed, as in `call(int) int`, or named:

```text
call(value: int) int
```

When names are present, both names and types must match the supplied callable. This named form is what lets a callable
struct field declare and bind a [method receiver](objects.md).

Callable return contracts are useful for higher-order functions:

```silver
let make_adder = fn(amount: int) call(int) int {
    fn(value: int) int { value + amount }
}
```

## Typed error contracts

Any struct type can be an expected error alternative. List alternatives after a function's successful return type:

```silver
struct NotFound { message: str, path: str }
struct PermissionProblem { message: str }

let read = fn(path: str) str | NotFound | PermissionProblem {
    NotFound{"file does not exist", path}
}
```

Producing a declared error struct unwinds the function instead of returning it as an ordinary success value. The caller
can handle it with [`try`/`catch`](control_flow.md#try-and-catch).

If success is null, omit the success type and begin with `|`:

```silver
let save = fn() | PermissionProblem {
    return
}
```

An error that escapes a function must be listed in that function's own contract. Error alternatives must resolve to
struct definitions. A callable value with fewer possible errors satisfies a callable contract that permits more; a
callable that may produce an undeclared error does not.

Silver's interpreter failures are built-in error structs with `message: str`:

- `RuntimeError`
- `AssertionError`
- `TypeError`
- `ValueError`
- `ZeroDivisionError`
- `NameError`
- `AttributeError`
- `ImportError`
- `SyntaxError`
- `KeyError`
- `IndexError`
- `TaskError`

I/O and networking add types such as `IOError`, `FileNotFound`, `PermissionDenied`, `ConnectionError`, `ListenError`,
`ReadError`, and `WriteError`.

## Nominal types

Struct and enum declarations create distinct nominal definitions:

```silver
struct UserId { value: int }
struct ProductId { value: int }

let user: UserId = UserId{7}
# user = ProductId{7} # TypeError despite the identical field layout
```

Standard-library types such as `path.Path`, `time.Time`, and `collections.Deque` are nominal too. Values must originate
from the exact definition named by the annotation.

## First-class types

Primitive type names and nominal definitions are runtime values. `core.type(value)` returns the applicable definition:

```silver
let core = import("core")

core.type(42) == int
core.type("hello") == str

struct User { name: str }
core.type(User{"Ada"}) == User
```

Calling `core.type` on a struct or enum definition returns that same definition. Imported modules report the `module`
type, and callables report `call`.

[Documentation index](../table_of_contents.md)
