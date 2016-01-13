package app

import (
	"testing"
)

func TestNewCommunicator(t *testing.T) {
	tests := []struct{
		commType string
		expected comm
		err string
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
		OutputDir: "../test_files/ubuntu/out/ubuntu",
		SourceDir: "../test_files/src/ubuntu",
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
	"ssh_host": "host_string",
	"ssh_port": 22,
	"ssh_username": "vagrant",
	"ssh_password": "vagrant",
	"ssh_private_key_file": "path/to/key_file",
	"ssh_pty": true,
	"ssh_timeout": "10m",
	"ssh_handshake_attempts": 10,
	"ssh_disable_agent": true,
	"ssh_bastion_host": "bastion_host",
	"ssh_bastion_port": 22,
	"ssh_bastion_username": "vagrant",
	"ssh_bastion_password": "vagrant",
	"ssh_bastion_private_key_file": "path/to/bastion_key_file",
}

func TestSSHCommunicator(t *testing.T) {
	cm, _ := NewCommunicator("ssh")
	res, err := cm.processSettings(ssh.Builders["virtualbox-iso"].Settings, &ssh)
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}
	if MarshalJSONToString.Get(res) != MarshalJSONToString.Get(sshExpected) {
		t.Errorf("got %s, want %s", MarshalJSONToString.Get(res), MarshalJSONToString.Get(sshExpected))
	}
}

var winRM = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/ubuntu/out/ubuntu",
		SourceDir: "../test_files/src/ubuntu",
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
	"winrm_host": "host_string",
	"winrm_port": 22,
	"winrm_username": "vagrant",
	"winrm_password": "vagrant",
	"winrm_timeout": "10m",
	"winrm_use_ssl": true,
	"winrm_insecure": true,
}
func TestWinRMCommunicator(t *testing.T) {
	cm, _ := NewCommunicator("winrm")
	res, err := cm.processSettings(winRM.Builders["virtualbox-iso"].Settings, &winRM)
	if err != nil {
		t.Errorf("unexpected error: %q", err)
		return
	}
	if MarshalJSONToString.Get(res) != MarshalJSONToString.Get(winRMExpected) {
		t.Errorf("got %s, want %s", MarshalJSONToString.Get(res), MarshalJSONToString.Get(winRMExpected))
	}
}
