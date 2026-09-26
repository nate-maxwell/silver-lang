# Silver Standard Library

The standard library contains modules implemented in both Go and Silver.
Go modules primarily support Silver primitives, or lower-level implementation details
that silver primitives cannot produce themselves (like sockets, io streams, etc.).

Tests for the standard library, both Go and Silver implementations, are written in Go
for a consistent output format.

Stub files (.stb) exist for tools to understand Go and Silver implementations.

Every bundled package has a `package.yaml`. Packages with Silver sources keep it beside
their sources in `silver/<package>/`; packages implemented entirely in Go keep it in
`native/<package>/`. The interpreter embeds these manifests, so imports work from
any working directory without a filesystem copy of the standard library.

The manifests use the same schema as user packages:

```yaml
package: http
members: []
export:
  - ./http.slv
  - ./client.slv
  - ./server.slv
  - ./cookies.slv
  - ./cookiejar.slv
```

Exported files automatically belong to the package. Add internal `.slv` files to
`members` to share the package's operator grammar without creating public import
names. Embedded members may import each other by relative file path. An exported
`<package>.slv` supplies the bare package name; other exported paths supply
`<package>/<path-without-extension>`, such as `http/client`.

Packages implemented entirely in Go have empty file lists. Their `package` name
selects the Go implementation, which supplies the module's bindings. A package
can also combine Go bindings with an exported `<package>.slv` entry file. Those
bindings are available directly in that entry file, and its `export` block
controls the public API. For example, the single `networking` package combines
native socket functions and types with a Silver protocol enum and `dial` function.
A native implementation without a manifest is a build configuration error
reported during library setup.

Each package has its own operator scope: HTTP modules share one scope, and JSON
has another. Adding a Silver entry point requires updating its manifest, not a
Go registration list. Run `go test ./...` after changing Silver sources.
