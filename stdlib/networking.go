package stdlib

import (
	"fmt"
	"io"
	"net"
	"silver/ast"
	"silver/object"
	"sync"
)

// networkingDefinitions contains the TCP and UDP entry points exported by
// import("networking"). Connections themselves are ordinary Silver structs
// whose call fields close over the native socket.
func networkingDefinitions(null *object.Null) []definition {
	return []definition{
		{name: "dial", fn: builtinDial(null), signature: callSignature(
			[]string{"network", "address"},
			[]*ast.TypeAnnotation{namedType("str"), namedType("str")},
			namedType("Connection"),
			"ConnectionError",
		)},
		{name: "listen", fn: builtinListen(null), signature: callSignature(
			[]string{"network", "address"},
			[]*ast.TypeAnnotation{namedType("str"), namedType("str")},
			namedType("Listener"),
			"ListenError",
		)},
	}
}

// nativeConnection is shared by the callable fields of a Connection value.
// TCP uses stream; UDP uses packet plus a default peer resolved by dial.
type nativeConnection struct {
	mu             sync.RWMutex
	network        string
	stream         net.Conn
	packet         *net.UDPConn
	packetNetwork  string
	defaultAddress *net.UDPAddr
	closed         bool
	null           *object.Null
}

type nativeListener struct {
	mu       sync.RWMutex
	listener net.Listener
	closed   bool
	null     *object.Null
}

func builtinDial(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		network, err := requireString("dial", 0, args[0])
		if err != nil {
			return err
		}
		address, err := requireString("dial", 1, args[1])
		if err != nil {
			return err
		}

		switch network {
		case "tcp":
			connection, dialErr := net.Dial("tcp", address)
			if dialErr != nil {
				return networkingError("ConnectionError", dialErr)
			}
			state := &nativeConnection{network: network, stream: connection, null: null}
			return state.value(address)
		case "udp":
			remote, resolveErr := net.ResolveUDPAddr("udp", address)
			if resolveErr != nil {
				return networkingError("ConnectionError", resolveErr)
			}
			packetNetwork := udpNetworkForAddress(remote)
			local, localErr := udpLocalAddress(packetNetwork, remote)
			if localErr != nil {
				return networkingError("ConnectionError", localErr)
			}
			connection, listenErr := net.ListenUDP(packetNetwork, local)
			if listenErr != nil {
				return networkingError("ConnectionError", listenErr)
			}
			state := &nativeConnection{
				network:        network,
				packet:         connection,
				packetNetwork:  packetNetwork,
				defaultAddress: remote,
				null:           null,
			}
			return state.value(udpConnectionAddress(connection, packetNetwork))
		default:
			return networkingErrorMessage("ConnectionError", fmt.Sprintf("unsupported network %q: expected tcp or udp", network))
		}
	}
}

func builtinListen(null *object.Null) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 2); err != nil {
			return err
		}
		network, err := requireString("listen", 0, args[0])
		if err != nil {
			return err
		}
		address, err := requireString("listen", 1, args[1])
		if err != nil {
			return err
		}
		if network != "tcp" {
			return networkingErrorMessage("ListenError", fmt.Sprintf("unsupported listener network %q: expected tcp", network))
		}

		listener, listenErr := net.Listen("tcp", address)
		if listenErr != nil {
			return networkingError("ListenError", listenErr)
		}
		state := &nativeListener{listener: listener, null: null}
		definition, _ := object.BuiltinStructDefinitionByName("Listener")
		return &object.StructInstance{
			Struct: definition,
			Values: map[string]object.Object{
				"address": &object.String{Value: listener.Addr().String()},
				"accept":  &object.Builtin{Fn: state.accept, Signature: listenerAcceptSignature()},
				"close":   &object.Builtin{Fn: state.close, Signature: connectionCloseSignature()},
			},
		}
	}
}

// udpLocalAddress asks the routing table which source address reaches the
// default peer, then binds the datagram socket to that interface. This keeps
// Connection.address usable as a write_to destination instead of exposing an
// unspecified address such as 0.0.0.0.
func udpLocalAddress(network string, remote *net.UDPAddr) (*net.UDPAddr, error) {
	if remote.IP == nil {
		return nil, nil
	}
	probe, err := net.DialUDP(network, nil, remote)
	if err != nil {
		return nil, err
	}
	local := probe.LocalAddr().(*net.UDPAddr)
	if err := probe.Close(); err != nil {
		return nil, err
	}
	return &net.UDPAddr{IP: local.IP, Zone: local.Zone}, nil
}

func udpNetworkForAddress(address *net.UDPAddr) string {
	if address.IP == nil {
		return "udp"
	}
	if address.IP.To4() != nil {
		return "udp4"
	}
	return "udp6"
}

func udpConnectionAddress(connection *net.UDPConn, network string) string {
	address := connection.LocalAddr().(*net.UDPAddr)
	if address.IP != nil && !address.IP.IsUnspecified() {
		return address.String()
	}
	if network == "udp4" {
		return (&net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: address.Port}).String()
	}
	return (&net.UDPAddr{IP: net.IPv6loopback, Port: address.Port}).String()
}

func (connection *nativeConnection) value(address string) *object.StructInstance {
	definition, _ := object.BuiltinStructDefinitionByName("Connection")
	return &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{
			"address":   &object.String{Value: address},
			"read":      &object.Builtin{Fn: connection.read, Signature: connectionReadSignature()},
			"write":     &object.Builtin{Fn: connection.write, Signature: connectionWriteSignature()},
			"write_to":  &object.Builtin{Fn: connection.writeTo, Signature: connectionWriteToSignature()},
			"read_from": &object.Builtin{Fn: connection.readFrom, Signature: connectionReadFromSignature()},
			"close":     &object.Builtin{Fn: connection.close, Signature: connectionCloseSignature()},
		},
	}
}

func (connection *nativeConnection) read(args ...object.Object) object.Object {
	size, err := networkingBufferSize("Connection.read", args)
	if err != nil {
		return err
	}
	if connection.isClosed() {
		return networkingErrorMessage("ReadError", "connection is closed")
	}

	buffer := make([]byte, size)
	var count int
	var readErr error
	if connection.network == "tcp" {
		count, readErr = connection.stream.Read(buffer)
	} else {
		count, _, readErr = connection.packet.ReadFromUDP(buffer)
	}
	if readErr != nil {
		return networkingError("ReadError", readErr)
	}
	return &object.String{Value: string(buffer[:count])}
}

func (connection *nativeConnection) write(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 1); err != nil {
		return err
	}
	data, err := requireString("Connection.write", 0, args[0])
	if err != nil {
		return err
	}
	if connection.isClosed() {
		return networkingErrorMessage("WriteError", "connection is closed")
	}

	var count int
	var writeErr error
	if connection.network == "tcp" {
		count, writeErr = io.WriteString(connection.stream, data)
	} else {
		count, writeErr = connection.packet.WriteToUDP([]byte(data), connection.defaultAddress)
	}
	if writeErr != nil {
		return networkingError("WriteError", writeErr)
	}
	if count != len(data) {
		return networkingError("WriteError", io.ErrShortWrite)
	}
	return connection.null
}

func (connection *nativeConnection) writeTo(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	data, err := requireString("Connection.write_to", 0, args[0])
	if err != nil {
		return err
	}
	address, err := requireString("Connection.write_to", 1, args[1])
	if err != nil {
		return err
	}
	if connection.network != "udp" {
		return networkingErrorMessage("WriteError", "write_to is only available on udp connections")
	}
	if connection.isClosed() {
		return networkingErrorMessage("WriteError", "connection is closed")
	}

	remote, resolveErr := net.ResolveUDPAddr(connection.packetNetwork, address)
	if resolveErr != nil {
		return networkingError("WriteError", resolveErr)
	}
	count, writeErr := connection.packet.WriteToUDP([]byte(data), remote)
	if writeErr != nil {
		return networkingError("WriteError", writeErr)
	}
	if count != len(data) {
		return networkingError("WriteError", io.ErrShortWrite)
	}
	return connection.null
}

func (connection *nativeConnection) readFrom(args ...object.Object) object.Object {
	size, err := networkingBufferSize("Connection.read_from", args)
	if err != nil {
		return err
	}
	if connection.network != "udp" {
		return networkingErrorMessage("ReadError", "read_from is only available on udp connections")
	}
	if connection.isClosed() {
		return networkingErrorMessage("ReadError", "connection is closed")
	}

	buffer := make([]byte, size)
	count, address, readErr := connection.packet.ReadFromUDP(buffer)
	if readErr != nil {
		return networkingError("ReadError", readErr)
	}
	definition, _ := object.BuiltinStructDefinitionByName("ReadFromResult")
	return &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{
			"data":    &object.String{Value: string(buffer[:count])},
			"address": &object.String{Value: address.String()},
		},
	}
}

func (connection *nativeConnection) close(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	connection.mu.Lock()
	if connection.closed {
		connection.mu.Unlock()
		return networkingErrorMessage("ConnectionError", "connection is closed")
	}
	connection.closed = true
	connection.mu.Unlock()

	var closeErr error
	if connection.network == "tcp" {
		closeErr = connection.stream.Close()
	} else {
		closeErr = connection.packet.Close()
	}
	if closeErr != nil {
		return networkingError("ConnectionError", closeErr)
	}
	return connection.null
}

func (connection *nativeConnection) isClosed() bool {
	connection.mu.RLock()
	defer connection.mu.RUnlock()
	return connection.closed
}

func (listener *nativeListener) accept(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	if listener.isClosed() {
		return networkingErrorMessage("ConnectionError", "listener is closed")
	}
	connection, acceptErr := listener.listener.Accept()
	if acceptErr != nil {
		return networkingError("ConnectionError", acceptErr)
	}
	state := &nativeConnection{network: "tcp", stream: connection, null: listener.null}
	return state.value(connection.RemoteAddr().String())
}

func (listener *nativeListener) close(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	listener.mu.Lock()
	if listener.closed {
		listener.mu.Unlock()
		return networkingErrorMessage("ConnectionError", "listener is closed")
	}
	listener.closed = true
	listener.mu.Unlock()
	if err := listener.listener.Close(); err != nil {
		return networkingError("ConnectionError", err)
	}
	return listener.null
}

func (listener *nativeListener) isClosed() bool {
	listener.mu.RLock()
	defer listener.mu.RUnlock()
	return listener.closed
}

func networkingBufferSize(name string, args []object.Object) (int, *object.Error) {
	if err := requireArgumentCount(args, 1); err != nil {
		return 0, err
	}
	size, err := requireInteger(name, args[0])
	if err != nil {
		return 0, err
	}
	if size < 0 {
		return 0, newError(object.RuntimeErrorKindValue, "argument to `%s` must be nonnegative", name)
	}
	converted := int(size)
	if int64(converted) != size {
		return 0, newError(object.RuntimeErrorKindValue, "argument to `%s` is too large", name)
	}
	return converted, nil
}

func networkingError(name string, err error) *object.StructInstance {
	return networkingErrorMessage(name, err.Error())
}

func networkingErrorMessage(name, message string) *object.StructInstance {
	definition, _ := object.BuiltinStructDefinitionByName(name)
	return &object.StructInstance{
		Struct: definition,
		Values: map[string]object.Object{"message": &object.String{Value: message}},
	}
}

func connectionReadSignature() *ast.TypeAnnotation {
	return callSignature([]string{"bytes"}, []*ast.TypeAnnotation{namedType("int")}, namedType("str"), "ReadError")
}

func connectionWriteSignature() *ast.TypeAnnotation {
	return callSignature([]string{"data"}, []*ast.TypeAnnotation{namedType("str")}, nil, "WriteError")
}

func connectionWriteToSignature() *ast.TypeAnnotation {
	return callSignature(
		[]string{"data", "address"},
		[]*ast.TypeAnnotation{namedType("str"), namedType("str")},
		nil,
		"WriteError",
	)
}

func connectionReadFromSignature() *ast.TypeAnnotation {
	return callSignature([]string{"bytes"}, []*ast.TypeAnnotation{namedType("int")}, namedType("ReadFromResult"), "ReadError")
}

func connectionCloseSignature() *ast.TypeAnnotation {
	return callSignature(nil, nil, nil, "ConnectionError")
}

func listenerAcceptSignature() *ast.TypeAnnotation {
	return callSignature(nil, nil, namedType("Connection"), "ConnectionError")
}
