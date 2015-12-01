// Main entry point for Rancher.
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"os"

	"github.com/mohae/cli"
	"github.com/mohae/rancher/app"
	jww "github.com/spf13/jwalterweatherman"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	// Logging to temp first
	app.SetTempLogging()
	err := app.SetCfgFile()
	if err != nil {
		jww.ERROR.Printf("%s", err.Error())
	}
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
		jww.ERROR.Printf("Rancher encountered an error: %s", err.Error())

	}
	return exitCode
}
