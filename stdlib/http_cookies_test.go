package stdlib_test

import "testing"

func TestHTTPCookiesParseAndSerialize(t *testing.T) {
	input := `let cookies = import("http/cookies")
let cookie = cookies.parse_set_cookie(
    "session=abc==; Path=/app; Domain=Example.com; Max-Age=60; Secure; HttpOnly; SameSite=Lax"
)
assert cookie.name == "session"
assert cookie.value == "abc=="
assert cookie.path == "/app"
assert cookie.domain == "example.com"
assert cookie.max_age == 60
assert cookie.secure && cookie.http_only
assert cookie.same_site == cookies.SameSite.Lax
assert cookies.to_set_cookie(cookie) == "session=abc==; Path=/app; Domain=example.com; Max-Age=60; Secure; HttpOnly; SameSite=Lax"

let request_cookies = cookies.parse("theme=dark; session=abc==")
assert request_cookies[0].name == "theme"
assert request_cookies[1].value == "abc=="
cookies.to_cookie_header(request_cookies) == "theme=dark; session=abc=="`
	testBooleanObject(t, testEval(input), true)
}

func TestHTTPCookieJarMatchesAndRemovesCookies(t *testing.T) {
	input := `let jar_module = import("http/cookiejar")
let jar = jar_module.new()
jar.set_from_header("http://example.com/app/login", "session=one; Path=/app; HttpOnly")
jar.set_from_header("https://example.com/", "secure=yes; Secure")
jar.set_from_header("http://example.com/", "site=wide; Domain=example.com")

assert jar.header("http://example.com/app/page") == "session=one; site=wide"
assert jar.header("http://example.com/other") == "site=wide"
assert jar.header("http://example.com/application") == "site=wide"
assert jar.header("https://sub.example.com/") == "site=wide"
assert jar.header("https://example.com/") == "secure=yes; site=wide"

jar.set_from_header("http://example.com/app/login", "session=gone; Path=/app; Max-Age=0")
assert jar.header("http://example.com/app/page") == "site=wide"
jar.clear()
jar.header("http://example.com/") == ""`
	testBooleanObject(t, testEval(input), true)
}

func TestHTTPCookieErrorsAreTyped(t *testing.T) {
	input := `let cookies = import("http/cookies")
try {
    cookies.parse_set_cookie("invalid")
    False
} catch cookies.CookieError err {
    err.message != ""
}`
	testBooleanObject(t, testEval(input), true)
}
