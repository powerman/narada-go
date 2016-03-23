package narada

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const configDir = "config/"

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
		if content, ok := configs[name[len(configDir):]]; ok {
			return ioutil.NopCloser(strings.NewReader(content)), nil
		}
		return os.Open(name)
	}
}

// GetConfig returns contents of file "config/"+path.
// If file not exists it will return nil without any error.
// Panics on invalid config name.
func GetConfig(path string) ([]byte, error) {
	if invalidName.MatchString(path) || !validName.MatchString(path) {
		panic("invalid config name: " + path)
	}
	lock, err := SharedLock(0)
	if err != nil {
		return nil, err
	}
	defer lock.UnLock()
	file, err := open(configDir + path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()
	return ioutil.ReadAll(file)
}

// GetConfigLine returns first line of file "config/"+path.
// If file not exists it will return empty string.
// Panics if unable to read file or it contains more than one line.
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

// GetConfigInt returns integer from first line of file "config/"+path.
// If file not exists or empty it will return 0.
// Panics if unable to read file or it contains more than one line or
// that line doesn't contain one integer.
func GetConfigInt(path string) int {
	str := GetConfigLine(path)
	if str == "" {
		return 0
	}
	i, err := strconv.Atoi(strings.TrimSpace(str))
	if err != nil {
		panic("config " + path + " must contain integer")
	}
	return i
}

// GetConfigIntBetween panics if value returned by GetConfigInt(path)
// is less than min or greater than max.
func GetConfigIntBetween(path string, min, max int) int {
	i := GetConfigInt(path)
	if i < min {
		panic(fmt.Sprintf("config %s must contain integer >= %d", path, min))
	}
	if i > max {
		panic(fmt.Sprintf("config %s must contain integer <= %d", path, max))
	}
	return i
}

// GetConfigDuration returns duration parsed from first line of file "config/"+path.
// Panics if file not exists or empty or unable to read file or
// file contains more than one line or that line doesn't contain duration
// (see time.ParseDuration).
func GetConfigDuration(path string) time.Duration {
	d, err := time.ParseDuration(GetConfigLine(path))
	if err != nil {
		panic("config " + path + " must contain duration")
	}
	return d
}

// GetConfigDurationBetween panics if value returned by GetConfigDuration(path)
// is less than min or greater than max.
func GetConfigDurationBetween(path string, min, max time.Duration) time.Duration {
	d := GetConfigDuration(path)
	if d < min {
		panic(fmt.Sprintf("config %s must contain duration >= %s", path, min))
	}
	if d > max {
		panic(fmt.Sprintf("config %s must contain duration <= %s", path, max))
	}
	return d
}
