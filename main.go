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
	"os"
	"time"

	log "github.com/cihub/seelog"
	"github.com/mitchellh/cli"
	"github.com/mohae/rancher/ranchr"
)

var Logger log.LoggerInterface

func init() {
	// Set by default
}

func main() {
	// main wraps runMain() and ensures that the log gets flushed prior to exit.
	// Exit with return code from runMain()
	defer ranchr.FlushLog()
	defer log.Flush()
	rc := runMain()
	// wait for things to catch up because some log output seems to be missing
	// dev only? TODO hmmm, this helped me discover I've wasted a lot of time
	// and explains the craziness >.>
	os.Exit(rc)
}

func runMain() int {
	// runMain parses the Flag for glog, sets up CLI stuff for the supported sub-
	// commands and runs Rancher.
	var err error
	if err = ranchr.SetEnv(); err != nil {
		fmt.Println("An error while processing Rancher ranchr.Environment variables: ", err.Error())
		return -1
	}
	if err = SetLogging(); err != nil {
		fmt.Println("An error occurred while configuring application logging: ", err.Error())
		return -1
	}

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
	time.Sleep(time.Millisecond * 10000)
	log.Info("Rancher exiting with an exit code of %v", exitCode)
	return exitCode
}

func SetLogging() error {
	logger, err := log.LoggerFromConfigAsFile("seelog.xml")
	if err != nil {
		return err
	}
	ranchr.UseLogger(logger)
	log.ReplaceLogger(logger)
	return nil
}
