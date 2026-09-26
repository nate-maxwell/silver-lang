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

Initialize a `package.yaml` manifest in the current directory with a package name and empty author, membership, and export lists:

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
let io = import("core:io")
let square = fn(value: int) int { return value * value }
io.println(square(9))
```

The REPL keeps bindings between entries. Send end-of-file to exit (`Ctrl+D` on macOS/Linux or `Ctrl+Z`, then Enter, on
Windows).

## Run a program

Create `hello.slv`:

```silver
let io = import("core:io")

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
authors: []
export:
  - ./math_helpers.slv
```

Then register the manifest and import the module. Run this example from the directory containing `package.yaml`:

```silver
let system = import("core:system")
system.append_path("./package.yaml")
let helpers = import("math_example:math_helpers")
let io = import("core:io")

io.println(helpers.double(helpers.answer))
```

Imports use `<package>:<module>`, with `/` for nested modules and no `.slv` extension.
Register package YAML paths in `SILVER_PATH` or add them at runtime with `system.append_path`.
Only `export` files are public modules; the optional `members` list adds internal files that share
the package's operators. Standard-library imports use the `core` package, such as `core:io`
and `core:arrays`, and need no path registration.

Imports are evaluated once per interpreter session and then cached. Circular imports
are reported as errors.
See [Modules and imports](language_guide/modules.md) for the complete resolution
and module-member rules.

## Learn the library

Standard-library modules are ordinary values returned by `import`:

```silver
let arrays = import("core:arrays")
let print = import("core:io").print

let values = arrays.sort([3, 1, 2])
print(values) # [1, 2, 3]
```

Continue with the [Language Guide](language_guide/language_guide.md) or browse the
[standard-library reference](table_of_contents.md#standard-library).

[Documentation index](table_of_contents.md) · [Project README](../README.md)
