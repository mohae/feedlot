package log

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/mohae/contour"
	"github.com/mohae/feedlot/conf"
)

// Logging errors is always on and defaults output to os.Stderr. This can be
// set to a file with the 'logfile' flag.
//
// Additionally, a loglevel can be specified for more detailed logging:
//    info   basic operational information.
//    debug  detailed information to help debug what the application is doing.
//
// Also, verbose is an option that will write information about what operations
// are being performed to os.Stdout.

var (
	level Level
)

//go:generate stringer -type=Level
type Level int

const (
	LogNone Level = iota
	LogError
	LogInfo
	LogDebug
)

type LevelErr struct {
	s string
}

func (l LevelErr) Error() string {
	return fmt.Sprintf("unknown log level: %s", l.s)
}

func parseLevel(s string) (Level, error) {
	v := strings.ToLower(s)
	switch v {
	case "none":
		return LogNone, nil
	case "error":
		return LogError, nil
	case "info":
		return LogInfo, nil
	case "debug":
		return LogDebug, nil
	default:
		return LogNone, LevelErr{s}
	}
}

type UnknownLogFlagErr struct {
	v string
}

func (e UnknownLogFlagErr) Error() string {
	return "unknown log flag: " + e.v
}

func ParseLogFlag(s string) (l int, err error) {
	v := strings.ToLower(s)
	switch v {
	case "ldate":
		return log.Ldate, nil
	case "ltime":
		return log.Ltime, nil
	case "lmicroseconds":
		return log.Lmicroseconds, nil
	case "llongfile":
		return log.Llongfile, nil
	case "lshortfile":
		return log.Lshortfile, nil
	case "lutc":
		return log.LUTC, nil
	case "lstdflags":
		return log.LstdFlags, nil
	case "none":
		return 0, nil
	}
	return 0, UnknownLogFlagErr{s}
}

// SetLogging sets application logging settings and verbose output.
func Set() error {
	var err error
	level, err = parseLevel(contour.GetString(conf.LogLevel))
	if err != nil {
		return err
	}
	if level == LogNone {
		log.SetOutput(ioutil.Discard)
		return nil
	}
	if contour.GetString(conf.LogFile) != "stdout" {
		f, err := os.OpenFile(contour.GetString(conf.File), os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0664)
		if err != nil {
			return fmt.Errorf("open logfile: %s", err)
		}
		log.SetOutput(f)
	}
	// set the log flags
	val := contour.GetString(conf.LogFlags)
	if val == "" { // if no flags were specified use log's default.
		return nil
	}
	if strings.ToLower(val) == "none" { // if no flags, unset
		log.SetFlags(0)
		return nil
	}
	// otherwise process the flags
	vals := strings.Split(val, ",")
	var flg int
	for _, v := range vals {
		i, err := ParseLogFlag(strings.TrimSpace(v))
		if err != nil {
			return err
		}
		flg |= i
	}
	log.SetFlags(flg)
	return nil
}

// Error writes an error entry to the log. If the Level == LogNone, nothing
// will be written. Even though
func Error(v interface{}) {
	if level == LogNone {
		return
	}
	log.Printf("error: %v", v)
}

// Errorf writes an error entry to the log using the provided format and data.
// If the Level == LogNone, nothing will be written. Even though
func Errorf(format string, v ...interface{}) {
	if level == LogNone {
		return
	}
	log.Printf(fmt.Sprintf("error: %s", format), v...)
}

// Info writes an info entry to the log. If the Level < LogInfo, nothing will
// be written.
func Info(v interface{}) {
	if level < LogInfo {
		return
	}
	log.Printf("info: %v", v)
}

// Infof writes an info entry to the log using the provided format and data.
// If the Level < LogInfo, nothing will be written.
func Infof(format string, v ...interface{}) {
	if level < LogInfo {
		return
	}
	log.Printf(fmt.Sprintf("info: %s", format), v...)
}

// Debug writes a debug entry to the log. If the Level < LogDebug, nothing
// will be written.
func Debug(v interface{}) {
	if level < LogInfo {
		return
	}
	log.Printf("info: %v", v)
}

// Debugf writes a debug entry to the log using the provided format and data.
// If the Level < LogDebug, nothing will be written.
func Debugf(format string, v ...interface{}) {
	if level < LogInfo {
		return
	}
	log.Printf(fmt.Sprintf("debug: %s", format), v...)
}
