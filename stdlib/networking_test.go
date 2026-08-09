package stdlib_test

import (
	"silver/object"
	"strings"
	"testing"
)

const networkingImport = `let net = import("networking")
`

func TestNetworkingTCPRoundTrip(t *testing.T) {
	input := `let listener: Listener = net.listen("tcp", "127.0.0.1:0")
let echo = fn() {
    let connection: Connection = listener.accept()
    let data = connection.read(1024)
    connection.write(data)
    connection.close()
}
let server = task echo
let connection: Connection = net.dial("tcp", listener.address)
connection.write("hello over tcp")
let response = connection.read(1024)
connection.close()
collect server
listener.close()
response`

	result, ok := testEval(networkingImport + input).(*object.String)
	if !ok {
		t.Fatalf("result is %T (%v), want a string", result, result)
	}
	if got, want := result.Value, "hello over tcp"; got != want {
		t.Fatalf("response is %q, want %q", got, want)
	}
}

func TestNetworkingUDPRoundTrip(t *testing.T) {
	input := `let receiver = net.dial("udp", "127.0.0.1:9")
let sender = net.dial("udp", "127.0.0.1:9")
sender.write_to("hello over udp", receiver.address)
let packet: ReadFromResult = receiver.read_from(1024)
sender.close()
receiver.close()
packet.data`

	result, ok := testEval(networkingImport + input).(*object.String)
	if !ok {
		t.Fatalf("result is %T (%v), want a string", result, result)
	}
	if got, want := result.Value, "hello over udp"; got != want {
		t.Fatalf("packet data is %q, want %q", got, want)
	}
}

func TestNetworkingUDPReadFromReportsSender(t *testing.T) {
	input := `let receiver = net.dial("udp", "127.0.0.1:9")
let sender = net.dial("udp", "127.0.0.1:9")
sender.write_to("packet", receiver.address)
let packet = receiver.read_from(64)
let matches = packet.address == sender.address
sender.close()
receiver.close()
matches`
	testBooleanObject(t, testEval(networkingImport+input), true)
}

func TestNetworkingUDPWriteUsesDefaultPeer(t *testing.T) {
	input := `let receiver = net.dial("udp", "127.0.0.1:9")
let sender = net.dial("udp", receiver.address)
sender.write("default peer")
let packet = receiver.read_from(64)
sender.close()
receiver.close()
packet.data`
	result, ok := testEval(networkingImport + input).(*object.String)
	if !ok || result.Value != "default peer" {
		t.Fatalf("default-peer packet is %#v, want %q", result, "default peer")
	}
}

func TestNetworkingExposesDeclaredSignatures(t *testing.T) {
	input := `let dialer: call(network: str, address: str) Connection | ConnectionError = net.dial
let listener_factory: call(network: str, address: str) Listener | ListenError = net.listen
let connection = net.dial("udp", "127.0.0.1:9")
let reader: call(bytes: int) str | ReadError = connection.read
let writer: call(data: str) | WriteError = connection.write
let write_to: call(data: str, address: str) | WriteError = connection.write_to
let read_from: call(bytes: int) ReadFromResult | ReadError = connection.read_from
let closer: call() | ConnectionError = connection.close
connection.close()
True`
	testBooleanObject(t, testEval(networkingImport+input), true)
}

func TestNetworkingOperationsReturnTypedErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "dial",
			input: `try {
net.dial("invalid", "localhost:80")
False
} catch ConnectionError err {
err.message != ""
}`,
		},
		{
			name: "listen",
			input: `try {
net.listen("udp", "127.0.0.1:0")
False
} catch ListenError err {
err.message != ""
}`,
		},
		{
			name: "read after close",
			input: `let connection = net.dial("udp", "127.0.0.1:9")
connection.close()
try {
connection.read(1)
False
} catch ReadError err {
err.message != ""
}`,
		},
		{
			name: "write after close",
			input: `let connection = net.dial("udp", "127.0.0.1:9")
connection.close()
try {
connection.write("data")
False
} catch WriteError err {
err.message != ""
}`,
		},
		{
			name: "close twice",
			input: `let connection = net.dial("udp", "127.0.0.1:9")
connection.close()
try {
connection.close()
False
} catch ConnectionError err {
err.message != ""
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testBooleanObject(t, testEval(networkingImport+tt.input), true)
		})
	}
}

func TestNetworkingRejectsInvalidArguments(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{input: `net.dial("tcp")`, message: "wrong number of arguments. got=1, want=2"},
		{input: `net.dial(1, "localhost:80")`, message: "argument 1 to `dial` must be STRING, got INTEGER"},
		{input: `net.listen("tcp", 1)`, message: "argument 2 to `listen` must be STRING, got INTEGER"},
		{input: `let connection = net.dial("udp", "127.0.0.1:9")
connection.read(-1)`, message: "argument to `Connection.read` must be nonnegative"},
	}

	for _, tt := range tests {
		result, ok := testEval(networkingImport + tt.input).(*object.Error)
		if !ok {
			t.Fatalf("%s returned %T, want *object.Error", tt.input, result)
		}
		if result.MessageText() != tt.message {
			t.Fatalf("error is %q, want %q", result.MessageText(), tt.message)
		}
	}
}

func TestNetworkingConnectionProtocolErrors(t *testing.T) {
	input := `let listener = net.listen("tcp", "127.0.0.1:0")
let accept_once = fn() {
    let connection = listener.accept()
    connection.close()
}
let server = task accept_once
let connection = net.dial("tcp", listener.address)
let handled = try {
    connection.write_to("data", listener.address)
    False
} catch WriteError err {
    err.message != ""
}
connection.close()
collect server
listener.close()
handled`
	testBooleanObject(t, testEval(networkingImport+input), true)

	result := testEval(networkingImport + `try {
net.listen("tcp", "not-an-address")
} catch ListenError err {
err.message
}`)
	message, ok := result.(*object.String)
	if !ok || strings.TrimSpace(message.Value) == "" {
		t.Fatalf("listen error is %#v, want a non-empty message", result)
	}
}
