package narada

import (
	"bufio"
	"bytes"
	"errors"
	"log"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"testing"
	"time"
)

var fromSyslog = make(chan string, 64)
var cmd *exec.Cmd

func fakeLog() {
	cmd = exec.Command("socat", "UNIX-RECV:var/log.sock,mode=666,unlink-early", "STDOUT")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	go func() {
		reader := bufio.NewReader(stdout)
		format := regexp.MustCompile(`\A(<\d+>)[^\]]*\](: .*)\n\z`)
		for {
			line, err := reader.ReadString('\n')
			line = format.ReplaceAllString(line, "$1$2")
			fromSyslog <- line
			if err != nil {
				break
			}
		}
		close(fromSyslog)
	}()
WAIT:
	for {
		switch _, err := os.Stat("var/log.sock"); {
		case err == nil:
			break WAIT
		case os.IsNotExist(err):
			time.Sleep(time.Millisecond)
		default:
			log.Fatal(err)
		}
	}
	InitLogError = initLog()
}

func TestInitLog(t *testing.T) {
	cases := []struct {
		setup   func()
		level   LogLevel
		ready   bool
		wanterr error
	}{
		{
			func() {},
			LogDEBUG, false, errors.New("require non-empty config/log/output"),
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": ""})
				InitLogError = initLog()
			},
			LogDEBUG, false, errors.New("unsupported config/log/level: "),
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": "bad"})
				InitLogError = initLog()
			},
			LogDEBUG, false, errors.New("unsupported config/log/level: bad"),
		},
		{
			func() {
				FakeConfig(map[string]string{"log/type": "bad"})
				InitLogError = initLog()
			},
			LogDEBUG, false, errors.New("unsupported config/log/type: bad"),
		},
		{
			func() {
				FakeConfig(map[string]string{"log/type": "file"})
				InitLogError = initLog()
			},
			LogDEBUG, false, errors.New("unsupported config/log/type: file"),
		},
		{
			func() {
				FakeConfig(map[string]string{"log/type": ""})
				InitLogError = initLog()
			},
			LogDEBUG, false, errors.New("dial unixgram var/log.sock: no such file or directory"),
		},
		{
			func() {
				FakeConfig(map[string]string{"log/type": "syslog"})
				InitLogError = initLog()
			},
			LogDEBUG, false, errors.New("dial unixgram var/log.sock: no such file or directory"),
		},
		{
			func() { fakeLog() },
			LogINFO, true, nil,
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": "ERR"})
				InitLogError = initLog()
			},
			LogERR, true, nil,
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": "WARN"})
				InitLogError = initLog()
			},
			LogWARN, true, nil,
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": "NOTICE"})
				InitLogError = initLog()
			},
			LogNOTICE, true, nil,
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": "INFO"})
				InitLogError = initLog()
			},
			LogINFO, true, nil,
		},
		{
			func() {
				FakeConfig(map[string]string{"log/level": "DEBUG"})
				InitLogError = initLog()
			},
			LogDEBUG, true, nil,
		},
	}
	for _, c := range cases {
		c.setup()
		if logLevel != c.level {
			t.Errorf("logLevel = %v, want %v", logLevel, c.level)
		}
		if (syslogLogger != nil) != c.ready {
			if c.ready {
				t.Errorf("syslogLogger = %v, want !=nil", syslogLogger)
			} else {
				t.Errorf("syslogLogger = %v, want nil", syslogLogger)
			}
		}
		if (InitLogError == nil) != (c.wanterr == nil) || InitLogError != nil && InitLogError.Error() != c.wanterr.Error() {
			t.Errorf("InitLogError = %v, want %v", InitLogError, c.wanterr)
		}
	}
}

func TestLogLevel(t *testing.T) {
	var level LogLevel
	cases := []struct {
		want LogLevel
		str  string
		next LogLevel
	}{
		{LogDEBUG, "DEBUG", LogINFO},
		{LogINFO, "INFO", LogNOTICE},
		{LogNOTICE, "NOTICE", LogWARN},
		{LogWARN, "WARN", LogERR},
		{LogERR, "ERR", LogERR + 1},
	}
	for _, c := range cases {
		if level != c.want {
			t.Errorf("level = %v, want %v", level, c.want)
		}
		if level.String() != c.str {
			t.Errorf("level.String() = %s, want %s", level, c.str)
		}
		if level >= c.next {
			t.Errorf("level = %v must be less than %v", level, c.next)
		}
		level++
	}
	if level.String() != "UNKNOWN" {
		t.Errorf("level.String() = %s, want UNKNOWN", level)
	}
}

func TestNewLog(t *testing.T) {
	l := NewLog("")
	if l.Prefix() != "" {
		t.Errorf("l.Prefix() = %v, want ``", l.Prefix())
	}
	l = NewLog("pre fix:")
	if l.Prefix() != "pre fix:" {
		t.Errorf("l.Prefix() = %v, want `pre fix:`", l.Prefix())
	}
}

const (
	lineERR    = "<11>: "
	lineWARN   = "<12>: "
	lineNOTICE = "<13>: "
	lineINFO   = "<14>: "
	lineDEBUG  = "<15>: "
)

func TestLog(t *testing.T) {
	l := NewLog("")

	l.ERR("---8<---")
	line := <-fromSyslog
	if line != lineERR+"---8<---" {
		t.Fatalf("fromSyslog = %v, want %v", line, lineERR+"---8<---")
	}

	wantlines := []string{}
	lines := getLines()
	if !reflect.DeepEqual(lines, wantlines) {
		t.Errorf("no log\nexp: %#v\ngot: %#v", wantlines, lines)
	}

	logLevel = LogDEBUG
	l.ERR("%%10")
	l.WARN("%%20")
	l.NOTICE("%%30")
	l.INFO("%%40")
	l.DEBUG("%%50")
	l.Print("%%60")
	l.Printf("%%61")
	l.Println("%%62")
	if pnk := getpnk(func() { l.Panic("%%70") }); !reflect.DeepEqual(pnk, "%%70") {
		t.Errorf("all+plain, panic=%#v, want %#v", pnk, "%%70")
	}
	if pnk := getpnk(func() { l.Panicf("%%71") }); !reflect.DeepEqual(pnk, "%71") {
		t.Errorf("all+plain, panic=%#v, want %#v", pnk, "%71")
	}
	if pnk := getpnk(func() { l.Panicln("%%72") }); !reflect.DeepEqual(pnk, "%%72\n") {
		t.Errorf("all+plain, panic=%#v, want %#v", pnk, "%%72\n")
	}
	wantlines = []string{
		lineERR + "%%10", lineWARN + "%%20", lineNOTICE + "%%30", lineINFO + "%%40", lineDEBUG + "%%50",
		lineNOTICE + "%%60", lineNOTICE + "%61", lineNOTICE + "%%62",
		lineERR + "%%70", lineERR + "%71", lineERR + "%%72",
	}
	lines = getLines()
	if !reflect.DeepEqual(lines, wantlines) {
		t.Errorf("all+plain\nexp: %#v\ngot: %#v", wantlines, lines)
	}

	logLevel = LogNOTICE
	l.ERR("%%1%d", 1)
	l.WARN("%%2%d", 1)
	l.NOTICE("%%3%d", 1)
	l.INFO("%%4%d", 1)
	l.DEBUG("%%5%d", 1)
	l.Print("%%6%d", 0)
	l.Printf("%%6%d", 1)
	l.Println("%%6%d", 2)
	if pnk := getpnk(func() { l.Panic("%%7%d", 0) }); !reflect.DeepEqual(pnk, "%%7%d0") {
		t.Errorf("level+sprintf, panic=%#v, want %#v", pnk, "%%7%d0")
	}
	if pnk := getpnk(func() { l.Panicf("%%7%d", 1) }); !reflect.DeepEqual(pnk, "%71") {
		t.Errorf("level+sprintf, panic=%#v, want %#v", pnk, "%71")
	}
	if pnk := getpnk(func() { l.Panicln("%%7%d", 2) }); !reflect.DeepEqual(pnk, "%%7%d 2\n") {
		t.Errorf("level+sprintf, panic=%#v, want %#v", pnk, "%%7%d 2\n")
	}
	wantlines = []string{
		lineERR + "%11", lineWARN + "%21", lineNOTICE + "%31",
		lineNOTICE + "%%6%d0", lineNOTICE + "%61", lineNOTICE + "%%6%d 2",
		lineERR + "%%7%d0", lineERR + "%71", lineERR + "%%7%d 2",
	}
	lines = getLines()
	if !reflect.DeepEqual(lines, wantlines) {
		t.Errorf("level+sprintf\nexp: %#v\ngot: %#v", wantlines, lines)
	}

	logLevel = LogWARN
	l.Print("63")
	if pnk := getpnk(func() { l.Panic("73") }); !reflect.DeepEqual(pnk, "73") {
		t.Errorf("level+sprint, panic=%#v, want %#v", pnk, "73")
	}
	wantlines = []string{lineERR + "73"}
	lines = getLines()
	if !reflect.DeepEqual(lines, wantlines) {
		t.Errorf("level+sprint\nexp: %#v\ngot: %#v", wantlines, lines)
	}

	l = NewLog("pfx")
	l.ERR("12")
	l.WARN("2%d", 2)
	wantlines = []string{lineERR + "pfx12", lineWARN + "pfx22"}
	lines = getLines()
	if !reflect.DeepEqual(lines, wantlines) {
		t.Errorf("prefix\nexp: %#v\ngot: %#v", wantlines, lines)
	}

	buf := bytes.NewBufferString("")
	log.SetOutput(buf)
	origSyslogLogger := syslogLogger
	syslogLogger = nil
	l.ERR("13")
	l.WARN("2%d", 3)
	syslogLogger = origSyslogLogger
	wantlines = []string{}
	lines = getLines()
	if !reflect.DeepEqual(lines, wantlines) {
		t.Errorf("fallback\nexp: %#v\ngot: %#v", wantlines, lines)
	}
	want := "ERR: pfx13\nWARN: pfx23\n"
	if buf.String() != want {
		t.Errorf("fallback buf=%q, want %q", buf.String(), want)
	}
	buf.Reset()

	err := cmd.Process.Kill()
	if err != nil {
		t.Error(err)
	}
	cmd.Wait()
	l.ERR("14")
	l.WARN("2%d", 4)
	want = "ERR: pfx14\nWARN: pfx24\n"
	if buf.String() != want {
		t.Errorf("fallback buf=%q, want %q", buf.String(), want)
	}

	InitLogError = initLog()
}

func getLines() []string {
	NewLog("").ERR("---8<---")
	var res = make([]string, 0)
	for {
		line := <-fromSyslog
		if line == lineERR+"---8<---" {
			break
		}
		res = append(res, line)
	}
	return res
}

func getpnk(f func()) (pnk interface{}) {
	defer func() {
		pnk = recover()
	}()
	f()
	return
}
