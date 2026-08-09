# `time`

`time` supplies nominal `Time` and `Duration` values backed by the host clock.

```silver
let time = import("time")

let started = time.now()
time.sleep(time.duration(50, "ms"))
let elapsed = time.diff(started, time.now())
elapsed.milliseconds
```

## Exports

| Export                                                              | Description                                |
| ------------------------------------------------------------------- | ------------------------------------------ |
| `now() Time`                                                        | Current local time.                        |
| `unix() int`                                                        | Current Unix timestamp in whole seconds.   |
| `format(value: Time, format: str) str`                              | Format a time with Silver tokens.          |
| `parse(value: str, format: str) Time`                               | Parse a time using the same tokens.        |
| `duration(value, unit) Duration`                                    | Build from an integer or float and a unit. |
| `add(value: Time, duration: Duration) Time`                         | Add a duration.                            |
| `diff(from: Time, to: Time) Duration`                               | Compute `to - from`.                       |
| `sleep(duration: Duration)`                                         | Block for the duration.                    |
| `before(left, right)` / `after(left, right)` / `equal(left, right)` | Compare times.                             |
| `Time`, `Duration`                                                  | Nominal definitions for annotations.       |

Accepted duration units are nanoseconds (`ns`), microseconds (`us`), milliseconds (`ms`), seconds (`s`), minutes (`m`),
hours (`h`), and days (`d`), including singular and plural names.

`Time` fields are `year`, `month`, `day`, `hour`, `minute`, `second`, `nanosecond`, and `timezone`. `Duration` exposes
total `hours`, `minutes`, `seconds`, and `milliseconds` as floats, and total `nanoseconds` as an integer.

## Format tokens

| Token                        | Meaning                                  |
| ---------------------------- | ---------------------------------------- |
| `YYYY`, `YY`                 | Four- or two-digit year.                 |
| `MM`, `DD`                   | Month and day.                           |
| `HH`, `mm`, `ss`             | 24-hour time.                            |
| `SSS`, `SSSSSS`, `SSSSSSSSS` | Milliseconds, microseconds, nanoseconds. |
| `Z`, `ZZ`                    | Numeric offset with or without a colon.  |
| `z`                          | Zone abbreviation.                       |

Text inside `[brackets]` is literal, which is useful when ordinary text contains token letters:

```silver
time.format(time.now(), "YYYY-MM-DD[T]HH:mm:ssZ")
```

[Standard library index](../table_of_contents.md#standard-library)
