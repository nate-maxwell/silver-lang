# Command-line interface

The examples below use `silver` as if the executable is on `PATH`. When running a locally built executable, use
`./silver` on macOS and Linux or `silver.exe` on Windows.

```text
silver [file]
silver astgen <path>
silver frmt <file>
silver package init <package_name>
silver version
```

## Start the REPL

```console
silver
```

Running Silver without arguments starts an interactive session. Definitions remain available until the session exits.
Send end-of-file to exit (`Ctrl+D` on macOS and Linux or `Ctrl+Z`, then Enter, on Windows).

## Run a source file

```console
silver program.slv
```

A single argument that is not a command name is treated as a source path. Silver evaluates the file and writes only
output requested by the program to standard output. Uncaught errors and tracebacks are written to standard error.

## Generate AST caches

```console
silver astgen <path>
```

`astgen` parses a source file and writes a sibling `.astc` cache. If `<path>` is a directory, Silver recursively
generates caches for every `.slv` file below it. The command prints the path of each generated cache.

AST caches are implementation details and can be safely regenerated. Normal file execution also creates or refreshes
a cache when needed.

## Format a source file

```console
silver frmt <file>
```

`frmt` formats one source file in place. It prints the file path when the file changes and produces no output when the
file is already formatted.

## Initialize a package

```console
silver package init <package_name>
```

`package init` creates `package.yaml` in the current directory. Package names must begin with an ASCII letter or
underscore; subsequent characters may also be digits. The command refuses to overwrite an existing `package.yaml`.

For example, `silver package init my_library` creates:

```yaml
package: my_library
export: []
```

Add relative `.slv` paths to the `export` list to expose modules from the package. See
[Modules and imports](language_guide/modules.md#packages-and-yaml-manifests) for manifest rules and the
[package example](../examples/packages/package_b/demo.slv) for a complete consumer.

## Print the version

```console
silver version
```

`version` prints the interpreter name and semantic version, such as `silver 0.8.0`.

## Exit statuses

Silver uses status `0` for success, `1` when execution or a command fails, and `2` for invalid command usage.

[Documentation index](table_of_contents.md) | [Getting Started](getting_started.md) | [Project README](../README.md)
