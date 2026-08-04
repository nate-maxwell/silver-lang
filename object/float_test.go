package object

import (
	"math"
	"testing"
)

func TestFloatInspection(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{1, "1.0"},
		{1.5, "1.5"},
		{math.Copysign(0, -1), "-0.0"},
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
