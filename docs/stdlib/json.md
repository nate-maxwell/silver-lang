# `json`

`json` converts between JSON and Silver's null, boolean, string, integer, float, array, and string-keyed map values.
Parsing and encoding are implemented in Silver; `load` and `dump` delegate only file reads and writes to `io`.

```silver
let json = import("json")

let value = json.loads("{\"name\":\"Silver\",\"ready\":true}")
let pretty = json.dumps(value, 2)
```

| Function                     | Description                                                                     |
|------------------------------| ------------------------------------------------------------------------------- |
| `loads(document)`            | Decode one complete JSON document.                                              |
| `load(file)`                 | Read and decode a native file-like object.                                      |
| `dumps(value, indent?)`      | Encode to a string. Optional indent is an integer number of spaces or a string. |
| `dump(value, file, indent?)` | Encode and write to a native file-like object; returns null.                    |

___

`JSONDecodeError` is raised whenever invalid JSON is encountered.

| Field     | Description                                                                    |
|-----------|--------------------------------------------------------------------------------|
| `message` | Complete error message, including the line, column, and character position.    |
| `msg`     | Error description without location information.                                |
| `doc`     | Complete JSON document that failed to decode.                                   |
| `pos`     | Zero-based character position where decoding failed.                           |
| `lineno`  | One-based line number where decoding failed.                                    |
| `colno`   | One-based column number where decoding failed.                                  |

Invalid syntax and extra non-whitespace data produce this error. Integers must fit Silver's signed 64-bit range.

Only string map keys are JSON-serializable. Functions, structs, modules, and other values are rejected, as are
non-finite floats and circular arrays/maps.

Integer indentation values at or below zero add line breaks without leading spaces; positive indentation is capped at
1,000 spaces.

[Standard library index](../table_of_contents.md#standard-library)
