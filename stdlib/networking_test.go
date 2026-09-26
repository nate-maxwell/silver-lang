package stdlib_test

import (
	"silver/object"
	"strings"
	"testing"
)

const networkingImport = `let net = import("core:networking")
`

func TestNetworkingExportsDocumentedTypes(t *testing.T) {
	evaluated := testEval(`import("core:networking")`)
	module, ok := evaluated.(*object.Module)
	if !ok {
		t.Fatalf("import failed: %s", evaluated.Inspect())
	}
	for _, name := range []string{"Connection", "Listener", "ReadFromResult", "ConnectionError", "ListenError", "ReadError", "WriteError"} {
		definition, _ := object.BuiltinStructDefinitionByName(name)
		if module.Exports[name] != definition {
			t.Errorf("networking.%s does not expose the socket's nominal type", name)
		}
	}
	if _, exists := module.Exports["_native"]; exists {
		t.Fatal("implementation module leaked into public exports")
	}
	testBooleanObject(t, testEval(networkingImport+`
let same_module = import("core:networking")
assert net == same_module
try {
    net.dial_tcp("not-an-address")
    False
} catch net.ConnectionError err {
    err.message != ""
}`), true)
}

func TestNetworkingTCPRoundTrip(t *testing.T) {
	input := `let listener: net.Listener = net.listen("127.0.0.1:0")
let connection: net.Connection = net.dial(net.Network.TCP, listener.address)
let peer: net.Connection = listener.accept()
connection.write("hello over tcp")
let data = peer.read(1024)
peer.write(data)
peer.close()
let response = connection.read(1024)
connection.close()
listener.close()
response`

	for name, dial := range map[string]string{
		"generic": `net.dial(net.Network.TCP, listener.address)`,
		"tcp":     `net.dial_tcp(listener.address)`,
	} {
		t.Run(name, func(t *testing.T) {
			program := strings.Replace(input, `net.dial(net.Network.TCP, listener.address)`, dial, 1)
			evaluated := testEval(networkingImport + program)
			result, ok := evaluated.(*object.String)
			if !ok || result.Value != "hello over tcp" {
				t.Fatalf("response is %s, want hello over tcp", evaluated.Inspect())
			}
		})
	}
}

func TestNetworkingUDPRoundTrip(t *testing.T) {
	input := `let receiver = net.dial(net.Network.UDP, "127.0.0.1:9")
let sender = net.dial(net.Network.UDP, "127.0.0.1:9")
sender.write_to("hello over udp", receiver.address)
let packet: net.ReadFromResult = receiver.read_from(1024)
sender.close()
receiver.close()
packet.data`

	for name, dial := range map[string]string{
		"generic": `net.dial(net.Network.UDP, `,
		"udp":     `net.dial_udp(`,
	} {
		t.Run(name, func(t *testing.T) {
			program := strings.ReplaceAll(input, `net.dial(net.Network.UDP, `, dial)
			evaluated := testEval(networkingImport + program)
			result, ok := evaluated.(*object.String)
			if !ok || result.Value != "hello over udp" {
				t.Fatalf("packet data is %s, want hello over udp", evaluated.Inspect())
			}
		})
	}
}

func TestNetworkingUDPReadFromReportsSender(t *testing.T) {
	input := `let receiver = net.dial(net.Network.UDP, "127.0.0.1:9")
let sender = net.dial(net.Network.UDP, "127.0.0.1:9")
sender.write_to("packet", receiver.address)
let packet = receiver.read_from(64)
let matches = packet.address == sender.address
sender.close()
receiver.close()
matches`
	testBooleanObject(t, testEval(networkingImport+input), true)
}

func TestNetworkingUDPWriteUsesDefaultPeer(t *testing.T) {
	input := `let receiver = net.dial(net.Network.UDP, "127.0.0.1:9")
let sender = net.dial(net.Network.UDP, receiver.address)
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
	input := `let dialer: call(network: net.Network, address: str) net.Connection | net.ConnectionError = net.dial
let tcp_dialer: call(address: str) net.Connection | net.ConnectionError = net.dial_tcp
let udp_dialer: call(address: str) net.Connection | net.ConnectionError = net.dial_udp
let listener_factory: call(address: str) net.Listener | net.ListenError = net.listen
let connection = net.dial(net.Network.UDP, "127.0.0.1:9")
let reader: call(bytes: int) str | net.ReadError = connection.read
let writer: call(data: str) | net.WriteError = connection.write
let write_to: call(data: str, address: str) | net.WriteError = connection.write_to
let read_from: call(bytes: int) net.ReadFromResult | net.ReadError = connection.read_from
let closer: call() | net.ConnectionError = connection.close
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
net.dial(net.Network.TCP, "not-an-address")
False
} catch ConnectionError err {
err.message != ""
}`,
		},
		{
			name: "listen",
			input: `try {
net.listen("not-an-address")
False
} catch ListenError err {
err.message != ""
}`,
		},
		{
			name: "read after close",
			input: `let connection = net.dial(net.Network.UDP, "127.0.0.1:9")
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
			input: `let connection = net.dial(net.Network.UDP, "127.0.0.1:9")
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
			input: `let connection = net.dial(net.Network.UDP, "127.0.0.1:9")
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
		{input: `net.dial(net.Network.TCP)`, message: "wrong number of arguments. got=1, want=2"},
		{input: `net.dial("tcp", "localhost:80")`, message: `type mismatch for parameter "network": expected Network, got str`},
		{input: `net.listen(1)`, message: "argument 1 to `listen` must be STRING, got INTEGER"},
		{input: `let connection = net.dial(net.Network.UDP, "127.0.0.1:9")
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
	input := `let listener = net.listen("127.0.0.1:0")
let connection = net.dial(net.Network.TCP, listener.address)
let peer = listener.accept()
peer.close()
let handled = try {
    connection.write_to("data", listener.address)
    False
} catch WriteError err {
    err.message != ""
}
connection.close()
listener.close()
handled`
	testBooleanObject(t, testEval(networkingImport+input), true)

	result := testEval(networkingImport + `try {
net.listen("not-an-address")
} catch ListenError err {
err.message
}`)
	message, ok := result.(*object.String)
	if !ok || strings.TrimSpace(message.Value) == "" {
		t.Fatalf("listen error is %#v, want a non-empty message", result)
	}
}
