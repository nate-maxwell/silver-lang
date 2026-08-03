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

func TestBuiltinIOErrorStructsExposeMessage(t *testing.T) {
	for _, name := range []string{"IOError", "FileNotFound", "PermissionDenied"} {
		definition, ok := BuiltinStructDefinitionByName(name)
		if !ok {
			t.Fatalf("%s definition is not registered", name)
		}
		if len(definition.Fields) != 1 || definition.Fields[0] != "message" || definition.FieldTypes[0].String() != "str" {
			t.Fatalf("%s does not have the expected message: str field", name)
		}
	}
}
