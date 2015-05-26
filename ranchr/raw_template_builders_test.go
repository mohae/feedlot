// raw_template_builders_test.go: tests for builders.
package ranchr

import (
	"testing"
)

var testUbuntu = &rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: "commands",
		OutDir:         "../test_files/ubuntu/out/ubuntu",
		SrcDir:         "../test_files/src/ubuntu",
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
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 300m",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=4096",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection{
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
		},
		PostProcessors: map[string]*postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"keep_input_artifact = false",
						"output = out/someComposedBoxName.box",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"shell",
		},
		Provisioners: map[string]*provisioner{
			"shell": {
				templateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"base_test.sh",
							"vagrant_test.sh",
							"cleanup_test.sh",
							"zerodisk_test.sh",
						},
					},
				},
			},
		},
	},
}

var testCentOS = &rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: "commands",
		OutDir:         "../test_files/out/centos",
		SrcDir:         "../test_files/src/centos",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test build template for salt provisioner using CentOS6",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
		BaseURL:   "",
	},
	Distro:  "centos",
	Arch:    "x86_64",
	Image:   "minimal",
	Release: "6",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"virtualbox-ovf",
			"vmware-iso",
			"vmware-vmx",
		},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=4096",
						},
					},
				},
			},
			"virtualbox-ovf": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=4096",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpuid.coresPerSocket=1",
							"memsize=1024",
							"numvcpus=1",
						},
					},
				},
			},
			"vmware-vmx": {
				templateSection{
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
		},
		PostProcessors: map[string]*postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"keep_input_artifact = false",
						"output = out/someComposedBoxName.box",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"shell",
			"salt",
		},
		Provisioners: map[string]*provisioner{
			"salt": {
				templateSection{
					Settings: []string{
						"local_state_tree = ~/saltstates/centos6/salt",
						"skip_bootstrap = true",
					},
				},
			},
			"shell": {
				templateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"base_test.sh",
							"vagrant_test.sh",
							"cleanup_test.sh",
							"zerodisk_test.sh",
						},
					},
				},
			},
		},
	},
}

var testAllBuilders = &rawTemplate{
	varVals: make(map[string]string),
	vars:    make(map[string]string),
	files:   make(map[string]string),
	dirs:    make(map[string]string),
	IODirInf: IODirInf{
		OutDir: "../test_files/out",
		SrcDir: "../test_files/src",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test build template for all builders",
	},
	BuildInf: BuildInf{
		Name:      "docker-alt",
		BuildName: "",
		BaseURL:   "",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "minimal",
	Release: "14.04",
	build: build{
		BuilderTypes: []string{
			"amazon-ebs",
			"digitalocean",
			"docker",
			"googlecompute",
			"virtualbox-iso",
			"virtualbox-ovf",
			"vmware-iso",
			"vmware-vmx",
		},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"amazon-ebs": {
				templateSection{
					Settings: []string{
						"access_key=AWS_ACCESS_KEY",
						"ami_description=AMI_DESCRIPTION",
						"ami_name=AMI_NAME",
						"associate_public_ip_address=false",
						"availability_zone=us-east-1b",
						"enhanced_networking=false",
						"instance_type=m3.medium",
						"iam_instance_profile=INSTANCE_PROFILE",
						"region=us-east-1",
						"secret_key=AWS_SECRET_ACCESS_KEY",
						"security_group_id=GROUP_ID",
						"source_ami=SOURCE_AMI",
						"spot_price=auto",
						"spot_price_auto_product=Linux/Unix",
						"ssh_port=22",
						"ssh_username=ssh_user",
						"ssh_private_key_file=myKey",
						"ssh_timeout=5m",
						"temporary_key_pair_name=TMP_KEYPAIR",
						"token=AWS_SECURITY_TOKEN",
						"user_data=SOME_USER_DATA",
						"user_data_file=amazon.userdata",
						"vpc_id=VPC_ID",
					},
					Arrays: map[string]interface{}{
						"ami_groups": []string{
							"AGroup",
						},
						"ami_product_codes": []string{
							"ami-d4e356aa",
						},
						"ami_regions": []string{
							"us-east-1",
						},
						"ami_users": []string{
							"ami-account",
						},
						"security_group_ids": []string{
							"SECURITY_GROUP",
						},
					},
				},
			},
			"digitalocean": {
				templateSection{
					Settings: []string{
						"api_token=DIGITALOCEAN_API_TOKEN",
						"api_url=https://api.digitalocean.com",
						"droplet_name=ocean-drop",
						"image=ubuntu-12-04-x64",
						"private_networking=false",
						"region=nyc3",
						"size=512mb",
						"snapshot_name=my-snapshot",
						"ssh_port=22",
						"ssh_timeout=5m",
						"ssh_username=user",
						"state_timeout=6m",
					},
				},
			},
			"docker": {
				templateSection{
					Settings: []string{
						"commit=true",
						"export_path=export/path",
						"image=baseImage",
						"login=true",
						"login_email=test@test.com",
						"login_username=username",
						"login_password=password",
						"login_server=127.0.0.1",
						"pull=true",
					},
					Arrays: map[string]interface{}{
						"run_command": []string{
							"-d",
							"-i",
							"-t",
							"{{.Image}}",
							"/bin/bash",
						},
					},
				},
			},
			"googlecompute": {
				templateSection{
					Settings: []string{
						"account_file=account.json",
						"image_name=packer-{{timestamp}}",
						"image_description=test image",
						"instance_name=packer-{{uuid}}",
						"machine_type=nl-standard-1",
						"network=default",
						"project_id=projectID",
						"source_image=centos-6",
						"ssh_timeout=5m",
						"state_timeout=5m",
						"zone=us-central1-a",
					},
					Arrays: map[string]interface{}{
						"tags": []string{
							"tag1",
						},
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Settings: []string{
						"format = ovf",
						"guest_additions_mode=upload",
						"guest_additions_path=path/to/additions",
						"guest_additions_sha256=89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
						"guest_additions_url=file://guest-additions",
						"hard_drive_interface=ide",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_interface=ide",
						"iso_url=http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_key_path=key/path",
						"ssh_port=22",
						"virtualbox_version_file=.vbox_version",
						"vm_name=test-vb-iso",
					},
					Arrays: map[string]interface{}{
						"boot_command": []string{
							"<bs>",
							"<del>",
							"<enter><return>",
							"<esc>",
						},
						"export_opts": []string{
							"opt1",
						},
						"floppy_files": []string{
							"disk1",
						},
						"vboxmanage": []string{
							"--cpus=1",
							"--memory=4096",
						},
						"vboxmanage_post": []string{
							"something=value",
						},
					},
				},
			},
			"virtualbox-ovf": {
				templateSection{
					Settings: []string{
						"format = ovf",
						"guest_additions_mode=upload",
						"guest_additions_path=path/to/additions",
						"guest_additions_sha256=89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
						"guest_additions_url=file://guest-additions",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"import_opts=keepallmacs",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"source_path=source/path",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_key_path=key/path",
						"ssh_port=22",
						"virtualbox_version_file=.vbox_version",
						"vm_name=test-vb-ovf",
					},
					Arrays: map[string]interface{}{
						"boot_command": []string{
							"<bs>",
							"<del>",
							"<enter><return>",
							"<esc>",
						},
						"export_opts": []string{
							"opt1",
						},
						"floppy_files": []string{
							"disk1",
						},
						"vboxmanage": []string{
							"--cpus=1",
							"--memory=4096",
						},
						"vboxmanage_post": []string{
							"something=value",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection{
					Settings: []string{
						"disk_type_id=1",
						"fusion_app_path=/Applications/VMware Fusion.app",
						"hard_drive_interface=ide",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_url=http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
						"output_directory=out/dir",
						"remote_cache_datastore=datastore1",
						"remote_cache_directory=packer_cache",
						"remote_datastore=datastore1",
						"remote_host=remoteHost",
						"remote_password=rpassword",
						"remote_type=esx5",
						"shutdown_timeout=5m",
						"ssh_host=127.0.0.1",
						"ssh_key_path=key/path",
						"ssh_port=22",
						"tools_upload_flavor=linux",
						"tools_upload_path={{.Flavor}}.iso",
						"version=9",
						"vm_name=packer-BUILDNAME",
						"vmdk_name=packer",
						"vmx_template_path=template/path",
						"vnc_port_min=5900",
						"vnc_port_max=6000",
					},
					Arrays: map[string]interface{}{
						"boot_command": []string{
							"<bs>",
							"<del>",
							"<enter><return>",
							"<esc>",
						},
						"floppy_files": []string{
							"disk1",
						},
						"vmx_data": []string{
							"cpuid.coresPerSocket=1",
							"memsize=1024",
							"numvcpus=1",
						},
						"vmx_data_post": []string{
							"something=value",
						},
					},
				},
			},
			"vmware-vmx": {
				templateSection{
					Settings: []string{
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"skip_compaction=false",
						"source_path=source/path/file.vmx",
						"ssh_key_path=key/path",
						"ssh_port=22",
						"ssh_skip_request_pty=false",
						"vm_name=packer-BUILDNAME",
						"vnc_port_min=5900",
						"vnc_port_max=6000",
					},
					Arrays: map[string]interface{}{
						"boot_command": []string{
							"<bs>",
							"<del>",
							"<enter><return>",
							"<esc>",
						},
						"floppy_files": []string{
							"disk1",
						},
						"vmx_data": []string{
							"cpuid.coresPerSocket=1",
							"memsize=1024",
							"numvcpus=1",
						},
						"vmx_data_post": []string{
							"something=value",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
		},
		PostProcessors: map[string]*postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"keep_input_artifact = false",
						"output = out/someComposedBoxName.box",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"salt",
		},
		Provisioners: map[string]*provisioner{
			"salt": {
				templateSection{
					Settings: []string{
						"local_state_tree = ~/saltstates/centos6/salt",
						"skip_bootstrap = true",
					},
				},
			},
		},
	},
}

var testDigtialOceanAPIV1 = &rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: "commands",
		OutDir:         "../test_files/ubuntu/out/ubuntu",
		SrcDir:         "../test_files/src/ubuntu",
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
	build: build{
		BuilderTypes: []string{
			"digitalocean",
		},
		Builders: map[string]*builder{
			"digitalocean": {
				templateSection{
					Settings: []string{
						"api_key=DIGITALOCEAN_API_KEY",
						"client_id=DIGITALOCEAN_CLIENT_ID",
						"api_url=https://api.digitalocean.com",
						"image=ubuntu-12-04-x64",
						"droplet_name=ocean-drop",
						"private_networking=false",
						"region=nyc3",
						"size=512mb",
						"snapshot_name=my-snapshot",
						"ssh_port=22",
						"ssh_timeout=5m",
						"ssh_username=user",
						"state_timeout=6m",
					},
				},
			},
		},
	},
}

var testDigtialOceanNoAPI = &rawTemplate{
	IODirInf: IODirInf{
		OutDir: "../test_files/ubuntu/out/ubuntu",
		SrcDir: "../test_files/src/ubuntu",
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
	build: build{
		BuilderTypes: []string{
			"digitalocean",
		},
		Builders: map[string]*builder{
			"digitalocean": {
				templateSection{
					Settings: []string{
						"api_url=https://api.digitalocean.com",
						"droplet_name=ocean-drop",
						"image=ubuntu-12-04-x64",
						"private_networking=false",
						"region=nyc3",
						"size=512mb",
						"snapshot_name=my-snapshot",
						"ssh_port=22",
						"ssh_timeout=5m",
						"ssh_username=user",
						"state_timeout=6m",
					},
				},
			},
		},
	},
}
var testDockerRunComandFile = &rawTemplate{
	varVals: make(map[string]string),
	vars:    make(map[string]string),
	files:   make(map[string]string),
	dirs:    make(map[string]string),
	IODirInf: IODirInf{
		OutDir: "../test_files/out",
		SrcDir: "../test_files/src",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test build template for all builders",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
		BaseURL:   "",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "minimal",
	Release: "14.04",
	build: build{
		BuilderTypes: []string{
			"docker",
		},
		Builders: map[string]*builder{
			"docker": {
				templateSection{
					Settings: []string{
						"commit=true",
						"export_path=export/path",
						"image=baseImage",
						"login=true",
						"login_email=test@test.com",
						"login_username=username",
						"login_password=password",
						"login_server=127.0.0.1",
						"pull=true",
						"run_command=docker.command",
					},
					Arrays: map[string]interface{}{},
				},
			},
		},
	},
}

// This should still result in only 1 command array, using the array value and not the
// file
var testDockerRunComand = &rawTemplate{
	varVals: make(map[string]string),
	vars:    make(map[string]string),
	files:   make(map[string]string),
	dirs:    make(map[string]string),
	IODirInf: IODirInf{
		OutDir: "../test_files/out",
		SrcDir: "../test_files/src",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test build template for all builders",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
		BaseURL:   "",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "minimal",
	Release: "14.04",
	build: build{
		BuilderTypes: []string{
			"docker",
		},
		Builders: map[string]*builder{
			"docker": {
				templateSection{
					Settings: []string{
						"commit=true",
						"export_path=export/path",
						"image=baseImage",
						"login=true",
						"login_email=test@test.com",
						"login_username=username",
						"login_password=password",
						"login_server=127.0.0.1",
						"pull=true",
						"run_command=docker.command",
					},
					Arrays: map[string]interface{}{
						"run_command": []string{
							"-d",
							"-i",
							"-t",
							"{{.Image}}",
							"/bin/bash",
						},
					},
				},
			},
		},
	},
}
var builderOrig = map[string]*builder{
	"common": {
		templateSection{
			Settings: []string{
				"boot_command = boot_test.command",
				"boot_wait = 5s",
				"disk_size = 20000",
				"http_directory = http",
				"iso_checksum_type = sha256",
				"shutdown_command = shutdown_test.command",
				"ssh_password = vagrant",
				"ssh_port = 22",
				"ssh_username = vagrant",
				"ssh_wait_timeout = 300m",
			},
			Arrays: map[string]interface{}{},
		},
	},
	"virtualbox-iso": {
		templateSection{
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"cpus=1",
					"memory=4096",
				},
			},
		},
	},
	"vmware-iso": {
		templateSection{
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
	},
}

var builderNew = map[string]*builder{
	"common": {
		templateSection{
			Settings: []string{
				"boot_command = boot_test.command",
				"boot_wait = 15s",
				"disk_size = 20000",
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
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"cpus=1",
					"memory=2048",
				},
			},
		},
	},
}

var builderMerged = map[string]*builder{
	"common": {
		templateSection{
			Settings: []string{
				"boot_command = boot_test.command",
				"boot_wait = 15s",
				"disk_size = 20000",
				"http_directory = http",
				"iso_checksum_type = sha256",
				"shutdown_command = shutdown_test.command",
				"ssh_password = vagrant",
				"ssh_port = 22",
				"ssh_username = vagrant",
				"ssh_wait_timeout = 240m",
			},
			Arrays: map[string]interface{}{},
		},
	},
	"virtualbox-iso": {
		templateSection{
			Settings: []string{},
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"cpus=1",
					"memory=2048",
				},
			},
		},
	},
	"vmware-iso": {
		templateSection{
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
	},
}

var vbB = &builder{
	templateSection{
		Settings: []string{
			"boot_wait=5s",
			"disk_size = 20000",
			"ssh_port= 22",
			"ssh_username =vagrant",
		},
		Arrays: map[string]interface{}{
			"vm_settings": []string{
				"cpuid.coresPerSocket=1",
				"memsize=2048",
			},
		},
	},
}

func TestCreateBuilders(t *testing.T) {
	_, _, err := testRawTemplateBuilderOnly.createBuilders()
	if err == nil {
		t.Error("Expected error \"unable to create builders: none specified\", got nil")
	} else {
		if err.Error() != "unable to create builders: none specified" {
			t.Errorf("Expected \"unable to create builders: none specified\", got %q", err.Error())
		}
	}

	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"amazon-ebs\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"amazon-ebs\"" {
			t.Errorf("Expected error \"no configuration found for \"amazon-ebs\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "digitalocean"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"digitalocean\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"digitalocean\"" {
			t.Errorf("Expected error \"no configuration found for \"digitalocean\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "docker"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"docker\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"docker\"" {
			t.Errorf("Expected error \"no configuration found for \"docker\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "googlecompute"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"googlecompute\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"googlecompute\"" {
			t.Errorf("Expected error \"no configuration found for \"googlecompute\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "virtualbox-iso"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"virtualbox-iso\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"virtualbox-iso\"" {
			t.Errorf("Expected error \"no configuration found for \"virtualbox-iso\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "virtualbox-ovf"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"virtualbox-ovf\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"virtualbox-ovf\"" {
			t.Errorf("Expected error \"no configuration found for \"virtualbox-ovf\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "vmware-iso"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"vmware-iso\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"vmware-iso\"" {
			t.Errorf("Expected error \"no configuration found for \"vmware-iso\"\", got %q", err.Error())
		}
	}

	testRawTemplateWOSection.build.BuilderTypes[0] = "vmware-vmx"
	_, _, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected error \"no configuration found for \"vmware-vmx\"\", got nil")
	} else {
		if err.Error() != "no configuration found for \"vmware-vmx\"" {
			t.Errorf("Expected error \"no configuration found for \"vmware-vmx\"\", got %q", err.Error())
		}
	}

	r := &rawTemplate{}
	r = testDistroDefaultUbuntu
	r.files = make(map[string]string)
	var bldrs []interface{}
	//first merge the variables so that create builders will work
	r.mergeVariables()
	bldrs, _, err = r.createBuilders()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	/*else {
		if MarshalJSONToString.Get(bldrs[0]) != MarshalJSONToString.Get(merged[0]) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(merged[0]), MarshalJSONToString.Get(bldrs[0]))
		}
		if MarshalJSONToString.Get(bldrs[1]) != MarshalJSONToString.Get(merged[1]) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(merged[1]), MarshalJSONToString.Get(bldrs[1]))
		}
	}
	*/
	_ = bldrs

	r.BuilderTypes[0] = "unsupported"
	bldrs, _, err = r.createBuilders()
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "Builder, \"unsupported\", is not supported by Rancher" {
			t.Errorf("Expected \"tBuilder, \"unsupported\", is not supported by Rancher\"), got %q", err.Error())
		}
	}

	r.BuilderTypes = nil
	bldrs, _, err = r.createBuilders()
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "unable to create builders: none specified" {
			t.Errorf("Expected \"unable to create builders: none specified\"), got %q", err.Error())
		}
	}
}

func TestRawTemplateUpdatebuilders(t *testing.T) {
	testUbuntu.updateBuilders(nil)
	if MarshalJSONToString.Get(testUbuntu.Builders) != MarshalJSONToString.Get(builderOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(builderOrig), MarshalJSONToString.Get(testUbuntu.Builders))
	}

	testUbuntu.updateBuilders(builderNew)
	if MarshalJSONToString.Get(testUbuntu.Builders) != MarshalJSONToString.Get(builderMerged) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(builderMerged), MarshalJSONToString.Get(testUbuntu.Builders))
	}
}

func TestRawTemplateUpdateBuilderCommon(t *testing.T) {
	testUbuntu.updateCommon(builderNew["common"])
	if MarshalJSONToString.Get(testUbuntu.Builders["common"]) != MarshalJSONToString.Get(builderMerged["common"]) {
		t.Errorf("expected %q, got %q", MarshalJSONToString.Get(builderMerged["common"]), MarshalJSONToString.Get(testUbuntu.Builders["common"]))
	}
}

func TestRawTemplateBuildersSettingsToMap(t *testing.T) {
	settings := vbB.settingsToMap(testRawTpl)
	if settings["boot_wait"] != "5s" {
		t.Errorf("Expected \"5s\", got %q", settings["boot_wait"])
	}
	if settings["disk_size"] != "20000" {
		t.Errorf("Expected \"20000\", got %q", settings["disk_size"])
	}
	if settings["ssh_port"] != "22" {
		t.Errorf("Expected \"22\", got %q", settings["ssh_port"])
	}
	if settings["ssh_username"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_username"])
	}
}

func TestAmazonEBSBuilder(t *testing.T) {
	expected := map[string]interface{}{
		"access_key":      "AWS_ACCESS_KEY",
		"ami_description": "AMI_DESCRIPTION",
		"ami_groups": []string{
			"AGroup",
		},
		"ami_name": "AMI_NAME",
		"ami_product_codes": []string{
			"ami-d4e356aa",
		},
		"ami_regions": []string{
			"us-east-1",
		},
		"associate_public_ip_address": false,
		"availability_zone":           "us-east-1b",
		"enhanced_networking":         false,
		"iam_instance_profile":        "INSTANCE_PROFILE",
		"instance_type":               "m3.medium",
		"region":                      "us-east-1",
		"secret_key":                  "AWS_SECRET_ACCESS_KEY",
		"security_group_id":           "GROUP_ID",
		"security_group_ids": []string{
			"SECURITY_GROUP",
		},
		"source_ami":              "SOURCE_AMI",
		"spot_price":              "auto",
		"spot_price_auto_product": "Linux/Unix",
		"ssh_port":                22,
		"ssh_username":            "ssh_user",
		"ssh_private_key_file":    "../test_files/src/amazon-ebs/myKey",
		"ssh_timeout":             "5m",
		"temporary_key_pair_name": "TMP_KEYPAIR",
		"token":                   "AWS_SECURITY_TOKEN",
		"type":                    "amazon-ebs",
		"user_data":               "SOME_USER_DATA",
		"user_data_file":          "../test_files/src/amazon-ebs/amazon.userdata",
		"vpc_id":                  "VPC_ID",
	}
	bldr, _, err := testAllBuilders.createAmazonEBS()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestDigitalOceanBuilder(t *testing.T) {
	expectedV1 := map[string]interface{}{
		"api_key":            "DIGITALOCEAN_API_KEY",
		"client_id":          "DIGITALOCEAN_CLIENT_ID",
		"api_url":            "https://api.digitalocean.com",
		"droplet_name":       "ocean-drop",
		"image":              "ubuntu-12-04-x64",
		"private_networking": false,
		"region":             "nyc3",
		"size":               "512mb",
		"snapshot_name":      "my-snapshot",
		"ssh_port":           22,
		"ssh_timeout":        "5m",
		"ssh_username":       "user",
		"state_timeout":      "6m",
		"type":               "digitalocean",
	}
	expectedV2 := map[string]interface{}{
		"api_token":          "DIGITALOCEAN_API_TOKEN",
		"api_url":            "https://api.digitalocean.com",
		"droplet_name":       "ocean-drop",
		"image":              "ubuntu-12-04-x64",
		"private_networking": false,
		"region":             "nyc3",
		"size":               "512mb",
		"snapshot_name":      "my-snapshot",
		"ssh_port":           22,
		"ssh_timeout":        "5m",
		"ssh_username":       "user",
		"state_timeout":      "6m",
		"type":               "digitalocean",
	}
	bldr, _, err := testAllBuilders.createDigitalOcean()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedV2) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedV2), MarshalJSONToString.Get(bldr))
		}
	}
	bldr, _, err = testDigtialOceanAPIV1.createDigitalOcean()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedV1) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedV1), MarshalJSONToString.Get(bldr))
		}
	}
	_, _, err = testDigtialOceanNoAPI.createDigitalOcean()
	if err == nil {
		t.Errorf("Expected an error, got nil")
	} else {
		if err.Error() != "required Digital Ocean API information not set" {
			t.Errorf("Expected \"required Digital Ocean API information not set\", got %q", err.Error())
		}
	}
}

func TestDockerBuilder(t *testing.T) {
	expected := map[string]interface{}{
		"commit":         true,
		"export_path":    "export/path",
		"image":          "baseImage",
		"login":          true,
		"login_email":    "test@test.com",
		"login_username": "username",
		"login_password": "password",
		"login_server":   "127.0.0.1",
		"pull":           true,
		"run_command": []string{
			"-d",
			"-i",
			"-t",
			"{{.Image}}",
			"/bin/bash",
		},
		"type": "docker",
	}
	expectedCommandFile := map[string]interface{}{
		"commit":         true,
		"export_path":    "export/path",
		"image":          "baseImage",
		"login":          true,
		"login_email":    "test@test.com",
		"login_username": "username",
		"login_password": "password",
		"login_server":   "127.0.0.1",
		"pull":           true,
		"run_command": []string{
			"-d",
			"-i",
			"-t",
			"{{.Image}}",
			"/bin/bash",
			"/invalid",
		},
		"type": "docker",
	}
	bldr, _, err := testAllBuilders.createDocker()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
	bldr, _, err = testDockerRunComandFile.createDocker()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedCommandFile) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedCommandFile), MarshalJSONToString.Get(bldr))
		}
	}
	bldr, _, err = testDockerRunComand.createDocker()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestDockerGoogleCompute(t *testing.T) {
	expected := map[string]interface{}{
		"account_file":      "account.json",
		"disk_size":         20000,
		"image_name":        "packer-{{timestamp}}",
		"image_description": "test image",
		"instance_name":     "packer-{{uuid}}",
		"machine_type":      "nl-standard-1",
		"network":           "default",
		"project_id":        "projectID",
		"source_image":      "centos-6",
		"ssh_port":          22,
		"ssh_timeout":       "5m",
		"ssh_username":      "vagrant",
		"state_timeout":     "5m",
		"tags": []string{
			"tag1",
		},
		"type": "googlecompute",
		"zone": "us-central1-a",
	}

	bldr, _, err := testAllBuilders.createGoogleCompute()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestCreateBuilderVirtualboxISO(t *testing.T) {
	expected := map[string]interface{}{
		"boot_command": []string{
			"<bs>",
			"<del>",
			"<enter><return>",
			"<esc>",
		},
		"boot_wait": "5s",
		"disk_size": 20000,
		"export_opts": []string{
			"opt1",
		},
		"floppy_files": []string{
			"disk1",
		},
		"format":                 "ovf",
		"guest_additions_mode":   "upload",
		"guest_additions_path":   "path/to/additions",
		"guest_additions_sha256": "89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
		"guest_additions_url":    "file://guest-additions",
		"guest_os_type":          "Ubuntu_64",
		"hard_drive_interface":   "ide",
		"headless":               true,
		"http_directory":         "http",
		"http_port_max":          9000,
		"http_port_min":          8000,
		"iso_checksum":           "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type":      "sha256",
		"iso_interface":          "ide",
		"iso_url":                "http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		"output_directory":       "out/dir",
		"shutdown_command":       "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":       "5m",
		"ssh_host_port_max":      40,
		"ssh_host_port_min":      22,
		"ssh_key_path":           "key/path",
		"ssh_password":           "vagrant",
		"ssh_port":               22,
		"ssh_username":           "vagrant",
		"ssh_wait_timeout":       "300m",
		"type":                   "virtualbox-iso",
		"vboxmanage": [][]string{
			[]string{
				"modifyvm",
				"{{.Name}}",
				"--cpus",
				"1",
			},
			[]string{
				"modifyvm",
				"{{.Name}}",
				"--memory",
				"4096",
			},
		},
		"vboxmanage_post": [][]string{
			[]string{
				"modifyvm",
				"{{.Name}}",
				"something",
				"value",
			},
		},
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-iso",
	}

	settings, _, err := testAllBuilders.createVirtualBoxISO()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestCreateBuilderVirtualboxOVF(t *testing.T) {
	expected := map[string]interface{}{
		"boot_command": []string{
			"<bs>",
			"<del>",
			"<enter><return>",
			"<esc>",
		},
		"boot_wait": "5s",
		"export_opts": []string{
			"opt1",
		},
		"floppy_files": []string{
			"disk1",
		},
		"format":                 "ovf",
		"guest_additions_mode":   "upload",
		"guest_additions_path":   "path/to/additions",
		"guest_additions_sha256": "89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
		"guest_additions_url":    "file://guest-additions",
		"headless":               true,
		"http_directory":         "http",
		"http_port_max":          9000,
		"http_port_min":          8000,
		"import_opts":            "keepallmacs",
		"output_directory":       "out/dir",
		"shutdown_command":       "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":       "5m",
		"source_path":            "source/path",
		"ssh_host_port_max":      40,
		"ssh_host_port_min":      22,
		"ssh_key_path":           "key/path",
		"ssh_password":           "vagrant",
		"ssh_port":               22,
		"ssh_username":           "vagrant",
		"ssh_wait_timeout":       "300m",
		"type":                   "virtualbox-ovf",
		"vboxmanage": [][]string{
			[]string{
				"modifyvm",
				"{{.Name}}",
				"--cpus",
				"1",
			},
			[]string{
				"modifyvm",
				"{{.Name}}",
				"--memory",
				"4096",
			},
		},
		"vboxmanage_post": [][]string{
			[]string{
				"modifyvm",
				"{{.Name}}",
				"something",
				"value",
			},
		},
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-ovf",
	}
	testAllBuilders.files = make(map[string]string)
	settings, _, err := testAllBuilders.createVirtualBoxOVF()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestCreateBuilderVMWareISO(t *testing.T) {
	expected := map[string]interface{}{
		"boot_command": []string{
			"<bs>",
			"<del>",
			"<enter><return>",
			"<esc>",
		},
		"boot_wait":    "5s",
		"disk_size":    20000,
		"disk_type_id": "1",
		"floppy_files": []string{
			"disk1",
		},
		"fusion_app_path":        "/Applications/VMware Fusion.app",
		"guest_os_type":          "Ubuntu_64",
		"headless":               true,
		"http_directory":         "http",
		"http_port_max":          9000,
		"http_port_min":          8000,
		"iso_checksum":           "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type":      "sha256",
		"iso_url":                "http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		"output_directory":       "out/dir",
		"remote_cache_datastore": "datastore1",
		"remote_cache_directory": "packer_cache",
		"remote_datastore":       "datastore1",
		"remote_host":            "remoteHost",
		"remote_password":        "rpassword",
		"remote_type":            "esx5",
		"shutdown_command":       "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":       "5m",
		"ssh_host":               "127.0.0.1",
		"ssh_key_path":           "key/path",
		"ssh_password":           "vagrant",
		"ssh_port":               22,
		"ssh_username":           "vagrant",
		"ssh_wait_timeout":       "300m",
		"tools_upload_flavor":    "linux",
		"tools_upload_path":      "{{.Flavor}}.iso",
		"type":                   "vmware-iso",
		"vmx_data": map[string]string{
			"cpuid.coresPerSocket": "1",
			"memsize":              "1024",
			"numvcpus":             "1",
		},
		"vmx_data_post": map[string]string{
			"something": "value",
		},
		"vm_name":           "packer-BUILDNAME",
		"vmdk_name":         "packer",
		"vmx_template_path": "template/path",
		"vnc_port_max":      6000,
		"vnc_port_min":      5900,
	}

	settings, _, err := testAllBuilders.createVMWareISO()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}

func TestCreateBuilderVMWareVMX(t *testing.T) {
	expected := map[string]interface{}{
		"boot_command": []string{
			"<bs>",
			"<del>",
			"<enter><return>",
			"<esc>",
		},
		"boot_wait": "5s",
		"floppy_files": []string{
			"disk1",
		},
		"fusion_app_path":      "/Applications/VMware Fusion.app",
		"headless":             true,
		"http_directory":       "http",
		"http_port_max":        9000,
		"http_port_min":        8000,
		"output_directory":     "out/dir",
		"shutdown_command":     "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":     "5m",
		"skip_compaction":      false,
		"source_path":          "source/path/file.vmx",
		"ssh_key_path":         "key/path",
		"ssh_password":         "vagrant",
		"ssh_port":             22,
		"ssh_skip_request_pty": false,
		"ssh_username":         "vagrant",
		"ssh_wait_timeout":     "300m",
		"type":                 "vmware-vmx",
		"vmx_data": map[string]string{
			"cpuid.coresPerSocket": "1",
			"memsize":              "1024",
			"numvcpus":             "1",
		},
		"vmx_data_post": map[string]string{
			"something": "value",
		},
		"vm_name":      "packer-BUILDNAME",
		"vnc_port_max": 6000,
		"vnc_port_min": 5900,
	}

	settings, _, err := testAllBuilders.createVMWareVMX()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
}
func TestDeepCopyMapStringPBuilder(t *testing.T) {
	cpy := DeepCopyMapStringPBuilder(testDistroDefaults.Templates[Ubuntu].Builders)
	if MarshalJSONToString.Get(cpy["common"]) != MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu].Builders["common"]) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu].Builders["common"]), MarshalJSONToString.Get(cpy["common"]))
	}
}
