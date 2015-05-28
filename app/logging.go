package app

import (
	jww "github.com/spf13/jwalterweatherman"
)

// SetLogging sets application logging settings.
func SetLogging() {
	// By default, jww has sane log level config; e.g. only log Warn and Above when a
	// io.writer is present. Error and above get printed to stdout. Info and lower
	// get sent to /dev/null.
	// Set custom levels for output, if they are set
	if AppConfig.LogLevelStdout != "" {
		res := getJWWLevel(AppConfig.LogLevelStdout)
		jww.SetStdoutThreshold(res)
	}
	if AppConfig.LogLevelFile != "" {
		res := getJWWLevel(AppConfig.LogLevelFile)
		jww.SetLogThreshold(res)
	}
	// Take care of log output stuff
	if AppConfig.LogToFile {
		// if the filename isn't set, use a temp log file
		if AppConfig.LogFilename == "" {
			jww.UseTempLogFile("rancher")
		} else {
			jww.SetLogFile(AppConfig.LogFilename)
		}
	}
	return
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
