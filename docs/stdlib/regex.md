# `regex`

`regex` provides regular-expression searching, capture groups, replacement, splitting, and reusable compiled
expressions. It uses Go's RE2 engine, so matching runs in linear time with respect to the input.

```silver
let regex = import("regex")

let match = regex.search("(?P<name>[A-Za-z]+):(?P<value>[0-9]+)", "port:8080")
match.group("name")  # "port"
match.group("value") # "8080"
match.span()         # [0, 9]
```

A search that finds nothing returns null. `findall`, `findlist`, and `split` return empty arrays when appropriate.
Malformed patterns raise a `ValueError`.

## Module functions

| Function                                      | Description                                                                                  |
| --------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `match(pattern, string) MatchObject \| null`  | Match only at the beginning of `string`.                                                     |
| `search(pattern, string) MatchObject \| null` | Find the first match anywhere in `string`.                                                   |
| `fullmatch(pattern, string) MatchObject \| null` | Match only when the pattern consumes the entire string.                                   |
| `findall(pattern, string) array`              | Return every non-overlapping complete match as a string.                                     |
| `findlist(pattern, string) array`              | Return every non-overlapping complete match as a `MatchObject`.                              |
| `sub(pattern, replacement, string) str`        | Replace every non-overlapping match.                                                         |
| `subn(pattern, replacement, string) array`     | Return `[replaced_string, replacement_count]`.                                               |
| `split(pattern, string) array`                 | Split around non-overlapping matches. Captured separators are not included.                  |
| `escape(string) str`                           | Escape regex metacharacters so the result matches the input literally.                      |
| `compile(pattern) Expression`                  | Compile a pattern once and return a reusable expression object.                              |

`findall` always returns complete match strings, even when the pattern contains capture groups. Use `findlist` to
inspect captures from every match.

Replacement strings use RE2 expansion syntax. `$1` inserts a numbered capture, `${name}` inserts a named capture,
and `$$` inserts a literal dollar sign:

```silver
regex.sub("([A-Za-z]+), ([A-Za-z]+)", "$2 $1", "Maxwell, Nate") # "Nate Maxwell"
regex.subn("[0-9]+", "#", "12 apples and 3 pears")              # ["# apples and # pears", 2]
```

## `MatchObject`

`MatchObject` is the nominal type returned by successful matching operations.

| Method or property | Description                                                                                         |
| ------------------ | --------------------------------------------------------------------------------------------------- |
| `group()`          | Return the complete matched string.                                                                 |
| `group(index)`     | Return a numbered capture; group `0` is the complete match.                                         |
| `group(name)`      | Return a named capture.                                                                             |
| `groups()`         | Return all capture groups, excluding group `0`, as an array.                                        |
| `groupmap()`       | Return a map from capture names to their matched strings.                                           |
| `start(group?)`    | Return the match or selected group's zero-based start byte index.                                   |
| `end(group?)`      | Return the match or selected group's exclusive end byte index.                                      |
| `span(group?)`     | Return `[start, end]` for the match or selected group.                                               |
| `string`           | The complete original string that was searched.                                                     |

The optional argument to `start`, `end`, and `span` may be a group index or name. An optional capture that did not
participate returns null from `group`, appears as null in `groups` and `groupmap`, and has the span `[-1, -1]`.
Requesting a group that does not exist raises an `IndexError`.

Indexes are UTF-8 byte indexes, consistent with `core.len` and the indexes returned by `string.find`.

## Compiled expressions

`compile(pattern)` validates and stores a pattern in an `Expression`. Its matching methods omit the pattern argument:

```silver
let number = regex.compile("[0-9]+")

number.search("version 23").group() # "23"
number.findall("1, 22, 333")        # ["1", "22", "333"]
number.sub("#", "room 12")         # "room #"
number.fullmatch("2026").group()    # "2026"
```

| Module call                                  | Compiled equivalent                 |
| -------------------------------------------- | ----------------------------------- |
| `match(pattern, string)`                     | `expression.match(string)`          |
| `search(pattern, string)`                    | `expression.search(string)`         |
| `findall(pattern, string)`                   | `expression.findall(string)`        |
| `findlist(pattern, string)`                  | `expression.findlist(string)`       |
| `sub(pattern, replacement, string)`          | `expression.sub(replacement, string)` |
| `subn(pattern, replacement, string)`         | `expression.subn(replacement, string)` |
| `split(pattern, string)`                     | `expression.split(string)`          |
| `fullmatch(pattern, string)`                 | `expression.fullmatch(string)`      |
| `escape(string)`                             | `expression.escape(string)`         |

`MatchObject` and `Expression` are exported nominal types and can be used in type annotations.

## Pattern syntax

RE2 supports character classes, repetition, alternation, anchors, numbered capture groups, and named groups written
as `(?P<name>...)` or `(?<name>...)`. It deliberately does not support lookahead, lookbehind, or pattern
backreferences.

Backslashes must also survive Silver string parsing. Write two backslashes in source when the regex needs one. For
example, the RE2 digit pattern `\d+` is written as `"\\d+"` in Silver. Character classes such as `"[0-9]+"` avoid
this extra escaping.

[Standard library index](../table_of_contents.md#standard-library)
