# `io`

`io` exposes the evaluator's process streams, printing, and read/write files.

```silver
let io = import("io")

io.print("hello", "Silver") # arguments separated by spaces, then newline

let file = io.open("notes.txt")
defer file.close()
let old_contents = file.read()
file.write(old_contents + "\nupdated")
```

## Exports

- `print(...values)`: Print inspected values separated by spaces and followed by a newline; return null.
- `open(path) File | FileNotFound | PermissionDenied`: Open an existing file for reading and writing. It does not create
  a missing file.
- `stdin`: An `IOStream` with `read() str | IOError`. Reading consumes all remaining input.
- `stdout`: An `IOStream` with `write(data: str) | IOError`.
- `stderr`: An `IOStream` with `write(data: str) | IOError`.

`File` has `path`, `read() str | IOError`, `write(contents: str) | IOError`, and `close() | IOError`. `write` truncates
and replaces the complete file. Standard streams cannot be closed.

`IOError`, `FileNotFound`, and `PermissionDenied` have a `message: str` field and can be handled with `try`/`catch`.

[Standard library index](../table_of_contents.md#standard-library)
