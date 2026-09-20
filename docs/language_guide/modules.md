# Modules and Imports

Every Silver source file is a module. A module is an isolated namespace whose members are the top-level bindings
created while that file is evaluated.

Imported source files must be `.slv` members of a package declared by a YAML manifest.

## Importing modules

`import(expression)` evaluates its argument, requires a string, loads the corresponding module, and returns a value of
type `module`:

```silver
let io = import("io")
let helpers = import("./helpers.slv")

io.println(helpers.answer)
```

The path can be computed at runtime:

```silver
let module_path = "./helpers.slv"
let helpers = import(module_path)
```

A non-string path raises `TypeError`. A module that cannot be read or resolved raises `ImportError`.
Every file import requires an owning package manifest, including relative and absolute paths. A file absent from both
`members` and `export`, or a file whose manifest is missing, raises `ImportError` before evaluation or a cached module
can be returned.

## Path resolution

Silver resolves module names in this order:

| Import form                                   | Resolution                                                                      |
|-----------------------------------------------|---------------------------------------------------------------------------------|
| Bundled standard-library name, such as `"io"` | Load the module embedded in the interpreter.                                    |
| Absolute filesystem path                      | Load that exact path.                                                           |
| Relative or bare path in a `.slv` file        | Check the directory containing that `.slv` file.                                |
| Relative or bare path entered in the REPL     | Check the process working directory.                                            |
| Unresolved source path                        | Check manifest exports in `SILVER_PATH` using the platform's path-list separator. |

Use an explicit relative path such as `./testing.slv` when a user file has the same name as a bundled module.

## Packages and YAML manifests

A package requires a YAML manifest. Use `package.yaml` at the package root and add that directory to `SILVER_PATH`:

```yaml
package: my_library
members:
  - ./internal/helper.slv
export:
  - ./this_module.slv
  - ./that_module.slv
  - ./that_other/module.slv
```

The manifest is a single YAML document with a required `package` string and `export` list, plus an optional `members`
list. Unknown fields are rejected. Every exported file is automatically a package member; `members` adds files that
share the package's language context without exposing them through search-path lookup. A file may appear in both lists,
but duplicate entries within either list and ownership by multiple manifests are errors.

Each exported file becomes importable by its declared relative path or filename. Silver resolves these as canonical full
paths internally; it does not rewrite the process environment. A `SILVER_PATH` directory exposes only exports from the
`.yaml` or `.yml` manifests immediately inside it. A manifest file can also be added directly. Directories without
manifests expose nothing, and individual `.slv` files are no longer accepted as `SILVER_PATH` entries. To migrate an
existing search directory, add a manifest listing its public files in `export`.

Member and export paths must be relative `.slv` files inside the package root and must exist when the manifest is loaded. The
manifest's file exports are separate from a source module's `export { Name }` declaration: the manifest controls which
modules can be found, while the source declaration controls which bindings a loaded module exposes.

All members of one manifest share a package-local custom-operator grammar. Operator declarations are discovered across
all members before the package is parsed, regardless of list order or which files are public. Their runtime definitions
still execute normally when the declaring module is imported. Package preparation parses every member, so syntax errors
in any member prevent the package from loading.

For example, an API can import `./ops.slv` and `./helper.slv`, and the helper can use an operator defined in `ops.slv`:

```yaml
package: library
members:
  - ./helper.slv
export:
  - ./ops.slv
  - ./api.slv
```

The helper cannot be discovered through `SILVER_PATH`, but explicit filesystem imports can load it because it is a
declared package member. Silver first checks registered package membership, then looks for the nearest enclosing
`package.yaml`. Other manifest filenames must be registered through `SILVER_PATH`. Moving or deleting the owning
manifest makes subsequent imports fail, including imports of already-cached modules.

A script passed directly to the CLI may run without a manifest, but every source file it imports must belong to a
manifest package. An import never falls back to treating a file as standalone.

The bundled standard library also uses one `package.yaml` per package. For example, `http` and its `http/client`,
`http/server`, and cookie modules share the HTTP package, while `json` has its own package and operator scope. Existing
bundled import names are unchanged. See the [standard-library layout](../../stdlib/README.md).

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
let geometry = import("./geometry.slv")
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
let first = import("testing")
let second = import("testing")
first == second # True
```

File identity uses the resolved absolute path with normalized separators and dot segments. On Windows, path casing is
also ignored: `./Identity.slv` and `./identity.slv` return the same module and the same nominal struct and enum definitions.
Bundled names remain case-sensitive on every platform. Symbolic links and hard links are not resolved for identity.

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
let template = ```{import("./counter.slv").next()}```
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

announce(import("io"))
```

Nominal definitions belonging to a module are named through the binding that holds the module:

```silver
let paths = import("path")
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

announce(import("io"))
```

An argument that satisfies a `module` parameter remains intact and is not destructured. The complete binding algorithm
is described under [Object destructuring](objects.md#object-destructuring).

## Failures and diagnostics

Failures raised while evaluating a module propagate through `import`. Tracebacks retain the imported file's path and
include module and function frames. See [Errors and diagnostics](errors.md#runtime-diagnostics).

[Language guide](language_guide.md) | [Documentation index](../table_of_contents.md)
