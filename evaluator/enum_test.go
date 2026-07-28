package evaluator

import (
	"path/filepath"
	"silver/object"
	"testing"
)

func TestEnumValue(t *testing.T) {
	evaluated := testEval(`
enum Direction { North, East, South, West }
Direction.North;
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
		{"enum Direction { North, South }\nDirection.North == Direction.North", true},
		{"enum Direction { North, South }\nDirection.North == Direction.South", false},
		{"enum Direction { North, South }\nDirection.North != Direction.South", true},
		{"enum First { Value }\nenum Second { Value }\nFirst.Value == Second.Value", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			testBooleanObject(t, testEval(tt.input), tt.want)
		})
	}
}

func TestEnumValuesAreHashable(t *testing.T) {
	evaluated := testEval(`
enum Direction { North, South }
let labels = { Direction.North: "north", Direction.South: "south" };
labels[Direction.South];
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
	evaluated := testEval("enum Direction { North }\nDirection.South")
	err, ok := evaluated.(*object.Error)
	if !ok {
		t.Fatalf("result is %T, want *object.Error", evaluated)
	}
	if err.Message != `enum "Direction" has no member "South"` {
		t.Fatalf("error message is %q", err.Message)
	}
}

func TestEnumExportedFromModule(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library.lib")
	mainPath := filepath.Join(dir, "main.slvr")
	writeMonkeyFile(t, libraryPath, `enum Status { Ready, Busy }`)
	writeMonkeyFile(t, mainPath, `
let library = import("./library.lib");
library.Status.Ready;
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
