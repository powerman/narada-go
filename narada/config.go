package narada

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const dir = "config/"

var invalidName = regexp.MustCompile(`(?:\A|/)[.][.]?/`)
var validName = regexp.MustCompile(`\A(?:[\w.-]+/)*[\w.-]+\z`)

var open = func(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// FakeConfig make GetConfig() return values from configs (keys are config
// names) instead of real files. If configs doesn't have a key for some
// config file - it will work as usually, by reading real file.
func FakeConfig(configs map[string]string) {
	open = func(name string) (io.ReadCloser, error) {
		if content, ok := configs[name[len(dir):]]; ok {
			return ioutil.NopCloser(strings.NewReader(content)), nil
		}
		return os.Open(name)
	}
}

// GetConfig return contents of file "config/"+path.
// If file not exists it will return nil without any error.
// Panics on invalid config name.
func GetConfig(path string) ([]byte, error) {
	if invalidName.MatchString(path) || !validName.MatchString(path) {
		panic("invalid config name: " + path)
	}
	file, err := open(dir + path)
	defer file.Close()
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

// GetConfigInt return integer from first line of file "config/"+path.
// If file not exists it will return 0.
// Panics if unable to read config or it contain more than one line or
// that line doesn't contain one integer.
func GetConfigInt(path string) int {
	str := GetConfigLine(path)
	if str == "" {
		return 0
	}
	i, err := strconv.Atoi(strings.Trim(str, " \t\r"))
	if err != nil {
		panic("config " + path + " must contain integer")
	}
	return i
}
