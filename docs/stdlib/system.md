# `system`

`system` reports host information and reads or changes the current process environment.

```silver
let system = import("system")
let io = import("io")

io.print(system.system(), system.release(), system.machine())
io.print("Silver " + system.VERSION)
io.print(system.getenv("PATH"))

system.append_path("./modules")
io.print(system.getenv(system.ENV_SILVER_PATH))
```

| Function or property  | Description                                                              |
|-----------------------|--------------------------------------------------------------------------|
| `MAJOR`               | Silver's integer major-version component.                                |
| `MINOR`               | Silver's integer minor-version component.                                |
| `PATCH`               | Silver's integer patch-version component.                                |
| `VERSION`             | Silver's version string in `MAJOR.MINOR.PATCH` form.                     |
| `ENV_SILVER_PATH`     | The string `"SILVER_PATH"`, Silver's source-module search-path variable. |
| `get_path_sep()`      | Platform path-list separator: `";"` on Windows and `":"` otherwise.      |
| `append_path(path)`   | Append an entry to `SILVER_PATH` and return null.                        |
| `system()`            | Friendly OS name such as `"Windows"`, `"Linux"`, or `"Darwin"`.          |
| `release()`           | OS/kernel release when available.                                        |
| `machine()`           | Machine architecture.                                                    |
| `processor()`         | Processor model/identifier when available.                               |
| `node()`              | Host name.                                                               |
| `getenv(name)`        | Environment value, or `""` when unset.                                   |
| `setenv(name, value)` | Set an environment variable for this process; returns null.              |
| `environment()`       | Snapshot map of all environment variables.                               |

Host-information queries return an empty string when the platform cannot provide a value. Changes made by `setenv`
or `append_path` affect later code and child processes, not the parent shell. When `SILVER_PATH` is empty,
`append_path` sets it directly rather than adding a leading separator.

[Standard library index](../table_of_contents.md#standard-library)
