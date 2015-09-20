package narada

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"
)

func fakeLog() {
	cmd := exec.Command("socat", "UNIX-RECV:var/log.sock,mode=666,unlink-early", "STDOUT")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	go ioutil.ReadAll(stdout)
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
	if logLevel != LogDEBUG {
		t.Errorf("logLevel = %v, want DEBUG", logLevel)
	}
	if syslogLogger != nil {
		t.Errorf("syslogLogger = %v, want nil", syslogLogger)
	}
	wanterr := errors.New("require non-empty config/log/output")
	if fmt.Sprintf("%#v", InitLogError) != fmt.Sprintf("%#v", wanterr) {
		t.Errorf("InitLogError = %#v, want %#v", InitLogError, wanterr)
	}

	InitLogError = initLog()
	if logLevel != LogINFO {
		t.Errorf("logLevel = %v, want INFO", logLevel)
	}
	if syslogLogger != nil {
		t.Errorf("syslogLogger = %v, want nil", syslogLogger)
	}
	wanterr = errors.New("dial unixgram var/log.sock: no such file or directory")
	if InitLogError == nil || InitLogError.Error() != wanterr.Error() {
		t.Errorf("InitLogError = %s, want %s", InitLogError, wanterr)
	}

	fakeLog()
	if logLevel != LogINFO {
		t.Errorf("logLevel = %v, want INFO", logLevel)
	}
	if syslogLogger == nil {
		t.Errorf("syslogLogger = nil, want non-nil")
	}
	if InitLogError != nil {
		t.Errorf("InitLogError = %s, want nil", InitLogError)
	}
}
