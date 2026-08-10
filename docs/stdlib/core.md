# `core`

`core` contains the small set of general-purpose operations that Silver keeps out of the global namespace.

```silver
let core = import("core")

core.len([10, 20, 30]) # 3
core.range(2, 5)       # [2, 3, 4]
core.type(42) == int   # True
```

| Function            | Description                                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `len(value)`        | Return the number of array elements, map pairs, or string bytes. Use `string.chars` to count Unicode code points.              |
| `range(start, end)` | Return integers from inclusive `start` to exclusive `end`; return `[]` when `end <= start`. The limit is 1,000,000 elements.   |
| `type(value)`       | Return the first-class primitive type or exact nominal struct/enum definition.                                                 |

[Standard library index](../table_of_contents.md#standard-library)
