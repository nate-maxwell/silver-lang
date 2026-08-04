package object

import "testing"

func TestBuiltinFileStructShape(t *testing.T) {
	file, ok := BuiltinStructDefinitionByName("File")
	if !ok {
		t.Fatal("File definition is not registered")
	}
	wantFields := []string{"path", "read", "write", "close"}
	wantTypes := []string{
		"str",
		"call() str | IOError",
		"call(contents: str) | IOError",
		"call() | IOError",
	}
	if len(file.Fields) != len(wantFields) {
		t.Fatalf("File has %d fields, want %d", len(file.Fields), len(wantFields))
	}
	for index, want := range wantFields {
		if file.Fields[index] != want {
			t.Fatalf("field %d is %q, want %q", index, file.Fields[index], want)
		}
		if got := file.FieldTypes[index].String(); got != wantTypes[index] {
			t.Fatalf("field %q has type %q, want %q", want, got, wantTypes[index])
		}
	}
}

func TestBuiltinErrorStructsExposeMessage(t *testing.T) {
	names := []string{"IOError", "FileNotFound", "PermissionDenied"}
	for _, name := range runtimeErrorStructNames {
		names = append(names, name)
	}
	for _, name := range names {
		definition, ok := BuiltinStructDefinitionByName(name)
		if !ok {
			t.Fatalf("%s definition is not registered", name)
		}
		if len(definition.Fields) != 1 || definition.Fields[0] != "message" || definition.FieldTypes[0].String() != "str" {
			t.Fatalf("%s does not have the expected message: str field", name)
		}
	}
}

func TestRuntimeErrorStructTableConstructsEveryKind(t *testing.T) {
	want := map[RuntimeErrorKind]string{
		RuntimeErrorKindRuntime:      "RuntimeError",
		RuntimeErrorKindType:         "TypeError",
		RuntimeErrorKindValue:        "ValueError",
		RuntimeErrorKindZeroDivision: "ZeroDivisionError",
		RuntimeErrorKindName:         "NameError",
		RuntimeErrorKindAttribute:    "AttributeError",
		RuntimeErrorKindImport:       "ImportError",
		RuntimeErrorKindSyntax:       "SyntaxError",
		RuntimeErrorKindKey:          "KeyError",
		RuntimeErrorKindIndex:        "IndexError",
		RuntimeErrorKindTask:         "TaskError",
	}
	if len(runtimeErrorStructNames) != len(want) {
		t.Fatalf("runtime error table has %d entries, want %d", len(runtimeErrorStructNames), len(want))
	}
	for kind, name := range want {
		if got := runtimeErrorStructNames[kind]; got != name {
			t.Fatalf("runtime error kind %q maps to %q, want %q", kind, got, name)
		}
		err := NewError(kind, "boom")
		if err.Value.Struct.Name != name || err.MessageText() != "boom" || !err.IsRuntimeError() {
			t.Fatalf("constructed %q as %#v", kind, err)
		}
	}
}
