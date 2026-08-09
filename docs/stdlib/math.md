# `math`

`math` supplies integer helpers, floating-point functions, comparisons, and constants. Functions described as numeric
accept `int` or `float`.

```silver
let math = import("math")

math.factorial(5)       # 120
math.gcd([18, 24, 30]) # 6
math.sqrt(2)            # 1.414...
math.sin(math.pi / 2)   # 1.0
```

## Integer and rounding functions

- `factorial(value)`: Return a nonnegative integer's factorial; overflow is an error.
- `gcd(values)`: Return the greatest common divisor of an integer array; the identity for `[]` is `0`.
- `lcm(values)`: Return the least common multiple of an integer array; the identity for `[]` is `1`.
- `lcd(values)`: Compatibility spelling with the same behavior as `lcm`.
- `isqrt(value)`: Return the floor of the square root of a nonnegative integer.
- `abs(value)`: Return the absolute value, preserving integer/float representation.
- `ceil(value)` / `floor(value)` / `trunc(value)`: Return an integer. `truc` is retained as an alias for `trunc`.

## Floating-point functions

| Group           | Exports                                                                       |
| --------------- | ----------------------------------------------------------------------------- |
| Trigonometry    | `acos`, `asin`, `atan`, `cos`, `sin`, `tan`                                   |
| Roots/exponents | `cbrt`, `sqrt`, `exp`, `exp2`, `expm1`                                        |
| Logarithms      | `log(value, base)`, `log1p`, `log2`, `log10`                                  |
| Remainders      | `fmod(left, right)`, `remainder(left, right)`; `remainer` is a retained alias |
| Angles          | `degrees(radians)`, `radians(degrees)`                                        |

`modf(value)` returns `[fractional_part, integer_part]`, both floats.

`min(left, right)` and `max(left, right)` accept mixed numeric values and return one of the original operands,
preserving its type.

## Constants

`pi`, `e`, `tau`, and `nan` are float values. Domain behavior follows Go's IEEE-754 math functions, so an operation such
as `sqrt(-1)` produces NaN rather than a catchable domain exception.

[Standard library index](../table_of_contents.md#standard-library)
