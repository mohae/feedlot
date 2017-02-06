package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

// holds the tmpLogFilename, not used after SetLogging()
var tmpLogFile string

func init() {
	jww.SetLogThreshold(getJWWLevel(contour.GetString(LogLevelFile)))
	jww.SetStdoutThreshold(getJWWLevel(contour.GetString(LogLevelStdOut)))
}

// SetLogging sets application logging settings.
func SetLogging() error {
	// Set LogLevels
	jww.SetLogThreshold(getJWWLevel(contour.GetString(LogLevelFile)))
	jww.SetStdoutThreshold(getJWWLevel(contour.GetString(LogLevelStdOut)))

	// set output
	logfile := contour.GetString(LogFile)
	jww.FEEDBACK.Printf("Log output will be written to %s\n", logfile)
	if logfile == stderr {
		jww.SetLogOutput(os.Stderr)
		return nil
	}
	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: error: open logfile: %s ", Name, err)
		return err
	}
	jww.SetLogOutput(f)
	return nil
}

func getJWWLevel(level string) jww.Threshold {
	level = strings.ToUpper(level)
	switch level {
	case "TRACE":
		return jww.LevelTrace
	case "DEBUG":
		return jww.LevelDebug
	case "INFO":
		return jww.LevelInfo
	case "WARN":
		return jww.LevelWarn
	case "ERROR":
		return jww.LevelError
	case "CRITICAL":
		return jww.LevelCritical
	case "FATAL":
		return jww.LevelFatal
	}
	// It should never get to this...but if it does return a valid level
	return jww.LevelError
}
