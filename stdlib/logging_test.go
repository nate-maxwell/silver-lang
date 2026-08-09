package stdlib_test

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const loggingImport = `let logging = import("logging")
`

func TestLoggingStreamNullAndLogger(t *testing.T) {
	var stdout, stderr bytes.Buffer
	input := loggingImport + `let io = import("io")
let stream = logging.new_stream_handler(io.stdout)
stream.handle("direct")
let discard = logging.new_null_handler()
discard.handle("discarded")
let logger = logging.new_logger(logging.Level.Info, logging.standard_format, stream)
logger.handle("formatted")
True`
	testBooleanObject(t, testEvalWithStreams(input, strings.NewReader(""), &stdout, &stderr), true)
	if got := stdout.String(); !strings.HasPrefix(got, "direct\n") || !strings.Contains(got, "[Level.Info]") || !strings.HasSuffix(got, " formatted\n") {
		t.Fatalf("stdout is %q, want direct and formatted records", got)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr is %q, want no output", stderr.String())
	}
}

func TestLoggingFileHandlerAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	input := loggingImport + `let path = import("path")
let handler = logging.new_file_handler(path.new(` + silverString(path) + `))
handler.handle("first")
handler.handle("second")`
	testNullObject(t, testEval(input))
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(contents), "first\nsecond\n"; got != want {
		t.Fatalf("file contents are %q, want %q", got, want)
	}
}

func TestLoggingRotatingFileHandler(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "app.log")
	input := loggingImport + `let path = import("path")
let handler = logging.new_rotating_file_handler(path.new(` + silverString(path) + `), 0)
handler.handle("first")
handler.handle("second")`
	testNullObject(t, testEval(input))

	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(contents), "second\n"; got != want {
		t.Fatalf("active file contents are %q, want %q", got, want)
	}
	rotated, err := filepath.Glob(path + ".*")
	if err != nil {
		t.Fatal(err)
	}
	if len(rotated) != 1 {
		t.Fatalf("rotated files are %v, want one", rotated)
	}
	contents, err = os.ReadFile(rotated[0])
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(contents), "first\n"; got != want {
		t.Fatalf("rotated file contents are %q, want %q", got, want)
	}
}

func TestLoggingTCPSocketHandler(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	received := make(chan string, 1)
	go func() {
		connection, acceptErr := listener.Accept()
		if acceptErr != nil {
			received <- "accept error: " + acceptErr.Error()
			return
		}
		defer connection.Close()
		buffer := make([]byte, 64)
		count, readErr := connection.Read(buffer)
		if readErr != nil {
			received <- "read error: " + readErr.Error()
			return
		}
		received <- string(buffer[:count])
	}()

	input := loggingImport + `let handler = logging.new_tcp_socket_handler(` + silverString(listener.Addr().String()) + `)
handler.handle("tcp record")`
	testNullObject(t, testEval(input))
	select {
	case got := <-received:
		if want := "tcp record\n"; got != want {
			t.Fatalf("TCP record is %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for TCP record")
	}
}

func TestLoggingUDPSocketHandler(t *testing.T) {
	connection, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	input := loggingImport + `let handler = logging.new_udp_socket_handler(` + silverString(connection.LocalAddr().String()) + `)
handler.handle("udp record")`
	testNullObject(t, testEval(input))
	if err := connection.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 64)
	count, _, err := connection.ReadFrom(buffer)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(buffer[:count]), "udp record\n"; got != want {
		t.Fatalf("UDP record is %q, want %q", got, want)
	}
}
