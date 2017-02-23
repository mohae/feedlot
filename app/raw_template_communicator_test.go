package app

import (
	"fmt"
	"testing"
)

func TestCommunicatorFromString(t *testing.T) {
	tests := []struct {
		val      string
		expected Communicator
	}{
		{"", InvalidCommunicator},
		{"nada", InvalidCommunicator},
		{"none", NoCommunicator},
		{"None", NoCommunicator},
		{"ssh", SSHCommunicator},
		{"SSH", SSHCommunicator},
		{"winrm", WinRMCommunicator},
		{"WinRM", WinRMCommunicator},
	}
	for i, test := range tests {
		comm := CommunicatorFromString(test.val)
		if comm != test.expected {
			t.Errorf("%d:  %q, want %q", i, comm, test.expected)
		}
	}
}

func TestNewCommunicator(t *testing.T) {
	tests := []struct {
		commType string
		expected comm
		err      string
	}{
		{"", nil, "invalid communicator"},
		{"none", nil, ""},
		{"NONE", nil, ""},
		{"ssh", SSH{}, ""},
		{"SSH", SSH{}, ""},
		{"winrm", WinRM{}, ""},
		{"WinRM", WinRM{}, ""},
	}
	for i, test := range tests {
		res, err := NewCommunicator(test.commType)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("%d: got %q, expected %q", i, err, test.err)
			}
			continue
		}
		if res != test.expected {
			t.Errorf("%d: got %q, expected %q", i, res, test.expected)
		}
	}
}

var ssh = rawTemplate{
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/ubuntu/out/ubuntu",
		PackerOutputDir:   "boxes/:distro/:build_name",
		SourceDir:         "../test_files/src/ubuntu",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test build template",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
		BaseURL:   "http://releases.ubuntu.com/",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "desktop",
	Release: "12.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Settings: []string{
						"communicator=ssh",
						"ssh_host=host_string",
						"ssh_port=22",
						"ssh_username=vagrant",
						"ssh_password=vagrant",
						"ssh_private_key_file=path/to/key_file",
						"ssh_pty=true",
						"ssh_timeout=10m",
						"ssh_handshake_attempts=10",
						"ssh_disable_agent=true",
						"ssh_bastion_host=bastion_host",
						"ssh_bastion_port=22",
						"ssh_bastion_username=vagrant",
						"ssh_bastion_password=vagrant",
						"ssh_bastion_private_key_file=path/to/bastion_key_file",
					},
				},
			},
		},
	},
}

var sshExpected = map[string]interface{}{
	"ssh_host":                     "host_string",
	"ssh_port":                     22,
	"ssh_username":                 "vagrant",
	"ssh_password":                 "vagrant",
	"ssh_private_key_file":         "path/to/key_file",
	"ssh_pty":                      true,
	"ssh_timeout":                  "10m",
	"ssh_handshake_attempts":       10,
	"ssh_disable_agent":            true,
	"ssh_bastion_host":             "bastion_host",
	"ssh_bastion_port":             22,
	"ssh_bastion_username":         "vagrant",
	"ssh_bastion_password":         "vagrant",
	"ssh_bastion_private_key_file": "path/to/bastion_key_file",
}

func TestSSHCommunicator(t *testing.T) {
	cm, _ := NewCommunicator("ssh")
	settings := map[string]interface{}{}
	err := cm.processSettings(ssh.Builders["virtualbox-iso"].Settings, &ssh, settings)
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}
	if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(sshExpected) {
		t.Errorf("got %s, want %s", MarshalJSONToString.Get(settings), MarshalJSONToString.Get(sshExpected))
	}
}

var winRM = rawTemplate{
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/ubuntu/out/ubuntu",
		PackerOutputDir:   "packer_boxes/ubuntu",
		SourceDir:         "../test_files/src/ubuntu",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test build template",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
		BaseURL:   "http://releases.ubuntu.com/",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "desktop",
	Release: "12.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Settings: []string{
						"communicator=winrm",
						"winrm_host=host_string",
						"winrm_port=22",
						"winrm_username=vagrant",
						"winrm_password=vagrant",
						"winrm_timeout=10m",
						"winrm_use_ssl=true",
						"winrm_insecure=true",
					},
				},
			},
		},
	},
}

var winRMExpected = map[string]interface{}{
	"winrm_host":     "host_string",
	"winrm_port":     22,
	"winrm_username": "vagrant",
	"winrm_password": "vagrant",
	"winrm_timeout":  "10m",
	"winrm_use_ssl":  true,
	"winrm_insecure": true,
}

func TestWinRMCommunicator(t *testing.T) {
	cm, _ := NewCommunicator("winrm")
	settings := map[string]interface{}{}
	err := cm.processSettings(winRM.Builders["virtualbox-iso"].Settings, &winRM, settings)
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}
	if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(winRMExpected) {
		t.Errorf("got %s, want %s", MarshalJSONToString.Get(settings), MarshalJSONToString.Get(winRMExpected))
	}
}

func TestProcessCommunicator(t *testing.T) {
	tests := []struct {
		vals     []string
		settings map[string]interface{}
		prefix   string
		err      string
	}{
		{[]string{}, map[string]interface{}{}, "", ""},
		{[]string{"a=b"}, map[string]interface{}{}, "", ""},
		{[]string{"communicator=nada"}, map[string]interface{}{}, "", "test 2.communicator: nada: invalid communicator"},
		{[]string{"a=b", "communicator=none"}, map[string]interface{}{"communicator": "none"}, "", ""},
		{[]string{"a=b", "communicator=None"}, map[string]interface{}{"communicator": "none"}, "", ""},
		{[]string{"a=b", "communicator=NONE"}, map[string]interface{}{"communicator": "none"}, "", ""},
		{[]string{"communicator=ssh", "ssh_username=vagrant", "ssh_password=vagrant"}, map[string]interface{}{"communicator": "ssh", "ssh_username": "vagrant", "ssh_password": "vagrant"}, "ssh", ""},
		{[]string{"communicator=SSH", "ssh_username=vagrant", "ssh_password=vagrant"}, map[string]interface{}{"communicator": "ssh", "ssh_username": "vagrant", "ssh_password": "vagrant"}, "ssh", ""},
		{[]string{"communicator=winrm", "winrm_username=vagrant"}, map[string]interface{}{"communicator": "winrm", "winrm_username": "vagrant"}, "winrm", ""},
		{[]string{"communicator=WinRM", "winrm_username=vagrant"}, map[string]interface{}{"communicator": "winrm", "winrm_username": "vagrant"}, "winrm", ""},
	}
	for i, test := range tests {
		r := rawTemplate{}
		settings := map[string]interface{}{}
		prefix, err := r.processCommunicator(fmt.Sprintf("test %d", i), test.vals, settings)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("%d: got %q, want %q", i, err, test.err)
			}
			continue
		}
		if test.err != "" {
			t.Errorf("%d: got no error, want %q", i, test.err)
			continue
		}
		if prefix != test.prefix {
			t.Errorf("%d: got %q, want %q", i, prefix, test.prefix)
			continue
		}
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(test.settings) {
			t.Errorf("%d: got %q, want %q", i, settings, test.settings)
		}
	}
}
