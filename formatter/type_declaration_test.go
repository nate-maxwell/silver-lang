package formatter

import "testing"

func TestSourceFormatsTypeDeclarations(t *testing.T) {
	source := []byte(`type Point=struct{
x:float
y:float
}
type Direction=enum{North,East,South,West}
type Names=array[str]
type Transform=call(Point)Point
let origin:Point=Point{0.0,0.0}
`)
	want := `type Point = struct {
    x: float
    y: float
}
type Direction = enum { North, East, South, West }
type Names = array[str]
type Transform = call(Point) Point
let origin: Point = Point{0.0, 0.0}
`
	formatted, err := Source("types.slv", source)
	if err != nil {
		t.Fatal(err)
	}
	if string(formatted) != want {
		t.Fatalf("formatted source:\n%s\nwant:\n%s", formatted, want)
	}
	again, err := Source("types.slv", formatted)
	if err != nil || string(again) != want {
		t.Fatalf("formatting is not stable: %s, %v", again, err)
	}
}
