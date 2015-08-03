package app

import (
	"fmt"
	"log"
	"os"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

// holds the tmpLogFilename, not used after SetLogging()
var tmpLogFile string

// SetTempLogging creates a temp logfile in the /tmp and enables logging. This
// is to support logging of operations prior to processing the command-line
// flags, at which point SetLogging will either move this to the actual log
// location or remove the temp logfile.
func SetTempLogging() {
	// use temp logfile
	jww.UseTempLogFile("rancher")
	jww.SetLogThreshold(getJWWLevel(contour.GetString(LogLevelFile)))
	jww.SetStdoutThreshold(getJWWLevel(contour.GetString(LogLevelStdOut)))
	jww.TRACE.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	jww.DEBUG.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	jww.INFO.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	jww.WARN.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	jww.ERROR.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	jww.CRITICAL.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	jww.FATAL.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

// SetLogging sets application logging settings.
func SetLogging() error {
	// Check to see if logging is enabled, if not, discard the temp logfile and remove.

	if !contour.GetBool(Log) {
		tmpFile := jww.LogHandle.(*os.File).Name()
		jww.LogHandle.(*os.File).Close()
		jww.DiscardLogging()
		os.Remove(tmpFile)
		return nil
	}
	logfile := contour.GetString(LogFile)
	fname, err := getUniqueFilename(logfile, "2006-01-02")
	if err != nil {
		err = fmt.Errorf("unable to continue: cannot obtain unique log filename: %s", err)
		jww.FEEDBACK.Println(err)
		return err
	}
	// if the names aren't the same, the logfile already exists. Rename it to the fname
	if fname != logfile {
		err := os.Rename(logfile, fname)
		if err != nil {
			err = fmt.Errorf("unable to continuecannot rename existing logfile: %s", err)
			jww.FEEDBACK.Println(err)
			return err
		}
	}
	// make the tmpLogFile the actual logfile
	err = os.Rename(jww.LogHandle.(*os.File).Name(), logfile)
	if err != nil {
		err = fmt.Errorf("unable to contineu: cannot rename the temp log to %s", err)
		jww.FEEDBACK.Println(err)
		return err
	}

	// Set LogLevels
	jww.SetLogThreshold(getJWWLevel(contour.GetString(LogLevelFile)))
	jww.SetStdoutThreshold(getJWWLevel(contour.GetString(LogLevelStdOut)))
	return nil
}

func getJWWLevel(level string) jww.Level {
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
	return jww.LevelInfo
}
