package stdlib_test

import (
	"bytes"
	"silver/object"
	"strings"
	"testing"
)

const terminalImport = "let terminal = import(\"terminal\")\n"

func TestTerminalControlOutput(t *testing.T) {
	var output bytes.Buffer
	input := terminalImport + `terminal.clear()
terminal.colored("color", "#12abCF")
terminal.style("bold", terminal.Style.Bold)
terminal.cursor_move(3, 4)
terminal.cursor_hide()
terminal.cursor_show()
terminal.clear_line()
terminal.alternate_screen()
terminal.main_screen()`
	testNullObject(t, testEvalWithStreams(input, strings.NewReader(""), &output, &bytes.Buffer{}))
	want := "\x1b[2J\x1b[H" +
		"\x1b[38;2;18;171;207mcolor\x1b[39m" +
		"\x1b[1mbold\x1b[0m" +
		"\x1b[5;4H" +
		"\x1b[?25l\x1b[?25h\x1b[0K\x1b[?1049h\x1b[?1049l"
	if got := output.String(); got != want {
		t.Fatalf("terminal output is %q, want %q", got, want)
	}
}

func TestTerminalDimensionsUseEnvironmentFallback(t *testing.T) {
	t.Setenv("COLUMNS", "132")
	t.Setenv("LINES", "43")
	result := testEvalWithStreams(
		terminalImport+`[terminal.width(), terminal.height()]`,
		strings.NewReader(""),
		&bytes.Buffer{},
		&bytes.Buffer{},
	)
	array := terminalResultArray(t, result, 2)
	testIntegerObject(t, array[0], 132)
	testIntegerObject(t, array[1], 43)
}

func TestTerminalCursorSaveReturnsZeroBasedCoordinates(t *testing.T) {
	var output bytes.Buffer
	result := testEvalWithStreams(
		terminalImport+`terminal.cursor_save()`,
		strings.NewReader("\x1b[12;34R"),
		&output,
		&bytes.Buffer{},
	)
	array := terminalResultArray(t, result, 2)
	testIntegerObject(t, array[0], 33)
	testIntegerObject(t, array[1], 11)
	if got, want := output.String(), "\x1b[6n"; got != want {
		t.Fatalf("cursor query is %q, want %q", got, want)
	}
}

func TestTerminalReadKeyReadsOneUTF8Character(t *testing.T) {
	result := testEvalWithStreams(
		terminalImport+`terminal.read_key()`,
		strings.NewReader("éremaining"),
		&bytes.Buffer{},
		&bytes.Buffer{},
	)
	text, ok := result.(*object.String)
	if !ok || text.Value != "é" {
		t.Fatalf("read_key returned %#v, want é", result)
	}
}

func TestTerminalFlushUsesOutputFlusher(t *testing.T) {
	output := &flushBuffer{}
	testNullObject(t, testEvalWithStreams(
		terminalImport+`terminal.flush()`,
		strings.NewReader(""),
		output,
		&bytes.Buffer{},
	))
	if output.flushes != 1 {
		t.Fatalf("flush count is %d, want 1", output.flushes)
	}
}

func TestTerminalRejectsInvalidArgumentsAndUnavailableRawMode(t *testing.T) {
	tests := []struct {
		input   string
		message string
	}{
		{`terminal.colored("text", "red")`, "argument 2 to `colored` must be a hex color in #RRGGBB form"},
		{`terminal.cursor_move(-1, 0)`, "arguments to `cursor_move` must be nonnegative"},
		{`terminal.style("text", "Bold")`, "argument 2 to `style` must be Style, got STRING"},
		{`terminal.raw_mode()`, "raw_mode requires terminal input"},
	}
	for _, test := range tests {
		result := testEvalWithStreams(
			terminalImport+test.input,
			strings.NewReader(""),
			&bytes.Buffer{},
			&bytes.Buffer{},
		)
		err, ok := result.(*object.Error)
		if !ok || err.MessageText() != test.message {
			t.Fatalf("%s returned %#v, want %q", test.input, result, test.message)
		}
	}
}

type flushBuffer struct {
	bytes.Buffer
	flushes int
}

func (buffer *flushBuffer) Flush() error {
	buffer.flushes++
	return nil
}

func terminalResultArray(t *testing.T, result object.Object, wantLength int) []object.Object {
	t.Helper()
	array, ok := result.(*object.Array)
	if !ok {
		t.Fatalf("result is %T, want *object.Array", result)
	}
	if len(array.Elements) != wantLength {
		t.Fatalf("array has %d elements, want %d", len(array.Elements), wantLength)
	}
	return array.Elements
}
