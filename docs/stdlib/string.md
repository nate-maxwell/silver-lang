# `string`

`string` contains Unicode-aware transformations, classification, searching, splitting, and joining. Functions take
strings explicitly; Silver does not add methods to `str` values.

```silver
let strings = import("string")

strings.upper("Silver")                    # "SILVER"
strings.split("one,two,three", ",")        # ["one", "two", "three"]
strings.join(["one", "two", "three"], ", ") # "one, two, three"
```

## Transformations

- `capitalize(value)`: Uppercase the first rune and lowercase the rest.
- `lower(value)` / `upper(value)`: Apply Unicode case conversion.
- `swapcase(value)` / `title(value)`: Swap letter case or title-case words.
- `strip(value)` / `lstrip(value)` / `rstrip(value)`: Remove Unicode whitespace.
- `removeprefix(value, prefix)` / `removesuffix(value, suffix)`: Remove one matching affix.
- `replace(value, old, new)`: Replace all occurrences.
- `repeat(value, count)`: Repeat a nonnegative number of times, up to a 1,000,000-byte result.
- `reverse(value)`: Reverse by Unicode code point.

## Splitting and composition

| Export                    | Description                                                   |
| ------------------------- | ------------------------------------------------------------- |
| `chars(value)`            | Array of Unicode code-point strings.                          |
| `fields(value)`           | Split around runs of Unicode whitespace.                      |
| `split(value, separator)` | Split on a nonempty separator.                                |
| `splitlines(value)`       | Split `LF`, `CRLF`, or `CR` lines without a final empty line. |
| `join(values, separator)` | Join an array containing only strings.                        |

## Search and comparison

| Export                                                  | Description                               |
| ------------------------------------------------------- | ----------------------------------------- |
| `compare(left, right)`                                  | `-1`, `0`, or `1` in lexical order.       |
| `contains(value, substring)`                            | Whether a substring occurs.               |
| `count(value, substring)`                               | Non-overlapping occurrence count.         |
| `find(value, substring)` / `rfind(value, substring)`    | First/last byte index, or `-1`.           |
| `startswith(value, prefix)` / `endswith(value, suffix)` | Affix predicates.                         |
| `equal_fold(left, right)`                               | Unicode simple case-insensitive equality. |

## Classification

The one-argument predicates are `isalnum`, `isalpha`, `isascii`, `isdecimal`, `isdigit`, `islower`, `isnumeric`,
`isprintable`, `isspace`, `istitle`, and `isupper`. They use Unicode character properties except `isascii`. Empty input
is false for all classifiers except `isascii` and `isprintable`.

Indexes returned by `find` and `rfind` are UTF-8 byte indexes; use `chars` when code-point positions are needed.

[Standard library index](../table_of_contents.md#standard-library)
