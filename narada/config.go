package narada

import (
	"bytes"
	"io/ioutil"
	"os"
	"regexp"
)

var invalidName = regexp.MustCompile(`\A[.][.]?/`)
var validName = regexp.MustCompile(`\A(?:[\w.-]+/)*[\w.-]+\z`)

// GetConfig return contents of file "config/"+path.
// If file not exists it will return nil without any error.
// Panics on invalid config name.
func GetConfig(path string) ([]byte, error) {
	if invalidName.MatchString(path) || !validName.MatchString(path) {
		panic("invalid config name: " + path)
	}
	file, err := os.Open("config/" + path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	return ioutil.ReadAll(file)
}

// GetConfigLine return first line of file "config/"+path.
// If file not exists it will return empty string.
// Panics if unable to read config or it contain more than one line.
func GetConfigLine(path string) string {
	cfg, err := GetConfig(path)
	if err != nil {
		panic(err)
	}
	if cfg == nil {
		return ""
	}
	if n := bytes.IndexByte(cfg, byte('\n')); n >= 0 {
		if len(bytes.TrimSpace(cfg[n:])) != 0 {
			panic("config " + path + " contain more than one line")
		}
		cfg = cfg[:n]
	}
	return string(cfg)
}
