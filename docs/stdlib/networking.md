# `networking`

`networking` provides blocking TCP/UDP connections and TCP listeners. Native sockets are wrapped in ordinary Silver
structs with typed callable fields.

```silver
let net = import("networking")

let connection = net.dial("tcp", "example.com:80")
defer connection.close()
connection.write("GET / HTTP/1.0\r\nHost: example.com\r\n\r\n")
let response = connection.read(4096)
```

## Entry points

| Export                                                           | Description                                   |
| ---------------------------------------------------------------- | --------------------------------------------- |
| `dial(network: str, address: str) Connection \| ConnectionError` | Connect using `"tcp"` or `"udp"`.             |
| `listen(network: str, address: str) Listener \| ListenError`     | Create a listener. Only `"tcp"` is supported. |

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
