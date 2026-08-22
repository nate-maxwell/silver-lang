# `http/cookies`

`http/cookies` parses and serializes HTTP `Cookie` and `Set-Cookie` header values.

```silver
let cookies = import("http/cookies")

let session = cookies.parse_set_cookie(
    "session=abc123; Path=/account; Max-Age=3600; HttpOnly; SameSite=Lax"
)
session.secure = True

let header = cookies.to_set_cookie(session)
```

See the complete runnable [cookie example](../../examples/std_lib/std_cookies.slv).

## Cookie values

| `Cookie` field | Type | Description |
| --- | --- | --- |
| `name` | `str` | Cookie name. |
| `value` | `str` | Cookie value. |
| `path` | `str` | Path attribute, or `""` when absent. |
| `domain` | `str` | Lowercase Domain attribute, or `""` for a host-only cookie. |
| `expires` | `str` | Expires attribute exactly as received, or `""`. |
| `max_age` | `int` | Positive lifetime in seconds, `0` when absent, or negative for deletion. |
| `secure` | `bool` | Whether the Secure attribute is present. |
| `http_only` | `bool` | Whether the HttpOnly attribute is present. |
| `same_site` | `SameSite` | `Default`, `Lax`, `Strict`, or `None`. |

`new(name, value) Cookie | CookieError` validates the name and value and supplies default attributes. Fields on the
returned struct are mutable, so attributes can be assigned before serialization.

## Parsing and serialization

| Function | Description |
| --- | --- |
| `parse(header) array \| CookieError` | Parse a request `Cookie` header into cookies in header order. |
| `parse_set_cookie(header) Cookie \| CookieError` | Parse one response `Set-Cookie` value. |
| `to_cookie_header(cookies) str \| CookieError` | Build a request `Cookie` value. |
| `to_set_cookie(cookie) str \| CookieError` | Build one response `Set-Cookie` value. |

`parse_set_cookie` recognizes `Path`, `Domain`, `Expires`, `Max-Age`, `Secure`, `HttpOnly`, and `SameSite`
case-insensitively. Unknown extension attributes are ignored. `Max-Age=0` becomes a negative `max_age` so a
[`cookiejar`](http_cookiejar.md) can remove the stored cookie.

`CookieError` contains `message: str` and reports invalid names, values, pairs, and `Max-Age` values.

## Limits

Quoted values have their surrounding quotes removed, but quoted-string escape processing is not provided. Expires is
preserved for round-trip serialization; the cookie jar currently expires cookies through `Max-Age`, not by parsing
Expires dates.

[Cookie jar](http_cookiejar.md) | [`http` constants](http.md) |
[Standard library index](../table_of_contents.md#standard-library)
