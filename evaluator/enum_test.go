package evaluator

import (
	"path/filepath"
	"silver/object"
	"testing"
)

func TestEnumValue(t *testing.T) {
	evaluated := testEval(`
type Direction = enum { North, East, South, West }
Direction.North
`)

	value, ok := evaluated.(*object.EnumValue)
	if !ok {
		t.Fatalf("result is %T, want *object.EnumValue", evaluated)
	}
	if value.EnumName != "Direction" || value.Member != "North" {
		t.Fatalf("unexpected enum value: %+v", value)
	}
	if value.Inspect() != "Direction.North" {
		t.Fatalf("enum inspection is %q", value.Inspect())
	}
}

func TestEnumEquality(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"type Direction = enum { North, South }\nDirection.North == Direction.North", true},
		{"type Direction = enum { North, South }\nDirection.North == Direction.South", false},
		{"type Direction = enum { North, South }\nDirection.North != Direction.South", true},
		{"type First = enum { Value }\ntype Second = enum { Value }\nFirst.Value == Second.Value", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestEnumValuesAreHashable(t *testing.T) {
	evaluated := testEval(`
type Direction = enum { North, South }
let labels = { Direction.North: "north", Direction.South: "south" }
labels[Direction.South]
`)

	value, ok := evaluated.(*object.String)
	if !ok {
		t.Fatalf("result is %T, want *object.String", evaluated)
	}
	if value.Value != "south" {
		t.Fatalf("result is %q, want south", value.Value)
	}
}

func TestMissingEnumMember(t *testing.T) {
	evaluated := testEval("type Direction = enum { North }\nDirection.South")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if err.MessageText() != `enum "Direction" has no member "South"` {
		t.Fatalf("error message is %q", err.MessageText())
	}
}

func TestEnumExportedFromModule(t *testing.T) {
	dir := t.TempDir()
	writePackageManifest(t, dir, "library.slv")
	libraryPath := filepath.Join(dir, "library.slv")
	mainPath := filepath.Join(dir, "main.slv")
	writeSilverFile(t, libraryPath, `type Status = enum { Ready, Busy }`)
	writeSilverFile(t, mainPath, `
let library = import("./library.slv")
library.Status.Ready
`)

	evaluated := New().EvalFile(mainPath, object.NewEnvironment())
	value, ok := evaluated.(*object.EnumValue)
	if !ok {
		t.Fatalf("result is %T, want *object.EnumValue", evaluated)
	}
	if value.Inspect() != "Status.Ready" {
		t.Fatalf("result is %q, want Status.Ready", value.Inspect())
	}
}
