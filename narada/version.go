package narada

import (
	"bytes"
	"io/ioutil"
	"log"
	"time"
)

// Project's version.
var VERSION string

func init() {
	Bootstrap(func() {
		buf, err := ioutil.ReadFile("VERSION")
		if err != nil {
			log.Fatalf("unable to read VERSION: %v", err)
		}
		VERSION = string(bytes.TrimRight(buf, " \r\n"))
	})
}

const BootstrapLockWait = 15 * time.Second

func Bootstrap(f func()) {
	lock, err := SharedLock(BootstrapLockWait)
	if err != nil {
		log.Fatal("can't get bootstrap lock")
	}
	defer lock.UnLock()
	f()
}
