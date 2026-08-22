# `http/cookiejar`

`http/cookiejar` stores cookies in memory and selects the cookies applicable to an HTTP or HTTPS URL.

```silver
let cookiejar = import("http/cookiejar")

let jar = cookiejar.new()
jar.set_from_header(
    "https://example.com/account/login",
    "session=abc123; Path=/account; Secure; HttpOnly"
)

let header = jar.header("https://example.com/account/profile")
# session=abc123
```

See the complete runnable [cookie-jar example](../../examples/std_lib/std_cookiejar.slv).

## Jar operations

`new() Jar` creates an empty jar. A `Jar` provides these bound methods:

| Method | Description |
| --- | --- |
| `set(url, cookie) \| JarError` | Store a parsed or programmatically built `http/cookies.Cookie`. |
| `set_from_header(url, header) \| JarError` | Parse and store one `Set-Cookie` value received from `url`. |
| `header(url) str \| JarError` | Return the matching request `Cookie` header value. |
| `clear()` | Remove every stored cookie. |

`cookies` is the jar's backing array. Prefer the methods above because its elements also contain the internal origin,
path, and expiration metadata used for matching.

## Matching behavior

When a cookie is stored, the jar:

- makes a cookie without Domain host-only;
- validates an explicit Domain against the source URL;
- uses the request directory when Path is absent or invalid;
- replaces an existing cookie with the same name, domain, and path;
- converts a positive `Max-Age` to an expiration time; and
- removes the matching cookie when `max_age` is negative, including parsed `Max-Age=0` cookies.

When `header(url)` is called, expired cookies are discarded. Host-only cookies require an exact host. Domain cookies
also match subdomains, Path limits requests by prefix, and Secure cookies are emitted only for `https://` URLs.

`JarError` contains `message: str`. It reports unsupported or malformed URLs, Domain attributes that do not match the
source host, and cookie parsing or serialization failures.

## Limits

The jar is in-memory and is not persisted. It does not implement a public-suffix list, IDNA canonicalization, IPv6 URL
authorities, capacity eviction, Expires-date evaluation, or SameSite policy enforcement. SameSite and HttpOnly remain
available on the stored `Cookie` for applications that need to inspect them.

[Cookie parsing](http_cookies.md) | [HTTP client](http_client.md) |
[Standard library index](../table_of_contents.md#standard-library)
