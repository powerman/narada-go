package narada

import (
	"bytes"
	"io/ioutil"
)

// Version returns project's version.
func Version() (version string, err error) {
	lock, err := SharedLock(0)
	if err != nil {
		return
	}
	buf, err := ioutil.ReadFile("VERSION")
	lock.UnLock()
	if err != nil {
		return
	}
	return string(bytes.TrimRight(buf, " \r\n")), nil
}
