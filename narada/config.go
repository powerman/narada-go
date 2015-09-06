package narada

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
)

var invalidName = regexp.MustCompile(`\A[.][.]?/`)
var validName = regexp.MustCompile(`\A(?:[\w.-]+/)*[\w.-]+\z`)

func GetConfig(name string) ([]byte, error) {
	if invalidName.MatchString(name) || !validName.MatchString(name) {
		panic("invalid config name: " + name)
	}
	file, err := os.Open("config/" + name)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return ioutil.ReadAll(file)
}

func GetConfigLine(name string) string {
	cfg, err := GetConfig(name)
	if err != nil {
		panic(err)
	}
	if cfg == nil {
		return ""
	}
	if n := bytes.IndexByte(cfg, byte('\n')); n >= 0 {
		if len(bytes.TrimSpace(cfg[n:])) != 0 {
			panic("config " + name + " contain more than one line")
		}
		cfg = cfg[:n]
	}
	return string(cfg)
}
