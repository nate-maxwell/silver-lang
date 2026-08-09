# `system`

`system` reports host information and reads or changes the current process environment.

```silver
let system = import("system")
let io = import("io")

io.print(system.system(), system.release(), system.machine())
io.print(system.getenv("PATH"))
```

| Export                | Description                                                     |
| --------------------- | --------------------------------------------------------------- |
| `system()`            | Friendly OS name such as `"Windows"`, `"Linux"`, or `"Darwin"`. |
| `release()`           | OS/kernel release when available.                               |
| `machine()`           | Machine architecture.                                           |
| `processor()`         | Processor model/identifier when available.                      |
| `node()`              | Host name.                                                      |
| `getenv(name)`        | Environment value, or `""` when unset.                          |
| `setenv(name, value)` | Set an environment variable for this process; returns null.     |
| `environment()`       | Snapshot map of all environment variables.                      |

Host-information queries return an empty string when the platform cannot provide a value. Changes made by `setenv`
affect later code and child processes, not the parent shell.

[Standard library index](../table_of_contents.md#standard-library)
