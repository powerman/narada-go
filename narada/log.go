package narada

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path"
)

type LogLevel byte

const (
	LogDEBUG LogLevel = iota
	LogINFO
	LogNOTICE
	LogWARN
	LogERR
)

type Logger struct {
	Level  LogLevel
	syslog *syslog.Writer
	opened bool
}

func (l Logger) write(level LogLevel, f func(string) error, format string, v ...interface{}) error {
	if !l.opened {
		panic("OpenLog() must be called first")
	}
	if l.Level > level {
		return nil
	}
	return f(fmt.Sprintf(format, v...))
}

func (l Logger) ERR(format string, v ...interface{}) error {
	return l.write(LogERR, l.syslog.Err, format, v...)
}

func (l Logger) WARN(format string, v ...interface{}) error {
	return l.write(LogWARN, l.syslog.Warning, format, v...)
}

func (l Logger) NOTICE(format string, v ...interface{}) error {
	return l.write(LogNOTICE, l.syslog.Notice, format, v...)
}

func (l Logger) INFO(format string, v ...interface{}) error {
	return l.write(LogINFO, l.syslog.Info, format, v...)
}

func (l Logger) DEBUG(format string, v ...interface{}) error {
	return l.write(LogDEBUG, l.syslog.Debug, format, v...)
}

var Log Logger

func init() {
	log.SetFlags(0)
}

func OpenLog() {
	Log.opened = true

	output := GetConfigLine("log/output")
	if len(output) == 0 {
		panic("require non-empty config/log/output")
	}

	logtype := GetConfigLine("log/type")
	switch logtype {
	case "":
		logtype = "syslog"
	case "syslog":
	default:
		panic("unsupported config/log/type: " + logtype)
	}

	switch level := GetConfigLine("log/level"); level {
	case "ERR":
		Log.Level = LogERR
	case "WARN":
		Log.Level = LogWARN
	case "NOTICE":
		Log.Level = LogNOTICE
	case "INFO":
		Log.Level = LogINFO
	case "DEBUG":
		Log.Level = LogDEBUG
	default:
		panic("unsupported config/log/level: " + level)
	}

	var err error
	Log.syslog, err = syslog.Dial("unixgram", output, syslog.LOG_NOTICE|syslog.LOG_USER, path.Base(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
}
