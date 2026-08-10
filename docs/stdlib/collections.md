# `collections`

`collections` provides three mutable structures: `Deque`, `Stack`, and `DefaultMap`. Their supported methods and
module-level operations differ as described below.

```silver
let collections = import("collections")

let queue = collections.deque([2, 3])
queue.appendleft(1)
queue.append(4)
collections.popleft(queue) # 1
queue[0]                   # 2

let stack = collections.stack()
stack.push("first")
stack.push("second")
stack.pop() # "second"
```

## Types and constructors

- `deque(initial?)` / `Deque`: Mutable deque copied from an optional array. Its methods are `append`, `appendleft`, and
  `pop`; bracket reads and writes are supported.
- `stack(initial?)` / `Stack`: Mutable stack copied from an optional array. Its methods are `push`, `peek`, and `pop`;
  bracket reads and writes are supported.
- `defaultmap(factory, initial?)` / `DefaultMap`: Map-like value that uses a zero-argument function to create
  and retain missing values. The optional initial value is a map.

The `?` after a parameter name means that argument is optional; it is documentation notation, not Silver syntax.

### Supported behavior

| Behavior                                  | `Deque` | `Stack` | `DefaultMap` |
| ----------------------------------------- | :-----: | :-----: | :----------: |
| Bracket reads and writes                  |   Yes   |   Yes   |     Yes      |
| `append`, `appendleft` methods            |   Yes   |   No    |      No      |
| `push`, `peek` methods                    |   No    |   Yes   |      No      |
| `pop` method                              |   Yes   |   Yes   |      No      |
| Missing-key creation through `.factory`   |   No    |   No    |     Yes      |
| Functions in **Mutable operations** below |   Yes   |   Yes   |      No      |

The backing map is available as `.values`. A default map also exposes `.factory`; `mapping[key]` creates a
missing value and `mapping[key] = value` stores one.

```silver
let make_zero = fn() int { 0 }
let counts = collections.defaultmap(make_zero)
counts["silver"] = counts["silver"] + 1
```

## Mutable operations

Every function in this section accepts `Deque` and `Stack`. None accepts `DefaultMap`. For `extend` and `extendleft`,
both `collection` and `other` may be a `Deque` or `Stack`.

| Export                             | Description                                                        |
| ---------------------------------- | ------------------------------------------------------------------ |
| `clear(collection)`                | Remove all values.                                                 |
| `copy(collection)`                 | Shallow copy, preserving deque/stack type.                         |
| `count(collection, value)`         | Number of equal values.                                            |
| `extend(collection, other)`        | Append all values from another collection.                         |
| `extendleft(collection, other)`    | Prepend values in reverse order, matching deque semantics.         |
| `index(collection, value)`         | Index of the first equal value; raises `ValueError` if absent.     |
| `insert(collection, index, value)` | Insert with negative-index normalization and end clamping.         |
| `popleft(collection)`              | Remove and return the first value; raises `IndexError` when empty. |
| `remove(collection, value)`        | Remove the first equal value; raises `ValueError` if absent.       |
| `reverse(collection)`              | Reverse in place.                                                  |
| `rotate(collection, amount)`       | Rotate right; a negative amount rotates left.                      |

Mutators return null. `pop`/`peek` on an empty sequence and out-of-range bracket access raise `IndexError`.

[Standard library index](../table_of_contents.md#standard-library)
