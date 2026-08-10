# `math`

`math` supplies integer helpers, floating-point functions, comparisons, and constants. The numeric-type column below
uses `inputs → result`; `numeric` means either `int` or `float`. Floating-point domain behavior follows Go's IEEE-754
math functions, so an operation such as `sqrt(-1)` produces NaN rather than a catchable domain exception.

```silver
let math = import("math")

math.factorial(5)       # 120
math.gcd([18, 24, 30])  # 6
math.sqrt(2)            # 1.414...
math.sin(math.pi / 2)   # 1.0
```

| Function or property       | Numeric type(s)                  | Description                                                                                   |
| -------------------------- | -------------------------------- | --------------------------------------------------------------------------------------------- |
| `factorial(value)`         | `int → int`                      | Return a nonnegative integer's factorial; overflow is an error.                               |
| `gcd(values)`              | `array[int] → int`               | Return the greatest common divisor; the identity for `[]` is `0`.                             |
| `lcm(values)`              | `array[int] → int`               | Return the least common multiple; the identity for `[]` is `1`.                               |
| `lcd(values)`              | `array[int] → int`               | Compatibility alias for `lcm`.                                                                |
| `isqrt(value)`             | `int → int`                      | Return the floor of the square root of a nonnegative integer.                                 |
| `abs(value)`               | `int → int`, `float → float`     | Return the absolute value while preserving its numeric type.                                  |
| `ceil(value)`              | `numeric → int`                  | Round upward to an integer.                                                                   |
| `floor(value)`             | `numeric → int`                  | Round downward to an integer.                                                                 |
| `trunc(value)`             | `numeric → int`                  | Truncate toward zero.                                                                         |
| `truc(value)`              | `numeric → int`                  | Compatibility alias for `trunc`.                                                              |
| `fmod(left, right)`        | `numeric, numeric → float`       | Return the remainder whose quotient is truncated toward zero.                                 |
| `remainder(left, right)`   | `numeric, numeric → float`       | Return the IEEE-754 remainder using the nearest integer quotient.                             |
| `remainer(left, right)`    | `numeric, numeric → float`       | Compatibility alias for `remainder`.                                                          |
| `modf(value)`              | `numeric → array[float]`         | Return `[fractional_part, integer_part]`.                                                      |
| `acos(value)`              | `numeric → float`                | Return the arc cosine in radians.                                                              |
| `asin(value)`              | `numeric → float`                | Return the arc sine in radians.                                                                |
| `atan(value)`              | `numeric → float`                | Return the arc tangent in radians.                                                             |
| `cos(value)`               | `numeric → float`                | Return the cosine of a radian angle.                                                           |
| `sin(value)`               | `numeric → float`                | Return the sine of a radian angle.                                                             |
| `tan(value)`               | `numeric → float`                | Return the tangent of a radian angle.                                                          |
| `cbrt(value)`              | `numeric → float`                | Return the cube root.                                                                          |
| `sqrt(value)`              | `numeric → float`                | Return the square root.                                                                        |
| `exp(value)`               | `numeric → float`                | Return `e` raised to `value`.                                                                  |
| `exp2(value)`              | `numeric → float`                | Return `2` raised to `value`.                                                                  |
| `expm1(value)`             | `numeric → float`                | Return `e` raised to `value`, minus `1`.                                                       |
| `log(value, base)`         | `numeric, numeric → float`       | Return the logarithm of `value` in the supplied base.                                          |
| `log1p(value)`             | `numeric → float`                | Return the natural logarithm of `1 + value`.                                                   |
| `log2(value)`              | `numeric → float`                | Return the base-2 logarithm.                                                                   |
| `log10(value)`             | `numeric → float`                | Return the base-10 logarithm.                                                                  |
| `degrees(radians)`         | `numeric → float`                | Convert radians to degrees.                                                                    |
| `radians(degrees)`         | `numeric → float`                | Convert degrees to radians.                                                                    |
| `min(left, right)`         | `numeric, numeric → numeric`     | Return the smaller original operand, preserving its type; mixed inputs are accepted.          |
| `max(left, right)`         | `numeric, numeric → numeric`     | Return the larger original operand, preserving its type; mixed inputs are accepted.           |
| `pi`                       | `float`                          | The ratio of a circle's circumference to its diameter.                                         |
| `e`                        | `float`                          | Euler's number, the base of natural logarithms.                                                 |
| `tau`                      | `float`                          | One full turn in radians, equal to `2 * pi`.                                                    |
| `nan`                      | `float`                          | An IEEE-754 not-a-number value.                                                                 |

[Standard library index](../table_of_contents.md#standard-library)
