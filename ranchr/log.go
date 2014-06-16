package ranchr

import (
	"errors"
	"io"
	
	seelog "github.com/cihub/seelog"
)
func init() {
	DisableLog()
}

// Logger setup stuff from:
//	github.com/cihub/seelog/wiki/Writing-libraries-with-Seelog
func DisableLog() {
	logger = seelog.Disabled
}

// UseLogger uses a specified seelog.LoggerInterface to output library log.
// This func is used when Seelog logging system is being used in app.
func UseLogger(newLogger seelog.LoggerInterface) {
	logger = newLogger
}

// SetLogWriter uses a specified io.Writer to output library log.
// Use this func if you are not using Seelog logging system in your app.
func SetLogWriter(writer io.Writer) error {
	if writer == nil {
		return errors.New("Nil writer")
	}
	
	newLogger, err := seelog.LoggerFromWriterWithMinLevel(writer, seelog.TraceLvl)
	if err != nil {
		return err
	}
	
	UseLogger(newLogger)
	return nil
}

// Call this before app shutdown.
func FlushLog() {
	logger.Flush()
}

