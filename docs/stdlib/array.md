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

- `append(values, value)`: Return a copy with `value` added at the end.
- `contains(values, value)`: Test value equality. Numeric comparison accepts mixed `int`/`float` values.
- `first(values)`: Return the first element, or null when empty.
- `last(values)`: Return the last element, or null when empty.
- `remove(values, index)`: Return a copy without the indexed element, or null when out of range.
- `rest(values)`: Return a copy without the first element, or null when empty.
- `reverse(values)`: Return a reversed copy.
- `sort(values)`: Return a stable ascending copy of an all-numeric or all-string array. Integers and floats may be
  mixed.

[Standard library index](../table_of_contents.md#standard-library)
