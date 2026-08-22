# `path`

`path` offers object-oriented, platform-native filesystem paths. Each `Path` stores computed properties and callable
operations; methods returning paths produce new `Path` values.
Path construction, lexical operations, matching, and traversal are implemented in Silver; host filesystem access is
delegated to low-level `io` operations.

```silver
let path = import("path")

let config = path.home().joinpath(".silver").joinpath("config.json")
if config.exists() {
    import("io").print(config.read_text())
}
```

## Constructors

| Function or type      | Description                                       |
| --------------------- | ------------------------------------------------- |
| `new(path: str) Path` | Clean and wrap a platform path.                   |
| `cwd() Path`          | Current working directory.                        |
| `home() Path`         | Current user's home directory.                    |
| `Path`                | Nominal path type for annotations and inspection. |

## Properties

| Property   | Description                                                                    |
| ---------- | ------------------------------------------------------------------------------ |
| `path`     | The cleaned, platform-native path string.                                      |
| `anchor`   | The drive and root that begin an absolute path.                                |
| `drive`    | The drive letter or volume name, when present.                                 |
| `root`     | The root directory separator for an absolute path, or an empty string.         |
| `name`     | The final path component.                                                      |
| `stem`     | The final component without its last suffix.                                   |
| `suffix`   | The final component's last file extension, including the leading dot.          |
| `parts`    | The path components as an array of strings.                                    |
| `suffixes` | All file extensions on the final component as an array of strings.             |

These properties are computed when the `Path` is constructed.

## Pure path methods

| Method                                                        | Description                                                               |
| ------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `parent()` / `parents()`                                      | Return the immediate parent or all ancestors.                             |
| `absolute()` / `resolve()`                                    | Return an absolute or symlink-resolved path.                              |
| `as_posix()` / `as_uri()`                                     | Return a slash-separated path or file URI.                                |
| `expanduser()`                                                | Expand a leading `~`.                                                      |
| `is_absolute()`                                               | Test platform-native absoluteness.                                         |
| `is_relative_to(other)` / `relative_to(other)`                | Test or compute lexical containment.                                       |
| `joinpath(parts: any...)`                                     | Append one or more string or `Path` components.                            |
| `match(pattern)`                                              | Match a path pattern.                                                      |
| `with_name(name)` / `with_stem(stem)` / `with_suffix(suffix)` | Return a path with one component changed.                                  |

## Queries and traversal

| Method                                                                    | Description                                                                    |
| ------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `exists()`                                                                | Test whether the path can be statted.                                           |
| `is_dir()` / `is_file()` / `is_symlink()`                                 | Common type predicates.                                                         |
| `is_block_device()` / `is_char_device()` / `is_fifo()` / `is_socket()`    | Special-file predicates.                                                        |
| `is_mount()`                                                              | Test whether the path is a mount point.                                         |
| `samefile(other)`                                                         | Test whether two paths identify the same file.                                  |
| `iterdir()`                                                               | Return immediate children as `Path` values.                                     |
| `glob(pattern)` / `rglob(pattern)`                                        | Return matching descendants as `Path` values.                                  |
| `walk()`                                                                  | Return recursive rows of `[directory_path, directory_names, file_names]`.       |
| `stat()` / `lstat()`                                                      | Return a metadata map, following or not following a final symlink.              |
| `readlink()`                                                              | Return the symlink target as a `Path`.                                          |

Filesystem type predicates return `False` for a missing path. Ordering from traversal methods follows the host
filesystem unless the operation documents a stronger order in the implementation.

## Filesystem changes and I/O

| Method                                               | Description                                                                                  |
| ---------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `chmod(mode)`                                        | Change permission bits.                                                                      |
| `mkdir()` / `rmdir()`                                | Create one directory or remove an empty directory.                                           |
| `touch()` / `unlink()`                               | Create/update a file or remove a non-directory.                                               |
| `rename(target)` / `replace(target)`                 | Move and return the target `Path`; `replace` permits replacement.                            |
| `hardlink_to(target)` / `symlink_to(target)`         | Create a link at this path pointing to `target`.                                              |
| `read_text()` / `read_bytes()`                       | Read complete contents as a Silver string. Byte APIs use strings as byte containers.         |
| `write_text(contents)` / `write_bytes(contents)`     | Replace contents and return the byte count.                                                   |
| `open() File \| FileNotFound \| PermissionDenied`   | Open through the `io` module's native `File` contract.                                       |

Most host failures outside the typed `open` contract surface as catchable runtime `ValueError`s. Path syntax and
behavior are intentionally platform-dependent.

[Standard library index](../table_of_contents.md#standard-library)
