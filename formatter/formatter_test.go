package formatter

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceFormatsSpacingIndentationAndComments(t *testing.T) {
	source := []byte(`# Keep this comment.
let add=fn(left:int,right:int)int{
return left+right # Keep this one too.
}
let result=add(20,22)
`)
	want := `# Keep this comment.
let add = fn(left: int, right: int) int {
    return left + right # Keep this one too.
}
let result = add(20, 22)
`
	got, err := Source("test.slv", source)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("formatted source is:\n%s\nwant:\n%s", got, want)
	}
}

func TestSourceFormatsVariadicParameter(t *testing.T) {
	formatted, err := Source("variadic.slv", []byte("let join=fn(prefix:str,parts:str...){\n}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(formatted), "let join = fn(prefix: str, parts: str...) {\n}\n"; got != want {
		t.Fatalf("formatted source is %q, want %q", got, want)
	}
}

func TestSourcePreservesLiteralContents(t *testing.T) {
	source := []byte("let text=\"# not a comment\"\nlet template=```keep  spacing {text}```\n")
	formatted, err := Source("test.slv", source)
	if err != nil {
		t.Fatal(err)
	}
	for _, literal := range []string{`"# not a comment"`, "```keep  spacing {text}```"} {
		if !strings.Contains(string(formatted), literal) {
			t.Fatalf("formatted source did not preserve %q: %s", literal, formatted)
		}
	}
}

func TestSourcePreservesNestedTemplates(t *testing.T) {
	source := []byte("let nested=```outer {```inner  text```.eval()}```\n")
	formatted, err := Source("test.slv", source)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(formatted), "let nested = ```outer {```inner  text```.eval()}```\n"; got != want {
		t.Fatalf("formatted source is %q, want %q", got, want)
	}
}

func TestSourceIndentsMultilineCollections(t *testing.T) {
	source := []byte("let values=[\n1,\n2\n]\n")
	want := "let values = [\n    1,\n    2\n]\n"
	formatted, err := Source("test.slv", source)
	if err != nil {
		t.Fatal(err)
	}
	if string(formatted) != want {
		t.Fatalf("formatted source is %q, want %q", formatted, want)
	}
}

func TestSourceFormatsExportDeclaration(t *testing.T) {
	source := []byte("export{\npublic_value,\nPublicType,\n}\n")
	want := "export {\n    public_value,\n    PublicType,\n}\n"
	formatted, err := Source("test.slv", source)
	if err != nil {
		t.Fatal(err)
	}
	if string(formatted) != want {
		t.Fatalf("formatted export is %q, want %q", formatted, want)
	}
}

func TestSourceFormatsEmbeddedStructField(t *testing.T) {
	formatted, err := Source("embedding.slv", []byte("struct Outer{\ninner :: Inner\n}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(formatted), "struct Outer {\n    inner:: Inner\n}\n"; got != want {
		t.Fatalf("formatted source is %q, want %q", got, want)
	}
}

func TestSourceRejectsInvalidCode(t *testing.T) {
	if _, err := Source("broken.slv", []byte("let value =\n")); err == nil {
		t.Fatal("Source accepted invalid code")
	}
}

func TestFileOnlyWritesChangedSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "main.slv")
	if err := os.WriteFile(path, []byte("let answer=42"), 0640); err != nil {
		t.Fatal(err)
	}
	changed, err := File(path)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("File reported no change")
	}
	changed, err = File(path)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatal("File changed already formatted source")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "let answer = 42\n" {
		t.Fatalf("file contains %q", got)
	}
}

func TestFileDoesNotRewriteInvalidSource(t *testing.T) {
	path := filepath.Join(t.TempDir(), "broken.slv")
	source := []byte("let value=\n")
	if err := os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	if changed, err := File(path); err == nil || changed {
		t.Fatalf("File returned changed=%t, err=%v", changed, err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(source) {
		t.Fatalf("invalid source changed from %q to %q", source, got)
	}
}

func TestRepositorySourcesFormatIdempotently(t *testing.T) {
	for _, root := range []string{"../examples", "../stdlib/silver"} {
		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil || entry.IsDir() || filepath.Ext(path) != ".slv" {
				return walkErr
			}
			source, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			formatted, err := Source(path, source)
			if err != nil {
				t.Errorf("format %s: %s", path, err)
				return nil
			}
			second, err := Source(path, formatted)
			if err != nil {
				t.Errorf("parse formatted %s: %s", path, err)
				return nil
			}
			if string(second) != string(formatted) {
				t.Errorf("formatting %s is not idempotent", path)
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}
