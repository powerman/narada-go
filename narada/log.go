package narada

import (
	"log"
	"log/syslog"
	"os"
	"path"
)

// TODO filter log levels
// TODO improve narada-viewlog
var Log *syslog.Writer

func init() {
	log.SetFlags(0)
}

func connectLog() {
	// TODO read from config files
	var err error
	Log, err = syslog.Dial("unixgram", "var/log.sock", syslog.LOG_INFO|syslog.LOG_USER, path.Base(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
}
