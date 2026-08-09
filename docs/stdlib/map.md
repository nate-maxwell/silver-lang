# `map`

`map` provides safe queries and copy-producing transformations. Direct index assignment remains available when mutation
is desired.

```silver
let maps = import("map")
let original = {"name": "Silver"}
let updated = maps.set(original, "year", 2026)

maps.get(original, "year")   # null
maps.get(updated, "year")    # 2026
maps.contains(updated, "name") # True
```

| Export                     | Description                                      |
| -------------------------- | ------------------------------------------------ |
| `get(mapping, key)`        | Associated value, or null when missing.          |
| `set(mapping, key, value)` | Copy with the pair inserted or replaced.         |
| `delete(mapping, key)`     | Copy without the key; a missing key is harmless. |
| `values(mapping)`          | Values in unspecified iteration order.           |
| `contains(mapping, key)`   | Whether the key is present.                      |

Keys must be hashable Silver values. Numeric keys normalize integer/float equality consistently.

[Standard library index](../table_of_contents.md#standard-library)
