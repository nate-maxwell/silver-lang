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

Running Silver without arguments starts the REPL; passing one source path runs that file:

```text
silver [file]
```

Silver also provides package initialization and version commands. See the
[command-line interface reference](cli.md) for the complete list.

Initialize a `package.yaml` manifest in the current directory with a package name and empty membership and export lists:

```text
silver package init <package_name>
```

## Use the REPL

Start Silver without a file:

```console
./silver
```

Then evaluate expressions and statements one line at a time:

```silver
let io = import("io")
let square = fn(value: int) int { return value * value }
io.println(square(9))
```

The REPL keeps bindings between entries. Send end-of-file to exit (`Ctrl+D` on macOS/Linux or `Ctrl+Z`, then Enter, on
Windows).

## Run a program

Create `hello.slv`:

```silver
let io = import("io")

type Person = struct {
    name: str
    greet: call(self: Person) str
}

let greet = fn(self: Person) str {
    return "Hello, " + self.name + "!"
}

let person = Person{"Silver", greet}
io.println(person.greet())
```

Run it:

```console
./silver hello.slv
```

Silver source uses the `.slv` extension. Successful file execution produces only the output requested by the program;
uncaught errors and tracebacks go to standard error and produce a nonzero exit status.

## Split code into modules

Every source file is a module. Suppose `math_helpers.slv` contains:

```silver
let double = fn(value: int) int { return value * 2 }
let answer = 42
```

Create `package.yaml` in the same directory to declare the imported module:

```yaml
package: math_example
members:
  - ./math_helpers.slv
export: []
```

Then import it from a file in that directory:

```silver
let helpers = import("./math_helpers.slv")
let io = import("io")

io.println(helpers.double(helpers.answer))
```

Relative imports resolve from the importing file. A non-relative file import first
checks the importing file's directory and then package exports in the platform-separated
`SILVER_PATH` environment variable. Every imported file requires a package YAML
manifest, including relative and absolute imports. Only `export` files are discoverable
through `SILVER_PATH`. The optional `members` list adds
internal files that share the package's operators. Bare standard-library names such
as `"io"` and `"arrays"` resolve to embedded modules.

Imports are evaluated once per interpreter session and then cached. Circular imports
are reported as errors.
See [Modules and imports](language_guide/modules.md) for the complete resolution
and module-member rules.

## Learn the library

Standard-library modules are ordinary values returned by `import`:

```silver
let arrays = import("arrays")
let print = import("io").print

let values = arrays.sort([3, 1, 2])
print(values) # [1, 2, 3]
```

Continue with the [Language Guide](language_guide/language_guide.md) or browse the
[standard-library reference](table_of_contents.md#standard-library).

[Documentation index](table_of_contents.md) · [Project README](../README.md)
