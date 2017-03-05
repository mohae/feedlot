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
	level   Level
	verbose bool
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

// SetLogging sets application logging settings and verbose output.
func Set() error {
	if contour.GetBool(conf.Verbose) {
		verbose = true
	}

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
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.LUTC)
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
	log.Printf("error: "+format, v...)
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
	log.Printf(format, v...)
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
	log.Printf(format, v...)
}

// Verbose writes the value to stdout as a line if verbose output is enabled.
func Verbose(v interface{}) {
	if !verbose {
		return
	}
	fmt.Printf("%v\n", v)
}

// Verbose writes the value to stdout using the provided format and data, if
// verbose output is enabled.
func Verbosef(format string, v ...interface{}) {
	if !verbose {
		return
	}
	fmt.Printf(format, v...)
}
