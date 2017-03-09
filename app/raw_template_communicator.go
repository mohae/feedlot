package app

import (
	"strconv"
	"strings"
)

// This supports standard Packer communicators.  Builders with custom
// communicators are associated with their builders.

// Communicator constants
const (
	InvalidCommunicator Communicator = iota
	NoCommunicator
	SSHCommunicator
	WinRMCommunicator
)

// Communicator is a Packer supported communicator.
type Communicator int

var communicators = [...]string{
	"invalid communicator",
	"none",
	"ssh",
	"winrm",
}

func (c Communicator) String() string { return communicators[c] }

// ParseCommunicator returns the communicator for the received s. If the value
// does not match a known communicator, InvalidCommunicator is returned.
func ParseCommunicator(s string) Communicator {
	s = strings.ToLower(s)
	switch s {
	case "none":
		return NoCommunicator
	case "ssh":
		return SSHCommunicator
	case "winrm":
		return WinRMCommunicator
	default:
		return InvalidCommunicator
	}
}

// Comm is an interface for communicators.
type comm interface {
	isCommunicator() bool
	processSettings([]string, *RawTemplate, map[string]interface{}) error
}

// NewCommunicator returns the communicator for s.  If the communicator is
// 'none', a nil is returned.  If the specified communicator does not match
// a valid communicator, an invalidCommunicatorErr is returned.
func NewCommunicator(s string) (comm, error) {
	typ := ParseCommunicator(s)
	switch typ {
	case NoCommunicator:
		return nil, nil
	case SSHCommunicator:
		return SSH{}, nil
	case WinRMCommunicator:
		return WinRM{}, nil
	default:
		return nil, InvalidComponentErr{cTyp: "communicator", s: s}
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

// Needed to fulfill the comm interface.
func (s SSH) isCommunicator() bool {
	return true
}

// ProcessSettings extracts the key value pairs that are relevant to the
// SSH communicator.
func (s SSH) processSettings(vals []string, r *RawTemplate, settings map[string]interface{}) error {
	for _, val := range vals {
		k, v := parseVar(val)
		v = r.replaceVariables(v)
		switch k {
		case "ssh_host":
			settings[k] = v
		case "ssh_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				return SettingErr{Key: k, Value: v, err: err}
			}
			settings[k] = i
		case "ssh_username":
			settings[k] = v
		case "ssh_password":
			settings[k] = v
		case "ssh_private_key_file":
			settings[k] = v
		case "ssh_pty":
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_timeout":
			settings[k] = v
		case "ssh_handshake_attempts":
			i, err := strconv.Atoi(v)
			if err != nil {
				return SettingErr{Key: k, Value: v, err: err}
			}
			settings[k] = i
		case "ssh_disable_agent":
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_bastion_host":
			settings[k] = v
		case "ssh_bastion_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				return SettingErr{Key: k, Value: v, err: err}
			}
			settings[k] = i
		case "ssh_bastion_username":
			settings[k] = v
		case "ssh_bastion_password":
			settings[k] = v
		case "ssh_bastion_private_key_file":
			settings[k] = v
		}
	}
	return nil
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

// Needed to fulfill the comm interface.
func (w WinRM) isCommunicator() bool {
	return true
}

// ProcessSettings extracts the key value pairs that are relevant to the
// SSH communicator.
func (w WinRM) processSettings(vals []string, r *RawTemplate, settings map[string]interface{}) error {
	for _, val := range vals {
		k, v := parseVar(val)
		v = r.replaceVariables(v)
		switch k {
		case "winrm_host":
			settings[k] = v
		case "winrm_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				return SettingErr{Key: k, Value: v, err: err}
			}
			settings[k] = i
		case "winrm_username":
			settings[k] = v
		case "winrm_password":
			settings[k] = v
		case "winrm_timeout":
			settings[k] = v
		case "winrm_use_ssl":
			settings[k], _ = strconv.ParseBool(v)
		case "winrm_insecure":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	return nil
}

func (r *RawTemplate) processCommunicator(id string, vals []string, settings map[string]interface{}) (prefix string, err error) {
	var hasComm bool
	var k, v string
	for _, val := range vals {
		k, v = parseVar(val)
		if k == "communicator" {
			hasComm = true
			break
		}
	}
	// If the slice doesn't have a communicator, just return.  The orignal
	// settings must be returned.
	if !hasComm {
		return "", nil
	}
	c, err := NewCommunicator(v)
	if err != nil {
		return "", err
	}
	// add the communicator
	settings[k] = strings.ToLower(v)
	// nil means the comm type was "none".  Treat it as the same as !hasComm.
	if c == nil {
		return "", nil
	}
	switch c.(type) {
	case SSH:
		prefix = "ssh"
	case WinRM:
		prefix = "winrm"
	}
	// get a map of the settings related to this communicator
	err = c.processSettings(vals, r, settings)
	if err != nil {
		return "", Error{slug: communicators[int(ParseCommunicator(v))], err: err}
	}
	return prefix, nil
}
