# `http/client`

`http/client` is a blocking HTTP/1.1 client. Each request opens one TCP connection, sends one request, reads one
response, and closes the connection.

```silver
let client = import("http/client")
let println = import("io").println

try {
    let response = client.get("http://127.0.0.1:8080/")
    println(response.status, response.body)
} catch client.ClientError err {
    println("request failed:", err.message)
}
```

See the runnable [client example](../../examples/std_lib/std_client.slv), which works with the
[server example](../../examples/std_lib/std_server.slv).

## Requests

`Request` contains `method: str`, `url: str`, `headers: map`, and `body: str`.

| Function | Description |
| --- | --- |
| `new_request(method, url, headers, body) Request` | Build a request without sending it. |
| `send(request) Response \| ClientError` | Send an existing `Request`. |
| `request(method, url, headers, body) Response \| ClientError` | Build and send a configurable request. |
| `get(url) Response \| ClientError` | Send a GET with no custom headers or body. |
| `post(url, body) Response \| ClientError` | Send a POST with no custom headers. |

Use constants from `http` when selecting a method:

```silver
let http = import("http")
let client = import("http/client")

let response = client.request(
    http.MethodPost,
    "http://127.0.0.1:8080/echo",
    {"content-type": "text/plain", "x-request-id": "demo"},
    "Hello over HTTP"
)
```

The client normalizes outgoing header names to lowercase, supplies `Host`, and replaces `Content-Length` and
`Connection`. URLs must use `http://`; a missing port defaults to port 80.

## Responses

| `Response` field | Type | Description |
| --- | --- | --- |
| `status` | `str` | Code and reason, such as `"200 OK"`. |
| `status_code` | `int` | Numeric status code. |
| `reason` | `str` | Reason phrase returned by the server. |
| `headers` | `map` | Response headers with lowercase names. |
| `body` | `str` | Response body. |

`ClientError` contains `message: str`. It reports malformed URLs or responses and wraps connection, read, and write
failures from `networking`.

## Protocol limits

The current client supports plain HTTP/1.1 with `Content-Length` framing. It does not provide HTTPS/TLS, redirects,
chunked transfer encoding, EOF-delimited response bodies, timeouts, connection pooling, or automatic content
encoding.

[`http` constants](http.md) | [HTTP server](http_server.md) |
[Standard library index](../table_of_contents.md#standard-library)
