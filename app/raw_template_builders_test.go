// raw_template_builders_test.go: tests for builders.
package app

import (
	"testing"

	"github.com/mohae/contour"
)

var testUbuntu = rawTemplate{
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
	Image:   "server",
	Release: "14.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
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
						"ssh_timeout = 30m",
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
		PostProcessorIDs: []string{
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"keep_input_artifact = false",
						"output = out/someComposedBoxName.box",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
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

var testCentOS = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/out/centos",
		SourceDir: "../test_files/src/centos",
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
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"virtualbox-iso",
			"virtualbox-ovf",
			"vmware-iso",
			"vmware-vmx",
		},
		Builders: map[string]builder{
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
						"ssh_timeout = 30m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"--cpus=1",
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
							"--memory=4096",
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
		PostProcessorIDs: []string{
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"keep_input_artifact = false",
						"output = out/someComposedBoxName.box",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
			"salt",
		},
		Provisioners: map[string]provisioner{
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

// Not all the settings in are valid for winrm, the invalid ones should not be included in the
// output.
var testAllBuilders = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/out",
		SourceDir: "../test_files/src",
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
	Image:   "server",
	Release: "14.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"amazon-ebs",
			"amazon-instance",
			"digitalocean",
			"docker",
			"googlecompute",
			"null",
			"virtualbox-iso",
			"virtualbox-ovf",
			"vmware-iso",
			"vmware-vmx",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
					Settings: []string{
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_timeout = 30m",
					},
				},
			},
			"amazon-chroot": {
				templateSection{
					Type: "amazon-chroot",
					Settings: []string{
						"access_key=AWS_ACCESS_KEY",
						"ami_description=AMI_DESCRIPTION",
						"ami_name=AMI_NAME",
						"ami_virtualization_type=paravirtual",
						"command_wrapper={{.Command}}",
						"device_path=/dev/xvdf",
						"enhanced_networking=false",
						"mount_path=packer-amazon-chroot-volumes/{{.Device}}",
						"secret_key=AWS_SECRET_ACCESS_KEY",
						"source_ami=SOURCE_AMI",
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
							"aws-account-1",
						},
						"chroot_mounts": []interface{}{
							[]string{
								"proc",
								"proc",
								"/proc",
							},
							[]string{
								"bind",
								"/dev",
								"/dev",
							},
						},
						"copy_files": []string{
							"/etc/resolv.conf",
						},
						"tags": map[string]string{
							"OS_Version": "Ubuntu",
							"Release":    "Latest",
						},
					},
				},
			},
			"amazon-ebs": {
				templateSection{
					Type: "amazon-ebs",
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
						"ssh_private_key_file=myKey",
						"temporary_key_pair_name=TMP_KEYPAIR",
						"token=AWS_SECURITY_TOKEN",
						"user_data=SOME_USER_DATA",
						"user_data_file=amazon.userdata",
						"vpc_id=VPC_ID",
					},
					Arrays: map[string]interface{}{
						"ami_block_device_mappings": []map[string]string{
							{
								"device_name":  "/dev/sdb",
								"virtual_name": "/ephemeral0",
							},
							{
								"device_name":  "/dev/sdc",
								"virtual_name": "/ephemeral1",
							},
						},
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
						"launch_block_device_mappings": []map[string]string{
							{
								"device_name":  "/dev/sdd",
								"virtual_name": "/ephemeral2",
							},
							{
								"device_name":  "/dev/sde",
								"virtual_name": "/ephemeral3",
							},
						},
						"security_group_ids": []string{
							"SECURITY_GROUP",
						},
						"run_tags": map[string]string{
							"foo": "bar",
							"fiz": "baz",
						},
						"tags": map[string]string{
							"OS_Version": "Ubuntu",
							"Release":    "Latest",
						},
					},
				},
			},
			"amazon-instance": {
				templateSection{
					Type: "amazon-instance",
					Settings: []string{
						"access_key=AWS_ACCESS_KEY",
						"account_id=YOUR_ACCOUNT_ID",
						"ami_description=AMI_DESCRIPTION",
						"ami_name=AMI_NAME",
						"ami_virtualization_type=paravirtual",
						"associate_public_ip_address=false",
						"availability_zone=us-east-1b",
						"bundle_destination=/tmp",
						"bundle_prefix=image--{{timestamp}}",
						"bundle_upload_command=bundle_upload.command",
						"bundle_vol_command=bundle_vol.command",
						"ebs_optimized=true",
						"enhanced_networking=false",
						"force_deregister=false",
						"iam_instance_profile=INSTANCE_PROFILE",
						"instance_type=m3.medium",
						"region=us-east-1",
						"s3_bucket=packer_bucket",
						"secret_key=AWS_SECRET_ACCESS_KEY",
						"security_group_id=GROUP_ID",
						"source_ami=SOURCE_AMI",
						"spot_price=auto",
						"spot_price_auto_product=Linux/Unix",
						"ssh_keypair_name=myKeyPair",
						"ssh_private_ip=true",
						"ssh_private_key_file=myKey",
						"ssh_username=vagrant",
						"subnet_id=subnet-12345def",
						"temporary_key_pair_name=TMP_KEYPAIR",
						"user_data=SOME_USER_DATA",
						"user_data_file=amazon.userdata",
						"vpc_id=VPC_ID",
						"windows_password_timeout=10m",
						"x509_cert_path=/path/to/x509/cert",
						"x509_key_path=/path/to/x509/key",
						"x509_upload_path=/etc/x509",
					},
					Arrays: map[string]interface{}{
						"ami_block_device_mappings": [][]string{
							[]string{
								"delete_on_termination=true",
								"device_name=/dev/sdb",
								"encrypted=true",
								"iops=1000",
								"no_device=false",
								"snapshot_id=SNAPSHOT",
								"virtual_name=ephemeral0",
								"volume_type=io1",
								"volume_size=10",
							},
							[]string{
								"device_name=/dev/sdc",
								"volume_type=io1",
								"volume_size=10",
							},
						},
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
						"launch_block_device_mappings": []map[string]string{
							{
								"device_name":  "/dev/sdd",
								"virtual_name": "/ephemeral2",
							},
							{
								"device_name":  "/dev/sde",
								"virtual_name": "/ephemeral3",
							},
						},
						"run_tags": map[string]string{
							"foo": "bar",
							"fiz": "baz",
						},
						"security_group_ids": []string{
							"SECURITY_GROUP",
						},
						"tags": map[string]string{
							"OS_Version": "Ubuntu",
							"Release":    "Latest",
						},
					},
				},
			},
			"digitalocean": {
				templateSection{
					Type: "digitalocean",
					Settings: []string{
						"api_token=DIGITALOCEAN_API_TOKEN",
						"droplet_name=ocean-drop",
						"image=ubuntu-12-04-x64",
						"private_networking=false",
						"region=nyc3",
						"size=512mb",
						"snapshot_name=my-snapshot",
						"state_timeout=6m",
						"user_data=userdata",
					},
				},
			},
			"docker": {
				templateSection{
					Type: "docker",
					Settings: []string{
						"commit=true",
						"discard=false",
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
						"volumes": map[string]string{
							"/var/data1": "/var/data",
							"/var/www":   "/var/www",
						},
					},
				},
			},
			"googlecompute": {
				templateSection{
					Type: "googlecompute",
					Settings: []string{
						"account_file=account.json",
						"address=ext-static",
						"disk_size=20",
						"image_name=packer-{{timestamp}}",
						"image_description=test image",
						"instance_name=packer-{{uuid}}",
						"machine_type=nl-standard-1",
						"network=default",
						"preemtible=true",
						"project_id=projectID",
						"source_image=centos-6",
						"state_timeout=5m",
						"use_internal_ip=true",
						"zone=us-central1-a",
					},
					Arrays: map[string]interface{}{
						"metadata": map[string]string{
							"key-1": "value-1",
							"key-2": "value-2",
						},
						"tags": []string{
							"tag1",
						},
					},
				},
			},
			"null": {
				templateSection{
					Type:     "null",
					Settings: []string{},
					Arrays:   map[string]interface{}{},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Settings: []string{
						"format = ovf",
						"guest_additions_mode=upload",
						"guest_additions_path=path/to/additions",
						"guest_additions_sha256=89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
						"guest_additions_url=file://guest-additions",
						"guest_os_type=Ubuntu_64",
						"hard_drive_interface=ide",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_interface=ide",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_private_key_file=key/path",
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
						"iso_urls": []string{
							"http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
							"http://2.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
						},
						"vboxmanage": []string{
							"--cpus=1",
							"memory=4096",
						},
						"vboxmanage_post": []string{
							"something=value",
						},
					},
				},
			},
			"virtualbox-ovf": {
				templateSection{
					Type: "virtualbox-ovf",
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
						"ssh_skip_nat_mapping=false",
						"source_path=source.ova",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_private_key_file=key/path",
						"ssh_skip_nat_mapping=true",
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
						"import_flags": []string{
							"--eula-accept",
						},
						"floppy_files": []string{
							"disk1",
						},
						"vboxmanage": []string{
							"cpus=1",
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
					Type: "vmware-iso",
					Settings: []string{
						"communicator=none",
						"disk_type_id=1",
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_target_path=../isocache/",
						"output_directory=out/dir",
						"remote_cache_datastore=datastore1",
						"remote_cache_directory=packer_cache",
						"remote_datastore=datastore1",
						"remote_host=remoteHost",
						"remote_password=rpassword",
						"remote_private_key_file=secret",
						"remote_type=esx5",
						"shutdown_timeout=5m",
						"skip_compaction=true",
						"ssh_host=127.0.0.1",
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
						"disk_additional_size": []string{
							"10000",
						},
						"floppy_files": []string{
							"disk1",
						},
						"iso_urls": []string{
							"http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
							"http://2.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
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
					Type: "vmware-vmx",
					Settings: []string{
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"skip_compaction=false",
						"source_path=source.vmx",
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
		PostProcessorIDs: []string{
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Type: "vagrant",
					Settings: []string{
						"keep_input_artifact = false",
						"output = out/someComposedBoxName.box",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"salt",
		},
		Provisioners: map[string]provisioner{
			"salt": {
				templateSection{
					Type: "salt",
					Settings: []string{
						"local_state_tree = ~/saltstates/centos6/salt",
						"skip_bootstrap = true",
					},
				},
			},
		},
	},
}

// Not all the settings in are valid for winrm, the invalid ones should not be included in the
// output.
var testAllBuildersSSH = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/out",
		SourceDir: "../test_files/src",
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
	Image:   "server",
	Release: "14.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"amazon-ebs",
			"amazon-instance",
			"digitalocean",
			"docker",
			"googlecompute",
			"null",
			"virtualbox-iso",
			"virtualbox-ovf",
			"vmware-iso",
			"vmware-vmx",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
					Settings: []string{
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
						"ssh_bastion_host=bastion.host",
						"ssh_bastion_port=2222",
						"ssh_bastion_username=packer",
						"ssh_bastion_password=packer",
						"ssh_bastion_private_key_file=secret",
						"ssh_disable_agent=true",
						"ssh_handshake_attempts=10",
						"ssh_host=127.0.0.1",
						"ssh_password=vagrant",
						"ssh_port=22",
						"ssh_private_key_file=key/path",
						"ssh_pty=true",
						"ssh_timeout=10m",
						"ssh_username=vagrant",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"amazon-chroot": {
				templateSection{
					Type: "amazon-chroot",
					Settings: []string{
						"access_key=AWS_ACCESS_KEY",
						"ami_description=AMI_DESCRIPTION",
						"ami_name=AMI_NAME",
						"ami_virtualization_type=paravirtual",
						"communicator=ssh",
						"command_wrapper={{.Command}}",
						"device_path=/dev/xvdf",
						"enhanced_networking=false",
						"mount_path=packer-amazon-chroot-volumes/{{.Device}}",
						"secret_key=AWS_SECRET_ACCESS_KEY",
						"source_ami=SOURCE_AMI",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"digitalocean": {
				templateSection{
					Type: "digitalocean",
					Settings: []string{
						"api_token=DIGITALOCEAN_API_TOKEN",
						"communicator=ssh",
						"droplet_name=ocean-drop",
						"image=ubuntu-12-04-x64",
						"private_networking=false",
						"region=nyc3",
						"size=512mb",
						"snapshot_name=my-snapshot",
						"state_timeout=6m",
						"user_data=userdata",
					},
				},
			},
			"docker": {
				templateSection{
					Type: "docker",
					Settings: []string{
						"commit=true",
						"communicator=ssh",
						"discard=false",
						"export_path=export/path",
						"image=baseImage",
						"login=true",
						"login_email=test@test.com",
						"login_username=username",
						"login_password=password",
						"login_server=127.0.0.1",
						"pull=true",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"googlecompute": {
				templateSection{
					Type: "googlecompute",
					Settings: []string{
						"account_file=account.json",
						"address=ext-static",
						"communicator=ssh",
						"disk_size=20",
						"image_name=packer-{{timestamp}}",
						"image_description=test image",
						"instance_name=packer-{{uuid}}",
						"machine_type=nl-standard-1",
						"network=default",
						"preemtible=true",
						"project_id=projectID",
						"source_image=centos-6",
						"state_timeout=5m",
						"use_internal_ip=true",
						"zone=us-central1-a",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"null": {
				templateSection{
					Type: "null",
					Settings: []string{
						"communicator=ssh",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Settings: []string{
						"communicator=ssh",
						"format = ovf",
						"guest_additions_mode=upload",
						"guest_additions_path=path/to/additions",
						"guest_additions_sha256=89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
						"guest_additions_url=file://guest-additions",
						"guest_os_type=Ubuntu_64",
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
						"ssh_private_key_file=key/path",
						"virtualbox_version_file=.vbox_version",
						"vm_name=test-vb-iso",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"virtualbox-ovf": {
				templateSection{
					Type: "virtualbox-ovf",
					Settings: []string{
						"communicator=ssh",
						"format = ovf",
						"guest_additions_mode=upload",
						"guest_additions_path=path/to/additions",
						"guest_additions_sha256=89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
						"guest_additions_url=file://guest-additions",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"import_opts=keepallmacs",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"ssh_private_key_file=key/path",
						"ssh_skip_nat_mapping=false",
						"source_path=source.ova",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_private_key_file=key/path",
						"ssh_skip_nat_mapping=true",
						"virtualbox_version_file=.vbox_version",
						"vm_name=test-vb-ovf",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"vmware-iso": {
				templateSection{
					Type: "vmware-iso",
					Settings: []string{
						"communicator=ssh",
						"disk_type_id=1",
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_target_path=../isocache/",
						"iso_url=http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
						"output_directory=out/dir",
						"remote_cache_datastore=datastore1",
						"remote_cache_directory=packer_cache",
						"remote_datastore=datastore1",
						"remote_host=remoteHost",
						"remote_password=rpassword",
						"remote_private_key_file=secret",
						"remote_type=esx5",
						"shutdown_timeout=5m",
						"skip_compaction=true",
						"ssh_skip_nat_mapping=false",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"tools_upload_flavor=linux",
						"tools_upload_path={{.Flavor}}.iso",
						"version=9",
						"vm_name=packer-BUILDNAME",
						"vmdk_name=packer",
						"vmx_template_path=template/path",
						"vnc_port_min=5900",
						"vnc_port_max=6000",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"vmware-vmx": {
				templateSection{
					Type: "vmware-vmx",
					Settings: []string{
						"communicator=ssh",
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"skip_compaction=false",
						"source_path=source.vmx",
						"vm_name=packer-BUILDNAME",
						"vnc_port_min=5900",
						"vnc_port_max=6000",
					},
					Arrays: map[string]interface{}{},
				},
			},
		},
	},
}

// Not all the settings in are valid for winrm, the invalid ones should not be included in the
// output.
var testAllBuildersWinRM = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/out",
		SourceDir: "../test_files/src",
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
	Image:   "server",
	Release: "14.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"amazon-ebs",
			"amazon-instance",
			"digitalocean",
			"docker",
			"googlecompute",
			"null",
			"virtualbox-iso",
			"virtualbox-ovf",
			"vmware-iso",
			"vmware-vmx",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
					Settings: []string{
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
						"winrm_host=host",
						"winrm_password = vagrant",
						"winrm_port = 22",
						"winrm_username = vagrant",
						"winrm_timeout=10m",
						"winrm_use_ssl=true",
						"winrm_insecure=true",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"amazon-chroot": {
				templateSection{
					Type: "amazon-chroot",
					Settings: []string{
						"access_key=AWS_ACCESS_KEY",
						"ami_description=AMI_DESCRIPTION",
						"ami_name=AMI_NAME",
						"ami_virtualization_type=paravirtual",
						"communicator=winrm",
						"command_wrapper={{.Command}}",
						"device_path=/dev/xvdf",
						"enhanced_networking=false",
						"mount_path=packer-amazon-chroot-volumes/{{.Device}}",
						"secret_key=AWS_SECRET_ACCESS_KEY",
						"source_ami=SOURCE_AMI",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"digitalocean": {
				templateSection{
					Type: "digitalocean",
					Settings: []string{
						"api_token=DIGITALOCEAN_API_TOKEN",
						"communicator=winrm",
						"droplet_name=ocean-drop",
						"image=ubuntu-12-04-x64",
						"private_networking=false",
						"region=nyc3",
						"size=512mb",
						"snapshot_name=my-snapshot",
						"state_timeout=6m",
						"user_data=userdata",
					},
				},
			},
			"docker": {
				templateSection{
					Type: "docker",
					Settings: []string{
						"commit=true",
						"communicator=winrm",
						"discard=false",
						"export_path=export/path",
						"image=baseImage",
						"login=true",
						"login_email=test@test.com",
						"login_username=username",
						"login_password=password",
						"login_server=127.0.0.1",
						"pull=true",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"googlecompute": {
				templateSection{
					Type: "googlecompute",
					Settings: []string{
						"account_file=account.json",
						"address=ext-static",
						"communicator=winrm",
						"disk_size=20",
						"image_name=packer-{{timestamp}}",
						"image_description=test image",
						"instance_name=packer-{{uuid}}",
						"machine_type=nl-standard-1",
						"network=default",
						"preemtible=true",
						"project_id=projectID",
						"source_image=centos-6",
						"state_timeout=5m",
						"use_internal_ip=true",
						"zone=us-central1-a",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"null": {
				templateSection{
					Type: "null",
					Settings: []string{
						"communicator=winrm",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Settings: []string{
						"communicator=winrm",
						"format = ovf",
						"guest_additions_mode=upload",
						"guest_additions_path=path/to/additions",
						"guest_additions_sha256=89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
						"guest_additions_url=file://guest-additions",
						"guest_os_type=Ubuntu_64",
						"hard_drive_interface=ide",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_interface=ide",
						"iso_url=http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"ssh_skip_nat_mapping=false",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_private_key_file=key/path",
						"virtualbox_version_file=.vbox_version",
						"vm_name=test-vb-iso",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"virtualbox-ovf": {
				templateSection{
					Type: "virtualbox-ovf",
					Settings: []string{
						"communicator=winrm",
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
						"source_path=source.ova",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"ssh_private_key_file=key/path",
						"ssh_skip_nat_mapping=true",
						"virtualbox_version_file=.vbox_version",
						"vm_name=test-vb-ovf",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"vmware-iso": {
				templateSection{
					Type: "vmware-iso",
					Settings: []string{
						"communicator=winrm",
						"disk_type_id=1",
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"iso_checksum=ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
						"iso_target_path=../isocache/",
						"iso_url=http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
						"output_directory=out/dir",
						"remote_cache_datastore=datastore1",
						"remote_cache_directory=packer_cache",
						"remote_datastore=datastore1",
						"remote_host=remoteHost",
						"remote_password=rpassword",
						"remote_private_key_file=secret",
						"remote_type=esx5",
						"shutdown_timeout=5m",
						"skip_compaction=true",
						"ssh_skip_nat_mapping=false",
						"ssh_host_port_min=22",
						"ssh_host_port_max=40",
						"tools_upload_flavor=linux",
						"tools_upload_path={{.Flavor}}.iso",
						"version=9",
						"vm_name=packer-BUILDNAME",
						"vmdk_name=packer",
						"vmx_template_path=template/path",
						"vnc_port_min=5900",
						"vnc_port_max=6000",
					},
					Arrays: map[string]interface{}{},
				},
			},
			"vmware-vmx": {
				templateSection{
					Type: "vmware-vmx",
					Settings: []string{
						"communicator=winrm",
						"fusion_app_path=/Applications/VMware Fusion.app",
						"headless=true",
						"http_port_min=8000",
						"http_port_max=9000",
						"output_directory=out/dir",
						"shutdown_timeout=5m",
						"skip_compaction=false",
						"source_path=source.vmx",
						"vm_name=packer-BUILDNAME",
						"vnc_port_min=5900",
						"vnc_port_max=6000",
					},
					Arrays: map[string]interface{}{},
				},
			},
		},
	},
}

var testDockerRunComandFile = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/out",
		SourceDir: "../test_files/src",
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
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"docker",
		},
		Builders: map[string]builder{
			"docker": {
				templateSection{
					Settings: []string{
						"commit=true",
						"discard=false",
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
var testDockerRunComand = rawTemplate{
	IODirInf: IODirInf{
		OutputDir: "../test_files/out",
		SourceDir: "../test_files/src",
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
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"docker",
		},
		Builders: map[string]builder{
			"docker": {
				templateSection{
					Settings: []string{
						"commit=true",
						"discard=false",
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
var builderOrig = map[string]builder{
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
				"ssh_timeout = 30m",
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

var builderNew = map[string]builder{
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
				"ssh_timeout = 240m",
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

var builderMerged = map[string]builder{
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
				"ssh_timeout = 240m",
			},
			Arrays: map[string]interface{}{},
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

var vbB = builder{
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

func init() {
	b := true
	testAllBuilders.IncludeComponentString = &b
	testAllBuildersSSH.IncludeComponentString = &b
	testAllBuildersWinRM.IncludeComponentString = &b
}

func TestCreateBuilders(t *testing.T) {
	_, err := testRawTemplateBuilderOnly.createBuilders()
	if err == nil {
		t.Error("Expected error \"unable to create builders: none specified\", got nil")
	} else {
		if err.Error() != "unable to create builders: none specified" {
			t.Errorf("Expected \"unable to create builders: none specified\", got %q", err)
		}
	}

	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for amazon-ebs not found\", got nil")
	} else {
		if err.Error() != "builder configuration for amazon-ebs not found" {
			t.Errorf("Expected \"builder configuration for amazon-ebs not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "digitalocean"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for digitalocean not found\", got nil")
	} else {
		if err.Error() != "builder configuration for digitalocean not found" {
			t.Errorf("Expected \"builder configuration for digitalocean not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "docker"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for docker not found\", got nil")
	} else {
		if err.Error() != "builder configuration for docker not found" {
			t.Errorf("Expected \"builder configuration for docker not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "googlecompute"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for googlecompute not found\", got nil")
	} else {
		if err.Error() != "builder configuration for googlecompute not found" {
			t.Errorf("Expected \"builder configuration for googlecompute not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "virtualbox-iso"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for virtualbox-iso not found\", got nil")
	} else {
		if err.Error() != "builder configuration for virtualbox-iso not found" {
			t.Errorf("Expected \"builder configuration for virtualbox-iso not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "virtualbox-ovf"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for virtualbox-ovf not found\", got nil")
	} else {
		if err.Error() != "builder configuration for virtualbox-ovf not found" {
			t.Errorf("Expected \"builder configuration for virtualbox-ovf not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "vmware-iso"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for vmware-iso not found\", got nil")
	} else {
		if err.Error() != "builder configuration for vmware-iso not found" {
			t.Errorf("Expected \"builder configuration for vmware-iso not found\", got %q", err)
		}
	}

	testRawTemplateWOSection.build.BuilderIDs[0] = "vmware-vmx"
	_, err = testRawTemplateWOSection.createBuilders()
	if err == nil {
		t.Error("Expected \"builder configuration for vmware-vmx not found\", got nil")
	} else {
		if err.Error() != "builder configuration for vmware-vmx not found" {
			t.Errorf("Expected \"builder configuration for vmware-vmx not found\", got %q", err)
		}
	}

	r := testDistroDefaultUbuntu
	r.BuilderIDs = nil
	_, err = r.createBuilders()
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "unable to create builders: none specified" {
			t.Errorf("Expected \"unable to create builders: none specified\"), got %q", err)
		}
	}
}

func TestRawTemplateUpdatebuilders(t *testing.T) {
	err := testUbuntu.updateBuilders(nil)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
	if MarshalJSONToString.Get(testUbuntu.Builders) != MarshalJSONToString.Get(builderOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(builderOrig), MarshalJSONToString.Get(testUbuntu.Builders))
	}

	err = testUbuntu.updateBuilders(builderNew)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
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

func TestCreateAmazonChroot(t *testing.T) {
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
		"ami_users": []string{
			"aws-account-1",
		},
		"ami_virtualization_type": "paravirtual",
		"chroot_mounts": []interface{}{
			[]string{
				"proc",
				"proc",
				"/proc",
			},
			[]string{
				"bind",
				"/dev",
				"/dev",
			},
		},
		"command_wrapper": "{{.Command}}",
		"copy_files": []string{
			"/etc/resolv.conf",
		},
		"device_path":         "/dev/xvdf",
		"enhanced_networking": false,
		"mount_path":          "packer-amazon-chroot-volumes/{{.Device}}",
		"secret_key":          "AWS_SECRET_ACCESS_KEY",
		"source_ami":          "SOURCE_AMI",
		"tags": map[string]string{
			"OS_Version": "Ubuntu",
			"Release":    "Latest",
		},
		"type": "amazon-chroot",
	}
	bldr, err := testAllBuilders.createAmazonChroot("amazon-chroot")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
	// SSH
	expectedSSH := map[string]interface{}{
		"access_key":                   "AWS_ACCESS_KEY",
		"ami_description":              "AMI_DESCRIPTION",
		"ami_name":                     "AMI_NAME",
		"ami_virtualization_type":      "paravirtual",
		"command_wrapper":              "{{.Command}}",
		"communicator":                 "ssh",
		"device_path":                  "/dev/xvdf",
		"enhanced_networking":          false,
		"mount_path":                   "packer-amazon-chroot-volumes/{{.Device}}",
		"secret_key":                   "AWS_SECRET_ACCESS_KEY",
		"source_ami":                   "SOURCE_AMI",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"type":                         "amazon-chroot",
	}
	bldr, err = testAllBuildersSSH.createAmazonChroot("amazon-chroot")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(bldr))
		}
	}
	// WinRM
	expectedWinRM := map[string]interface{}{
		"access_key":              "AWS_ACCESS_KEY",
		"ami_description":         "AMI_DESCRIPTION",
		"ami_name":                "AMI_NAME",
		"ami_virtualization_type": "paravirtual",
		"command_wrapper":         "{{.Command}}",
		"communicator":            "winrm",
		"device_path":             "/dev/xvdf",
		"enhanced_networking":     false,
		"mount_path":              "packer-amazon-chroot-volumes/{{.Device}}",
		"secret_key":              "AWS_SECRET_ACCESS_KEY",
		"source_ami":              "SOURCE_AMI",
		"type":                    "amazon-chroot",
		"winrm_host":              "host",
		"winrm_password":          "vagrant",
		"winrm_port":              22,
		"winrm_timeout":           "10m",
		"winrm_username":          "vagrant",
		"winrm_use_ssl":           true,
		"winrm_insecure":          true,
	}
	bldr, err = testAllBuildersWinRM.createAmazonChroot("amazon-chroot")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(bldr))
		}
	}
}

/*
func TestCreateAmazonEBS(t *testing.T) {
	expected := map[string]interface{}{
		"access_key": "AWS_ACCESS_KEY",
		"ami_block_device_mappings": []map[string]string{
			{
				"device_name":  "/dev/sdb",
				"virtual_name": "/ephemeral0",
			},
			{
				"device_name":  "/dev/sdc",
				"virtual_name": "/ephemeral1",
			},
		},

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
		"ami_users": []string{
			"ami-account",
		},
		"associate_public_ip_address": false,
		"availability_zone":           "us-east-1b",
		"enhanced_networking":         false,
		"iam_instance_profile":        "INSTANCE_PROFILE",
		"instance_type":               "m3.medium",
		"launch_block_device_mappings": []map[string]string{
			{
				"device_name":  "/dev/sdd",
				"virtual_name": "/ephemeral2",
			},
			{
				"device_name":  "/dev/sde",
				"virtual_name": "/ephemeral3",
			},
		},
		"region": "us-east-1",
		"run_tags": map[string]string{
			"foo": "bar",
			"fiz": "baz",
		},
		"secret_key":        "AWS_SECRET_ACCESS_KEY",
		"security_group_id": "GROUP_ID",
		"security_group_ids": []string{
			"SECURITY_GROUP",
		},
		"source_ami":              "SOURCE_AMI",
		"spot_price":              "auto",
		"spot_price_auto_product": "Linux/Unix",
		"ssh_port":                22,
		"ssh_username":            "vagrant",
		"ssh_private_key_file":    "myKey",
		"tags": map[string]string{
			"OS_Version": "Ubuntu",
			"Release":    "Latest",
		},
		"temporary_key_pair_name": "TMP_KEYPAIR",
		"token":                   "AWS_SECURITY_TOKEN",
		"type":                    "amazon-ebs",
		"user_data":               "SOME_USER_DATA",
		"user_data_file":          "amazon-ebs/amazon.userdata",
		"vpc_id":                  "VPC_ID",
	}
	bldr, err := testAllBuilders.createAmazonEBS("amazon-ebs")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
}
*/
func TestCreateAmazonInstance(t *testing.T) {
	expected := map[string]interface{}{
		"access_key":      "AWS_ACCESS_KEY",
		"account_id":      "YOUR_ACCOUNT_ID",
		"ami_description": "AMI_DESCRIPTION",
		"ami_block_device_mappings": []map[string]interface{}{
			{
				"delete_on_termination": true,
				"device_name":           "/dev/sdb",
				"encrypted":             true,
				"iops":                  1000,
				"no_device":             false,
				"snapshot_id":           "SNAPSHOT",
				"virtual_name":          "ephemeral0",
				"volume_type":           "io1",
				"volume_size":           10,
			},
		},
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
		"ami_users": []string{
			"ami-account",
		},
		"ami_virtualization_type":     "paravirtual",
		"associate_public_ip_address": false,
		"availability_zone":           "us-east-1b",
		"bundle_destination":          "/tmp",
		"bundle_prefix":               "image--{{timestamp}}",
		"bundle_upload_command":       "sudo -n ec2-bundle-vol -k {{.KeyPath}} -u {{.AccountId}} -c {{.CertPath}} -r {{.Architecture}} -e {{.PrivatePath}} -d {{.Destination}} -p {{.Prefix}} --batch --no-filter",
		"bundle_vol_command":          "sudo -n ec2-upload-bundle -b {{.BucketName}} -m {{.ManifestPath}} -a {{.AccessKey}} -s {{.SecretKey}} -d {{.BundleDirectory}} --batch --region {{.Region}} --retry",
		"ebs_optimized":               true,
		"enhanced_networking":         false,
		"force_deregister":            false,
		"iam_instance_profile":        "INSTANCE_PROFILE",
		"instance_type":               "m3.medium",
		"launch_block_device_mappings": []map[string]string{
			{
				"device_name":  "/dev/sdd",
				"virtual_name": "/ephemeral2",
			},
			{
				"device_name":  "/dev/sde",
				"virtual_name": "/ephemeral3",
			},
		},
		"region": "us-east-1",
		"run_tags": map[string]string{
			"foo": "bar",
			"fiz": "baz",
		},
		"s3_bucket":         "packer_bucket",
		"secret_key":        "AWS_SECRET_ACCESS_KEY",
		"security_group_id": "GROUP_ID",
		"security_group_ids": []string{
			"SECURITY_GROUP",
		},
		"source_ami":              "SOURCE_AMI",
		"spot_price":              "auto",
		"spot_price_auto_product": "Linux/Unix",
		"ssh_keypair_name":        "myKeyPair",
		"ssh_private_ip":          true,
		"ssh_private_key_file":    "myKey",
		"ssh_username":            "vagrant",
		"subnet_id":               "subnet-12345def",
		"temporary_key_pair_name": "TMP_KEYPAIR",
		"tags": map[string]string{
			"OS_Version": "Ubuntu",
			"Release":    "Latest",
		},
		"type":                     "amazon-instance",
		"user_data":                "SOME_USER_DATA",
		"user_data_file":           "amazon-instance/amazon.userdata",
		"vpc_id":                   "VPC_ID",
		"windows_password_timeout": "10m",
		"x509_cert_path":           "/path/to/x509/cert",
		"x509_key_path":            "/path/to/x509/key",
		"x509_upload_path":         "/etc/x509",
	}
	contour.UpdateString("source_dir", "../test_files/src")
	bldr, err := testAllBuilders.createAmazonInstance("amazon-instance")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestCreateDigitalOcean(t *testing.T) {
	expected := map[string]interface{}{
		"api_token":          "DIGITALOCEAN_API_TOKEN",
		"droplet_name":       "ocean-drop",
		"image":              "ubuntu-12-04-x64",
		"private_networking": false,
		"region":             "nyc3",
		"size":               "512mb",
		"snapshot_name":      "my-snapshot",
		"state_timeout":      "6m",
		"type":               "digitalocean",
		"user_data":          "userdata",
	}
	bldr, err := testAllBuilders.createDigitalOcean("digitalocean")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
	// SSH
	expectedSSH := map[string]interface{}{
		"api_token":                    "DIGITALOCEAN_API_TOKEN",
		"communicator":                 "ssh",
		"droplet_name":                 "ocean-drop",
		"image":                        "ubuntu-12-04-x64",
		"private_networking":           false,
		"region":                       "nyc3",
		"size":                         "512mb",
		"snapshot_name":                "my-snapshot",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"state_timeout":                "6m",
		"type":                         "digitalocean",
		"user_data":                    "userdata",
	}
	bldr, err = testAllBuildersSSH.createDigitalOcean("digitalocean")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(bldr))
		}
	}

	// WinRM
	expectedWinRM := map[string]interface{}{
		"api_token":          "DIGITALOCEAN_API_TOKEN",
		"communicator":       "winrm",
		"droplet_name":       "ocean-drop",
		"image":              "ubuntu-12-04-x64",
		"private_networking": false,
		"region":             "nyc3",
		"size":               "512mb",
		"snapshot_name":      "my-snapshot",
		"state_timeout":      "6m",
		"type":               "digitalocean",
		"user_data":          "userdata",
		"winrm_host":         "host",
		"winrm_password":     "vagrant",
		"winrm_port":         22,
		"winrm_timeout":      "10m",
		"winrm_username":     "vagrant",
		"winrm_use_ssl":      true,
		"winrm_insecure":     true,
	}
	bldr, err = testAllBuildersWinRM.createDigitalOcean("digitalocean")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestCreateDocker(t *testing.T) {
	expected := map[string]interface{}{
		"commit":         true,
		"discard":        false,
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
		"volumes": map[string]string{
			"/var/data1": "/var/data",
			"/var/www":   "/var/www",
		},
	}
	expectedCommand := map[string]interface{}{
		"commit":         true,
		"discard":        false,
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
		"discard":        false,
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
	bldr, err := testAllBuilders.createDocker("docker")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
	bldr, err = testDockerRunComandFile.createDocker("docker")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedCommandFile) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedCommandFile), MarshalJSONToString.Get(bldr))
		}
	}
	bldr, err = testDockerRunComand.createDocker("docker")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedCommand) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedCommand), MarshalJSONToString.Get(bldr))
		}
	}
	expectedSSH := map[string]interface{}{
		"commit":                       true,
		"communicator":                 "ssh",
		"discard":                      false,
		"export_path":                  "export/path",
		"image":                        "baseImage",
		"login":                        true,
		"login_email":                  "test@test.com",
		"login_username":               "username",
		"login_password":               "password",
		"login_server":                 "127.0.0.1",
		"pull":                         true,
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"type":                         "docker",
	}
	bldr, err = testAllBuildersSSH.createDocker("docker")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(bldr))
		}
	}
	expectedWinRM := map[string]interface{}{
		"commit":         true,
		"communicator":   "winrm",
		"discard":        false,
		"export_path":    "export/path",
		"image":          "baseImage",
		"login":          true,
		"login_email":    "test@test.com",
		"login_username": "username",
		"login_password": "password",
		"login_server":   "127.0.0.1",
		"pull":           true,
		"winrm_host":     "host",
		"winrm_password": "vagrant",
		"winrm_port":     22,
		"winrm_timeout":  "10m",
		"winrm_username": "vagrant",
		"winrm_use_ssl":  true,
		"winrm_insecure": true,
		"type":           "docker",
	}
	bldr, err = testAllBuildersWinRM.createDocker("docker")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestCreateGoogleCompute(t *testing.T) {
	expected := map[string]interface{}{
		"account_file":      "googlecompute/account.json",
		"address":           "ext-static",
		"disk_size":         20,
		"image_name":        "packer-{{timestamp}}",
		"image_description": "test image",
		"instance_name":     "packer-{{uuid}}",
		"machine_type":      "nl-standard-1",
		"metadata": map[string]string{
			"key-1": "value-1",
			"key-2": "value-2",
		},
		"network":       "default",
		"preemtible":    true,
		"project_id":    "projectID",
		"source_image":  "centos-6",
		"state_timeout": "5m",
		"tags": []string{
			"tag1",
		},
		"type":            "googlecompute",
		"use_internal_ip": true,
		"zone":            "us-central1-a",
	}

	bldr, err := testAllBuilders.createGoogleCompute("googlecompute")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(bldr))
		}
	}
	// ssh
	expectedSSH := map[string]interface{}{
		"account_file":                 "googlecompute/account.json",
		"address":                      "ext-static",
		"communicator":                 "ssh",
		"disk_size":                    20,
		"image_name":                   "packer-{{timestamp}}",
		"image_description":            "test image",
		"instance_name":                "packer-{{uuid}}",
		"machine_type":                 "nl-standard-1",
		"network":                      "default",
		"preemtible":                   true,
		"project_id":                   "projectID",
		"source_image":                 "centos-6",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"state_timeout":                "5m",
		"type":                         "googlecompute",
		"use_internal_ip":              true,
		"zone":                         "us-central1-a",
	}

	bldr, err = testAllBuildersSSH.createGoogleCompute("googlecompute")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(bldr))
		}
	}

	expectedWinRM := map[string]interface{}{
		"account_file":      "googlecompute/account.json",
		"address":           "ext-static",
		"communicator":      "winrm",
		"disk_size":         20,
		"image_name":        "packer-{{timestamp}}",
		"image_description": "test image",
		"instance_name":     "packer-{{uuid}}",
		"machine_type":      "nl-standard-1",
		"network":           "default",
		"preemtible":        true,
		"project_id":        "projectID",
		"source_image":      "centos-6",
		"state_timeout":     "5m",
		"type":              "googlecompute",
		"use_internal_ip":   true,
		"winrm_host":        "host",
		"winrm_password":    "vagrant",
		"winrm_port":        22,
		"winrm_timeout":     "10m",
		"winrm_username":    "vagrant",
		"winrm_use_ssl":     true,
		"winrm_insecure":    true,
		"zone":              "us-central1-a",
	}

	bldr, err = testAllBuildersWinRM.createGoogleCompute("googlecompute")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(bldr))
		}
	}
}

func TestBuilderNull(t *testing.T) {
	// a communicator of none or no communicator setting should result in an error
	expected := "null: null builder requires a communicator other than \"none\""
	_, err := testAllBuilders.createNull("null")
	if err == nil {
		t.Errorf("expected an error, got none")
	} else {
		if err.Error() != expected {
			t.Errorf("got %q, want %q", err, expected)
		}
	}
	// ssh
	expectedSSH := map[string]interface{}{
		"communicator":                 "ssh",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"type":                         "null",
	}
	bldr, err := testAllBuildersSSH.createNull("null")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(bldr))
		}
	}
	// winrm
	expectedWinRM := map[string]interface{}{
		"communicator":   "winrm",
		"winrm_host":     "host",
		"winrm_password": "vagrant",
		"winrm_port":     22,
		"winrm_timeout":  "10m",
		"winrm_username": "vagrant",
		"winrm_use_ssl":  true,
		"winrm_insecure": true,
		"type":           "null",
	}
	bldr, err = testAllBuildersWinRM.createNull("null")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(bldr) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(bldr))
		}
	}

}
func TestCreateVirtualboxISO(t *testing.T) {
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
		"iso_urls": []string{
			"http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
			"http://2.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		},
		"output_directory":  "out/dir",
		"shutdown_command":  "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":  "5m",
		"ssh_host_port_max": 40,
		"ssh_host_port_min": 22,
		"ssh_password":      "vagrant",
		"ssh_username":      "vagrant",
		"type":              "virtualbox-iso",
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
				"--something",
				"value",
			},
		},
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-iso",
	}
	testAllBuilders.BaseURL = "http://releases.ubuntu.com/"
	settings, err := testAllBuilders.createVirtualBoxISO("virtualbox-iso")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
	// ssh
	expectedSSH := map[string]interface{}{
		"boot_wait":                    "5s",
		"communicator":                 "ssh",
		"disk_size":                    20000,
		"format":                       "ovf",
		"guest_additions_mode":         "upload",
		"guest_additions_path":         "path/to/additions",
		"guest_additions_sha256":       "89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
		"guest_additions_url":          "file://guest-additions",
		"guest_os_type":                "Ubuntu_64",
		"hard_drive_interface":         "ide",
		"headless":                     true,
		"http_directory":               "http",
		"http_port_max":                9000,
		"http_port_min":                8000,
		"iso_checksum":                 "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type":            "sha256",
		"iso_interface":                "ide",
		"iso_url":                      "http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		"output_directory":             "out/dir",
		"shutdown_command":             "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":             "5m",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host_port_max":            40,
		"ssh_host_port_min":            22,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"type":                         "virtualbox-iso",
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-iso",
	}
	testAllBuildersSSH.BaseURL = "http://releases.ubuntu.com/"
	settings, err = testAllBuildersSSH.createVirtualBoxISO("virtualbox-iso")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(settings))
		}
	}

	// winrm communicator
	expectedWinRM := map[string]interface{}{
		"boot_wait":               "5s",
		"communicator":            "winrm",
		"disk_size":               20000,
		"format":                  "ovf",
		"guest_additions_mode":    "upload",
		"guest_additions_path":    "path/to/additions",
		"guest_additions_sha256":  "89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
		"guest_additions_url":     "file://guest-additions",
		"guest_os_type":           "Ubuntu_64",
		"hard_drive_interface":    "ide",
		"headless":                true,
		"http_directory":          "http",
		"http_port_max":           9000,
		"http_port_min":           8000,
		"iso_checksum":            "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type":       "sha256",
		"iso_interface":           "ide",
		"iso_url":                 "http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		"output_directory":        "out/dir",
		"shutdown_command":        "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":        "5m",
		"type":                    "virtualbox-iso",
		"winrm_host":              "host",
		"winrm_password":          "vagrant",
		"winrm_port":              22,
		"winrm_timeout":           "10m",
		"winrm_username":          "vagrant",
		"winrm_use_ssl":           true,
		"winrm_insecure":          true,
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-iso",
	}
	testAllBuildersWinRM.BaseURL = "http://releases.ubuntu.com/"
	settings, err = testAllBuildersWinRM.createVirtualBoxISO("virtualbox-iso")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(settings))
		}
	}
}

func TestCreateVirtualboxOVF(t *testing.T) {
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
		"import_flags": []string{
			"--eula-accept",
		},
		"import_opts":          "keepallmacs",
		"output_directory":     "out/dir",
		"shutdown_command":     "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":     "5m",
		"source_path":          "virtualbox-ovf/source.ova",
		"ssh_host_port_max":    40,
		"ssh_host_port_min":    22,
		"ssh_skip_nat_mapping": true,
		"ssh_username":         "vagrant",
		"type":                 "virtualbox-ovf",
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
				"--something",
				"value",
			},
		},
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-ovf",
	}
	testAllBuilders.files = make(map[string]string)
	settings, err := testAllBuilders.createVirtualBoxOVF("virtualbox-ovf")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
	// ssh
	expectedSSH := map[string]interface{}{
		"boot_wait":                    "5s",
		"communicator":                 "ssh",
		"format":                       "ovf",
		"guest_additions_mode":         "upload",
		"guest_additions_path":         "path/to/additions",
		"guest_additions_sha256":       "89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
		"guest_additions_url":          "file://guest-additions",
		"headless":                     true,
		"http_directory":               "http",
		"http_port_max":                9000,
		"http_port_min":                8000,
		"import_opts":                  "keepallmacs",
		"output_directory":             "out/dir",
		"shutdown_command":             "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":             "5m",
		"source_path":                  "virtualbox-ovf/source.ova",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host_port_max":            40,
		"ssh_host_port_min":            22,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_skip_nat_mapping":         true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"type":                         "virtualbox-ovf",
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-ovf",
	}
	testAllBuildersSSH.BaseURL = "http://releases.ubuntu.com/"
	settings, err = testAllBuildersSSH.createVirtualBoxOVF("virtualbox-ovf")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(settings))
		}
	}

	// winrm communicator
	expectedWinRM := map[string]interface{}{
		"boot_wait":               "5s",
		"communicator":            "winrm",
		"format":                  "ovf",
		"guest_additions_mode":    "upload",
		"guest_additions_path":    "path/to/additions",
		"guest_additions_sha256":  "89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597",
		"guest_additions_url":     "file://guest-additions",
		"headless":                true,
		"http_directory":          "http",
		"http_port_max":           9000,
		"http_port_min":           8000,
		"import_opts":             "keepallmacs",
		"output_directory":        "out/dir",
		"shutdown_command":        "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":        "5m",
		"source_path":             "virtualbox-ovf/source.ova",
		"type":                    "virtualbox-ovf",
		"winrm_host":              "host",
		"winrm_password":          "vagrant",
		"winrm_port":              22,
		"winrm_timeout":           "10m",
		"winrm_username":          "vagrant",
		"winrm_use_ssl":           true,
		"winrm_insecure":          true,
		"virtualbox_version_file": ".vbox_version",
		"vm_name":                 "test-vb-ovf",
	}
	testAllBuildersWinRM.BaseURL = "http://releases.ubuntu.com/"
	settings, err = testAllBuildersWinRM.createVirtualBoxOVF("virtualbox-ovf")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(settings))
		}
	}
}

func TestCreateVMWareISO(t *testing.T) {
	expected := map[string]interface{}{
		"boot_command": []string{
			"<bs>",
			"<del>",
			"<enter><return>",
			"<esc>",
		},
		"boot_wait":    "5s",
		"communicator": "none",
		"disk_additional_size": []int{
			10000,
		},
		"disk_size":    20000,
		"disk_type_id": "1",
		"floppy_files": []string{
			"disk1",
		},
		"fusion_app_path":   "/Applications/VMware Fusion.app",
		"guest_os_type":     "Ubuntu_64",
		"headless":          true,
		"http_directory":    "http",
		"http_port_max":     9000,
		"http_port_min":     8000,
		"iso_checksum":      "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type": "sha256",
		"iso_target_path":   "../isocache/",
		"iso_urls": []string{
			"http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
			"http://2.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		},
		"output_directory":        "out/dir",
		"remote_cache_datastore":  "datastore1",
		"remote_cache_directory":  "packer_cache",
		"remote_datastore":        "datastore1",
		"remote_host":             "remoteHost",
		"remote_password":         "rpassword",
		"remote_private_key_file": "secret",
		"remote_type":             "esx5",
		"shutdown_command":        "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":        "5m",
		"skip_compaction":         true,
		"ssh_username":            "vagrant",
		"tools_upload_flavor":     "linux",
		"tools_upload_path":       "{{.Flavor}}.iso",
		"type":                    "vmware-iso",
		"version":                 "9",
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

	testAllBuilders.BaseURL = "http://releases.ubuntu.com/"
	settings, err := testAllBuilders.createVMWareISO("vmware-iso")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}
	// SSH
	expectedSSH := map[string]interface{}{
		"boot_wait":                    "5s",
		"communicator":                 "ssh",
		"disk_size":                    20000,
		"disk_type_id":                 "1",
		"fusion_app_path":              "/Applications/VMware Fusion.app",
		"guest_os_type":                "Ubuntu_64",
		"headless":                     true,
		"http_directory":               "http",
		"http_port_max":                9000,
		"http_port_min":                8000,
		"iso_checksum":                 "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type":            "sha256",
		"iso_target_path":              "../isocache/",
		"iso_url":                      "http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		"output_directory":             "out/dir",
		"remote_cache_datastore":       "datastore1",
		"remote_cache_directory":       "packer_cache",
		"remote_datastore":             "datastore1",
		"remote_host":                  "remoteHost",
		"remote_password":              "rpassword",
		"remote_private_key_file":      "secret",
		"remote_type":                  "esx5",
		"shutdown_command":             "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":             "5m",
		"skip_compaction":              true,
		"tools_upload_flavor":          "linux",
		"tools_upload_path":            "{{.Flavor}}.iso",
		"type":                         "vmware-iso",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"version":                      "9",
		"vm_name":                      "packer-BUILDNAME",
		"vmdk_name":                    "packer",
		"vmx_template_path":            "template/path",
		"vnc_port_max":                 6000,
		"vnc_port_min":                 5900,
	}

	testAllBuildersSSH.BaseURL = "http://releases.ubuntu.com/"
	settings, err = testAllBuildersSSH.createVMWareISO("vmware-iso")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(settings))
		}
	}
	// winrm
	expectedWinRM := map[string]interface{}{
		"boot_wait":               "5s",
		"communicator":            "winrm",
		"disk_size":               20000,
		"disk_type_id":            "1",
		"fusion_app_path":         "/Applications/VMware Fusion.app",
		"guest_os_type":           "Ubuntu_64",
		"headless":                true,
		"http_directory":          "http",
		"http_port_max":           9000,
		"http_port_min":           8000,
		"iso_checksum":            "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388",
		"iso_checksum_type":       "sha256",
		"iso_target_path":         "../isocache/",
		"iso_url":                 "http://releases.ubuntu.com/14.04/ubuntu-14.04.1-server-amd64.iso",
		"output_directory":        "out/dir",
		"remote_cache_datastore":  "datastore1",
		"remote_cache_directory":  "packer_cache",
		"remote_datastore":        "datastore1",
		"remote_host":             "remoteHost",
		"remote_password":         "rpassword",
		"remote_private_key_file": "secret",
		"remote_type":             "esx5",
		"shutdown_command":        "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":        "5m",
		"skip_compaction":         true,
		"tools_upload_flavor":     "linux",
		"tools_upload_path":       "{{.Flavor}}.iso",
		"type":                    "vmware-iso",
		"version":                 "9",
		"vm_name":                 "packer-BUILDNAME",
		"vmdk_name":               "packer",
		"vmx_template_path":       "template/path",
		"vnc_port_max":            6000,
		"vnc_port_min":            5900,
		"winrm_host":              "host",
		"winrm_password":          "vagrant",
		"winrm_port":              22,
		"winrm_timeout":           "10m",
		"winrm_username":          "vagrant",
		"winrm_use_ssl":           true,
		"winrm_insecure":          true,
	}

	testAllBuildersWinRM.BaseURL = "http://releases.ubuntu.com/"
	settings, err = testAllBuildersWinRM.createVMWareISO("vmware-iso")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(settings))
		}
	}
}

func TestCreateVMWareVMX(t *testing.T) {
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
		"fusion_app_path":  "/Applications/VMware Fusion.app",
		"headless":         true,
		"http_directory":   "http",
		"http_port_max":    9000,
		"http_port_min":    8000,
		"output_directory": "out/dir",
		"shutdown_command": "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout": "5m",
		"skip_compaction":  false,
		"source_path":      "vmware-vmx/source.vmx",
		"ssh_username":     "vagrant",
		"type":             "vmware-vmx",
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

	settings, err := testAllBuilders.createVMWareVMX("vmware-vmx")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings))
		}
	}

	expectedSSH := map[string]interface{}{
		"boot_wait":                    "5s",
		"communicator":                 "ssh",
		"fusion_app_path":              "/Applications/VMware Fusion.app",
		"headless":                     true,
		"http_directory":               "http",
		"http_port_max":                9000,
		"http_port_min":                8000,
		"output_directory":             "out/dir",
		"shutdown_command":             "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout":             "5m",
		"skip_compaction":              false,
		"source_path":                  "vmware-vmx/source.vmx",
		"ssh_bastion_host":             "bastion.host",
		"ssh_bastion_port":             2222,
		"ssh_bastion_username":         "packer",
		"ssh_bastion_password":         "packer",
		"ssh_bastion_private_key_file": "secret",
		"ssh_disable_agent":            true,
		"ssh_handshake_attempts":       10,
		"ssh_host":                     "127.0.0.1",
		"ssh_password":                 "vagrant",
		"ssh_port":                     22,
		"ssh_private_key_file":         "key/path",
		"ssh_pty":                      true,
		"ssh_username":                 "vagrant",
		"ssh_timeout":                  "10m",
		"type":                         "vmware-vmx",
		"vm_name":                      "packer-BUILDNAME",
		"vnc_port_max":                 6000,
		"vnc_port_min":                 5900,
	}

	settings, err = testAllBuildersSSH.createVMWareVMX("vmware-vmx")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedSSH) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedSSH), MarshalJSONToString.Get(settings))
		}
	}
	// WinRM
	expectedWinRM := map[string]interface{}{
		"boot_wait":        "5s",
		"communicator":     "winrm",
		"fusion_app_path":  "/Applications/VMware Fusion.app",
		"headless":         true,
		"http_directory":   "http",
		"http_port_max":    9000,
		"http_port_min":    8000,
		"output_directory": "out/dir",
		"shutdown_command": "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'",
		"shutdown_timeout": "5m",
		"skip_compaction":  false,
		"source_path":      "vmware-vmx/source.vmx",
		"type":             "vmware-vmx",
		"vm_name":          "packer-BUILDNAME",
		"vnc_port_max":     6000,
		"vnc_port_min":     5900,
		"winrm_host":       "host",
		"winrm_password":   "vagrant",
		"winrm_port":       22,
		"winrm_timeout":    "10m",
		"winrm_username":   "vagrant",
		"winrm_use_ssl":    true,
		"winrm_insecure":   true,
	}

	settings, err = testAllBuildersWinRM.createVMWareVMX("vmware-vmx")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(settings) != MarshalJSONToString.Get(expectedWinRM) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expectedWinRM), MarshalJSONToString.Get(settings))
		}
	}
}

func TestDeepCopyMapStringBuilder(t *testing.T) {
	cpy := DeepCopyMapStringBuilder(testDistroDefaults.Templates[Ubuntu].Builders)
	if MarshalJSONToString.Get(cpy["common"]) != MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu].Builders["common"]) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu].Builders["common"]), MarshalJSONToString.Get(cpy["common"]))
	}
}

func TestProcessAMIBlockDeviceMappings(t *testing.T) {
	r := newRawTemplate()
	_, err := r.processAMIBlockDeviceMappings([]string{})
	if err == nil {
		t.Errorf("expected error, got none")
	} else {
		expected := "not in a supported format"
		if err.Error() != expected {
			t.Errorf("got %q; expected %q", err, expected)
		}
	}

	expected := []map[string]interface{}{
		{
			"delete_on_termination": true,
			"device_name":           "/dev/sdb",
			"encrypted":             true,
			"iops":                  1000,
			"no_device":             false,
			"snapshot_id":           "SNAPSHOT",
			"virtual_name":          "/ephemeral0",
			"volume_type":           "io1",
			"volume_size":           20,
		},
		{
			"device_name":  "/dev/sdc",
			"iops":         500,
			"virtual_name": "/ephemeral1",
			"volume_type":  "io1",
			"volume_size":  10,
		},
	}
	// test using array of block mappings: []map[string]interface{}
	mappings := []map[string]interface{}{
		{
			"delete_on_termination": true,
			"device_name":           "/dev/sdb",
			"encrypted":             true,
			"iops":                  1000,
			"no_device":             false,
			"snapshot_id":           "SNAPSHOT",
			"virtual_name":          "/ephemeral0",
			"volume_type":           "io1",
			"volume_size":           20,
		},
		{
			"device_name":  "/dev/sdc",
			"iops":         500,
			"virtual_name": "/ephemeral1",
			"volume_type":  "io1",
			"volume_size":  10,
		},
	}
	// expected is the same as mappings
	ret, err := r.processAMIBlockDeviceMappings(mappings)
	if err != nil {
		t.Errorf("got %q, expected no error", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(ret) {
			t.Errorf("Got %s; want %s", MarshalJSONToString.Get(ret), MarshalJSONToString.Get(expected))
		}
	}
	// test using array of block mappings: [][]string
	mappingsSlice := [][]string{
		[]string{
			"delete_on_termination=true",
			"device_name=/dev/sdb",
			"encrypted=true",
			"iops=1000",
			"no_device=false",
			"snapshot_id=SNAPSHOT",
			"virtual_name=/ephemeral0",
			"volume_type=io1",
			"volume_size=20",
		},
		[]string{
			"device_name=/dev/sdc",
			"iops=500",
			"virtual_name=/ephemeral1",
			"volume_type=io1",
			"volume_size=10",
		},
	}
	ret, err = r.processAMIBlockDeviceMappings(mappingsSlice)
	if err != nil {
		t.Errorf("got %q, expected no error", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(ret) {
			t.Errorf("Got %s; want %s", MarshalJSONToString.Get(ret), MarshalJSONToString.Get(expected))
		}
	}
}
