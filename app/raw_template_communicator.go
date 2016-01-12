package app

import (
	"strings"
)

// This supports standard Packer communicators.  Builders with custom
// communicators are associated with their builders.

// Communicator constants
const (
	None Communicator = iota
	SSH
	WinRM
)

// Communicator is a Packer supported communicator.
type Communicator int

var communicators = [...]string{
	"none",
	"ssh",
	"winrm"
}

func (c Communicator) String() string { return communicators[c] }

// CommunicatorFromString returns the communicator constant for the passed
// string or none.  Invalid values are treated as none.  All incoming
// strings are normalized to lowercase.
func CommunicatorFromString(s string) Communicator {
	s = strings.ToLower(s)
	switch s {
	case "ssh":
		return SSH
	case "winrm":
		return WinRM
	default:
		return None
	}
}

// SSH communicator.  In the templates, the actual field names are prefixed
// with ssh_, e.g. ssh_host.  The field comments are copied from
// https://www.packer.io/docs/templates/communicator.html
type SSH struct {
	// The address to SSH to. This usually is automatically configured by the builder.
	Host string
	// The port to connect to SSH. This defaults to 22.
	Port int
	// The username to connect to SSH with.
	Username string
	// A plaintext password to use to authenticate with SSH.
	Password string
	// Path to a PEM encoded private key file to use to authentiate with SSH.
	PrivateKeyFile string
	// If true, a PTY will be requested for the SSH connection. This defaults to false.
	PTY bool
	// The time to wait for SSH to become available. Packer uses this to determine when the
	// machine has booted so this is usually quite long. Example value: "10m"
	Timeout string
	// The number of handshakes to attempt with SSH once it can connect. This defaults to 10.
	HandshakeAttempts int
	//  If true, SSH agent forwarding will be disabled.
	DisableAgent bool
	// A bastion host to use for the actual SSH connection.
	BastionHost string
	// The port of the bastion host. Defaults to 22.
	BastionPort int
	// The username to connect to the bastion host.
	BastionUsername string
	// The password to use to authenticate with the bastion host.
	BastionPassword string
	// A private key file to use to authenticate with the bastion host.
	BastionPrivateKeyFile string
}

// WinRm communicator.  In the templates, the actual field names are prefixed
// with winrm_, e.g. winrm_host.  The field comments are copied from
// https://www.packer.io/docs/templates/communicator.html
type WinRM struct {
	// The address for WinRM to connect to.
	Host string
	// The WinRM port to connect to. This defaults to 5985.
	Port int
	// The username to use to connect to WinRM.
	Username string
	// The password to use to connect to WinRM.
	Password string
	// The amount of time to wait for WinRM to become available. This defaults to "30m"
	// since setting up a Windows machine generally takes a long time.
	Timeout string
	// If true, use HTTPS for WinRM
	UseSSL bool
	// If true, do not check server certificate chain and host name
	Insecure bool
}
