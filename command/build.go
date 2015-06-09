package command

import (
	"strings"

	"github.com/mohae/cli"
	"github.com/mohae/contour"
	"github.com/mohae/rancher/app"

	//jww "github.com/spf13/jwalterweatherman"
)

// BuildCommand is a Command implementation that generates Packer templates
// from named named builds and passed build arguments.
type BuildCommand struct {
	UI cli.Ui
}

// Help prints the help text for the build sub-command.
func (c *BuildCommand) Help() string {
	helpText := `
Usage: rancher build [options] <buildName...>

A Packer template will be created for each passed build name, if there are any.
Each build name, if there are more than one, must be separated by a space. Each
build must exist in the application's builds.toml file.
 
	$ rancher build <buildName...>
	$ rancher build 1204-amd64-server 1404-amd64-desktop

Options can also be passed to generate a build for a targeted supported distro.
This is done using the -distro flag:

	$ rancher build -distro=<distro name>
	$ rancher build -distro=ubuntu

For builds using the -distro flag, the -arch, -image, and -release flags are 
optional. If any of them are missing, the distribution's default value for that
flag will be used.

Options:
-distro=<distroName>	Create a Packer template from the distro's default
			settings. The -arch, -image, and -release flags can be
			used with this flag.

-arch=<architecture>	Override the distro's default architecture with this
			flag. The actual values are determined by the distro.

-image=<imageType>	Override the distro's default image with this flag. The
			actual values are determined by the distro.

-release=<releaseNum>	Override the distro's default release with this flag.
			The actual values are determined by the distro.
`
	return strings.TrimSpace(helpText)
}

// Run runs the build sub-command, handling all passed args and flags.
func (c *BuildCommand) Run(args []string) int {
	// Declare the command flag set and their values.
	contour.SetUsage(func() {
		c.UI.Output(c.Help())
	})

	var err error
	var filteredArgs []string
	filteredArgs, err = contour.FilterArgs(args)
	if err != nil {
		c.UI.Error(err.Error())
		return 1
	}
	/*
		jww.FEEDBACK.Printf("%#v\n", filteredArgs)
		b, _ := contour.GetBoolE("log")
		jww.FEEDBACK.Printf("log: %v\n", b)
		s, _ := contour.GetStringE("log_level_stdout")
		jww.FEEDBACK.Printf("log_level_stdout: %s\n", s)
		s, _ = contour.GetStringE("log_level_file")
		jww.FEEDBACK.Printf("log_level_file: %s\n", s)
		s, _ = contour.GetStringE("build_file")
		jww.FEEDBACK.Printf("build_file: %s\n", s)
	*/
	message, err := app.BuildBuilds(filteredArgs...)
	if err != nil {
		c.UI.Error(err.Error())
	}

	c.UI.Output(message)
	return 0
}

// Synopsis provides a precis of the build sub-command.
func (c *BuildCommand) Synopsis() string {
	return "Create a Packer template from either supported distro defaults or pre-defined Rancher build configurations."
}
