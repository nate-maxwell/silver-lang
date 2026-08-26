# User-defined infix operators

An `operator` declaration gives a symbolic infix spelling a two-argument function. The left expression is passed as
the first argument and the right expression as the second:

```silver
operator |> fn(left, right: call) {
    return right(left)
}

let double = fn(value: int) int { value * 2 }
let increment = fn(value: int) int { value + 1 }
let result = 20 |> double |> increment
```

Operator functions must have exactly two non-variadic parameters. Their parameter types, explicit return types, and
error alternatives use the ordinary function syntax. Unlike an ordinary unannotated function, an unannotated operator
function preserves its returned value, allowing the concise pipe declaration above.

Operator declarations are read in source order. Using a symbol before its declaration is a syntax error. Once declared,
the symbol belongs to the interpreter session rather than a lexical scope or module export list, so sources parsed later
can use it. A declaration executed inside a function or block becomes globally callable afterward. Declaring the same
symbol more than once is an error; operators already supplied by the language cannot be redefined either.

## Binding power

An operator uses addition's binding power, 60, by default and associates to the left. Put an integer between the symbol
and `fn` to customize it:

```silver
operator @ 55 fn(left, right) any {
    # ...
}
```

The binding power must be greater than 10. Built-in powers are spaced by ten so custom operators can fit between them:

| Power | Syntax               |
|------:|----------------------|
|    20 | `\|\|`               |
|    30 | `&&`                 |
|    40 | `==`, `!=`           |
|    50 | `<`, `>`, `<=`, `>=` |
|    60 | `+`, `-`             |
|    70 | `*`, `/`, `//`, `%`  |
|    80 | prefix `-`, `!`      |
|    90 | `**`                 |
|   100 | calls                |
|   110 | indexing             |
|   120 | member access        |

Symbols may contain the punctuation characters `!$%&*+-./:;<=>?@^|~`. Existing delimiter spellings such as `::` can
therefore be declared for ordinary expressions without affecting their established declaration syntax.

## Examples

### Pipeline Operator
```silver
operator |> 55 fn(left, right: call) any {
    return right(left)
}

let first_function = fn() str {
    return "hello"
}

let second_function = fn(s: str) str {
    return s + ", world"
}

let third_function = fn(s: str) str {
    return s + "!"
}

io.println( first_function() |> second_function |> third_function )
```
```
>> hello, world!
```

### Boilerplate Operators

#### Appending
```silver
operator <> 15 fn(left: array, right: any) array {
    return arrays.append(left, right)
}

let foo = []
io.println( foo <> "new" )
```
```
>> [new]
```

#### Range
```silver
operator .. 60 fn(left: int, right: int) array {
    return core.range(left, right)
}

for i in 2..10 {
    io.print(i)
}
```
```
>> 23456789
```

#### null Coalescing
Think `let name = user.name ?? "anonymous"`
```silver
operator ?? fn(left: any, right: any) any {
    if core.type(left) == null {
        return right
    }
    return left
}

io.println(null ?? 5)
```
```
>> 5
```

### Destructuring Utilization
```silver
struct Vector {
    x: float
    y: float
}

struct Outer {
    v: Vector
}

operator $ fn(v: Vector, right: int) {
    v.x = v.x + right
    v.y = v.y + right
}

let v = Vector{1.0, 1.0}
let o = Outer{v}

o $ 5
io.print(v)
```
```
>> Vector{x: 6.0, y: 6.0}
```
