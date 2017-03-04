package log

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mohae/contour"
	"github.com/mohae/feedlot/app"
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
	return fmt.Sprintf("unknown loglevel: %s", l.s)
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

func init() {
	log.SetPrefix(filepath.Base(os.Args[0]))
}

// SetLogging sets application logging settings and verbose output.
func Set() error {
	if contour.GetBool(app.Verbose) {
		verbose = true
	}

	var err error
	level, err = parseLevel(contour.GetString(app.LogLevel))
	if err != nil {
		return err
	}
	if level == LogNone {
		log.SetOutput(ioutil.Discard)
		return nil
	}
	if contour.GetString(app.LogFile) != "stdout" {
		f, err := os.OpenFile(contour.GetString(app.LogFile), os.O_CREATE|os.O_APPEND|os.O_RDONLY, 0664)
		if err != nil {
			return fmt.Errorf("open logfile: %s", err)
		}
		log.SetOutput(f)
	}
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.LUTC)
	return nil
}

// Error writes an error entry to the log. If the Level == LogNone, nothing
// will be written.
func Error(s string) {
	if level == LogNone {
		return
	}
	log.Printf("%s: error: %s", app.Name, s)
}

// Info writes an info entry to the log. If the Level < LogInfo, nothing will
// be written.
func Info(s string) {
	if level < LogInfo {
		return
	}
	log.Printf("%s: info: %s", app.Name, s)
}

// Debug writes a debug entry to the log. If the Level < LogDebug, nothing
// will be written.
func Debug(s string) {
	if level < LogInfo {
		return
	}
	log.Printf("%s: info: %s", app.Name, s)
}

// Verbose writes the string to stdout if verbose output is enabled.
func Verbose(s string) {
	if !verbose {
		return
	}
	fmt.Println(s)
}
