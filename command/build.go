package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/mohae/rancher/app"
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
	var distroFilter, archFilter, imageFilter, releaseFilter, logDirFilter string
	// Declare the command flag set and their values.
	cmdFlags := flag.NewFlagSet("build", flag.ContinueOnError)
	cmdFlags.Usage = func() {
		c.UI.Output(c.Help())
	}
	cmdFlags.StringVar(&distroFilter, "distro", "", "distro filter")
	cmdFlags.StringVar(&archFilter, "arch", "", "arch filter")
	cmdFlags.StringVar(&imageFilter, "image", "", "image filter")
	cmdFlags.StringVar(&releaseFilter, "release", "", "release filter")
	cmdFlags.StringVar(&logDirFilter, "log_dir", "", "log directory")
	// Parse the passed args for flags.
	if err := cmdFlags.Parse(args); err != nil {
		c.UI.Error(fmt.Sprintf("Parse of command-line arguments failed: %s", err))
		return 1
	}
	// Remaining flags are build names
	buildArgs := cmdFlags.Args()
	// If the distro option was passed, create the Packer template from distro defaults
	if distroFilter != "" {
		args := ranchr.ArgsFilter{Arch: archFilter, Distro: distroFilter, Image: imageFilter, Release: releaseFilter}
		err := ranchr.BuildDistro(args)
		if err != nil {
			c.UI.Output(err.Error())
			return 1
		}
	}

	// If there were any builds passed, build them.
	if len(buildArgs) > 0 {
		var message string
		var err error
		message, err = ranchr.BuildBuilds(buildArgs...)
		if err != nil {
			c.UI.Error(err.Error())
			return 1
		}
		c.UI.Output(message)
	}

	c.UI.Output("Rancher Build complete.")
	return 0
}

// Synopsis provides a precis of the build sub-command.
func (c *BuildCommand) Synopsis() string {
	return "Create a Packer template from either supported distro defaults or pre-defined Rancher build configurations."
}
