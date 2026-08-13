# `random`

`random` provides pseudo-random numbers, collection choices, and in-place array shuffling. Its generator is seeded
automatically; call `seed` when a repeatable sequence is needed.

```silver
let random = import("random")

random.seed(42)
random.random()                 # float in [0.0, 1.0)
random.randint(1, 6)            # either endpoint may be returned
random.randelem(["a", "b"])    # "a" or "b"

let values = [1, 2, 3]
random.shuffle(values)          # values is changed in place
```

| Function          | Description                                                                 |
|-------------------| --------------------------------------------------------------------------- |
| `random()`        | Return a random float greater than or equal to `0.0` and less than `1.0`.   |
| `seed(int)`       | Seed the generator with an integer and return null.                         |
| `randint(a, b)`   | Return a random integer in the inclusive range from `a` through `b`.        |
| `randelem(array)` | Return a random array element. An empty array raises `IndexError`.          |
| `randkey(map)`    | Return a random map key. An empty map raises `KeyError`.                    |
| `shuffle(array)`  | Shuffle an array in place and return null.                                  |

* `randint` raises `ValueError` when `a` is greater than `b`.
* Map iteration order is unspecified, so seeded calls to `randkey` are repeatable as random index choices but
not guaranteed to select the same key across separate runs.

[Standard library index](../table_of_contents.md#standard-library)
