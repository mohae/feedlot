package command

import (
	"strings"

	"github.com/mohae/cli"
	"github.com/mohae/contour"
	"github.com/mohae/feedlot/app"
	"github.com/mohae/feedlot/log"
)

// RunCommand is a Command implementation that generates Packer templates
// from passed build list names.
type RunCommand struct {
	UI cli.Ui
}

// Help prints the help text for the run sub-command.
func (c *RunCommand) Help() string {
	helpText := `
    Usage: feedlot run <BuildList names...>

        Generates Packer templates. At minimum, this command needs to be run
        with at least one BuildList name. Multiple BuildList names can be
	specified by using a space separated list.

            $ feedlot run example1
            $ feedlot run example1 example2

        The above command generates Packer templates from all of the feedlot
	Builds that have been specified within the RunList 'example1'.

	Options:
	-eg=bool           true/false: create builds from examples; generates
                       example Packer templates.
`

	return strings.TrimSpace(helpText)
}

// Run runs the run sub-command; the args are a variadic list of build list names.
func (c *RunCommand) Run(args []string) int {
	// Declare the command flag set and their values.
	contour.SetUsage(func() {
		c.UI.Output(c.Help())
	})
	// set flags/filter rgs
	var err error
	var filteredArgs []string
	filteredArgs, err = contour.FilterArgs(args)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	err = log.Set()
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if len(filteredArgs) == 0 {
		c.UI.Error("Nothing to do: no build list names were received.")
		return 1
	}
	// the remaining args are build names: build those templates.
	messages, errs := app.Run(filteredArgs...)
	// message and err are slices, []string and []error, respectively.
	// go through them and print out what's appropriate.
	// TODO: there better matching of returned results to build list. I am assuming the
	//       worst when I write this TODO, it may be that the returned stuff already does
	//       this.
	if messages != nil {
		for _, message := range messages {
			if message != "" {
				c.UI.Output(message)
			}
		}
	}
	if errs != nil {
		for _, err := range errs {
			if err != nil {
				c.UI.Error(err.Error())
			}
		}
	}
	return 0

}

// Synopsis provides a precis of the run sub-command.
func (c *RunCommand) Synopsis() string {
	return "Create Packer templates from the feedlot Build templates specified in the passed BuildList(s)."
}
