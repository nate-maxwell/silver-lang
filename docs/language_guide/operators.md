# User-defined infix operators

An `operator` declaration gives a symbolic infix spelling a two-argument function. The left expression is passed as
the first argument and the right expression as the second:

```silver
operator |> = fn(left, right: call) {
    return right(left)
}

let double = fn(value: int) int { return value * 2 }
let increment = fn(value: int) int { return value + 1 }
let result = 20 |> double |> increment
```

Operator functions must have exactly two non-variadic parameters. Their parameter types, explicit return types, and
error alternatives use the ordinary function syntax. Unlike an ordinary unannotated function, an unannotated operator
function preserves its returned value, allowing the concise pipe declaration above.

In standalone files, operator declarations are read in source order. Using a symbol before its declaration is a syntax
error. Once declared, the symbol belongs to the standalone interpreter session, so standalone sources parsed later can
use it. In a [manifest package](modules.md#packages-and-yaml-manifests), declarations are discovered across all package
members before parsing. Internal members share this grammar even when they are omitted from the manifest's `export`
list. Other packages have separate operator scopes.

A declaration's runtime definition becomes callable when that declaration executes, including inside a function or
block. Declaring the same symbol more than once in one operator scope is an error; operators already supplied by the
language cannot be redefined either.

## Precedence

Every user-defined operator has the same precedence: lower than every built-in operator.
User-defined operator precedence cannot be customized.

Symbols may contain the punctuation characters `!$%&*+-./:;<=>?@^|~`. Existing delimiter spellings such as `::` can
therefore be declared for ordinary expressions without affecting their established declaration syntax.

## Examples

### Pipeline Operator
```silver
operator |> = fn(left, right: call) any {
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
operator <> = fn(left: array, right: any) array {
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
operator .. = fn(left: int, right: int) array {
    return core.range(left, right)
}

io.print(2..10)
```
```
>> [2, 3, 4, 5, 6, 7, 8, 9]
```

#### null Coalescing
Think `let name = user.name ?? "anonymous"`
```silver
operator ?? = fn(left: any, right: any) any {
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
type Vector = struct {
    x: float
    y: float
}

type Outer = struct {
    v: Vector
}

operator $ = fn(v: Vector, right: int) {
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
