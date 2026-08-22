# `http`

`http` provides method names, status codes, reason phrases, and status-class helpers shared by
[`http/client`](http_client.md) and [`http/server`](http_server.md).

```silver
let http = import("http")

http.MethodGet
http.StatusNotFound
http.status_description(http.StatusNotFound)
http.is_client_error(http.StatusNotFound)
```

## Methods

| Constant | Value |
| --- | --- |
| `MethodGet` | `GET` |
| `MethodHead` | `HEAD` |
| `MethodPost` | `POST` |
| `MethodPut` | `PUT` |
| `MethodPatch` | `PATCH` |
| `MethodDelete` | `DELETE` |
| `MethodConnect` | `CONNECT` |
| `MethodOptions` | `OPTIONS` |
| `MethodTrace` | `TRACE` |

## Status codes

`status_description(code)` returns the description shown below, or `"Status"` for an unknown code.

| Code | Constant | Description |
| ---: | --- | --- |
| 100 | `StatusContinue` | Continue |
| 101 | `StatusSwitchingProtocols` | Switching Protocols |
| 102 | `StatusProcessing` | Processing |
| 103 | `StatusEarlyHints` | Early Hints |
| 200 | `StatusOK` | OK |
| 201 | `StatusCreated` | Created |
| 202 | `StatusAccepted` | Accepted |
| 203 | `StatusNonAuthoritativeInfo` | Non-Authoritative Information |
| 204 | `StatusNoContent` | No Content |
| 205 | `StatusResetContent` | Reset Content |
| 206 | `StatusPartialContent` | Partial Content |
| 207 | `StatusMultiStatus` | Multi-Status |
| 208 | `StatusAlreadyReported` | Already Reported |
| 226 | `StatusIMUsed` | IM Used |
| 300 | `StatusMultipleChoices` | Multiple Choices |
| 301 | `StatusMovedPermanently` | Moved Permanently |
| 302 | `StatusFound` | Found |
| 303 | `StatusSeeOther` | See Other |
| 304 | `StatusNotModified` | Not Modified |
| 305 | `StatusUseProxy` | Use Proxy |
| 307 | `StatusTemporaryRedirect` | Temporary Redirect |
| 308 | `StatusPermanentRedirect` | Permanent Redirect |
| 400 | `StatusBadRequest` | Bad Request |
| 401 | `StatusUnauthorized` | Unauthorized |
| 402 | `StatusPaymentRequired` | Payment Required |
| 403 | `StatusForbidden` | Forbidden |
| 404 | `StatusNotFound` | Not Found |
| 405 | `StatusMethodNotAllowed` | Method Not Allowed |
| 406 | `StatusNotAcceptable` | Not Acceptable |
| 407 | `StatusProxyAuthRequired` | Proxy Authentication Required |
| 408 | `StatusRequestTimeout` | Request Timeout |
| 409 | `StatusConflict` | Conflict |
| 410 | `StatusGone` | Gone |
| 411 | `StatusLengthRequired` | Length Required |
| 412 | `StatusPreconditionFailed` | Precondition Failed |
| 413 | `StatusRequestEntityTooLarge` | Request Entity Too Large |
| 414 | `StatusRequestURITooLong` | Request URI Too Long |
| 415 | `StatusUnsupportedMediaType` | Unsupported Media Type |
| 416 | `StatusRequestedRangeNotSatisfiable` | Requested Range Not Satisfiable |
| 417 | `StatusExpectationFailed` | Expectation Failed |
| 418 | `StatusTeapot` | I'm a teapot |
| 421 | `StatusMisdirectedRequest` | Misdirected Request |
| 422 | `StatusUnprocessableEntity` | Unprocessable Entity |
| 423 | `StatusLocked` | Locked |
| 424 | `StatusFailedDependency` | Failed Dependency |
| 425 | `StatusTooEarly` | Too Early |
| 426 | `StatusUpgradeRequired` | Upgrade Required |
| 428 | `StatusPreconditionRequired` | Precondition Required |
| 429 | `StatusTooManyRequests` | Too Many Requests |
| 431 | `StatusRequestHeaderFieldsTooLarge` | Request Header Fields Too Large |
| 451 | `StatusUnavailableForLegalReasons` | Unavailable For Legal Reasons |
| 500 | `StatusInternalServerError` | Internal Server Error |
| 501 | `StatusNotImplemented` | Not Implemented |
| 502 | `StatusBadGateway` | Bad Gateway |
| 503 | `StatusServiceUnavailable` | Service Unavailable |
| 504 | `StatusGatewayTimeout` | Gateway Timeout |
| 505 | `StatusHTTPVersionNotSupported` | HTTP Version Not Supported |
| 506 | `StatusVariantAlsoNegotiates` | Variant Also Negotiates |
| 507 | `StatusInsufficientStorage` | Insufficient Storage |
| 508 | `StatusLoopDetected` | Loop Detected |
| 510 | `StatusNotExtended` | Not Extended |
| 511 | `StatusNetworkAuthenticationRequired` | Network Authentication Required |

## Status classes

| Function | True for |
| --- | --- |
| `is_informational(code: int)` | `100-199` |
| `is_success(code: int)` | `200-299` |
| `is_redirection(code: int)` | `300-399` |
| `is_client_error(code: int)` | `400-499` |
| `is_server_error(code: int)` | `500-599` |

[HTTP client](http_client.md) | [Cookies](http_cookies.md) | [Cookie jar](http_cookiejar.md) |
[HTTP server](http_server.md) |
[Standard library index](../table_of_contents.md#standard-library)
