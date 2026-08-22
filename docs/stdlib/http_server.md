# `http/server`

`http/server` is a blocking HTTP/1.1 server. A handler receives a parsed `Request` and returns a `Response`.

```silver
let http = import("http")
let server = import("http/server")

let handle = fn(request: server.Request) server.Response {
    if request.method == http.MethodGet && request.path == "/" {
        return server.response(http.StatusOK, "Hello from Silver!")
    }
    return server.response(http.StatusNotFound, "Not found")
}

let app = server.new("127.0.0.1:8080", handle)
defer app.close()
app.serve()
```

See the runnable [server example](../../examples/std_lib/std_server.slv). Run the
[client example](../../examples/std_lib/std_client.slv) in another terminal to send requests to it.

## Requests and handlers

The handler passed to `new` has the signature `call(request: Request) Response`.

| `Request` field | Type | Description |
| --- | --- | --- |
| `method` | `str` | Request method, such as `GET`. |
| `path` | `str` | Request target, including its query string. |
| `version` | `str` | Protocol version, such as `HTTP/1.1`. |
| `headers` | `map` | Request headers with lowercase names. |
| `body` | `str` | Request body. |
| `remote_address` | `str` | TCP address of the client. |

## Responses

| Function | Description |
| --- | --- |
| `response(status_code, body) Response` | Build a response without custom headers. |
| `response_with_headers(status_code, headers, body) Response` | Build a response with custom headers. |

Both constructors obtain the reason phrase from `http.status_description`. A `Response` contains `status_code: int`,
`reason: str`, `headers: map`, and `body: str`. The server normalizes header names and supplies `Content-Length` and
`Connection: close` when writing it.

## Servers

`new(address, handler) Server | ListenError` binds a TCP listener. Use `127.0.0.1:0` to select an available local
port, then read the chosen address from `Server.address`.

| `Server` member | Description |
| --- | --- |
| `address` | Bound listener address. |
| `serve_once()` | Accept and handle one connection, or return `ServerError`. |
| `serve()` | Repeatedly call `serve_once`; blocks until an error occurs. |
| `close()` | Close the listener, or return `ConnectionError`. |

`ServerError` contains `message: str` and wraps malformed requests and accept/read/write failures. `new` retains the
underlying `ListenError`, while `close` retains `ConnectionError`.

## Protocol limits

The current server handles one HTTP/1.1 request per connection and closes that connection after responding. Serving
is sequential. Request bodies use `Content-Length`; chunked transfer encoding, TLS, keep-alive, request timeouts, and
automatic routing or middleware are not provided.

[`http` constants](http.md) | [HTTP client](http_client.md) |
[Standard library index](../table_of_contents.md#standard-library)
