package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
)

// BuildCommand is a Command implementation that generates Packer templates
// from named Builds.
type RunCommand struct {
	Ui cli.Ui
}

func (c *RunCommand) Help() string {
	helpText := `
    Usage: rancher Run <BuildList names...>

        Generates Packer templates. At minimum, this command needs to be run
        with at least one BuildList name. Multiple BuildList names can be
	specified by using a space separated list.

            rancher run example1

        The above command generates Packer templates from all of the Rancher
	Builds that have been specified within the RunList 'example1'.
`
	return strings.TrimSpace(helpText)
}

func (c *RunCommand) Run(args []string) int {
	var logLevel string

	cmdFlags := flag.NewFlagSet("run", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&logLevel, "log-level", "INFO", "log level")

	fmt.Printf("%+v\n", args)

	return 0

}

func (c *RunCommand) Synopsis() string {
	return "Create Packer templates from the Rancher Build templates specified in the passed BuildList(s)."
}
