package stdlib_test

import (
	"fmt"
	"silver/evaluator"
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
    return server_module.response_with_headers(http.StatusCreated, {"content-type": "text/plain"}, "created")
}
server_module.new("127.0.0.1:0", handler)`
	serverValue := testEval(httpImport + input)
	server, ok := serverValue.(*object.StructInstance)
	if !ok {
		t.Fatalf("server creation returned %T (%v), want a server", serverValue, serverValue)
	}
	address, _ := server.Get("address")
	listener, _ := server.Get("listener")
	closeListener, _ := listener.(*object.StructInstance).Get("close")

	program, parseError := evaluator.ParseSource("http_server_test.slv", []byte("server.serve_once()"))
	if parseError != nil {
		t.Fatal(parseError.Inspect())
	}
	env := object.NewEnvironment()
	env.Set("server", server)
	serverEvaluator := evaluator.New()
	// The Go harness runs the blocking server while a separate interpreter
	// makes the client request. Each Silver program executes synchronously.
	served := make(chan object.Object, 1)
	done := make(chan struct{})
	go func() {
		defer close(done)
		served <- serverEvaluator.Eval(program, env)
	}()
	t.Cleanup(func() {
		closeListener.(*object.Builtin).Fn()
		<-done
	})

	input = fmt.Sprintf(`let result = client.post(%q, "silver")
assert result.status_code == http.StatusCreated
assert result.status == "201 Created"
assert result.headers["content-type"] == "text/plain"
result.body`, "http://"+address.(*object.String).Value+"/submit?source=test")

	evaluated := testEval(httpImport + input)
	result, ok := evaluated.(*object.String)
	if !ok || result.Value != "created" {
		if failure, isFailure := evaluated.(*object.Error); isFailure {
			t.Fatalf("round trip failed: %s", failure.MessageText())
		}
		t.Fatalf("round trip returned %#v, want %q", evaluated, "created")
	}
	if failure, ok := (<-served).(*object.Error); ok {
		t.Fatalf("server failed: %s", failure.Inspect())
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
