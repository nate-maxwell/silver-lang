package repl

import (
	"bytes"
	"strings"
	"testing"
)

func TestStandardStreamsShareREPLInputAndOutput(t *testing.T) {
	input := "import(\"io\").stdin.read()\npayload\n"
	var output bytes.Buffer
	Start(strings.NewReader(input), &output)

	if got := output.String(); !strings.Contains(got, "payload") {
		t.Fatalf("REPL output %q does not contain stdin payload", got)
	}
}

func TestStandardStreamReadLineSharesREPLInput(t *testing.T) {
	input := "import(\"io\").stdin.read_line()\npayload\n"
	var output bytes.Buffer
	Start(strings.NewReader(input), &output)

	if got := output.String(); !strings.Contains(got, "payload") {
		t.Fatalf("REPL output %q does not contain line read from stdin", got)
	}
}

func TestStandardErrorUsesREPLOutput(t *testing.T) {
	input := "import(\"io\").stderr.write(\"error output\")\n"
	var output bytes.Buffer
	Start(strings.NewReader(input), &output)

	if got := output.String(); !strings.Contains(got, "error output") {
		t.Fatalf("REPL output %q does not contain stderr text", got)
	}
}

func TestOperatorDeclarationPersistsAcrossREPLLines(t *testing.T) {
	input := "operator @ fn(left, right) int { return left + right }\n20 @ 22\n"
	var output bytes.Buffer
	Start(strings.NewReader(input), &output)

	if got := output.String(); !strings.Contains(got, "42") {
		t.Fatalf("REPL output %q does not contain operator result", got)
	}
}
