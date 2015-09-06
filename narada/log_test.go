package narada

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"time"
)

func init() {
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
	openLog()
}
