# `collections`

`collections` provides three mutable structures: `Deque`, `Stack`, and `DefaultMap`.

```silver
let collections = import("collections")
```

## `Deque`

A `Deque` is an ordered sequence designed for adding and removing values at either end. Use it for queues, work lists,
and other sequences that need efficient access to both the first and last value.

Create an empty deque with `deque()`, or pass an array to `deque(initial)` to copy its values into a new deque:

```silver
let queue = collections.deque(["compile", "test"])
queue.appendleft("format")
queue.append("package")

let first = collections.popleft(queue) # "format"
let last = queue.pop()                  # "package"
```

### Methods and indexing

| Method or operation       | Description                                                                            |
| ------------------------- | -------------------------------------------------------------------------------------- |
| `append(value)`           | Add a value to the right end.                                                          |
| `appendleft(value)`       | Add a value to the left end.                                                           |
| `pop()`                   | Remove and return the value at the right end. Raises `IndexError` when empty.          |
| `deque[index]`            | Read a value by zero-based index. Raises `IndexError` when out of range.               |
| `deque[index] = value`    | Replace a value by zero-based index. Raises `IndexError` when out of range.            |

### Functions

| Function                      | Description                                                                                |
| ----------------------------- | ------------------------------------------------------------------------------------------ |
| `clear(deque)`                | Remove all values.                                                                         |
| `copy(deque)`                 | Return a shallow `Deque` copy.                                                              |
| `count(deque, value)`         | Return the number of matching values.                                                       |
| `extend(deque, other)`        | Append all values from another `Deque` or `Stack`.                                          |
| `extendleft(deque, other)`    | Prepend all values from another `Deque` or `Stack` in reverse order.                        |
| `index(deque, value)`         | Return the first matching index. Raises `ValueError` when absent.                           |
| `insert(deque, index, value)` | Insert at an index, normalizing negative indexes and clamping at either end.                |
| `popleft(deque)`              | Remove and return the value at the left end. Raises `IndexError` when empty.                |
| `remove(deque, value)`        | Remove the first matching value. Raises `ValueError` when absent.                           |
| `reverse(deque)`              | Reverse the values in place.                                                                |
| `rotate(deque, amount)`       | Rotate right by `amount`; a negative amount rotates left.                                  |

Functions that mutate the deque return null.

## `Stack`

A `Stack` is a last-in, first-out sequence. Use it when the most recently added value must be read or removed first,
such as for history, traversal, or nested work.

Create an empty stack with `stack()`, or pass an array to `stack(initial)` to copy its values into a new stack:

```silver
let history = collections.stack()
history.push("open")
history.push("edit")

let current = history.peek() # "edit"
let previous = history.pop() # "edit"
```

### Methods and indexing

| Method or operation      | Description                                                                         |
| ------------------------ | ----------------------------------------------------------------------------------- |
| `push(value)`            | Add a value to the top.                                                             |
| `peek()`                 | Return the top value without removing it. Raises `IndexError` when empty.           |
| `pop()`                  | Remove and return the top value. Raises `IndexError` when empty.                    |
| `stack[index]`           | Read a value by zero-based index. Raises `IndexError` when out of range.            |
| `stack[index] = value`   | Replace a value by zero-based index. Raises `IndexError` when out of range.         |

### Functions

| Function                      | Description                                                                                         |
| ----------------------------- | --------------------------------------------------------------------------------------------------- |
| `clear(stack)`                | Remove all values.                                                                                  |
| `copy(stack)`                 | Return a shallow `Stack` copy.                                                                      |
| `count(stack, value)`         | Return the number of matching values.                                                               |
| `extend(stack, other)`        | Add all values from another `Deque` or `Stack` to the top in iteration order.                       |
| `extendleft(stack, other)`    | Prepend all values from another `Deque` or `Stack` in reverse order.                                |
| `index(stack, value)`         | Return the first matching index. Raises `ValueError` when absent.                                   |
| `insert(stack, index, value)` | Insert at an index, normalizing negative indexes and clamping at either end.                        |
| `popleft(stack)`              | Remove and return the bottom value. Raises `IndexError` when empty.                                 |
| `remove(stack, value)`        | Remove the first matching value. Raises `ValueError` when absent.                                   |
| `reverse(stack)`              | Reverse the values in place.                                                                        |
| `rotate(stack, amount)`       | Rotate right by `amount`; a negative amount rotates left.                                          |

Functions that mutate the stack return null. Prefer `push`, `peek`, and `pop` when ordinary stack behavior is sufficient.

## `DefaultMap`

A `DefaultMap` stores key-value pairs and creates a value when a missing key is first read. Use it for grouping,
counting, or other cases where every key should start with a predictable value.

Create one with `defaultmap(factory)`. The factory must be a zero-argument Silver function. Pass a map as the optional
second argument, `defaultmap(factory, initial)`, to copy existing pairs into the new default map:

```silver
let make_count = fn() int { 0 }
let counts = collections.defaultmap(make_count, {"silver": 1})

counts["go"] = counts["go"] + 1
counts["silver"] = counts["silver"] + 1
```

### Operations and fields

| Operation or property    | Description                                                                                         |
| ------------------------ | --------------------------------------------------------------------------------------------------- |
| `mapping[key]`           | Return the stored value. For a missing key, call `.factory`, store its result, and return it.       |
| `mapping[key] = value`   | Store or replace a value.                                                                           |
| `mapping.values`         | Access the backing map containing all stored pairs.                                                  |
| `mapping.factory`        | Access the zero-argument function used to create missing values.                                     |

`DefaultMap` has no collection methods. The module functions listed for `Deque` and `Stack` do not accept a
`DefaultMap`.

[Standard library index](../table_of_contents.md#standard-library)
