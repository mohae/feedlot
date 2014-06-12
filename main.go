// Main entry point for Rancher.
//
//
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"fmt"
	//flag"
	//"io/ioutil"
	"os"

	"github.com/mitchellh/cli"
	"github.com/mohae/rancher/ranchr"
	log "github.com/cihub/seelog"
)

var Logger log.LoggerInterface

func init() {
	// Set by default

}

// main wraps runMain() and ensures that the log gets flushed prior to exit.
func main() {
	// Exit with return code from runMain()
	os.Exit(runMain())
}

// runMain parses the Flag for glog, sets up CLI stuff for the supported sub-
// commands and runs Rancher.
func runMain() int {

	var err error

	if err = ranchr.SetEnv(); err != nil {
		fmt.Println("An error while processing Rancher ranchr.Environment variables: %s\n", err.Error())
		return -1
	}

	SetLogging()

	log.Info("Rancher starting with args: %v", os.Args[:])
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
		log.Error("Rancher encountered an error: %s\n", err.Error())
		fmt.Println("Rancher encountered an error: %s\n", err.Error())
	}

	log.Info("Rancher exiting with an exit code of %v", exitCode)

	return exitCode
}

func SetLogging() error {
	logger, err := log.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		return err
	}

	defer log.Flush()

	ranchr.UseLogger(logger)
	defer ranchr.FlushLog()

	log.ReplaceLogger(logger)
	return nil
}

/*
func getFormattedLogFilename() string {
	var suffix, tmpName, logFile string

	// Set beego.Log stuff
	// Logfile name is <filename>-<year>-<month>-<day>.Log
	logFile = os.Getenv(ranchr.EnvLogFile)

	// Find the extension and insert the date info before it.
	n := strings.LastIndex(logFile, ".")
	if n < 0 {
		// If there is no extension found, the suffix is .Log by default.
		suffix = ".Log"
		tmpName = logFile
	} else {
		// The last index is assumed to be the suffix of the filename.
		suffix = SubString(logFile, n, len(logFile))
		tmpName = SubString(logFile, 0, n)

	}

	// Create the formatted name and see if the file already exists. Act accordingly.
	logFile = fmt.Sprintf("%s-%04d-%02d-%02d%s", tmpName, time.Now().Year(), time.Now().Month(), time.Now().Day(), suffix)

	return logFile

}
*/
