package stdlib_test

import (
	"silver/object"
	"testing"
)

const httpImport = `let http = import("http")
let client = import("http/client")
let server_module = import("http/server")
`

func TestHTTPClientAndServerRoundTrip(t *testing.T) {
	input := `let handler = fn(request: server_module.Request) server_module.Response {
    assert request.method == http.MethodPost
    assert request.path == "/submit?source=test"
    assert request.body == "silver"
    server_module.response_with_headers(http.StatusCreated, {"content-type": "text/plain"}, "created")
}
let server = server_module.new("127.0.0.1:0", handler)
let serve = task server.serve_once
let result = client.post("http://" + server.address + "/submit?source=test", "silver")
collect serve
assert result.status_code == http.StatusCreated
assert result.status == "201 Created"
assert result.headers["content-type"] == "text/plain"
result.body`

	evaluated := testEval(httpImport + input)
	result, ok := evaluated.(*object.String)
	if !ok || result.Value != "created" {
		if failure, isFailure := evaluated.(*object.Error); isFailure {
			t.Fatalf("round trip failed: %s", failure.MessageText())
		}
		t.Fatalf("round trip returned %#v, want %q", evaluated, "created")
	}
}

func TestHTTPClientRejectsUnsupportedScheme(t *testing.T) {
	input := `try {
    client.get("https://example.com/")
    False
} catch client.ClientError err {
    err.message == "only http:// URLs are supported"
}`
	testBooleanObject(t, testEval(httpImport+input), true)
}

func TestHTTPStatusClassesAndMethods(t *testing.T) {
	input := `http.is_informational(http.StatusContinue) &&
http.is_success(http.StatusNoContent) &&
http.is_redirection(http.StatusPermanentRedirect) &&
http.is_client_error(http.StatusNotFound) &&
http.is_server_error(http.StatusBadGateway) &&
http.MethodGet == "GET" && http.MethodPatch == "PATCH" &&
http.status_description(http.StatusUnprocessableEntity) == "Unprocessable Entity" &&
server_module.response(http.StatusUnprocessableEntity, "").reason == "Unprocessable Entity"`
	testBooleanObject(t, testEval(httpImport+input), true)
}
