package main

import (
	"os"
	_ "os/signal"

	"github.com/mitchellh/cli"
	"github.com/mohae/rancher/command"
)

// Commands is the mapping of all available Rancher commands.
var Commands map[string]cli.CommandFactory

func init() {
	ui := &cli.BasicUi{Writer: os.Stdout}

	Commands = map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &command.BuildCommand{
				Ui: ui,
			}, nil
		},

		"run": func() (cli.Command, error) {
			return &command.RunCommand{
				Ui: ui,
			}, nil
		},

		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:          GitCommit,
				Version:           Version,
				VersionPrerelease: VersionPrerelease,
				Ui:                ui,
			}, nil
		},
	}
}
