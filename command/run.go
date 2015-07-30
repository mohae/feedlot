package command

import (
	"flag"
	"strings"

	"github.com/mitchellh/cli"
)

// RunCommand is a Command implementation that generates Packer templates
// from passed build list names.
type RunCommand struct {
	UI cli.Ui
}

// Help prints the help text for the run sub-command.
func (c *RunCommand) Help() string {
	helpText := `
    Usage: rancher Run <BuildList names...>

        Generates Packer templates. At minimum, this command needs to be run
        with at least one BuildList name. Multiple BuildList names can be
	specified by using a space separated list.

            $ rancher run example1
            $ rancher run example1 example2

        The above command generates Packer templates from all of the Rancher
	Builds that have been specified within the RunList 'example1'.
`
	return strings.TrimSpace(helpText)
}

// Run runs the run sub-command; the args are a variadic list of build list names.
func (c *RunCommand) Run(args []string) int {
	var logLevel string
	cmdFlags := flag.NewFlagSet("run", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.UI.Output(c.Help()) }
	cmdFlags.StringVar(&logLevel, "log-level", "INFO", "log level")
	return 0

}

// Synopsis provides a precis of the run sub-command.
func (c *RunCommand) Synopsis() string {
	return "Create Packer templates from the Rancher Build templates specified in the passed BuildList(s)."
}
