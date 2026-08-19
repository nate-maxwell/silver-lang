package version

import "testing"

func TestString(t *testing.T) {
	if got, want := String(), "0.4.0"; got != want {
		t.Fatalf("String() is %q, want %q", got, want)
	}
}
