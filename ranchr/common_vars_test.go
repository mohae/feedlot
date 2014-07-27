package ranchr

import (
	"time"

	json "github.com/mohae/customjson"
)

// Variables that are used in various tests, so they aren't scattered every-
// where. If a variable is only used locally, then it will not appear here--
// or that is the hope, but some of the various struct setup for GoConvey
// will be here too...which means that mostly the old table driven test data
// will remain in the same file.
// I know lack of locality, but I'm tired of 1000+ line tests with mostly var
// setup.
var testDir = "../test_files/"
var testRancherCfg = testDir + "rancher_test.cfg"
var testDefaultsFile = testDir + "conf/defaults_test.toml"
var testSupportedFile = testDir + "conf/supported_test.toml"
var testBuildsFile = testDir + "conf/builds_test.toml"
var testBuildListsFile = "../test_files/conf/build_lists_test.toml"
var today = time.Now().Local().Format("2006-01-02")
var testRawTemplate = newRawTemplate()
var JSONToStringMarshaller = json.NewMarshalToString()

var testProvisioners = map[string]provisioner{
	"shell": {
		templateSection{
			Settings: []string{
				"execute_command = :commands_src_dir/execute_test.command",
			},
			Arrays: map[string]interface{} {		
				"scripts": []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/base_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
					":scripts_dir/zerodisk_test.sh",
				},
			},
		},
	},
}

var testDefaults = &defaults {
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:type/:build_name",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:type",
	},
	PackerInf: PackerInf{
		Description:      "Test Default Rancher template",
		MinPackerVersion: "0.4.0",
	},
	BuildInf: BuildInf{
		BaseURL:   "",
		BuildName: "",
		Name:      ":build_name",
	},
	build: build{
		BuilderTypes: []string {
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder {
			"common": {
				templateSection {
					Settings: []string {
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
				templateSection {
					Settings: []string {
						"virtualbox_version_file = .vbox_version",
					},
					Arrays: map[string]interface{}{
						"vm_settings": []string {
							"cpus=1",
							"memory=1024",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection {
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
				templateSection {
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = out/rancher-packer.box",
					},
					Arrays: map[string]interface{}{
						"include": []string {
							"include1",
							"include2",
						},
					},
				},
			},
			"vagrant-cloud": {
				templateSection {
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
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"except": []string {
							"docker",
						},
						"only": []string {
							"virtualbox-iso",
						},
						"scripts": []string {
 							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
		},
	},
	loaded: true,
}

var testSupportedUbuntu = &distro{
	BuildInf: BuildInf{
		BaseURL: "http://releases.ubuntu.com/",
	},
	IODirInf: IODirInf{},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test supported distribution template",
	},
	Arch: []string{
		"i386",
		"amd64",
	},
	Image: []string{
		"desktop",
		"server",
		"alternate",
	},
	Release: []string{
		"10.04",
		"12.04",
		"12.10",
		"13.04",
		"13.10",
	},
	DefImage: []string{
		"release = 12.04",
		"image = server",
		"arch = amd64",
	},
	build: build{
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings: []string{
						"boot_command = ../test_files/src/ubuntu/commands/boot_test.command",
						"shutdown_command = :command_src_dir/shutdown_test.command",
					},
				},
			},
			"virtualbox-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{"memory=2048"},
					},
				},
			},
			"vmware-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{"memsize=2048"},
					},
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection {
					Settings: []string{
						"output = out/:build_name-packer.box",
					},
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
					Settings: []string{
						"execute_command = :command_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/base_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
		},
	},
}

var testSupportedCentOS = &distro{
	BuildInf: BuildInf{BaseURL: ""},
	IODirInf: IODirInf{
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test template config and Rancher options for CentOS",
	},
	Arch: []string{
		"i386",
		"x86_64",
	},
	Image: []string{
		"minimal",
		"netinstall",
	},
	Release: []string{
		"5",
		"6",
	},
	DefImage: []string{
		"release = 6",
		"image = minimal",
		"arch = x86_64",
	},
}

//var testRawPackerTemplate =
var testDistroDefaultUbuntu = rawTemplate{
	PackerInf: PackerInf{MinPackerVersion: "0.4.0", Description: "Test supported distribution template"},
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
		BaseURL:   "http://releases.ubuntu.org/",
	},
	date:    today,
	delim:   ":",
	Type:    "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "12.04",
	varVals: map[string]string{},
	vars:    map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection {
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
				templateSection {
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
				templateSection {
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
				templateSection {
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
					},
				},
			},
			"vagrant-cloud": {
				templateSection {
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar",
						"no_release = true",
						"version = 1.0.1",
					},
				},
			},
		},
		ProvisionerTypes: []string{ "shell" },
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
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
		},
	},
}

var testDistroDefaultCentOS = rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description: "Test template config and Rancher options for CentOS",
	},
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir: "http",
		HTTPSrcDir: ":src_dir/http",
		OutDir: "../test_files/out/:type/:build_name",
		ScriptsDir: "scripts",
		ScriptsSrcDir: ":src_dir/scripts",
		SrcDir: "../test_files/src/:type",
	},
	BuildInf: BuildInf{
		Name: ":build_name",
		BuildName: "",
		BaseURL: "",
	},
	date:     today,
	delim:    ":",
	Type:     "centos",
	Arch:     "x86_64",
	Image:    "minimal",
	Release:  "6",
	varVals:  map[string]string{},
	vars:     map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings:   []string{
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
				templateSection {
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
				templateSection {
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
				templateSection {
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
					},
				},
			},
			"vagrant-cloud": {
				templateSection {
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar",
						"no_release = true",
						"version = 1.0.1",
					},
				},
			},
		},
		ProvisionerTypes: []string {
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
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
						"scripts":  []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
		},
	},
}

var testBuildTest1 = rawTemplate{
	PackerInf: PackerInf{
		Description: "Test build template #1",
	},
	Type:    "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "1204",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
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
				templateSection {
					Settings: []string{
						"output = :out_dir/packer.box",
					},
					Arrays: map[string]interface{}{
						"except": []string {
							"docker",
						},
						"only": []string {
							"virtualbox-iso",
						},
					},
				},
			},
			"vagrant-cloud": {
				templateSection {
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar/baz",
						"no_release = false",
						"version = 1.0.2",
					},
				},
			},
		},	
		ProvisionerTypes: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
		},
	},
}

var testBuildTest2 = rawTemplate{
	PackerInf: PackerInf{
		Description: "Test build template #2: causes an error",
	},
	Type:    "ubuntuu",
	Arch:    "amd64",
	Image:   "desktop",
	Release: "1204",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
						},
					},
				},
			},
		},
	},
}

var testBuildCentOS6Salt = rawTemplate{
	PackerInf: PackerInf{
		Description: "Test build template for salt provisioner using CentOS6",
	},
	Type:    "centos",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
		},
		Provisioners: map[string]provisioner{
			"salt-masterless": {
				templateSection {
					Settings: []string{
						"local_state_tree = ~/saltstates/centos6/salt",
						"skip_bootstrap = true",
					},
				},
			},
		},
	},
}

var testMergedBuildTest1 = rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:type",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:type",
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
	Type:    "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "12.04",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings: []string{
						"boot_command = :commands_src_dir/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :commands_src_dir/shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=4096",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection {
					Settings: []string{
						"keep_input_artifact = false",
						"output = :out_dir/packer.box",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
		},
	},
}

var testMergedBuildTest2 = rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:type",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:type",
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
	Type:    "ubuntu",
	Arch:    "amd64",
	Image:   "desktop",
	Release: "12.04",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings: []string{
						"boot_command = :commands_src_dir/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :commands_src_dir/shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=4096",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection {
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
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection {
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
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection {
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/base_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/cleanup_test.sh",
							":scripts_dir/zerodisk_test.sh",
						},
					},
				},
			},
		},
	},
}

var testMergedBuildCentOS6Salt = rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:type",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:type",
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
	Type:    "centos",
	Arch:    "x86_64",
	Image:   "minimal",
	Release: "6",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection {
					Settings: []string{
						"boot_command = :commands_src_dir/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :commands_src_dir/shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection {
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=4096",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection {
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
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection {
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
		Provisioners: map[string]provisioner{
			"salt-masterless": {
				templateSection {
					Settings: []string {
						"local_state_tree = ~/saltstates/centos6/salt",
						"skip_bootstrap = true",
					},
				},
			},
			"shell": {
				templateSection {
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/base_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/cleanup_test.sh",
							":scripts_dir/zerodisk_test.sh",
						},
					},
				},
			},
		},
	},
}

var testSupported supported
var testMergedBuilds, testDistroDefaults map[string]rawTemplate
var testBuilds builds
var testDataSet bool

func setCommonTestData() {
	if testDataSet {
		return
	}
	testSupported.Distro = map[string]*distro{}
	testSupported.Distro["ubuntu"] = testSupportedUbuntu
	testSupported.Distro["centos"] = testSupportedCentOS

	testDistroDefaults = map[string]rawTemplate{}
	testDistroDefaults["ubuntu"] = testDistroDefaultUbuntu
	testDistroDefaults["centos"] = testDistroDefaultCentOS
	
	testBuilds.Build = map[string]rawTemplate{}
	testBuilds.Build["test1"] = testBuildTest1
	testBuilds.Build["test2"] = testBuildTest2
	testBuilds.Build["test-centos6-salt"] = testBuildCentOS6Salt

	testMergedBuilds = map[string]rawTemplate{}
	testMergedBuilds["test1"] = testMergedBuildTest1
	testMergedBuilds["test2"] = testMergedBuildTest2
	testMergedBuilds["test-centos6-salt"] = testMergedBuildCentOS6Salt

	testDataSet = true

	return
}
//BuildUbuntu
//BuildCentos
