// Initializes the Commands struct for the application.
// New commands need to be added to the CommandFactory map.
package main

import (
	"os"

	"github.com/mohae/cli"
	"github.com/mohae/rancher/command"
)

// Commands is the mapping of all available Rancher commands.
var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}
	Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &command.BuildCommand{
				UI: ui,
			}, nil
		},
		"run": func() (cli.Command, error) {
			return &command.RunCommand{
				UI: ui,
			}, nil
		},
	}
}
