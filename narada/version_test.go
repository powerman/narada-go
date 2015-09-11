package narada

import "testing"

func TestVERSION(t *testing.T) {
	want := "1.2.3+example-1234567890"
	if VERSION != want {
		t.Errorf("VERSION = %q, want %q", VERSION, want)
	}
}
