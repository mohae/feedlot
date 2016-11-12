package app

import "testing"

var testRawTemplateProvisioner = &rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description:      "Test template config and Feedlot options for CentOS",
	},
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/out/:build_name",
		PackerOutputDir:   "packer_boxes/:build_name",
		SourceDir:         "../test_files/src",
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
		BuilderIDs: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
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
					Type: "virtualbox-iso",
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
					Type:     "vmware-iso",
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
		PostProcessorIDs: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Type: "vagrant",
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = out/feedlot-packer.box",
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
					Type: "vagrant-cloud",
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar",
						"no_release = true",
						"version = 1.0.1",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell-test",
			"file",
		},
		Provisioners: map[string]provisioner{
			"shell-test": {
				templateSection{
					Type: "shell",
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
					Type: "file",
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
		Description:      "Test template config and Feedlot options for CentOS",
	},
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/out/:build_name",
		PackerOutputDir:   "packer_boxes/:build_name",
		SourceDir:         "../test_files/src",
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
		BuilderIDs: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
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
					Type: "virtualbox-iso",
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
					Type:     "vmware-iso",
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
		PostProcessorIDs: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Type: "vagrant",
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = out/feedlot-packer.box",
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
					Type: "vagrant-cloud",
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar",
						"no_release = true",
						"version = 1.0.1",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"ansible",
			"ansible-local",
			"file",
			"chef-client",
			"chef-solo",
			"puppet-client",
			"salt-masterless",
			"shell",
		},
		Provisioners: map[string]provisioner{
			"ansible": {
				templateSection{
					Type: "ansible",
					Settings: []string{
						"playbook_file= playbook.yml",
						"sftp_command =  /usr/lib/sftp-server -e",
						"host_alias = default",
						"local_port = 22",
						"ssh_authorized_key_file = .ssh/authorized_keys",
						"ssh_host_key_file = .ssh/host_key",
						"user = user",
					},
					Arrays: map[string]interface{}{
						"extra_arguments": []string{
							"arg1",
							"arg2",
						},
						"ansible_env_vars": []string{
							"ANSIBLE_HOST_KEY_CHECKING=False",
							"ANSIBLE_NOCOLOR=True",
						},
						"empty_groups": []string{
							"egroup1",
						},
						"groups": []string{
							"agroup1",
							"agroup2",
						},
					},
				},
			},
			"ansible-local": {
				templateSection{
					Type: "ansible-local",
					Settings: []string{
						"playbook_file= playbook.yml",
						"command = ansible_test.command",
						"inventory_file = inventory_file",
						"inventory_groups = my_group_1,my_group_2",
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
					Type: "chef-client",
					Settings: []string{
						"chef_environment=web",
						"client_key = client.pem",
						"config_template=chef.cfg",
						"encrypted_data_bag_secret_path = path/to/secret",
						"execute_command=execute.command",
						"guest_os_type = unix",
						"install_command=install.command",
						"node_name=test-chef",
						"prevent_sudo=false",
						"server_url=https://mychefserver.com",
						"skip_clean_client=true",
						"skip_clean_node=false",
						"skip_install=false",
						"ssl_verify_mode = verify_peer",
						"staging_directory=/tmp/chef/",
						"validation_client_name=some_value",
						"validation_key_path=/home/user/chef/chef-key",
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
					Type: "chef-solo",
					Settings: []string{
						"chef_environment = env",
						"config_template=chef.cfg",
						"data_bags_path=data_bag",
						"encrypted_data_bag_secret_path=/home/user/chef/secret_data_bag",
						"environments_path=environments",
						"execute_command=execute.command",
						"guest_os_type = unix",
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
			"file": {
				templateSection{
					Type: "file",
					Settings: []string{
						"source = app.tar.gz",
						"destination = /tmp/app.tar.gz",
					},
				},
			},
			"filedir": {
				templateSection{
					Type: "file",
					Settings: []string{
						"source = source/",
						"destination = /tmp/",
					},
				},
			},
			"puppet-masterless": {
				templateSection{
					Type: "puppet-masterless",
					Settings: []string{
						"manifest_filet=site.pp",
						"execute_command=execute.command",
						"hiera_config_path=hiera.yaml",
						"ignore_exit_codes = true",
						"manifest_dir=manifests",
						"manifest_file=site.pp",
						"prevent_sudo=false",
						"staging_directory=/tmp/puppet-masterless",
						"working_directory=work",
					},
					Arrays: map[string]interface{}{
						"extra_arguments": []string{
							"arg1",
						},
						"facter": map[string]string{
							"server_role": "webserver",
						},
						"module_paths": []string{
							"/etc/puppetlabs/puppet/modules",
							"/opt/puppet/share/puppet/modules",
						},
					},
				},
			},
			"puppet-server": {
				templateSection{
					Type: "puppet-server",
					Settings: []string{
						"client_cert_path = /etc/puppet/client.pem",
						"client_private_key_path=/home/puppet/.ssh/puppet_id_rsa",
						"ignore_exit_codes = true",
						"options=-v --detailed-exitcodes",
						"prevent_sudo= false",
						"puppet_node=vagrant-puppet-srv01",
						"puppet_server=server",
						"staging_directory=/tmp/puppet-server",
					},
					Arrays: map[string]interface{}{
						"facter": map[string]string{
							"server_role": "webserver",
						},
					},
				},
			},
			"salt-masterless": {
				templateSection{
					Type: "salt-masterless",
					Settings: []string{
						"bootstrap_args = args",
						"disable_sudo = true",
						"local_pillar_roots=pillar",
						"local_state_tree=salt",
						"no_exit_on_failure = true",
						"minion_config=salt",
						"skip_bootstrap=false",
						"temp_config_dir=/tmp",
					},
				},
			},
			"salt-masterless-remote-settings": {
				templateSection{
					Type: "salt-masterless",
					Settings: []string{
						"bootstrap_args = args",
						"disable_sudo = true",
						"local_pillar_roots=pillar",
						"local_state_tree=salt",
						"no_exit_on_failure = true",
						"remote_pillar_roots = /srv/pillar",
						"remote_state_tree = /srv/salt",
						"skip_bootstrap=false",
						"temp_config_dir=/tmp",
					},
				},
			},
			"salt-masterless-remote-pillar-minion-err": {
				templateSection{
					Type: "salt-masterless",
					Settings: []string{
						"bootstrap_args = args",
						"disable_sudo = true",
						"local_pillar_roots=pillar",
						"local_state_tree=salt",
						"no_exit_on_failure = true",
						"minion_config=salt",
						"remote_pillar_roots = /srv/pillar",
						"skip_bootstrap=false",
						"temp_config_dir=/tmp",
					},
				},
			},
			"salt-masterless-remote-state-minion-err": {
				templateSection{
					Type: "salt-masterless",
					Settings: []string{
						"bootstrap_args = args",
						"disable_sudo = true",
						"local_pillar_roots=pillar",
						"local_state_tree=salt",
						"no_exit_on_failure = true",
						"minion_config=salt",
						"remote_state_tree = /srv/salt",
						"skip_bootstrap=false",
						"temp_config_dir=/tmp",
					},
				},
			},
			"shell-local": {
				templateSection{
					Type: "shell-local",
					Settings: []string{
						"command = echo foo",
						"execute_command = execute_test.command",
					},
				},
			},
			"shell-required-missing": {
				templateSection{
					Type: "shell",
					Settings: []string{
						"binary = false",
						"execute_command = execute_test.command",
						"inline_shebang = /bin/sh",
						"remote_file = shell_test",
						"remote_folder = /tmp",
						"remote_path = /tmp/script.sh",
						"skip_clean = true",
						"start_retry_timeout = 5m",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
			"shell-inline": {
				templateSection{
					Type: "shell",
					Settings: []string{
						"binary = false",
						"execute_command = execute_test.command",
						"inline_shebang = /bin/sh",
						"remote_file = shell_test",
						"remote_folder = /tmp",
						"remote_path = /tmp/script.sh",
						"skip_clean = true",
						"start_retry_timeout = 5m",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
						"inline": []string{
							"apt-get update",
							"apt-get upgrade",
						},
					},
				},
			},
			"shell-script": {
				templateSection{
					Type: "shell",
					Settings: []string{
						"binary = false",
						"execute_command = execute_test.command",
						"inline_shebang = /bin/sh",
						"remote_file = shell_test",
						"remote_folder = /tmp",
						"remote_path = /tmp/script.sh",
						"script = vagrant_test.sh",
						"skip_clean = true",
						"start_retry_timeout = 5m",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
			"shell-scripts": {
				templateSection{
					Type: "shell",
					Settings: []string{
						"binary = false",
						"execute_command = execute_test.command",
						"inline_shebang = /bin/sh",
						"remote_file = shell_test",
						"remote_folder = /tmp",
						"remote_path = /tmp/script.sh",
						"skip_clean = true",
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
	"shell-test": provisioner{
		templateSection{
			Type: "shell",
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
			Type: "file",
			Settings: []string{
				"source = app.tar.gz",
				"destination = /tmp/app.tar.gz",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

var prNew = map[string]provisioner{
	"shell-test": provisioner{
		templateSection{
			Type:     "shell",
			Settings: []string{},
			Arrays: map[string]interface{}{
				"only": []string{
					"vmware-iso",
				},
				"except": []string{
					"digitalocean",
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
	"shell-test": provisioner{
		templateSection{
			Type: "shell",
			Settings: []string{
				"execute_command = execute_test.command",
			},
			Arrays: map[string]interface{}{
				"except": []string{
					"digitalocean",
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
			Type: "file",
			Settings: []string{
				"source = app.tar.gz",
				"destination = /tmp/app.tar.gz",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

func init() {
	b := true
	testRawTemplateProvisionersAll.IncludeComponentString = &b
}

func TestRawTemplateUpdateProvisioners(t *testing.T) {
	err := testRawTemplateProvisioner.updateProvisioners(nil)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
	if MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners) != MarshalJSONToString.Get(prOrig) {
		t.Errorf("Got %q, want %q", MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners), MarshalJSONToString.Get(prOrig))
	}

	err = testRawTemplateProvisioner.updateProvisioners(prNew)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
	if MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners) != MarshalJSONToString.GetIndented(prMerged) {
		t.Errorf("Got %q, want %q", MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners), MarshalJSONToString.Get(prMerged))
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
		"ansible_env_vars": []string{
			"ANSIBLE_HOST_KEY_CHECKING=False",
			"ANSIBLE_NOCOLOR=True",
		},
		"empty_groups": []string{
			"egroup1",
		},
		"extra_arguments": []string{
			"arg1",
			"arg2",
		},
		"groups": []string{
			"agroup1",
			"agroup2",
		},
		"host_alias":              "default",
		"sftp_command":            "/usr/lib/sftp-server -e",
		"local_port":              "22",
		"playbook_file":           "ansible/playbook.yml",
		"ssh_authorized_key_file": ".ssh/authorized_keys",
		"ssh_host_key_file":       ".ssh/host_key",
		"type":                    "ansible",
		"user":                    "user",
	}
	settings, err := testRawTemplateProvisionersAll.createAnsible("ansible")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestAnsibleLocalProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"command": "ansible_test.command",
		"extra_arguments": []string{
			"arg1",
			"arg2",
		},
		"group_vars":       "ansible-local/groupvars",
		"host_vars":        "ansible-local/hostvars",
		"inventory_file":   "ansible-local/inventory_file",
		"inventory_groups": "my_group_1,my_group_2",
		"playbook_dir":     "ansible-local/playbooks",
		"playbook_file":    "ansible-local/playbook.yml",
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
	settings, err := testRawTemplateProvisionersAll.createAnsibleLocal("ansible-local")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestChefClientProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"chef_environment":               "web",
		"client_key":                     "client.pem",
		"config_template":                "chef-client/chef.cfg",
		"encrypted_data_bag_secret_path": "path/to/secret",
		"execute_command":                "{{if .Sudo}}sudo {{end}}chef-client --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		"guest_os_type":                  "unix",
		"install_command":                "curl -L https://www.opscode.com/chef/install.sh | {{if .Sudo}}sudo{{end}} bash",
		"node_name":                      "test-chef",
		"prevent_sudo":                   false,
		"run_list": []string{
			"recipe[hello::default]",
			"recipe[world::default]",
		},
		"server_url":        "https://mychefserver.com",
		"skip_clean_client": true,
		"skip_clean_node":   false,
		"skip_install":      false,
		"ssl_verify_mode":   "verify_peer",
		"staging_directory": "/tmp/chef/",
		"type":              "chef-client",
		"validation_client_name": "some_value",
		"validation_key_path":    "/home/user/chef/chef-key",
	}
	settings, err := testRawTemplateProvisionersAll.createChefClient("chef-client")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestChefSoloProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"chef_environment": "env",
		"config_template":  "chef-solo/chef.cfg",
		"cookbook_paths": []string{
			"chef-solo/cookbook1",
			"chef-solo/cookbook2",
		},
		"data_bags_path":                 "chef-solo/data_bag",
		"encrypted_data_bag_secret_path": "/home/user/chef/secret_data_bag",
		"environments_path":              "chef-solo/environments",
		"execute_command":                "{{if .Sudo}}sudo {{end}}chef-client --no-color -c {{.ConfigPath}} -j {{.JsonPath}}",
		"guest_os_type":                  "unix",
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
	settings, err := testRawTemplateProvisionersAll.createChefSolo("chef-solo")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestFileProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"destination": "/tmp/app.tar.gz",
		"source":      "file/app.tar.gz",
		"type":        "file",
	}
	settings, err := testRawTemplateProvisionersAll.createFile("file")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
	expected = map[string]interface{}{
		"destination": "/tmp/",
		"source":      "file/source/",
		"type":        "file",
	}
	settings, err = testRawTemplateProvisionersAll.createFile("filedir")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestPuppetMasterlessProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"extra_arguments": []string{
			"arg1",
		},
		"facter": map[string]string{
			"server_role": "webserver",
		},
		"hiera_config_path": "puppet-masterless/hiera.yaml",
		"ignore_exit_codes": true,
		"manifest_dir":      "puppet-masterless/manifests",
		"manifest_file":     "puppet-masterless/site.pp",
		"module_paths": []string{
			"/etc/puppetlabs/puppet/modules",
			"/opt/puppet/share/puppet/modules",
		},
		"prevent_sudo":      false,
		"staging_directory": "/tmp/puppet-masterless",
		"type":              "puppet-masterless",
		"working_directory": "work",
	}
	settings, err := testRawTemplateProvisionersAll.createPuppetMasterless("puppet-masterless")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestPuppetServerProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"client_cert_path":        "/etc/puppet/client.pem",
		"client_private_key_path": "/home/puppet/.ssh/puppet_id_rsa",
		"facter": map[string]string{
			"server_role": "webserver",
		},
		"ignore_exit_codes": true,
		"options":           "-v --detailed-exitcodes",
		"prevent_sudo":      false,
		"puppet_node":       "vagrant-puppet-srv01",
		"puppet_server":     "server",
		"staging_directory": "/tmp/puppet-server",
		"type":              "puppet-server",
	}
	settings, err := testRawTemplateProvisionersAll.createPuppetServer("puppet-server")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestSaltProvisioner(t *testing.T) {
	tests := []struct {
		key         string
		expected    map[string]interface{}
		expectedErr string
	}{
		{
			key: "salt-masterless",
			expected: map[string]interface{}{
				"bootstrap_args":     "args",
				"disable_sudo":       true,
				"local_pillar_roots": "salt-masterless/pillar",
				"local_state_tree":   "salt-masterless/salt",
				"minion_config":      "salt-masterless/salt",
				"no_exit_on_failure": true,
				"skip_bootstrap":     false,
				"temp_config_dir":    "/tmp",
				"type":               "salt-masterless",
			},
			expectedErr: "",
		},
		{
			key: "salt-masterless-remote-settings",
			expected: map[string]interface{}{
				"bootstrap_args":      "args",
				"disable_sudo":        true,
				"local_pillar_roots":  "salt-masterless/pillar",
				"local_state_tree":    "salt-masterless/salt",
				"no_exit_on_failure":  true,
				"remote_pillar_roots": "/srv/pillar",
				"remote_state_tree":   "/srv/salt",
				"skip_bootstrap":      false,
				"temp_config_dir":     "/tmp",
				"type":                "salt-masterless",
			},
			expectedErr: "",
		},
		{
			key: "salt-masterless-remote-pillar-minion-err",
			expected: map[string]interface{}{
				"bootstrap_args":      "args",
				"disable_sudo":        true,
				"local_pillar_roots":  "salt-masterless/pillar",
				"local_state_tree":    "salt-masterless/salt",
				"minion_config":       "salt-masterless/salt",
				"no_exit_on_failure":  true,
				"remote_pillar_roots": "/srv/pillar",
				"skip_bootstrap":      false,
				"temp_config_dir":     "/tmp",
				"type":                "salt-masterless",
			},
			expectedErr: "salt-masterless-remote-pillar-minion-err: remote_pillar_roots: /srv/pillar: cannot be used with the 'minon_config' setting",
		},
		{
			key: "salt-masterless-remote-state-minion-err",
			expected: map[string]interface{}{
				"bootstrap_args":     "args",
				"disable_sudo":       true,
				"local_pillar_roots": "salt-masterless/pillar",
				"local_state_tree":   "salt-masterless/salt",
				"minion_config":      "salt-masterless/salt",
				"no_exit_on_failure": true,
				"remote_state_tree":  "/srv/salt",
				"skip_bootstrap":     false,
				"temp_config_dir":    "/tmp",
				"type":               "salt-masterless",
			},
			expectedErr: "salt-masterless-remote-state-minion-err: remote_state_tree: /srv/salt: cannot be used with the 'minon_config' setting",
		},
	}
	for i, test := range tests {
		settings, err := testRawTemplateProvisionersAll.createSalt(test.key)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: got %q, want %q", i, err, test.expectedErr)
			}
			continue
		}
		if err == nil && test.expectedErr != "" {
			t.Errorf("%d: wanted %q; got no error", i, test.expectedErr)
			continue
		}
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(test.expected) {
			t.Errorf("%d: expected %q, got %q", i, MarshalJSONToString.Get(test.expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestShellProvisionerRequiredMissing(t *testing.T) {
	expected := "shell-required-missing.inline, script, scripts: required setting"
	_, err := testRawTemplateProvisionersAll.createShell("shell-required-missing")
	if err == nil {
		t.Error("Expected error, got none")
	} else {
		if err.Error() != expected {
			t.Errorf("Expected %q, got %q", expected, err)
		}
	}
}

func TestShellProvisionerInline(t *testing.T) {
	expected := map[string]interface{}{
		"binary": false,
		"except": []string{
			"docker",
		},
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"inline": []string{
			"apt-get update",
			"apt-get upgrade",
		},
		"inline_shebang": "/bin/sh",
		"only": []string{
			"virtualbox-iso",
		},
		"remote_file":         "shell_test",
		"remote_folder":       "/tmp",
		"remote_path":         "/tmp/script.sh",
		"skip_clean":          true,
		"start_retry_timeout": "5m",
		"type":                "shell",
	}
	settings, err := testRawTemplateProvisionersAll.createShell("shell-inline")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestShellProvisionerScript(t *testing.T) {
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
		"remote_file":         "shell_test",
		"remote_folder":       "/tmp",
		"remote_path":         "/tmp/script.sh",
		"script":              "shell/vagrant_test.sh",
		"skip_clean":          true,
		"start_retry_timeout": "5m",
		"type":                "shell",
	}
	settings, err := testRawTemplateProvisionersAll.createShell("shell-script")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestShellProvisionerScripts(t *testing.T) {
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
		"remote_file":   "shell_test",
		"remote_folder": "/tmp",
		"remote_path":   "/tmp/script.sh",
		"scripts": []string{
			"shell/setup_test.sh",
			"shell/vagrant_test.sh",
			"shell/sudoers_test.sh",
			"shell/cleanup_test.sh",
		},
		"skip_clean":          true,
		"start_retry_timeout": "5m",
		"type":                "shell",
	}
	settings, err := testRawTemplateProvisionersAll.createShell("shell-scripts")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestShellLocalProvisioner(t *testing.T) {
	expected := map[string]interface{}{
		"command":         "echo foo",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"type":            "shell-local",
	}
	settings, err := testRawTemplateProvisionersAll.createShellLocal("shell-local")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}
