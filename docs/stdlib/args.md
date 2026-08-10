# `args`

`args` is a command-line parser written in Silver. The caller supplies the argument array, which keeps parsing
deterministic and testable. The current CLI does not yet expose process arguments automatically.

```silver
let args = import("args")

let parser = args.new("backup", "Copy a source to a destination")
parser.add(args.positional("source", "File to copy"))
parser.add(args.option_with_default("destination", "d", "Output path", "backup.out"))
parser.add(args.flag("force", "f", "Replace an existing output"))

let parsed = parser.parse(["input.txt", "--destination", "copy.txt", "-f"])
parsed.values["source"]      # "input.txt"
parsed.values["destination"] # "copy.txt"
parsed.values["force"]       # True
```

## Constructors

| Function                                          | Description                                    |
| ------------------------------------------------- | ---------------------------------------------- |
| `new(program, description) Parser`                | Create an empty parser.                        |
| `option(name, short, help) Argument`              | Optional string-valued option.                 |
| `option_with_default(name, short, help, default)` | Optional string option with a default.         |
| `required_option(name, short, help)`              | Required string option.                        |
| `flag(name, short, help)`                         | Boolean flag, defaulting to `False`.           |
| `count(name, short, help)`                        | Integer occurrence counter, defaulting to `0`. |
| `positional(name, help)`                          | Required positional argument.                  |
| `optional_positional(name, help, default)`        | Optional positional with a default.            |

Names do not include leading dashes; a short name is empty or one character. `-h` and `--help` are reserved. Required
positional arguments must be added before optional ones.

## `Parser`

| Method                                    | Description                                          |
| ----------------------------------------- | ---------------------------------------------------- |
| `add(argument) \| ArgumentError`          | Validate and append a declaration.                   |
| `parse(argv) ParsedArgs \| ArgumentError` | Parse an array of strings.                           |
| `help() str`                              | Render usage, description, positionals, and options. |

Long (`--output`) and separate short (`-o`) forms are supported; combined short flags and `--name=value` are not. `--`
disables option parsing. `-h` or `--help` returns early with `help_requested == True`.

`ParsedArgs` contains `values: map` and `help_requested: bool`. `ArgumentError` contains `message: str`. The types
`Argument`, `ArgumentKind`, `ParsedArgs`, `Parser`, and `ArgumentError` are also available for annotations and inspection.

[Standard library index](../table_of_contents.md#standard-library)
