// Main entry point for Rancher. 
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
// Application logger variables
var (
	AppLogConfigFilename = "RANCHER_LOG_CONFIG_FILENAME"
	defaultLogConfigFilename = "seelog.xml"
	Logger log.LoggerInterface
	LogConfigFilename string
)
// Anything non-IO that should be set by default.
func init() {
}
func main() {
	// main wraps runMain() and ensures that the log gets flushed prior to exit.
	// Exit with return code from runMain()
	defer ranchr.FlushLog()
	defer log.Flush()
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
		log.Errorf("Rancher encountered an error: %s\n", err.Error())
		fmt.Printf("Rancher encountered an error: %s\n", err.Error())
	}
	log.Info("Rancher exiting with an exit code of %v", exitCode)
	// TODO Force sleep because log messages seem to be dropped...the needs to be addressed
	time.Sleep(time.Millisecond * 5000)
	return exitCode
}

// Set application logging. If the 
func SetLogging() error {
	var err error
	LogConfigFilename = os.Getenv(AppLogConfigFilename)
	if LogConfigFilename == "" {
		LogConfigFilename = defaultLogConfigFilename
	}
	Logger, err = log.LoggerFromConfigAsFile(LogConfigFilename)
	if err != nil {
		return err
	}
	ranchr.UseLogger(Logger)
	log.ReplaceLogger(Logger)
	return nil
}
