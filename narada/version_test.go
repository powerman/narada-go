package narada

import "testing"

func TestVersion(t *testing.T) {
	want := "1.2.3+example-1234567890"
	if ver, err := Version(); err != nil {
		t.Errorf("Version(), err = %v", err)
	} else if ver != want {
		t.Errorf("Version() = %q, want %q", ver, want)
	}
}
