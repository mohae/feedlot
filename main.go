// Main entry point for Rancher.
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/mohae/cli"
	"github.com/mohae/rancher/app"
	jww "github.com/spf13/jwalterweatherman"
)

func main() {
	cpus := runtime.NumCPU()
	if cpus > 1 {
		runtime.GOMAXPROCS(runtime.NumCPU() - 1)
	}
	os.Exit(realMain())
}

func realMain() int {
	// set the environment variables from the application configuration file, if applicable.
	err := app.SetEnv()
	if err != nil {
		fmt.Printf("An error while processing the Rancher file and Environment variables: %s\n", err)
		return -1
	}
	// Logging setup
	app.SetLogging()
	args := os.Args[1:]
	cli := &cli.CLI{
		Name:     "rancher",
		Version:  Version,
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc("rancher"),
	}
	exitCode, err := cli.Run()
	if err != nil {
		jww.ERROR.Printf("Rancher encountered an error: %s\n", err)

	}
	return exitCode
}
