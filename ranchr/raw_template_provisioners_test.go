// raw_template_provisioners_test.go: tests for provisioners.
package ranchr

import (
	"testing"
)

var testRawTemplateProvisioner = &rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description:      "Test template config and Rancher options for CentOS",
	},
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:type/:build_name",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:type",
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
	vars:    map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = :commands_src_dir/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :commands_src_dir/shutdown_test.command",
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
		PostProcessors: map[string]*postProcessor{
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
			"shell-scripts",
			"file-uploads",
		},
		Provisioners: map[string]*provisioner{
			"shell-scripts": {
				templateSection{
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
			"file-uploads": {
				templateSection{
					Settings: []string{
						"source = src/",
						"destination = dst/",
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
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:type/:build_name",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:type",
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
	vars:    map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = :commands_src_dir/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :commands_src_dir/shutdown_test.command",
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
		PostProcessors: map[string]*postProcessor{
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
			"salt-masterless",
			"shell-scripts",
			"file-uploads",
		},
		Provisioners: map[string]*provisioner{
			"ansible-local": {
				templateSection{
					Settings: []string{
						"playbook_file= :src_dir/ansible/playbook.yml",
						"command =  :commands_src_dir/ansible_test.command",
						"inventory_file = :src_dir/ansible/inventory_file",
						"group_vars = groupvars",
						"host_vars = hostvars",
						"playbook_dir = :src_dir/ansible/playbooks",
						"staging_directory = staging/directory",
					},
					Arrays: map[string]interface{}{
						"extra_arguments": []string{
							"arg1",
							"arg2",
						},
						"playbook_paths": []string{
							"../ansible/playbook/",
						},
						"role_paths": []string{
							"../ansible/roles1",
							"../ansible/roles2",
						},
					},
				},
			},
			"salt-masterless": {
				templateSection{
					Settings: []string{
						"bootstrap_args = args",
						"local_pillar_roots=/srv/pillar/",
						"local_state_tree=/srv/salt/",
						"minion_config=minion",
						"skip_bootstrap=false",
						"temp_config_dir=/tmp",
					},
				},
			},
			"shell-scripts": {
				templateSection{
					Settings: []string{
						"binary = false",
						"execute_command = :commands_src_dir/execute_test.command",
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
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
			"file-uploads": {
				templateSection{
					Settings: []string{
						"source = /src/",
						"destination = /dst/",
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
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/virtualbox.sh",
						"scripts/cleanup.sh",
					},
				},
				"scripts": []string{
					"scripts/base.sh",
					"scripts/vagrant.sh",
					"scripts/cleanup.sh",
				},
			},
		},
	},
}

var prOrig = map[string]*provisioner{
	"shell-scripts": &provisioner{
		templateSection{
			Settings: []string{
				"execute_command = :commands_src_dir/execute_test.command",
			},
			Arrays: map[string]interface{}{
				"except": []string{
					"docker",
				},
				"only": []string{
					"virtualbox-iso",
				},
				"scripts": []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/sudoers_test.sh",
					":scripts_dir/cleanup_test.sh",
				},
			},
		},
	},
	"file-uploads": {
		templateSection{
			Settings: []string{
				"source = src/",
				"destination = dst/",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

var prNew = map[string]*provisioner{
	"shell-scripts": &provisioner{
		templateSection{
			Settings: []string{},
			Arrays: map[string]interface{}{
				"only": []string{
					"vmware-iso",
				},
				"override": map[string]interface{}{
					"vmware-iso": map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/vmware_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
				"scripts": []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/sudoers_test.sh",
					":scripts_dir/cleanup_test.sh",
				},
			},
		},
	},
}

var prMerged = map[string]*provisioner{
	"shell-scripts": &provisioner{
		templateSection{
			Settings: []string{
				"execute_command = :commands_src_dir/execute_test.command",
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
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/vmware_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
				"scripts": []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/sudoers_test.sh",
					":scripts_dir/cleanup_test.sh",
				},
			},
		},
	},
	"file-uploads": {
		templateSection{
			Settings: []string{
				"source = src/",
				"destination = dst/",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

func TestRawTemplateUpdateProvisioners(t *testing.T) {
	testRawTemplateProvisioner.updateProvisioners(nil)
	if MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners) != MarshalJSONToString.Get(prOrig) {
		t.Errorf("Got %q, want %q", MarshalJSONToString.Get(prOrig), MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners))
	}

	testRawTemplateProvisioner.updateProvisioners(prNew)
	if MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners) != MarshalJSONToString.GetIndented(prMerged) {
		t.Errorf("Got %q, want %q", MarshalJSONToString.GetIndented(prMerged), MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners))
	}
}

func TestCreateProvisioners(t *testing.T) {
	_, _, err := testRawTemplateBuilderOnly.createProvisioners()
	if err == nil {
		t.Error("Expected error \"unable to create provisioners: none specified\", got nil")
	} else {
		if err.Error() != "unable to create provisioners: none specified" {
			t.Errorf("Expected \"unable to create provisioners: none specified\", got %q", err.Error())
		}
	}

	_, _, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"ansible-local\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"ansible-local\"" {
			t.Errorf("Expected error \"no configuration found for \"ansible-local\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.ProvisionerTypes[0] = "file-uploads"
	_, _, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"file-uploads\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"file-uploads\"" {
			t.Errorf("Expected error \"no configuration found for \"file-uploads\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.ProvisionerTypes[0] = "salt-masterless"
	_, _, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"salt-masterless\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"salt-masterless\"" {
			t.Errorf("Expected error \"no configuration found for \"salt-masterless\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.ProvisionerTypes[0] = "shell"
	_, _, err = testRawTemplateWOSection.createProvisioners()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"shell\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"shell\"" {
			t.Errorf("Expected error \"no configuration found for \"shell\"\", got %q", err.Error())
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
		"command": ":commands_src_dir/ansible_test.command",
		"extra_arguments": []string{
			"arg1",
			"arg2",
		},
		"group_vars":     "groupvars",
		"host_vars":      "hostvars",
		"inventory_file": ":src_dir/ansible/inventory_file",
		"playbook_dir":   ":src_dir/ansible/playbooks",
		"playbook_file":  ":src_dir/ansible/playbook.yml",
		"playbook_paths": []string{
			"../ansible/playbook/",
		},
		"role_paths": []string{
			"../ansible/roles1",
			"../ansible/roles2",
		},
		"staging_directory": "staging/directory",
		"type":              "ansible-local",
	}
	settings, _, err := testRawTemplateProvisionersAll.createAnsibleLocal()
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
		"local_pillar_roots": "/srv/pillar/",
		"local_state_tree":   "/srv/salt/",
		"minion_config":      "minion",
		"skip_bootstrap":     false,
		"temp_config_dir":    "/tmp",
		"type":               "salt-masterless",
	}
	settings, _, err := testRawTemplateProvisionersAll.createSaltMasterless()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

// TODO: elided until refactor is done
/*
func TestShellProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"binary": false,
		"except": []string{
			"docker",
		},
		"execute_command": ":commands_src_dir/execute_test.command",
		"inline_shebang":  "/bin/sh",
		"only": []string{
			"virtualbox-iso",
		},
		"remote_path": "/tmp/script.sh",
		"scripts": []string{
			":scripts_dir/setup_test.sh",
			":scripts_dir/vagrant_test.sh",
			":scripts_dir/sudoers_test.sh",
			":scripts_dir/cleanup_test.sh",
		},
		"start_retry_timeout": "5m",
		"type":                "shell-scripts",
	}
	settings, _, err := testRawTemplateProvisionersAll.createShellScripts()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}
*/

func TestFileUploadsProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"destination": "/dst/",
		"source":      "/src/",
		"type":        "file-uploads",
	}
	settings, _, err := testRawTemplateProvisionersAll.createFileUploads()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}
