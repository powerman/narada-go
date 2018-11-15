package narada

import (
	"bytes"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestFakeConfig(t *testing.T) {
	type configCases []struct {
		path    string
		want    []byte
		wanterr error
	}
	cases := []struct {
		fake    map[string]string
		configs configCases
	}{
		{
			map[string]string{
				"fake":     "FAKE1",
				"dir/fake": "Fake\n2\n",
			},
			configCases{
				{"nosuch", nil, nil},
				{"fake", []byte("FAKE1"), nil},
				{"file", []byte("REAL1"), nil},
				{"unreadable", nil, &os.PathError{Op: "open", Path: "config/unreadable", Err: syscall.EACCES}},
				{"dir", []byte{}, &os.PathError{Op: "read", Path: "config/dir", Err: syscall.EISDIR}},
				{"dir/nosuch", nil, nil},
				{"dir/fake", []byte("Fake\n2\n"), nil},
				{"dir/file", []byte("Real2\n"), nil},
			},
		},
		{
			map[string]string{
				"fake2": "FAKE2\n",
			},
			configCases{
				{"nosuch", nil, nil},
				{"fake", nil, nil},
				{"fake2", []byte("FAKE2\n"), nil},
				{"file", []byte("REAL1"), nil},
				{"unreadable", nil, &os.PathError{Op: "open", Path: "config/unreadable", Err: syscall.EACCES}},
				{"dir", []byte{}, &os.PathError{Op: "read", Path: "config/dir", Err: syscall.EISDIR}},
				{"dir/nosuch", nil, nil},
				{"dir/fake", nil, nil},
				{"dir/file", []byte("Real2\n"), nil},
			},
		},
		{
			nil,
			configCases{
				{"nosuch", nil, nil},
				{"fake", nil, nil},
				{"fake2", nil, nil},
				{"file", []byte("REAL1"), nil},
				{"unreadable", nil, &os.PathError{Op: "open", Path: "config/unreadable", Err: syscall.EACCES}},
				{"dir", []byte{}, &os.PathError{Op: "read", Path: "config/dir", Err: syscall.EISDIR}},
				{"dir/nosuch", nil, nil},
				{"dir/fake", nil, nil},
				{"dir/file", []byte("Real2\n"), nil},
			},
		},
	}
	origOpen := open
	for _, c := range cases {
		FakeConfig(c.fake)
		for _, c := range c.configs {
			buf, err := GetConfig(c.path)
			if (buf == nil) != (c.want == nil) || !bytes.Equal(buf, c.want) {
				t.Errorf("FakeConfig(%q) = %#v, want = %#v", c.path, buf, c.want)
			}
			if fmt.Sprintf("%#v", err) != fmt.Sprintf("%#v", c.wanterr) {
				t.Errorf("FakeConfig(%q), err = %#v, want %#v", c.path, err, c.wanterr)
			}
		}
	}
	open = origOpen
}

func TestGetConfig(t *testing.T) {
	cases := []struct {
		path    string
		want    []byte
		wanterr error
	}{
		{"nosuch", nil, nil},
		{"empty", []byte{}, nil},
		{"file", []byte("REAL1"), nil},
		{"unreadable", nil, &os.PathError{Op: "open", Path: "config/unreadable", Err: syscall.EACCES}},
		{"log", []byte{}, &os.PathError{Op: "read", Path: "config/log", Err: syscall.EISDIR}},
		{"log/nosuch", nil, nil},
		{"log/level", []byte("INFO\n"), nil},
		{"log/no-such_dir.123/no-such_file.123", nil, nil},
	}
	for _, c := range cases {
		buf, err := GetConfig(c.path)
		if (buf == nil) != (c.want == nil) || !bytes.Equal(buf, c.want) {
			t.Errorf("GetConfig(%q) = %#v, want = %#v", c.path, buf, c.want)
		}
		if fmt.Sprintf("%#v", err) != fmt.Sprintf("%#v", c.wanterr) {
			t.Errorf("GetConfig(%q), err = %#v, want %#v", c.path, err, c.wanterr)
		}
	}
}

func TestGetConfigBadName(t *testing.T) {
	cases := []string{
		"",
		" ",
		"./empty",
		"../config/empty",
		"log/./level",
		"log/../empty",
		"bad:name",
		"bad name",
	}
	for _, path := range cases {
		var pnk interface{}
		func() {
			defer func() { pnk = recover() }()
			GetConfig(path)
		}()
		wantpnk := "invalid config name: " + path
		if fmt.Sprintf("%#v", pnk) != fmt.Sprintf("%#v", wantpnk) {
			t.Errorf("GetConfig(%q), panic = %#v, want %#v", path, pnk, wantpnk)
		}
	}
}

func TestGetConfigLine(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"nosuch", ""},
		{"empty", ""},
		{"log/type", "syslog"},
		{"single_line", "line1"},
		{"int", " 42 "},
	}
	for _, c := range cases {
		line := GetConfigLine(c.path)
		if line != c.want {
			t.Errorf("GetConfigLine(%q) = %#v, want = %#v", c.path, line, c.want)
		}
	}
}

func TestGetConfigLineBad(t *testing.T) {
	cases := []struct {
		path    string
		wantpnk interface{}
	}{
		{"log", &os.PathError{Op: "read", Path: "config/log", Err: syscall.EISDIR}},
		{"multi_line", "config multi_line contain more than one line"},
	}
	for _, c := range cases {
		var pnk interface{}
		func() {
			defer func() { pnk = recover() }()
			GetConfigLine(c.path)
		}()
		if fmt.Sprintf("%#v", pnk) != fmt.Sprintf("%#v", c.wantpnk) {
			t.Errorf("GetConfigLine(%q), panic = %#v, want %#v", c.path, pnk, c.wantpnk)
		}
	}
}

func TestGetConfigInt(t *testing.T) {
	cases := []struct {
		path string
		want int
	}{
		{"nosuch", 0},
		{"empty", 0},
		{"int", 42},
	}
	for _, c := range cases {
		i := GetConfigInt(c.path)
		if i != c.want {
			t.Errorf("GetConfigInt(%q) = %#v, want = %#v", c.path, i, c.want)
		}
	}
}

func TestGetConfigIntBad(t *testing.T) {
	cases := []struct {
		path    string
		wantpnk string
	}{
		{"multi_line", "config multi_line contain more than one line"},
		{"log/level", "config log/level must contain integer"},
		{"badint", "config badint must contain integer"},
		{"twoint", "config twoint must contain integer"},
		{"float", "config float must contain integer"},
	}
	for _, c := range cases {
		var pnk interface{}
		func() {
			defer func() { pnk = recover() }()
			GetConfigInt(c.path)
		}()
		if fmt.Sprintf("%#v", pnk) != fmt.Sprintf("%#v", c.wantpnk) {
			t.Errorf("GetConfigInt(%q), panic = %#v, want %#v", c.path, pnk, c.wantpnk)
		}
	}
}

func TestGetConfigIntBetween(t *testing.T) {
	cases := []struct {
		path string
		min  int
		max  int
		want int
	}{
		{"nosuch", 0, 5, 0},
		{"empty", -5, 0, 0},
		{"int", 40, 45, 42},
		{"int", 42, 45, 42},
		{"int", 40, 42, 42},
		{"int", 42, 42, 42},
	}
	for _, c := range cases {
		i := GetConfigIntBetween(c.path, c.min, c.max)
		if i != c.want {
			t.Errorf("GetConfigIntBetween(%q,%d,%d) = %#v, want = %#v", c.path, c.min, c.max, i, c.want)
		}
	}
}

func TestGetConfigIntBetweenBad(t *testing.T) {
	cases := []struct {
		path    string
		min     int
		max     int
		wantpnk string
	}{
		{"nosuch", 1, 5, "config nosuch must contain integer >= 1"},
		{"empty", -5, -1, "config empty must contain integer <= -1"},
		{"int", 43, 45, "config int must contain integer >= 43"},
		{"int", 40, 41, "config int must contain integer <= 41"},
	}
	for _, c := range cases {
		var pnk interface{}
		func() {
			defer func() { pnk = recover() }()
			GetConfigIntBetween(c.path, c.min, c.max)
		}()
		if fmt.Sprintf("%#v", pnk) != fmt.Sprintf("%#v", c.wantpnk) {
			t.Errorf("GetConfigIntBetween(%q,%d,%d), panic = %#v, want %#v", c.path, c.min, c.max, pnk, c.wantpnk)
		}
	}
}

func TestGetConfigDuration(t *testing.T) {
	cases := []struct {
		path string
		want time.Duration
	}{
		{"duration", 3 * time.Second},
	}
	for _, c := range cases {
		i := GetConfigDuration(c.path)
		if i != c.want {
			t.Errorf("GetConfigDuration(%q) = %#v, want = %#v", c.path, i, c.want)
		}
	}
}

func TestGetConfigDurationBad(t *testing.T) {
	cases := []struct {
		path    string
		wantpnk string
	}{
		{"nosuch", "config nosuch must contain duration"},
		{"empty", "config empty must contain duration"},
		{"badint", "config badint must contain duration"},
	}
	for _, c := range cases {
		var pnk interface{}
		func() {
			defer func() { pnk = recover() }()
			GetConfigDuration(c.path)
		}()
		if fmt.Sprintf("%#v", pnk) != fmt.Sprintf("%#v", c.wantpnk) {
			t.Errorf("GetConfigDuration(%q), panic = %#v, want %#v", c.path, pnk, c.wantpnk)
		}
	}
}

func TestGetConfigDurationBetween(t *testing.T) {
	cases := []struct {
		path string
		min  time.Duration
		max  time.Duration
		want time.Duration
	}{
		{"duration", 0, time.Minute, 3 * time.Second},
		{"duration", 2 * time.Second, 4 * time.Second, 3 * time.Second},
		{"duration", 3 * time.Second, 3 * time.Second, 3 * time.Second},
	}
	for _, c := range cases {
		i := GetConfigDurationBetween(c.path, c.min, c.max)
		if i != c.want {
			t.Errorf("GetConfigDurationBetween(%q,%s,%s) = %#v, want = %#v", c.path, c.min, c.max, i, c.want)
		}
	}
}

func TestGetConfigDurationBetweenBad(t *testing.T) {
	cases := []struct {
		path    string
		min     time.Duration
		max     time.Duration
		wantpnk string
	}{
		{"duration", 4 * time.Second, time.Minute, "config duration must contain duration >= 4s"},
		{"duration", 1 * time.Second, 2 * time.Second, "config duration must contain duration <= 2s"},
	}
	for _, c := range cases {
		var pnk interface{}
		func() {
			defer func() { pnk = recover() }()
			GetConfigDurationBetween(c.path, c.min, c.max)
		}()
		if fmt.Sprintf("%#v", pnk) != fmt.Sprintf("%#v", c.wantpnk) {
			t.Errorf("GetConfigDurationBetween(%q,%s,%s), panic = %#v, want %#v", c.path, c.min, c.max, pnk, c.wantpnk)
		}
	}
}
