package object

import "testing"

func TestStringHashKey(t *testing.T) {
	hello1 := &String{Value: "Hello World"}
	hello2 := &String{Value: "Hello World"}
	diff1 := &String{Value: "My name is johnny"}
	diff2 := &String{Value: "My name is johnny"}

	if hello1.HashKey() != hello2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if diff1.HashKey() != diff2.HashKey() {
		t.Errorf("strings with same content have different hash keys")
	}

	if hello1.HashKey() == diff1.HashKey() {
		t.Errorf("strings with different content have same hash keys")
	}
}

func TestFloatInspection(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{1, "1.0"},
		{1.5, "1.5"},
		{-0.0, "0.0"},
	}

	for _, tt := range tests {
		if got := (&Float{Value: tt.value}).Inspect(); got != tt.want {
			t.Errorf("float %g inspects as %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestIntegralFloatSharesIntegerHashKey(t *testing.T) {
	integer := (&Integer{Value: 42}).HashKey()
	float := (&Float{Value: 42.0}).HashKey()
	if integer != float {
		t.Fatalf("equal numeric values have different hash keys: %+v != %+v", integer, float)
	}
}
