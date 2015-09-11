package narada

import (
	"bytes"
	"errors"
	"testing"
)

func TestGetConfig(t *testing.T) {
	cases := []struct {
		path    string
		want    []byte
		wanterr error
	}{
		{"log/level", []byte("INFO"), nil},
		{"log/empty", nil, nil},
		{"log", []byte{}, errors.New("read config/log: is a directory")},
		{"log/no-such_dir.123/no-such_file.123", nil, nil},
	}
	for _, c := range cases {
		buf, err := GetConfig(c.path)
		if buf != nil {
			buf = bytes.TrimRight(buf, "\n")
		}
		if err == nil && c.wanterr != nil || err != nil && c.wanterr == nil || err != nil && err.Error() != c.wanterr.Error() {
			t.Errorf("GetConfig(%q), err = %v", c.path, err)
		}
		if buf == nil && c.want != nil || buf != nil && c.want == nil || bytes.Compare(buf, c.want) != 0 {
			t.Errorf("GetConfig(%q) = %#v, want = %#v", c.path, buf, c.want)
		}
	}
}

func TestGetConfigBadName(t *testing.T) {
	cases := []struct{ path, msg string }{
		{"", "invalid config name: "},
		{" ", "invalid config name:  "},
		{"./empty", "invalid config name: ./empty"},
		{"../config/empty", "invalid config name: ../config/empty"},
		{"log/./level", "invalid config name: log/./level"},
		{"log/../empty", "invalid config name: log/../empty"},
		{"bad:name", "invalid config name: bad:name"},
		{"bad name", "invalid config name: bad name"},
	}
	for _, c := range cases {
		var msg interface{}
		func() {
			defer func() { msg = recover() }()
			GetConfig(c.path)
		}()
		if msg == nil || msg.(string) != c.msg {
			t.Errorf("GetConfig(%q), panic = %v, want %v", c.path, msg, c.msg)
		}
	}
}

func TestGetConfigLine(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"single_line", "line1"},
		{"log/empty", ""},
	}
	for _, c := range cases {
		line := GetConfigLine(c.path)
		if line != c.want {
			t.Errorf("GetConfigLine(%q) = %#v, want = %#v", c.path, line, c.want)
		}
	}
}

func TestGetConfigLineBad(t *testing.T) {
	cases := []struct{ path, msg string }{
		{"log", "read config/log: is a directory"},
		{"multi_line", "config multi_line contain more than one line"},
	}
	for _, c := range cases {
		var msg interface{}
		func() {
			defer func() { msg = recover() }()
			GetConfigLine(c.path)
		}()
		switch v := msg.(type) {
		case error:
			msg = v.Error()
		}
		if msg == nil || msg.(string) != c.msg {
			t.Errorf("GetConfigLine(%q), panic = %v, want %v", c.path, msg, c.msg)
		}
	}
}

func TestGetConfigInt(t *testing.T) {
	cases := []struct {
		path string
		want int
	}{
		{"empty", 0},
		{"log/nosuch", 0},
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
	cases := []struct{ path, msg string }{
		{"log/level", "config log/level must contain integer"},
	}
	for _, c := range cases {
		var msg interface{}
		func() {
			defer func() { msg = recover() }()
			GetConfigInt(c.path)
		}()
		switch v := msg.(type) {
		case error:
			msg = v.Error()
		}
		if msg == nil || msg.(string) != c.msg {
			t.Errorf("GetConfigInt(%q), panic = %v, want %v", c.path, msg, c.msg)
		}
	}
}
