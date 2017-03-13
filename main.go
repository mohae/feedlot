// Main entry point for Rancher.
//
// Notes on code in Main: some of the code in runMain is copied from the copy-
// right holder, Mitchell Hashimoto (github.com/mitchellh), as I am using his
// cli package.
package main

import (
	"fmt"
	"os"

	"github.com/mohae/cli"
	"github.com/mohae/feedlot/conf"
)

func main() {
	os.Exit(realMain())
}

func realMain() int {
	err := conf.SetAppConfFile()
	if err != nil {
		fmt.Println(err)
		return 1
	}
	args := os.Args[1:]
	cli := &cli.CLI{
		Name:     conf.Name,
		Version:  Version,
		Args:     args,
		Commands: Commands,
		HelpFunc: cli.BasicHelpFunc(conf.Name),
	}
	exitCode, err := cli.Run()
	if err != nil {
		fmt.Println(err)
	}
	return exitCode
}
