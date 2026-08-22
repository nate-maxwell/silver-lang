<p align="center">"I hate all languages equally" - Nate Maxwell</p>

# Silver

Silver is an interpreted, struct-centric programming language where behavior is data. A struct describes both
the values an object carries and the callable fields that give it behavior. Methods, operator overloads, indexing,
dependency injection, and even standard-library objects all grow from that small idea.

Silver exists to explore a middle ground between scripting-language immediacy and explicit program contracts:

- write a script and run it immediately;
- add runtime-checked types where explicitness is required;
- model expected failures in function signatures and catch them by nominal type;
- compose behavior from ordinary functions stored in ordinary structs;
- use concurrency without losing track of task results.

The implementation is written in Go and includes a REPL, source modules, cached ASTs, tracebacks, and a standard
library implemented in both Go and Silver. Silver is currently a young language: it is a good place to experiment,
learn, and contribute, but its syntax and APIs may still evolve.

Read the [documentation table of contents](docs/table_of_contents.md), start with [Getting Started](docs/getting_started.md), or dive into the
[Language Guide](docs/language_guide/language_guide.md).


## Basic syntax

Import a module, create values, and use familiar control flow:

```silver
let io = import("io")
let core = import("core")

let total = 0
for number in core.range(1, 6) {
    total = total + number
}

io.println("1 + 2 + 3 + 4 + 5 =", total)
```

Types are optional and checked when the program runs:

```silver
let apply = fn(operation: call(int) int, value: int) int {
    operation(value)
}

let double = fn(value: int) int { value * 2 }
apply(double, 21) # 42
```

## Behavior is data

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
print(location.x, location.y, location.z)

>> 5.0 5.0 5.0
```

Because `move` has a detailed `call(self: MovableLocation)` field contract, reading `location.move` binds `location`
as its first argument. Nothing special was declared outside the struct: the method is an ordinary function stored
in an ordinary field. Silver builds operator overloading and custom indexing on the same foundation.

## Errors are part of the signature

Any struct can be an expected error type. Returning one of a function's declared error alternatives unwinds to
a matching `catch`:

```silver
struct MissingUser { message: str, id: int }

let find_user = fn(id: int) str | MissingUser {
    if id == 7 {
        return "Ada"
    } else {
        return MissingUser{"user not found", id}
    }
}

let result = try {
    find_user(42)
} catch MissingUser err {
    err.message
}
```

Runtime failures use the same path, so errors such as `TypeError`, `KeyError`, and `AssertionError` are catchable too.

## Lazy templates and structured concurrency

Triple-backtick templates capture their lexical scope and evaluate only when asked:

````silver
let name = "Silver"
let greeting: TemplateString = ```Hello, {name}!```
name = "world"
greeting.eval() # "Hello, world!"
````

Tasks run zero-argument callables concurrently. `collect` joins them and names each non-null result after its
handle:

```silver
let answer = fn() int { return 6 * 7 }
let greeting = fn() str { return "hello" }

let calculation = task answer
let message = task greeting
let results = collect calculation, message

results.calculation # 42
results.message     # "hello"
```

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
the command prints the filepath when it changes the file.

## Project status and contributing

The test suite is the executable specification for the language today:

```console
go test ./...
```

Changes to syntax or semantics should include parser/evaluator tests and matching documentation. See the
repository's [MIT license](LICENSE).
