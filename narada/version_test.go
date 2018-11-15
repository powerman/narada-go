package narada

import (
	"fmt"
	"os"
	"syscall"
	"testing"
)

func TestVersion(t *testing.T) {
	want := "1.2.3+example-1234567890"
	ver, err := Version()
	if err != nil {
		t.Errorf("Version(), err = %v", err)
	}
	if ver != want {
		t.Errorf("Version() = %q, want %q", ver, want)
	}

	err = os.Chmod("VERSION", 0)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chmod("VERSION", 0644)

	wanterr := &os.PathError{Op: "open", Path: "VERSION", Err: syscall.EACCES}
	_, err = Version()
	if fmt.Sprintf("%#v", err) != fmt.Sprintf("%#v", wanterr) {
		t.Errorf("Version(), err = %#v, want %#v", err, wanterr)
	}
}
