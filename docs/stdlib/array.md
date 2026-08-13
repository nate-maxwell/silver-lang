# `array`

`array` provides non-mutating operations over native arrays. Except for the values already contained in an array,
returned arrays are fresh copies.

```silver
let arrays = import("array")
let values = [3, 1, 2]

arrays.append(values, 4) # [3, 1, 2, 4]
arrays.sort(values)      # [1, 2, 3]
values                   # still [3, 1, 2]
```

| Function                  | Description                                                                                           |
|---------------------------| ----------------------------------------------------------------------------------------------------- |
| `append(array, value)`    | Return a copy with `value` added at the end.                                                          |
| `contains(array, value)` | Test value equality. Numeric comparison accepts mixed `int`/`float` values.                           |
| `first(array)`           | Return the first element, or null when empty.                                                         |
| `last(array)`            | Return the last element, or null when empty.                                                          |
| `remove(array, index)`   | Return a copy without the indexed element, or null when out of range.                                 |
| `rest(array)`            | Return a copy without the first element, or null when empty.                                          |
| `reverse(array)`         | Return a reversed copy.                                                                               |
| `sort(array)`            | Return a stable ascending copy of an all-numeric or all-string array. Integers and floats may be mixed. |

[Standard library index](../table_of_contents.md#standard-library)
