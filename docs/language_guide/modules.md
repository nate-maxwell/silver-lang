# Modules and Imports

Every Silver source file is a module. A module is an isolated namespace whose members are the top-level bindings
created while that file is evaluated.

Imported source files must be `.slv` members of a package declared by a YAML manifest.

## Importing modules

`import(expression)` evaluates its argument, requires a string, loads the corresponding module, and returns a value of
type `module`. All imports use `<package>:<module>[/<optional_sub_namespace>]`:

```silver
let io = import("core:io")
let logging = import("core:logging")
let cookiejar = import("core:http/cookiejar")
let println = import("core:io").println

println("hello")
```

The path can be computed at runtime:

```silver
let module_path = "my_library:helpers"
let helpers = import(module_path)
```

A non-string path raises `TypeError`. A module that cannot be read or resolved raises `ImportError`.
Every imported source file must belong to a registered package. Bare names such as `import("io")` and
filesystem imports such as `import("./helpers.slv")` are invalid. Use `import("core:io")` or
`import("my_library:helpers")`.

## Path resolution

The package name selects the namespace; the remaining path selects an exported module:

| Import form                                   | Resolution                                                                      |
|-----------------------------------------------|---------------------------------------------------------------------------------|
| `core:io`, `core:http/cookiejar` | Load the named standard-library module embedded in the interpreter. |
| `my_library:that_module` | Find `package: my_library` on `SILVER_PATH`, then its `that_module.slv` export. |
| `my_library:that_other/module` | Find the same package's `that_other/module.slv` export. |

Qualified names are case-sensitive on every platform and omit `.slv`. Use `/` for namespace separators; empty,
`.` and `..` components are invalid. Qualified imports never fall back to neighboring files or exports from another
package. `core` is reserved for the bundled standard library, including `core:core` for `len`, `type`, and `range`.

## Packages and YAML manifests

A package requires a YAML manifest. Use `package.yaml` at the package root and register the path to that file:

```yaml
package: my_library
authors: []
members:
  - ./internal/helper.slv
export:
  - ./this_module.slv
  - ./that_module.slv
  - ./that_other/module.slv
```

The manifest is a single YAML document with a required `package` string and `export` list, plus optional `authors`
and `members` lists. Authors are strings. Unknown fields are rejected. Every exported file is automatically a package
member; `members` adds files that share the package's language context without exposing them through search-path lookup.
A file may appear in both lists,
but duplicate entries within either list and ownership by multiple manifests are errors.

Each exported file becomes importable as `<package>:<declared-path-without-.slv>`. The leading `./` is optional in
the manifest. Nested paths retain their namespace: `my_library:module` does not match `./that_other/module.slv`.

```silver
let system = import("core:system")
system.append_path("D:/my_library/package.yaml")

let that_module = import("my_library:that_module")
let foo = import("my_library:that_module").foo
let nested = import("my_library:that_other/module")
that_module.foo()
```

`system.append_path` adds the manifest path to the process's `SILVER_PATH`; subsequent imports see the addition immediately.
Relative search-path entries resolve against the process working directory. You can also set `SILVER_PATH` before
starting Silver, separating entries with the os path separator character. If multiple registered manifests declare the
same package name, the first one owns that entire namespace; later manifests do not supply missing exports.

Every `SILVER_PATH` entry must point to an existing file named `package.yaml`. Directories, individual `.slv`
files, other YAML filenames, and missing manifest files are errors. `system.append_path` validates the manifest
before adding it and leaves `SILVER_PATH` unchanged if validation fails.

Member and export paths must be relative `.slv` files inside the package root and must exist when the manifest is loaded. The
manifest's file exports are separate from a source module's `export { Name }` declaration: the manifest controls which
modules can be found, while the source declaration controls which bindings a loaded module exposes.

All members of one manifest share a package-local custom-operator grammar. Operator declarations are discovered across
all members before the package is parsed, regardless of list order or which files are public. Their runtime definitions
still execute normally when the declaring module is imported. Package preparation parses every member, so syntax errors
in any member prevent the package from loading.

For example, an API can import `library:ops` and `library:helper`, and the helper can use an operator defined in `ops.slv`:

```yaml
package: library
members:
  - ./helper.slv
export:
  - ./ops.slv
  - ./api.slv
```

The helper is an internal member: `import("library:helper")` works only inside modules belonging to this
registered package. Other packages can import only its `export` files. Running a member directly discovers the
nearest enclosing `package.yaml` for its operator scope; the manifest must still be explicitly registered for imports.
Moving or deleting a registered manifest makes subsequent imports fail, including imports of already-cached modules.

A script passed directly to the CLI may run without a manifest, but every source file it imports must belong to a
manifest package. An import never falls back to treating a file as standalone.

The bundled standard library keeps internal manifests for each implementation group. `core:http`,
`core:http/client`, `core:http/server`, and the cookie modules share the HTTP operator scope; `core:json` has its
own scope. All are publicly imported through `core`. See the [standard-library layout](../../stdlib/README.md).

## Module members and exports

By default, every top-level binding becomes a public module member, including `let` bindings and struct or enum
definitions. Files without an export declaration therefore retain the original export-all behavior:

```silver
# geometry.slv
type Point = struct { x: int, y: int }

let origin = Point{0, 0}
let translate = fn(point: Point, amount: int) Point {
    return Point{point.x + amount, point.y + amount}
}
```

Add one top-level `export` block to make only its listed bindings public:

```silver
# geometry.slv
export {
    Point,
    translate,
}

type Point = struct { x: int, y: int }
let origin = Point{0, 0} # Private to this module.
let translate = fn(point: Point, amount: int) Point {
    return Point{point.x + amount, point.y + amount}
}
```

The declaration may appear before or after the bindings it names. `export {}` creates a module with no public
members. Listing a name that is not defined as a top-level binding raises `NameError` when the module is imported.
Only one export block is allowed, and it is not valid inside a function or another block.

Importers access those bindings through member syntax:

```silver
let geometry = import("my_library:geometry")
let point: geometry.Point = geometry.translate(geometry.origin, 5)
```

Top-level bindings stay isolated from the importing scope. Importing `geometry.slv` does not introduce `Point`,
`origin`, or `translate` as unqualified names. Accessing an absent or private member raises `AttributeError`.

Module members cannot be replaced through member assignment. Member assignment is reserved for mutable struct fields:

```silver
# geometry.origin = geometry.Point{1, 1} # TypeError
```

## Evaluation, caching, and state

A successful module is evaluated once per interpreter session and cached by its bundled name or canonical absolute
path. Repeated imports return the same module object:

```silver
let first = import("core:testing")
let second = import("core:testing")
first == second # True
```

File identity uses the resolved absolute path with normalized separators and dot segments. On Windows, path casing is
also ignored when registering a manifest. Package and module names remain case-sensitive on every platform;
`my_library:Identity` and `my_library:identity` are distinct names. Symbolic links and hard links are not resolved for identity.

Consequently, module functions share the top-level bindings they captured during evaluation. This is useful for modules
such as the stateful [`testing`](../stdlib/testing.md) runner, but libraries should make shared mutation deliberate:

```silver
# counter.slv
let count = 0
let next = fn() int {
    count = count + 1
    return count
}
```

Every importer of `counter.slv` observes the same sequence through `next()`. Reassigning `count` updates the binding seen
by the module's closures; it does not replace the value already exposed as `counter.count`. Mutating an exposed array,
map, or struct is visible through every import because those importers hold the same value.

Lazy template evaluations share this session state, including modules first imported after the template was created:

````silver
let template = ```{import("my_library:counter").next()}```
[template.eval(), template.eval()] # ["1", "2"]
````

A module is not cached when its evaluation fails. An import cycle detected while modules are loading raises
`ImportError` rather than exposing a partially initialized namespace, including when the import runs inside a template.

## Module values and qualified types

Use the primitive `module` annotation when a function expects an intact module value:

```silver
let announce = fn(library: module) {
    library.print("ready")
}

announce(import("core:io"))
```

Nominal definitions belonging to a module are named through the binding that holds the module:

```silver
let paths = import("core:path")
let current: paths.Path = paths.cwd()
```

The qualification matters: nominal types are identified by their exact definitions, not merely by their names or
fields.

## Module destructuring

Modules implement the same named-value protocol as structs. If a module argument does not satisfy the parameter at its
position, its members can fill remaining parameters with matching names:

```silver
let announce = fn(print: call) {
    print("ready")
}

announce(import("core:io"))
```

An argument that satisfies a `module` parameter remains intact and is not destructured. The complete binding algorithm
is described under [Object destructuring](objects.md#object-destructuring).

## Failures and diagnostics

Failures raised while evaluating a module propagate through `import`. Tracebacks retain the imported file's path and
include module and function frames. See [Errors and diagnostics](errors.md#runtime-diagnostics).

[Language guide](language_guide.md) | [Documentation index](../table_of_contents.md)
