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
	defer lock.UnLock()
	buf, err := ioutil.ReadFile("VERSION")
	if err != nil {
		return
	}
	return string(bytes.TrimRight(buf, " \r\n")), nil
}
