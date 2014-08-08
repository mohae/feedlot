// version.go handles the version sub-command.
package command

import (
	"bytes"
	"fmt"

	"github.com/mitchellh/cli"
)

// VersionCommand is a Command implementation that prints the version.
type VersionCommand struct {
	Revision          string
	Version           string
	VersionPrerelease string
	UI                cli.Ui
}

// Help prints the Help text for the version sub-command
func (c *VersionCommand) Help() string {
	return "Prints Rancher's version information."
}

// Run runs the version sub-command.
func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Rancher v%s", c.Version)
	if c.VersionPrerelease != "" {
		fmt.Fprintf(&versionString, ".%s", c.VersionPrerelease)

		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}

	c.UI.Output(versionString.String())

	return 0
}

// Synopsis provides a precis of the version sub-command.
func (c *VersionCommand) Synopsis() string {
	return "Prints the Rancher version"
}
