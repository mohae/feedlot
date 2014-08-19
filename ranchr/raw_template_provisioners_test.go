// raw_template_provisioners_test.go: tests for provisioners.
package ranchr

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
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
	Type:    "centos",
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
			"shell",
			"file",
		},
		Provisioners: map[string]*provisioner{
			"shell": {
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
			"file": {
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
	Type:    "centos",
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
			"shell",
			"file",
		},
		Provisioners: map[string]*provisioner{
			"ansible-local": {
				templateSection{
					Settings: []string{
						"playbook_file= :src_dir/ansible/playbook_file",
						"command =  :commands_src_dir/ansible_test.command",
						"inventory_file = :src_dir/ansible/inventory_file",
						"playbook_dir = :src_dir/ansible/playbooks",
						"staging_directory = staging/directory",
					},
					Arrays: map[string]interface{}{
						"extra_arguments": []string{
							"arg1",
							"arg2",
						},
						"playbook_paths": []string{
							"path1",
							"path2",
						},
						"role_paths": []string{
							"rolepath1",
							"rolepath2",
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
			"shell": {
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
			"file": {
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
	"shell": &provisioner{
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
	"file": {
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
	"shell": &provisioner{
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
	"shell": &provisioner{
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
	"file": {
		templateSection{
			Settings: []string{
				"source = src/",
				"destination = dst/",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

var prCreatedString = "[{\"except\":[\"docker\"],\"execute_command = :commands_src_dir/execute_test.command\":\":commands_src_dir/execute_test.command\",\"only\":[\"virtualbox-iso\"],\"scripts\":[\":scripts_dir/setup_test.sh\",\":scripts_dir/vagrant_test.sh\",\":scripts_dir/sudoers_test.sh\",\":scripts_dir/cleanup_test.sh\"],\"type\":\"shell\"},{\"destination\":\"dst/\",\"source\":\"src/\",\"type\":\"file\"}]"

func TestRawTemplateUpdateProvisioners(t *testing.T) {
	Convey("Given a template", t, func() {
		Convey("Updating Provisioners with nil", func() {
			testRawTemplateProvisioner.updateProvisioners(nil)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.Get(testRawTemplateProvisioner.Provisioners), ShouldEqual, MarshalJSONToString.Get(prOrig))
			})
		})
		Convey("Updating Provisioners with new values", func() {
			testRawTemplateProvisioner.updateProvisioners(prNew)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners), ShouldEqual, MarshalJSONToString.GetIndented(prMerged))
			})
		})

	})
}

func TestProvisionersSettingsToMap(t *testing.T) {
	Convey("Given a provisioner and a raw template", t, func() {
		Convey("transform settingns map should result in", func() {
			res := pr.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type": "shell", "execute_command": "echo 'vagrant' | sudo -S sh '{{.Path}}'"})
			})
		})
	})
}

func TestRawTemplateCreateProvisioners(t *testing.T) {
	Convey("Given a template", t, func() {
		var prov interface{}
		var err error
		Convey("Creating Provisioners", func() {
			prov, _, err = testRawTemplateProvisionersAll.createProvisioners()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in Provisioners", func() {
				//				So(MarshalJSONToString.GetIndented(testRawTemplateProvisioner.Provisioners), ShouldEqual, MarshalJSONToString.GetIndented(prov))
				So(MarshalJSONToString.Get(prov), ShouldEqual, prCreatedString)
			})
		})
	})
}
