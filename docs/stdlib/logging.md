# `logging`

`logging` is implemented in Silver and builds log records from lazy template strings. A logger sends each rendered
record to one handler.

```silver
let logging = import("logging")
let io = import("io")

let handler = logging.new_stream_handler(io.stdout)
let logger = logging.new_logger(
    logging.Level.Info,
    logging.standard_format,
    handler
)

logger.handle("application started")
```

## Levels and loggers

`Level` is an enum with `Debug`, `Info`, `Warning`, `Error`, and `Critical` members. `LogContext` contains `level`,
`time`, and `message`. A format function receives that context and returns a `TemplateString`:

````silver
let compact = fn(context: logging.LogContext) TemplateString {
    ```{context.level}: {context.message}```
}
````

| Function                                    | Description                                                                  |
| ------------------------------------------- | ---------------------------------------------------------------------------- |
| `standard_format(context) TemplateString`   | Return `[{level}] {time} {message}`.                                          |
| `new_logger(level, format, handler) Logger` | Create a logger whose `handle(message)` renders and dispatches a record.      |

`Logger.level` labels the records; the current module does not perform threshold filtering.

## Handlers

| Function                                    | Description                                                                                 |
| ------------------------------------------- | ------------------------------------------------------------------------------------------- |
| `new_stream_handler(stream)`                | Write each record plus a newline to an `IOStream`.                                          |
| `new_file_handler(path)`                    | Append by reading and rewriting the complete `path.Path` file.                              |
| `new_rotating_file_handler(path, interval)` | Rotate after `interval` Unix seconds, suffixing the old path with a timestamp.               |
| `new_tcp_socket_handler(address)`           | Open a TCP connection per record, write it, and close it.                                   |
| `new_udp_socket_handler(address)`           | Open a UDP connection per record, write it, and close it.                                   |
| `new_null_handler()`                        | Discard records.                                                                            |

The corresponding nominal types—`Logger`, `LogContext`, `StreamHandler`, `FileHandler`, `RotatingFileHandler`,
`TCPSocketHandler`, `UDPSocketHandler`, and `NullHandler`—are available for annotations and inspection. Stream and
socket handlers may propagate the typed I/O/network errors declared by their `handle` fields.

File handlers favor a small transparent implementation over append efficiency and are not designed for concurrent
multi-process writes.

[Standard library index](../table_of_contents.md#standard-library)
