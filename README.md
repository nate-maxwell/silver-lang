### <div align="center"> "I hate all languages equally." </div>

# Silver

Silver is an interpreted, struct-centric programming language where behavior is data. A struct describes both
the values an object carries and the callable fields that give it behavior. Methods, operator overloads, indexing,
dependency injection, and even standard-library objects all grow from that small idea.

Silver exists to maximize code reuse without the drawbacks of inheritance, while minimizing boilerplate:

- ordinary structs have ordinary functions added to fields as struct methods;
- structs offer their fields when passed to functions that expect different types;
- composed structs can "embed" themselvevs, elevating their members to the outer struct namespace;
- users can define custom operators that are accessible package-wide;
- structs can overload builtin and custom operators.

The implementation is written in Go and includes a REPL, source modules, cached ASTs, tracebacks, and a standard
library implemented in both Go and Silver. Silver is currently a young language: it is a good place to experiment and learn, but its syntax and APIs may still evolve.

Read the [documentation table of contents](docs/table_of_contents.md), start with
[Getting Started](docs/getting_started.md), browse the [CLI reference](docs/cli.md), or dive into the
[Language Guide](docs/language_guide/language_guide.md).

## Behavior Is Data

Silver functions can destructure structs by parameter name. Here, `move` declares three float parameters, but
receives one `Location`:

```silver
struct Location {
    x: float
    y: float
    z: float
}

let move = fn(x: float, y: float, z: float) Location {
    return Location{ x + 5.0, y + 5.0, z + 5.0 }
}

let location = Location{ 0.0, 0.0, 0.0 }
location = move(location) # the actor moves diagonally by 5 units
```

`Location` is not a `float`, so it cannot bind directly to `x`. Silver instead offers its fields to the function's
unbound parameters, matching `x`, `y`, and `z` by name and checking each field's type. The one argument therefore
supplies all three parameters. If a parameter expected a `Location`, Silver would pass the value intact instead.

This destructuring mechanism allows silver code to be highly reusable. Functions can operator on multiple structs
without needing to reference their exact type or shape.

Struct methods extend this idea: functions can be data too. Add a callable field and store the behavior beside
the values it acts on:

```silver
let print = import("io").print

struct MovableLocation {
    x: float
    y: float
    z: float
    move: call(self: MovableLocation)
}

let move = fn(self: MovableLocation) {
    self.x = self.x + 5
    self.y = self.y + 5
    self.z = self.z + 5
}

let location = MovableLocation{0.0, 0.0, 0.0, move}
location.move()
io.print(location.x, location.y, location.z)
```
```
>> 5.0 5.0 5.0
```

Because `move` has a detailed `call(self: MovableLocation)` field contract, reading `location.move` binds `location`
as its first argument. Nothing special was declared outside the struct: the method is an ordinary function stored
in an ordinary field. Silver builds operator overloading and custom indexing on the same foundation and has a multitude
of ways structs can utilize callable fields.

## Struct Embedding

Struts can "embed" themsleves in other structs, raising their members to the outer struct's namespace.

```
struct Location {
    x: float
    y: float
}

struct Actor {
    name: str
    location:: Location
}

let foo = Location { 1.0, 2.0 }
let bar = Actor { "Ada", foo }

io.print(bar.x, bar.y)
```
```
>> 1.0 2.0
```

## Custom Operators

Silver allows users to create their own infix operators (operators with a left and
right expression, like `+`).

```
operator .. 65 fn(left: int, right: int) array {
    return core.range(left, right)
}

for i in 0..100 { io.println(i) }
```
```silver
operator ?? fn(left: any, right: any) any {
    if core.type(left) == null {
        return right
    }
    return left
}

let name = user.name ?? "anonymous"
```

## Operator Overloading

Structs overload an operator by declaring a callable field named after that operator. The left operand becomes
`self`, and the right operand is passed as the method's argument:

```silver
struct Vector {
    x: int
    y: int
    +: call(self: Vector, other: Vector) Vector
}

let add = fn(self: Vector, other: Vector) Vector {
    return Vector{self.x + other.x, self.y + other.y, add}
}

let first = Vector{2, 3, add}
let second = Vector{5, 8, add}
let sum = first + second
```
```
>> Vector{7, 11, add}
```

The same mechanism works for both built-in and custom infix operators.

Structs overload operators defined in their package of origin. This prevents dependent
packages from consuming all possible operator spellings.

## Get started

Silver currently builds from source and requires Go 1.25.2 or newer:

```console
git clone https://github.com/nate-maxwell/silver-lang.git
cd silver-lang
go build -o silver .
./silver
```

Run a source file with `./silver program.slv` (or `silver.exe program.slv` on Windows). During development,
`go run . program.slv` works too. Format one source file in place with `./silver frmt program.slv`; like `go fmt`,
the command prints the filepath when it changes the file. Initialize a package manifest in the current directory with
`./silver package init <package_name>`.

## Project status and contributing

The test suite is the executable specification for the language today:

```console
go test ./...
```

Changes to syntax or semantics should include parser/evaluator tests and matching documentation. See the
repository's [MIT license](LICENSE).
