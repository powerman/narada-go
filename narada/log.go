package narada

import (
	"errors"
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path"
)

var logLevel = LogDEBUG
var syslogLogger *syslog.Writer

var InitLogError = initLog()

// initLog is not thread-safe.
func initLog() (err error) {
	defer func() {
		if err != nil {
			logLevel = LogDEBUG
			syslogLogger = nil
		}
	}()

	lock, err := SharedLock(0)
	if err != nil {
		return err
	}
	defer lock.UnLock()

	output := GetConfigLine("log/output")
	if len(output) == 0 {
		return errors.New("require non-empty config/log/output")
	}

	switch level := GetConfigLine("log/level"); level {
	case "ERR":
		logLevel = LogERR
	case "WARN":
		logLevel = LogWARN
	case "NOTICE":
		logLevel = LogNOTICE
	case "INFO":
		logLevel = LogINFO
	case "DEBUG":
		logLevel = LogDEBUG
	default:
		return errors.New("unsupported config/log/level: " + level)
	}

	logtype := GetConfigLine("log/type")
	switch logtype {
	case "":
		logtype = "syslog"
	case "syslog":
	default:
		return errors.New("unsupported config/log/type: " + logtype)
	}

	syslogLogger, err = syslog.Dial("unixgram", output, syslog.LOG_NOTICE|syslog.LOG_USER, path.Base(os.Args[0]))
	if err != nil {
		return err
	}
	log.SetFlags(0)
	return nil
}

type LogLevel byte

const (
	LogDEBUG LogLevel = iota
	LogINFO
	LogNOTICE
	LogWARN
	LogERR
)

func (level LogLevel) String() string {
	switch level {
	case LogDEBUG:
		return "DEBUG"
	case LogINFO:
		return "INFO"
	case LogNOTICE:
		return "NOTICE"
	case LogWARN:
		return "WARN"
	case LogERR:
		return "ERR"
	}
	return "UNKNOWN"
}

type Log struct {
	prefix string
}

func NewLog(prefix string) *Log {
	return &Log{prefix: prefix}
}

func (l Log) Prefix() string {
	return l.prefix
}

func (l Log) Print(v ...interface{}) {
	l.write(LogNOTICE, fmt.Sprint(v...))
}

func (l Log) Printf(format string, v ...interface{}) {
	l.write(LogNOTICE, fmt.Sprintf(format, v...))
}

func (l Log) Println(v ...interface{}) {
	l.write(LogNOTICE, fmt.Sprintln(v...))
}

func (l Log) Fatal(v ...interface{}) {
	l.write(LogERR, fmt.Sprint(v...))
	os.Exit(1)
}

func (l Log) Fatalf(format string, v ...interface{}) {
	l.write(LogERR, fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l Log) Fatalln(v ...interface{}) {
	l.write(LogERR, fmt.Sprintln(v...))
	os.Exit(1)
}

func (l Log) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.write(LogERR, s)
	panic(s)
}

func (l Log) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.write(LogERR, s)
	panic(s)
}

func (l Log) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.write(LogERR, s)
	panic(s)
}

func (l Log) ERR(format string, v ...interface{}) {
	l.write(LogERR, format, v...)
}

func (l Log) WARN(format string, v ...interface{}) {
	l.write(LogWARN, format, v...)
}

func (l Log) NOTICE(format string, v ...interface{}) {
	l.write(LogNOTICE, format, v...)
}

func (l Log) INFO(format string, v ...interface{}) {
	l.write(LogINFO, format, v...)
}

func (l Log) DEBUG(format string, v ...interface{}) {
	l.write(LogDEBUG, format, v...)
}

func (l Log) write(level LogLevel, msg string, v ...interface{}) {
	if logLevel > level {
		return
	}
	if len(v) != 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	msg = l.prefix + msg

	if syslogLogger == nil {
		log.Print(level.String() + ": " + msg)
	} else {
		var err error
		switch level {
		case LogDEBUG:
			err = syslogLogger.Debug(msg)
		case LogINFO:
			err = syslogLogger.Info(msg)
		case LogNOTICE:
			err = syslogLogger.Notice(msg)
		case LogWARN:
			err = syslogLogger.Warning(msg)
		default:
			err = syslogLogger.Err(msg)
		}
		if err != nil {
			log.Print(level.String() + ": " + msg)
		}
	}
}
