# `io`

`io` exposes the evaluator's process streams, printing, and read/write files.

```silver
let io = import("io")

io.print("hello", "Silver") # arguments separated by spaces, then newline

let command = io.stdin.read_line() # waits for one line and removes its newline

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
| `stdin`                                                   | A readable `IOStream`; `read()` consumes all input and `read_line()` consumes one line.     |
| `stdout`                                                  | An `IOStream` with `write(data: str) \| IOError`.                                           |
| `stderr`                                                  | An `IOStream` with `write(data: str) \| IOError`.                                           |

`File` exposes its source path through the `path` field and provides these methods:

| Method                       | Description                                                                 |
| ---------------------------- | --------------------------------------------------------------------------- |
| `read() str \| IOError`      | Read and return the complete file contents from the beginning.              |
| `write(contents: str) \| IOError` | Truncate the file and replace its complete contents; return null on success. |
| `close() \| IOError`         | Close the file; later operations return `IOError`.                          |

Standard streams cannot be closed.

`IOStream.read_line() str | IOError` waits until a line is available and returns it without the trailing `\n` or
`\r\n`. A final line does not need a newline. At end-of-input it returns an empty string. It shares its buffered input
with `read()`, so a later `read()` returns everything remaining after the lines already consumed.

`IOError`, `FileNotFound`, and `PermissionDenied` have a `message: str` field and can be handled with `try`/`catch`.

[Standard library index](../table_of_contents.md#standard-library)
