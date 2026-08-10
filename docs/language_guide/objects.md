# Objects

Silver is struct-centric: data lives in named fields, behavior can live in callable fields, and functions can
destructure structs or modules into named parameters. These rules provide methods, operator overloading, custom
indexing, and capability-style dependency injection without a separate class system.

## Declaring and constructing structs

A struct declaration defines a nominal type and an ordered list of fields:

```silver
struct Location {
    x: float
    y: float
    z: float
}

let origin = Location{0.0, 0.0, 0.0}
```

Construction supplies exactly one value per field in declaration order. Fields may omit annotations, but annotated
fields are checked at construction and on later assignment.

```silver
origin.x = 5.0
# origin.x = "east" # TypeError: Location.x expects float
```

Struct instances are mutable reference values. Assigning an instance to another binding creates an alias, not a copy:

```silver
let current = origin
current.y = 2.0
origin.y # 2.0
```

Struct types are nominal. Two declarations with identical field names are still different types.

## Object destructuring

Functions bind ordinary positional arguments first. When an argument does not satisfy the parameter at its position and
the argument is a struct or module, Silver offers its named fields or members to the remaining unbound parameters.

```silver
struct Location {
    x: float
    y: float
    z: float
}

let move = fn(x: float, y: float, z: float) Location {
    Location{x + 5.0, y + 5.0, z + 5.0}
}

let location = Location{0.0, 0.0, 0.0}
location = move(location)
```

`location` cannot satisfy `x: float`, so destructuring begins. The fields `x`, `y`, and `z` match the remaining
parameters by name, not by declaration order. Each extracted value must satisfy its parameter's annotation.

An argument that already satisfies its parameter is kept intact:

```silver
let move_by = fn(location: Location, amount: float) Location {
    Location{
        location.x + amount,
        location.y + amount,
        location.z + amount
    }
}

move_by(location, 2.0)
```

This distinction makes both whole-object and field-oriented APIs possible. An unannotated parameter accepts the whole
argument immediately, so annotate the field parameters when destructuring is intended.

Destructuring can combine with normal positional arguments:

```silver
struct Point { x: int, y: int }

let encode = fn(offset: int, x: int, y: int) int {
    offset + x * 10 + y
}

encode(100, Point{2, 3}) # 123
```

It can also combine several objects. Extra fields are ignored; missing parameters eventually produce an arity error. A
field with the right name but wrong type produces a type error for that parameter.

Modules implement the same named-value interface. Passing `import("io")` to a function with a `print` parameter injects
the matching module member:

```silver
let announce = fn(print: call) {
    print("ready")
}

announce(import("io"))
```

A parameter annotated `module` accepts the module intact and therefore does not destructure it.

## Methods are callable fields

A struct field with a detailed `call(...)` annotation is a method slot. When that field contains a Silver function,
member access binds the struct instance as the function's first argument:

```silver
struct Counter {
    value: int
    increment: call(self: Counter, amount: int) int
}

let increment = fn(self: Counter, amount: int) int {
    self.value = self.value + amount
    self.value
}

let counter = Counter{0, increment}
counter.increment(3) # equivalent to increment(counter, 3)
```

The callable field's complete signature is checked when the struct is constructed. Named callable parameters are part of
that contract, so `self` and `amount` above must match the stored function's parameter names as well as its types.

Each lookup binds the function to the instance it came from. A method taken from one instance does not accidentally
operate on another.

A field annotated with bare `call` is an ordinary callback. It accepts any callable and does not bind a receiver:

```silver
let identity = fn(value: int) int { value }

struct Box {
    callback: call
}

let box = Box{identity}
box.callback(42) # 42; Box is not inserted as an argument
```

Native standard-library objects use callable fields too. `File.close`, `Path.exists`, and `Connection.write` look like
methods because their structs carry callables that close over native state.

## Operator protocols

A struct opts into a binary operator by storing a callable in the corresponding field:

| Operator | Field     | Operator | Field    |
| -------- | --------- | -------- | -------- |
| `+`      | `add`     | `-`      | `sub`    |
| `*`      | `mul`     | `/`      | `div`    |
| `//`     | `int_div` | `**`     | `pow`    |
| `==`     | `eq`      | `!=`     | `not_eq` |
| `<`      | `lt`      | `>`      | `gt`     |
| `<=`     | `lte`     | `>=`     | `gte`    |

The left operand provides the method; the right operand becomes its explicit argument:

```silver
struct Vector {
    x: int
    y: int
    add: call(self: Vector, other: Vector) Vector
}

let add = fn(self: Vector, other: Vector) Vector {
    Vector{self.x + other.x, self.y + other.y, add}
}

let left = Vector{2, 3, add}
let right = Vector{5, 8, add}
let sum = left + right
```

The operator method may return any value allowed by its declared signature. If the required field is absent or not
callable, the operation raises `TypeError`. Struct truthiness does not invoke an operator method; every struct instance
is truthy.

## Indexing protocols

`get_item` and `set_item` callable fields let a struct support bracket syntax:

```silver
struct IntBuffer {
    values: array
    get_item: call(self: IntBuffer, index: int) int
    set_item: call(self: IntBuffer, index: int, value: int)
}

let get_item = fn(self: IntBuffer, index: int) int {
    self.values[index]
}

let set_item = fn(self: IntBuffer, index: int, value: int) {
    self.values[index] = value
}

let buffer = IntBuffer{[10, 20, 30], get_item, set_item}
buffer[1] = 99
buffer[1] # 99
```

`value[key]` calls `get_item(key)`. `value[key] = replacement` calls `set_item(key, replacement)`. Errors and
return-type checks from those functions propagate normally.

Native arrays and maps implement indexing directly. Array indexes must be in range. Missing map keys raise `KeyError`,
while [`map.get`](../stdlib/map.md) returns null for absence.

## Enums

Enums define nominal singleton values:

```silver
enum Direction {
    North,
    East,
    South,
    West,
}

let direction: Direction = Direction.North
```

Members may be separated by commas or newlines, and a trailing comma is accepted. Access members through the enum
namespace. Values compare by identity, carry the enum's nominal type, and can be used as map keys.

## Modules

Module namespaces, member access, qualified types, caching, and shared state are covered in
[Modules and imports](modules.md). Modules participate in the destructuring rules described above but do not support
member assignment.

[Documentation index](../table_of_contents.md)
