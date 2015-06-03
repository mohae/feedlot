// raw_template_provisioners_test.go: tests for provisioners.
package app

import (
	"fmt"
	"testing"
)

var testRawTemplateProvisioner = &rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description:      "Test template config and Rancher options for CentOS",
	},
	IODirInf: IODirInf{
		OutDir: "../test_files/out/:build_name",
		SrcDir: "../test_files/src",
	},
	BuildInf: BuildInf{
		Name:      ":build_name",
		BuildName: "",
		BaseURL:   "",
	},
	date:    today,
	delim:   ":",
	Distro:  "centos",
	Arch:    "x86_64",
	Image:   "minimal",
	Release: "6",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Settings: []string{
						"virtualbox_version_file = .vbox_version",
					},
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=1024",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection{
					Settings: []string{},
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpuid.coresPerSocket=1",
							"memsize=1024",
							"numvcpus=1",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = out/rancher-packer.box",
					},
					Arrays: map[string]interface{}{
						"include": []string{
							"include1",
							"include2",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
			"vagrant-cloud": {
				templateSection{
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar",
						"no_release = true",
						"version = 1.0.1",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"shell",
			"file",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"sudoers_test.sh",
							"cleanup_test.sh",
						},
					},
				},
			},
			"file": {
				templateSection{
					Settings: []string{
						"source = app.tar.gz",
						"destination = /tmp/app.tar.gz",
					},
					Arrays: map[string]interface{}{},
				},
			},
		},
	},
}

var testRawTemplateProvisionersAll = &rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description:      "Test template config and Rancher options for CentOS",
	},
	IODirInf: IODirInf{
		IncludeComponentString: true,
		OutDir:                 "../test_files/out/:build_name",
		SrcDir:                 "../test_files/src",
	},
	BuildInf: BuildInf{
		Name:      ":build_name",
		BuildName: "",
		BaseURL:   "",
	},
	date:    today,
	delim:   ":",
	Distro:  "centos",
	Arch:    "x86_64",
	Image:   "minimal",
	Release: "6",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Settings: []string{
						"virtualbox_version_file = .vbox_version",
					},
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=1024",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection{
					Settings: []string{},
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpuid.coresPerSocket=1",
							"memsize=1024",
							"numvcpus=1",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = out/rancher-packer.box",
					},
					Arrays: map[string]interface{}{
						"include": []string{
							"include1",
							"include2",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
			"vagrant-cloud": {
				templateSection{
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar",
						"no_release = true",
						"version = 1.0.1",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"ansible-local",
			"chef-client",
			"chef-solo",
			"salt-masterless",
			"shell",
			"file",
		},
		Provisioners: map[string]provisioner{
			"ansible-local": {
				templateSection{
					Settings: []string{
						"playbook_file= playbook.yml",
						"command =  ansible_test.command",
						"inventory_file = inventory_file",
						"group_vars = groupvars",
						"host_vars = hostvars",
						"playbook_dir = playbooks",
						"staging_directory = staging/directory",
					},
					Arrays: map[string]interface{}{
						"extra_arguments": []string{
							"arg1",
							"arg2",
						},
						"playbook_paths": []string{
							"playbook1",
							"playbook2",
						},
						"role_paths": []string{
							"roles1",
							"roles2",
						},
					},
				},
			},
			"chef-client": {
				templateSection{
					Settings: []string{
						"chef_environment=web",
						"config_template=chef.cfg",
						"execute_command=execute.command",
						"install_command=install.command",
						"node_name=test-chef",
						"prevent_sudo=false",
						"server_url=https://mychefserver.com",
						"skip_clean_client=true",
						"skip_clean_node=false",
						"skip_install=false",
						"staging_directory=/tmp/chef/",
						"validation_client_name=some_value",
						"validation_key_path=chef-key",
					},
					Arrays: map[string]interface{}{
						"run_list": []string{
							"recipe[hello::default]",
							"recipe[world::default]",
						},
					},
				},
			},
			"chef-solo": {
				templateSection{
					Settings: []string{
						"config_template=chef.cfg",
						"data_bags_path=data_bag",
						"encrypted_data_bag_secret_path=secret_data_bag",
						"environments_path=environments",
						"execute_command=execute.command",
						"install_command=install.command",
						"prevent_sudo=false",
						"roles_path=roles",
						"skip_install=false",
						"staging_directory=/tmp/chef/",
					},
					Arrays: map[string]interface{}{
						"cookbook_paths": []string{
							"cookbook1",
							"cookbook2",
						},
						"remote_cookbook_paths": []string{
							"remote/path1",
							"remote/path2",
						},
						"run_list": []string{
							"recipe[hello::default]",
							"recipe[world::default]",
						},
					},
				},
			},
			"salt-masterless": {
				templateSection{
					Settings: []string{
						"bootstrap_args = args",
						"local_pillar_roots=pillar",
						"local_state_tree=salt",
						"minion_config=salt",
						"skip_bootstrap=false",
						"temp_config_dir=/tmp",
					},
				},
			},
			"shell": {
				templateSection{
					Settings: []string{
						"binary = false",
						"execute_command = execute_test.command",
						"inline_shebang = /bin/sh",
						"remote_path = /tmp/script.sh",
						"start_retry_timeout = 5m",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"sudoers_test.sh",
							"cleanup_test.sh",
						},
					},
				},
			},
			"file": {
				templateSection{
					Settings: []string{
						"source = app.tar.gz",
						"destination = /tmp/app.tar.gz",
					},
				},
			},
		},
	},
}

var pr = &provisioner{
	templateSection{
		Settings: []string{
			"execute_command= echo 'vagrant' | sudo -S sh '{{.Path}}'",
			"type = shell",
		},
		Arrays: map[string]interface{}{
			"override": map[string]interface{}{
				"virtualbox-iso": map[string]interface{}{
					"scripts": []string{
						"base.sh",
						"vagrant.sh",
						"virtualbox.sh",
						"cleanup.sh",
					},
				},
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"cleanup.sh",
				},
			},
		},
	},
}

var prOrig = map[string]provisioner{
	"shell": provisioner{
		templateSection{
			Settings: []string{
				"execute_command = execute_test.command",
			},
			Arrays: map[string]interface{}{
				"except": []string{
					"docker",
				},
				"only": []string{
					"virtualbox-iso",
				},
				"scripts": []string{
					"setup_test.sh",
					"vagrant_test.sh",
					"sudoers_test.sh",
					"cleanup_test.sh",
				},
			},
		},
	},
	"file": {
		templateSection{
			Settings: []string{
				"source = app.tar.gz",
				"destination = /tmp/app.tar.gz",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

var prNew = map[string]provisioner{
	"shell": provisioner{
		templateSection{
			Settings: []string{},
			Arrays: map[string]interface{}{
				"only": []string{
					"vmware-iso",
				},
				"override": map[string]interface{}{
					"vmware-iso": map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"vmware_test.sh",
							"cleanup_test.sh",
						},
					},
				},
				"scripts": []string{
					"setup_test.sh",
					"vagrant_test.sh",
					"sudoers_test.sh",
					"cleanup_test.sh",
				},
			},
		},
	},
}

var prMerged = map[string]provisioner{
	"shell": provisioner{
		templateSection{
			Settings: []string{
				"execute_command = execute_test.command",
			},
			Arrays: map[string]interface{}{
				"except": []string{
					"docker",
				},
				"only": []string{
					"vmware-iso",
				},
				"override": map[string]interface{}{
					"vmware-iso": map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"vmware_test.sh",
							"cleanup_test.sh",
						},
					},
				},
				"scripts": []string{
					"setup_test.sh",
					"vagrant_test.sh",
					"sudoers_test.sh",
					"cleanup_test.sh",
				},
			},
		},
	},
	"file": {
		templateSection{
			Settings: []string{
				"source = app.tar.gz",
				"destination = /tmp/app.tar.gz",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

func TestRawTemplateUpdateProvisioners(t *testing.T) {
	err := testRawTemplateProvisioner.updateProvisioners(nil)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err.Error())
	}
	if MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners) != MarshalJSONToString.Get(prOrig) {
		t.Errorf("Got %q, want %q", MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners), MarshalJSONToString.Get(prOrig))
	}

	err = testRawTemplateProvisioner.updateProvisioners(prNew)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err.Error())
	}
	if MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners) != MarshalJSONToString.GetIndented(prMerged) {
		t.Errorf("Got %q, want %q", MarshalJSONToString.GetIndented(prMerged), MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners))
	}
}

func TestCreateProvisioners(t *testing.T) {
	_, err := testRawTemplateBuilderOnly.createProvisioners()
	if err == nil {
		t.Error("Expected error \"unable to create provisioners: none specified\", got nil")
	} else {
		if err.Error() != "unable to create provisioners: none specified" {
			t.Errorf("Expected \"unable to create provisioners: none specified\", got %q", err.Error())
		}
	}

	_, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Errorf("Expected error \"no configuration found for %q\", got nil", Ansible.String())
	} else {
		if err.Error() != fmt.Sprintf("no configuration found for %q", Ansible.String()) {
			t.Errorf("Expected error \"no configuration found for %q\", got %q", Ansible.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.ProvisionerTypes[0] = FileUploads.String()
	_, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Errorf("Expected error \"no configuration found for %q\", got nil", FileUploads.String())
	} else {
		if err.Error() != fmt.Sprintf("no configuration found for %q", FileUploads.String()) {
			t.Errorf("Expected error \"no configuration found for %q\", got %q", FileUploads.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.ProvisionerTypes[0] = Salt.String()
	_, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Errorf("Expected error \"no configuration found for %q\", got nil", Salt.String())
	} else {
		if err.Error() != fmt.Sprintf("no configuration found for %q", Salt.String()) {
			t.Errorf("Expected error \"no configuration found for %q\", got %q", Salt.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.ProvisionerTypes[0] = ShellScript.String()
	_, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Errorf("Expected error \"no configuration found for %q\", got nil", ShellScript.String())
	} else {
		if err.Error() != "no configuration found for \"shell\"" {
			t.Errorf("Expected error \"no configuration found for %q\", got %q", ShellScript.String(), err.Error())
		}
	}

}

func TestProvisionersSettingsToMap(t *testing.T) {
	res := pr.settingsToMap("shell", testRawTpl)
	compare := map[string]interface{}{"type": "shell", "execute_command": "echo 'vagrant' | sudo -S sh '{{.Path}}'"}
	for k, v := range res {
		val, ok := compare[k]
		if !ok {
			t.Errorf("Expected to find entry for Key %s, none found", k)
			continue
		}
		if val != v {
			t.Errorf("Got %q, want %q", v, val)
		}
	}
}

func TestAnsibleProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"command": "ansible_test.command",
		"extra_arguments": []string{
			"arg1",
			"arg2",
		},
		"group_vars":     "ansible-local/groupvars",
		"host_vars":      "ansible-local/hostvars",
		"inventory_file": "ansible-local/inventory_file",
		"playbook_dir":   "ansible-local/playbooks",
		"playbook_file":  "ansible-local/playbook.yml",
		"playbook_paths": []string{
			"ansible-local/playbook1",
			"ansible-local/playbook2",
		},
		"role_paths": []string{
			"ansible-local/roles1",
			"ansible-local/roles2",
		},
		"staging_directory": "staging/directory",
		"type":              "ansible-local",
	}
	settings, err := testRawTemplateProvisionersAll.createAnsible()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestChefClientProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"chef_environment": "web",
		"config_template":  "chef-client/chef.cfg",
		"execute_command":  "{{if .Sudo}}sudo {{end}}chef-client --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		"install_command":  "curl -L https://www.opscode.com/chef/install.sh | {{if .Sudo}}sudo{{end}} bash",
		"node_name":        "test-chef",
		"prevent_sudo":     false,
		"run_list": []string{
			"recipe[hello::default]",
			"recipe[world::default]",
		},
		"server_url":        "https://mychefserver.com",
		"skip_clean_client": true,
		"skip_clean_node":   false,
		"skip_install":      false,
		"staging_directory": "/tmp/chef/",
		"type":              "chef-client",
		"validation_client_name": "some_value",
		"validation_key_path":    "chef-client/chef-key",
	}
	settings, err := testRawTemplateProvisionersAll.createChefClient()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestChefSoloProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"config_template": "chef-solo/chef.cfg",
		"cookbook_paths": []string{
			"chef-solo/cookbook1",
			"chef-solo/cookbook2",
		},
		"data_bags_path":                 "chef-solo/data_bag",
		"encrypted_data_bag_secret_path": "chef-solo/secret_data_bag",
		"environments_path":              "chef-solo/environments",
		"execute_command":                "{{if .Sudo}}sudo {{end}}chef-client --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		"install_command":                "curl -L https://www.opscode.com/chef/install.sh | {{if .Sudo}}sudo{{end}} bash",
		"prevent_sudo":                   false,
		"roles_path":                     "chef-solo/roles",
		"remote_cookbook_paths": []string{
			"remote/path1",
			"remote/path2",
		},
		"run_list": []string{
			"recipe[hello::default]",
			"recipe[world::default]",
		},
		"skip_install":      false,
		"staging_directory": "/tmp/chef/",
		"type":              "chef-solo",
	}
	settings, err := testRawTemplateProvisionersAll.createChefSolo()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestSaltProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"bootstrap_args":     "args",
		"local_pillar_roots": "salt-masterless/pillar",
		"local_state_tree":   "salt-masterless/salt",
		"minion_config":      "salt-masterless/salt",
		"skip_bootstrap":     false,
		"temp_config_dir":    "/tmp",
		"type":               "salt-masterless",
	}
	settings, err := testRawTemplateProvisionersAll.createSalt()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestShellProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"binary": false,
		"except": []string{
			"docker",
		},
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"inline_shebang":  "/bin/sh",
		"only": []string{
			"virtualbox-iso",
		},
		"remote_path": "/tmp/script.sh",
		"scripts": []string{
			"shell/setup_test.sh",
			"shell/vagrant_test.sh",
			"shell/sudoers_test.sh",
			"shell/cleanup_test.sh",
		},
		"start_retry_timeout": "5m",
		"type":                "shell",
	}
	settings, err := testRawTemplateProvisionersAll.createShellScript()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestFileUploadsProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"destination": "/tmp/app.tar.gz",
		"source":      "file/app.tar.gz",
		"type":        "file",
	}
	settings, err := testRawTemplateProvisionersAll.createFileUploads()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}
