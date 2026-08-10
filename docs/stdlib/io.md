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

## Functions and streams

| Function or property                                      | Description                                                                                 |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `print(...values)`                                        | Print inspected values separated by spaces and followed by a newline; return null.          |
| `open(path) File \| FileNotFound \| PermissionDenied`     | Open an existing file for reading and writing. It does not create a missing file.           |
| `stdin`                                                   | An `IOStream` whose `read() str \| IOError` consumes all remaining input.                   |
| `stdout`                                                  | An `IOStream` with `write(data: str) \| IOError`.                                           |
| `stderr`                                                  | An `IOStream` with `write(data: str) \| IOError`.                                           |

`File` exposes its source path through the `path` field and provides these methods:

| Method                       | Description                                                                 |
| ---------------------------- | --------------------------------------------------------------------------- |
| `read() str \| IOError`      | Read and return the complete file contents from the beginning.              |
| `write(contents: str) \| IOError` | Truncate the file and replace its complete contents; return null on success. |
| `close() \| IOError`         | Close the file; later operations return `IOError`.                          |

Standard streams cannot be closed.

`IOError`, `FileNotFound`, and `PermissionDenied` have a `message: str` field and can be handled with `try`/`catch`.

[Standard library index](../table_of_contents.md#standard-library)
