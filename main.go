// Main entry point for Rancher.
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"os"

	"github.com/mohae/cli"
	"github.com/mohae/feedlot/app"
	jww "github.com/spf13/jwalterweatherman"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	err := app.SetCfgFile()
	if err != nil {
		jww.ERROR.Printf("%s", err.Error())
	}
	args := os.Args[1:]
	cli := &cli.CLI{
		Name:     app.Name,
		Version:  Version,
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc(app.Name),
	}
	exitCode, err := cli.Run()
	if err != nil {
		jww.ERROR.Printf("%s encountered an error: %s", app.Name, err.Error())

	}
	return exitCode
}
