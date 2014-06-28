// Main entry point for Rancher.
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"fmt"
	"os"
	_"time"

	"github.com/mitchellh/cli"
	"github.com/mohae/rancher/ranchr"
	jww "github.com/spf13/jwalterweatherman"
)

func main() {
	// main wraps runMain() and ensures that the log gets flushed prior to exit.
	// Exit with return code from runMain()
	rc := runMain()
	os.Exit(rc)
}
func runMain() int {
	// runMain parses the Flag for glog, sets up CLI stuff for the supported
	// subcommands and runs Rancher.
	var err error
	if err = ranchr.SetEnv(); err != nil {
		fmt.Println("An error while processing Rancher Environment variables: ", err.Error())
		return -1
	}
	
	// Logging setup
	SetLogging()

	jww.INFO.Printf("Rancher starting with args: %v", os.Args[:])
	args := os.Args[1:]
	// Get the command line args. We shortcut "--version" and "-v" to
	// just show the version.
	for _, arg := range args {
		if arg == "-v" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}
	cli := &cli.CLI{
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc("rancher"),
	}
	exitCode, err := cli.Run()
	if err != nil {
		jww.ERROR.Printf("Rancher encountered an error: %s\n", err.Error())
	}
	jww.INFO.Printf("Rancher exiting with an exit code of %v", exitCode)
	// TODO Force sleep because log messages seem to be dropped...this needs to be addressed
	//	time.Sleep(time.Millisecond * 5000)
	return exitCode
}

// Set application logging. If the
func SetLogging() {
	// By default, jww has sane log level config; e.g. only log Warn and Above when a
	// io.writer is present. Error and above get printed to stdout. Info and lower
	// get sent to /dev/null.

	// Set custom levels for output, if they are set
	if ranchr.AppConfig.LogLevelStdout != "" {
		res := getJWWLevel(ranchr.AppConfig.LogLevelStdout)
		jww.SetStdoutThreshold(res)
	}

	if ranchr.AppConfig.LogLevelFile != "" {
		res := getJWWLevel(ranchr.AppConfig.LogLevelFile)
		jww.SetLogThreshold(res)
	}

	// Take care of log output stuff
	if ranchr.AppConfig.LogToFile {
		// if the filename isn't set, use a temp log file
		if ranchr.AppConfig.LogFilename == "" {
			jww.UseTempLogFile("rancher")
		} else {
			jww.SetLogFile(ranchr.AppConfig.LogFilename)
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
