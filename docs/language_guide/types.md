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
quotient; and `%` produces the remainder. Exponentiation uses `**`.

```silver
5 / 2   # 2.5
5 // 2  # 2
5 % 2   # 1
2 ** 10 # 1024
```

Division or remainder by numeric zero raises `ZeroDivisionError`. Integer arithmetic uses signed 64-bit values;
ordinary arithmetic overflow follows the implementation's wrapping integer behavior. Standard-library functions may
add stricter range checks where documented.

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

Triple-backtick [template strings](template_strings.md) are separate `TemplateString` values and become
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

Annotations can name primitive types, structs, enums, built-in nominal types, or qualified definitions belonging to
modules:

```silver
let paths = import("path")
let working_directory: paths.Path = paths.cwd()
```

An unknown type name raises `NameError` when the contract is resolved.

## Functions and errors

Parameter, return, and detailed `call(...)` contracts are covered in [Functions](functions.md). Error alternatives and
the nominal structs used by `try`/`catch` are covered in [Errors and diagnostics](errors.md).

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
