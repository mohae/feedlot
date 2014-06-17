package command

import (
	"flag"
	"fmt"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/mohae/rancher/ranchr"
)

// BuildCommand is a Command implementation that generates Packer templates
// from named named builds and passed build arguments.
type BuildCommand struct {
	Ui cli.Ui
}

// Rancher help text.
func (c *BuildCommand) Help() string {
	helpText := `
Usage: rancher build [options]

Generates Packer templates. At minimum, this command needs to be run with
either the -distro flag or a build name. The simplest way to generate a Packer
template with rancher is to build a template with just the target distribution
name. The distribution must be supported, i.e. exists within Rancher's
distros.toml file:

	% rancher build -distro=<ditribution name>
	% rancher build -distro=ubuntu

The above command generates a Packer template, targeting Ubuntu, using the
defaults for that distribution, which are found in the distros.toml configur-
ation. file. Each of the distro defaults can be selectively overridden using
some of the other flags listed in the Options section.

Rancher can also generate Packer templates using preconfigured Rancher build
templates via the builds.toml file. The name of the build is used to specify
which build configuration should be used:

	% rancher build <build template name...>
	% rancher build 1204-amd64-server 1310-amd64-desktop


The above command generates two Packer templates using the 1204-amd64-server
and 1310-amd64-desktop build templates. The list of build template names is
variadic, accepting 1 or more build template names. For builds using the
-distro flag, the -arch, -image, and -release flags are optional. If any of
them are missing, the distribution's default value for that flag will be used.

Options:

-distro=<distroName>	If provided, Rancher will generate a template for the
			passed distribution name, e.g. ubuntu. This flag can
			be used along with the -arch, -image, and -release
			flags to override the Distribution's default values
			for those settings.

-arch=<architecture>	Specify whether 32 or 64 bit code should be used,
			e.g."x32" or "amd64" for ubuntu. This flag is only
			valid when used with the -distro flag.

-image=<imageType>	The type of ISO image that this Packer template will
			target, e.g. server, desktop, minimal for ubuntu. If
			the -distro flag is used and this flag is not used,
			the distro's default imageType will be used. This flag
			is only valid when used with the -distro flag.

-release=<releaseNum>	The release number that this Packer template will
			target, e.g. 12.04, etc. Only the targeted distri-
			bution's supported releases are valid. This flag is
			only valid when used with the -distro flag.

-log_dir=<logDirPath>	The directory path in which logging files will be
			written. This will override the existing logging
			directory information.
`
	return strings.TrimSpace(helpText)
}

func (c *BuildCommand) Run(args []string) int {
	var distroFilter, archFilter, imageFilter, releaseFilter, logDirFilter string
	cmdFlags := flag.NewFlagSet("build", flag.ContinueOnError)
	cmdFlags.Usage = func() { c.Ui.Output(c.Help()) }
	cmdFlags.StringVar(&distroFilter, "distro", "", "distro filter")
	cmdFlags.StringVar(&archFilter, "arch", "", "arch filter")
	cmdFlags.StringVar(&imageFilter, "image", "", "image filter")
	cmdFlags.StringVar(&releaseFilter, "release", "", "release filter")
	cmdFlags.StringVar(&logDirFilter, "log_dir", "", "log directory")
	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Parse of command-line arguments failed: %s", err))
		return 1
	}
	buildArgs := cmdFlags.Args()
	if distroFilter != "" {
		args := ranchr.ArgsFilter{Arch: archFilter, Distro: distroFilter, Image: imageFilter, Release: releaseFilter}
		// TODO go it
		if err := ranchr.BuildDistro(args); err != nil {
			c.Ui.Output(err.Error())
			return 1
		}
	}
	if len(buildArgs) > 0 {
		var message string
		var err error
		if message, err = ranchr.BuildBuilds(buildArgs...); err != nil {
			c.Ui.Error(err.Error())
			return 1
		}
		c.Ui.Output(message)
	}
	c.Ui.Output("Rancher Build complete.")
	return 0
}

func (c *BuildCommand) Synopsis() string {
	return "Create a Packer template from either distribution defaults or pre-defined Rancher Build templates."
}
