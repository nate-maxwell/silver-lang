# `networking`

`networking` provides blocking TCP/UDP connections and TCP listeners. Native sockets are wrapped in ordinary Silver
structs with typed callable fields. All functions and types on this page are
available from `import("core:networking")`.

Import it as `import("core:networking")`; `_networking` is not an importable module.

```silver
let net = import("core:networking")

let connection: net.Connection = net.dial_tcp("example.com:80")
defer connection.close()
connection.write("GET / HTTP/1.0\r\nHost: example.com\r\n\r\n")
let response = connection.read(4096)
```

## Entry points

| Function                                                            | Description                             |
| ------------------------------------------------------------------- | --------------------------------------- |
| `dial(network: Network, address: str) Connection \| ConnectionError` | Connect using `Network.TCP` or `Network.UDP`. |
| `dial_tcp(address: str) Connection \| ConnectionError`              | Connect over TCP.                      |
| `dial_udp(address: str) Connection \| ConnectionError`              | Create a UDP socket with a default peer. |
| `listen(address: str) Listener \| ListenError`                       | Create a TCP listener.                  |

`net.Network` is an enum with `TCP` and `UDP` members. `net.dial(net.Network.TCP, address)` is equivalent to
`net.dial_tcp(address)`, and `net.dial(net.Network.UDP, address)` is equivalent to `net.dial_udp(address)`.

The module also exports `net.Connection`, `net.Listener`, `net.ReadFromResult`, `net.ConnectionError`,
`net.ListenError`, `net.ReadError`, and `net.WriteError`. Use these names in type annotations and catch clauses:

```silver
let net = import("core:networking")
let message = try {
    net.dial_tcp("not-an-address")
    "connected"
} catch net.ConnectionError err {
    err.message
}
```

## `Connection`

| Field                                               | Description                                             |
| --------------------------------------------------- | ------------------------------------------------------- |
| `address: str`                                      | Remote TCP address or usable local UDP address.         |
| `read(bytes: int) str \| ReadError`                 | Read at most the requested nonnegative byte count.      |
| `write(data: str) \| WriteError`                    | Write to the TCP peer or UDP connection's default peer. |
| `write_to(data: str, address: str) \| WriteError`   | Send a UDP datagram to an explicit address.             |
| `read_from(bytes: int) ReadFromResult \| ReadError` | Receive a UDP datagram and its sender.                  |
| `close() \| ConnectionError`                        | Close the socket.                                       |

`ReadFromResult` contains `data: str` and `address: str`. `write_to` and `read_from` report typed errors when used on
TCP.

## `Listener`

A listener has `address: str`, `accept() Connection | ConnectionError`, and `close() | ConnectionError`. `accept` blocks
until a connection arrives.

`ConnectionError`, `ListenError`, `ReadError`, and `WriteError` each contain `message: str`. Operations are blocking and
have no timeout option in the current API.

[Standard library index](../table_of_contents.md#standard-library)
