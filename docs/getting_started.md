# Getting Started

Silver is currently distributed from source. You need Git and Go 1.25.2 or newer.

## Build Silver

Clone the repository, run the tests, and build the interpreter:

```console
git clone https://github.com/nate-maxwell/silver-lang.git
cd silver-lang
go test ./...
go build -o silver .
```

On Windows the output is `silver.exe`; on macOS and Linux it is `silver`. You can also skip the build while developing
and replace `./silver` in the examples below with `go run .`.

The command accepts either no arguments or one source path:

```text
silver [file]
```

## Use the REPL

Start Silver without a file:

```console
./silver
```

Then evaluate expressions and statements one line at a time:

```silver
let io = import("io")
let square = fn(value: int) int { value * value }
io.print(square(9))
```

The REPL keeps bindings between entries. Send end-of-file to exit (`Ctrl+D` on macOS/Linux or `Ctrl+Z`, then Enter, on
Windows).

## Run a program

Create `hello.slv`:

```silver
let io = import("io")

struct Person {
    name: str
    greet: call(self: Person) str
}

let greet = fn(self: Person) str {
    "Hello, " + self.name + "!"
}

let person = Person{"Silver", greet}
io.print(person.greet())
```

Run it:

```console
./silver hello.slv
```

Silver source uses the `.slv` extension. Successful file execution produces only the output requested by the program;
uncaught errors and tracebacks go to standard error and produce a nonzero exit status.

Silver writes a sibling `.astc` cache after parsing a file. The cache is an implementation detail: it is validated
against the source and regenerated when missing, stale, or damaged. It does not need to be committed for ordinary user
programs.

## Split code into modules

Every source file is a module. Suppose `math_helpers.slv` contains:

```silver
let double = fn(value: int) int { value * 2 }
let answer = 42
```

Import it from a file in the same directory:

```silver
let helpers = import("./math_helpers.slv")
let io = import("io")

io.print(helpers.double(helpers.answer))
```

Relative imports resolve from the importing file. A non-relative file import first checks the importing file's directory
and then each directory in the platform-separated `SILVER_PATH` environment variable. Bare standard-library names such as
`"io"` and `"array"` resolve to embedded modules.

Imports are evaluated once per interpreter session and then cached. Circular imports are reported as errors.
See [Modules and imports](language_guide/modules.md) for the complete resolution and module-member rules.

## Learn the library

Standard-library modules are ordinary values returned by `import`:

```silver
let arrays = import("array")
let print = import("io").print

let values = arrays.sort([3, 1, 2])
print(values) # [1, 2, 3]
```

Continue with the [Language Guide](language_guide/language_guide.md) or browse the
[standard-library reference](table_of_contents.md#standard-library).

[Documentation index](table_of_contents.md) · [Project README](../README.md)
