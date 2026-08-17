package stdlib

import (
	"fmt"
	"io"
	"os"
	"silver/ast"
	"silver/object"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/term"
)

const escape = "\x1b"

type terminalState struct {
	mu       sync.Mutex
	in       io.Reader
	out      io.Writer
	null     *object.Null
	rawState *term.State
}

func terminalDefinitions(in io.Reader, out io.Writer, null *object.Null) []definition {
	state := &terminalState{in: in, out: out, null: null}
	return []definition{
		{name: "Style", value: terminalStyleEnum()},
		{name: "width", fn: state.width, signature: terminalSignature(nil, nil, "int")},
		{name: "height", fn: state.height, signature: terminalSignature(nil, nil, "int")},
		{name: "flush", fn: state.flush, signature: terminalSignature(nil, nil, "")},
		{name: "clear", fn: state.output("clear", escape+"[2J"+escape+"[H"), signature: terminalSignature(nil, nil, "")},
		{name: "colored", fn: state.colored, signature: terminalSignature([]string{"text", "color"}, []string{"str", "str"}, "")},
		{name: "style", fn: state.style, signature: terminalSignature([]string{"text", "style"}, []string{"str", "Style"}, "")},
		{name: "cursor_move", fn: state.cursorMove, signature: terminalSignature([]string{"x", "y"}, []string{"int", "int"}, "")},
		{name: "cursor_hide", fn: state.output("cursor_hide", escape+"[?25l"), signature: terminalSignature(nil, nil, "")},
		{name: "cursor_show", fn: state.output("cursor_show", escape+"[?25h"), signature: terminalSignature(nil, nil, "")},
		{name: "cursor_save", fn: state.cursorSave, signature: terminalSignature(nil, nil, "array")},
		{name: "read_key", fn: state.readKey, signature: terminalSignature(nil, nil, "str")},
		{name: "raw_mode", fn: state.rawMode, signature: terminalSignature(nil, nil, "")},
		{name: "normal_mode", fn: state.normalMode, signature: terminalSignature(nil, nil, "")},
		{name: "clear_line", fn: state.output("clear_line", escape+"[0K"), signature: terminalSignature(nil, nil, "")},
		{name: "alternate_screen", fn: state.output("alternate_screen", escape+"[?1049h"), signature: terminalSignature(nil, nil, "")},
		{name: "main_screen", fn: state.output("main_screen", escape+"[?1049l"), signature: terminalSignature(nil, nil, "")},
	}
}

func terminalStyleEnum() *object.Enum {
	style := &object.Enum{Name: "Style"}
	style.Members = make(map[string]*object.EnumValue, 5)
	for index, name := range []string{"Normal", "Bold", "Underline", "Italic", "Dim"} {
		style.Members[name] = &object.EnumValue{
			EnumName: "Style",
			Member:   name,
			// Native enum IDs count down from the top of uint64 space, while
			// source enum IDs count upward from 1 within an evaluator session.
			HashID: ^uint64(0) - uint64(index),
			Enum:   style,
		}
	}
	return style
}

func terminalSignature(parameterNames, parameterTypes []string, returnType string) *ast.TypeAnnotation {
	types := make([]*ast.TypeAnnotation, len(parameterTypes))
	for index, name := range parameterTypes {
		types[index] = namedType(name)
	}
	var result *ast.TypeAnnotation
	if returnType != "" {
		result = namedType(returnType)
	}
	return callSignature(parameterNames, types, result)
}

func (state *terminalState) width(args ...object.Object) object.Object {
	return state.dimension("width", "COLUMNS", true, args)
}

func (state *terminalState) height(args ...object.Object) object.Object {
	return state.dimension("height", "LINES", false, args)
}

func (state *terminalState) dimension(name, environment string, wantWidth bool, args []object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	if fd, ok := fileDescriptor(state.out); ok && term.IsTerminal(fd) {
		width, height, err := term.GetSize(fd)
		if err != nil {
			return terminalError("%s failed: %s", name, err)
		}
		if wantWidth {
			return &object.Integer{Value: int64(width)}
		}
		return &object.Integer{Value: int64(height)}
	}
	if value, err := strconv.ParseInt(os.Getenv(environment), 10, 64); err == nil && value > 0 {
		return &object.Integer{Value: value}
	}
	return terminalError("could not determine terminal %s: output is not a terminal", name)
}

type errorFlusher interface{ Flush() error }
type flusher interface{ Flush() }

func (state *terminalState) flush(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if err := state.flushLocked(); err != nil {
		return terminalError("flush failed: %s", err)
	}
	return state.null
}

func (state *terminalState) flushLocked() error {
	switch writer := state.out.(type) {
	case errorFlusher:
		return writer.Flush()
	case flusher:
		writer.Flush()
	}
	return nil
}

func (state *terminalState) output(name, sequence string) object.BuiltinFunction {
	return func(args ...object.Object) object.Object {
		if err := requireArgumentCount(args, 0); err != nil {
			return err
		}
		state.mu.Lock()
		defer state.mu.Unlock()
		if err := writeAll(state.out, sequence); err != nil {
			return terminalError("%s failed: %s", name, err)
		}
		return state.null
	}
}

func (state *terminalState) colored(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	text, err := requireString("colored", 0, args[0])
	if err != nil {
		return err
	}
	color, err := requireString("colored", 1, args[1])
	if err != nil {
		return err
	}
	if len(color) != 7 || color[0] != '#' {
		return newError(object.RuntimeErrorKindValue, "argument 2 to `colored` must be a hex color in #RRGGBB form")
	}
	value, parseErr := strconv.ParseUint(color[1:], 16, 24)
	if parseErr != nil {
		return newError(object.RuntimeErrorKindValue, "argument 2 to `colored` must be a hex color in #RRGGBB form")
	}
	red, green, blue := value>>16, value>>8&0xff, value&0xff
	sequence := fmt.Sprintf("%s[38;2;%d;%d;%dm%s%s[39m", escape, red, green, blue, text, escape)
	return state.write("colored", sequence)
}

func (state *terminalState) style(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	text, err := requireString("style", 0, args[0])
	if err != nil {
		return err
	}
	style, ok := args[1].(*object.EnumValue)
	if !ok || style.EnumName != "Style" {
		return newError(object.RuntimeErrorKindType, "argument 2 to `style` must be Style, got %s", args[1].Type())
	}
	codes := map[string]int{"Normal": 0, "Bold": 1, "Dim": 2, "Italic": 3, "Underline": 4}
	code, ok := codes[style.Member]
	if !ok {
		return newError(object.RuntimeErrorKindValue, "unsupported Style member %q", style.Member)
	}
	return state.write("style", fmt.Sprintf("%s[%dm%s%s[0m", escape, code, text, escape))
}

func (state *terminalState) cursorMove(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 2); err != nil {
		return err
	}
	x, ok := args[0].(*object.Integer)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument 1 to `cursor_move` must be INTEGER, got %s", args[0].Type())
	}
	y, ok := args[1].(*object.Integer)
	if !ok {
		return newError(object.RuntimeErrorKindType, "argument 2 to `cursor_move` must be INTEGER, got %s", args[1].Type())
	}
	if x.Value < 0 || y.Value < 0 {
		return newError(object.RuntimeErrorKindValue, "arguments to `cursor_move` must be nonnegative")
	}
	if x.Value == int64(^uint64(0)>>1) || y.Value == int64(^uint64(0)>>1) {
		return newError(object.RuntimeErrorKindValue, "argument to `cursor_move` is too large")
	}
	return state.write("cursor_move", fmt.Sprintf("%s[%d;%dH", escape, y.Value+1, x.Value+1))
}

func (state *terminalState) write(name, value string) object.Object {
	state.mu.Lock()
	defer state.mu.Unlock()
	if err := writeAll(state.out, value); err != nil {
		return terminalError("%s failed: %s", name, err)
	}
	return state.null
}

func writeAll(writer io.Writer, value string) error {
	written, err := io.WriteString(writer, value)
	if err != nil {
		return err
	}
	if written != len(value) {
		return io.ErrShortWrite
	}
	return nil
}

func (state *terminalState) readKey(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	restore, err := state.temporaryRawLocked()
	if err != nil {
		return terminalError("read_key failed: %s", err)
	}
	if restore != nil {
		defer restore()
	}
	value, err := readRune(state.in)
	if err != nil {
		return terminalError("read_key failed: %s", err)
	}
	return &object.String{Value: value}
}

func readRune(reader io.Reader) (string, error) {
	buffer := make([]byte, utf8.UTFMax)
	if _, err := io.ReadFull(reader, buffer[:1]); err != nil {
		return "", err
	}
	if buffer[0] < utf8.RuneSelf {
		return string(buffer[:1]), nil
	}
	size := 0
	switch {
	case buffer[0]&0xe0 == 0xc0:
		size = 2
	case buffer[0]&0xf0 == 0xe0:
		size = 3
	case buffer[0]&0xf8 == 0xf0:
		size = 4
	default:
		return "", fmt.Errorf("input is not valid UTF-8")
	}
	if _, err := io.ReadFull(reader, buffer[1:size]); err != nil {
		return "", err
	}
	if !utf8.Valid(buffer[:size]) {
		return "", fmt.Errorf("input is not valid UTF-8")
	}
	return string(buffer[:size]), nil
}

func (state *terminalState) rawMode(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.rawState != nil {
		return state.null
	}
	fd, ok := fileDescriptor(state.in)
	if !ok || !term.IsTerminal(fd) {
		return terminalError("raw_mode requires terminal input")
	}
	rawState, err := term.MakeRaw(fd)
	if err != nil {
		return terminalError("raw_mode failed: %s", err)
	}
	state.rawState = rawState
	return state.null
}

func (state *terminalState) normalMode(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	if state.rawState == nil {
		return state.null
	}
	fd, ok := fileDescriptor(state.in)
	if !ok {
		return terminalError("normal_mode cannot access terminal input")
	}
	if err := term.Restore(fd, state.rawState); err != nil {
		return terminalError("normal_mode failed: %s", err)
	}
	state.rawState = nil
	return state.null
}

func (state *terminalState) temporaryRawLocked() (func(), error) {
	if state.rawState != nil {
		return nil, nil
	}
	fd, ok := fileDescriptor(state.in)
	if !ok || !term.IsTerminal(fd) {
		return nil, nil
	}
	rawState, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}
	return func() { _ = term.Restore(fd, rawState) }, nil
}

func (state *terminalState) cursorSave(args ...object.Object) object.Object {
	if err := requireArgumentCount(args, 0); err != nil {
		return err
	}
	state.mu.Lock()
	defer state.mu.Unlock()
	restore, err := state.temporaryRawLocked()
	if err != nil {
		return terminalError("cursor_save failed: %s", err)
	}
	if restore != nil {
		defer restore()
	}
	if err := writeAll(state.out, escape+"[6n"); err != nil {
		return terminalError("cursor_save failed: %s", err)
	}
	if err := state.flushLocked(); err != nil {
		return terminalError("cursor_save failed: %s", err)
	}
	response, err := readCursorResponse(state.in)
	if err != nil {
		return terminalError("cursor_save failed: %s", err)
	}
	parts := strings.Split(response, ";")
	row, rowErr := strconv.ParseInt(parts[0], 10, 64)
	column, columnErr := strconv.ParseInt(parts[1], 10, 64)
	if rowErr != nil || columnErr != nil || row < 1 || column < 1 {
		return terminalError("cursor_save received invalid terminal response")
	}
	return &object.Array{Elements: []object.Object{
		&object.Integer{Value: column - 1},
		&object.Integer{Value: row - 1},
	}}
}

func readCursorResponse(reader io.Reader) (string, error) {
	response := make([]byte, 0, 16)
	for len(response) < 64 {
		var one [1]byte
		if _, err := io.ReadFull(reader, one[:]); err != nil {
			return "", err
		}
		response = append(response, one[0])
		if one[0] == 'R' {
			text := string(response)
			if !strings.HasPrefix(text, escape+"[") {
				return "", fmt.Errorf("invalid terminal response %q", text)
			}
			coordinates := strings.TrimSuffix(strings.TrimPrefix(text, escape+"["), "R")
			if strings.Count(coordinates, ";") != 1 {
				return "", fmt.Errorf("invalid terminal response %q", text)
			}
			return coordinates, nil
		}
	}
	return "", fmt.Errorf("terminal response exceeded 64 bytes")
}

func fileDescriptor(value any) (int, bool) {
	if provider, ok := value.(interface{ TerminalFileDescriptor() (uintptr, bool) }); ok {
		fd, available := provider.TerminalFileDescriptor()
		return int(fd), available
	}
	provider, ok := value.(interface{ Fd() uintptr })
	if !ok {
		return 0, false
	}
	return int(provider.Fd()), true
}

func terminalError(format string, args ...any) *object.Error {
	return newError(object.RuntimeErrorKindRuntime, format, args...)
}
