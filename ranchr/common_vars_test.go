package ranchr

import (
	"time"
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

var testProvisioners = map[string]provisioner{
	"shell": {
		Settings: []string{
			"execute_command = :commands_src_dir/execute_test.command",
			[]string{
				":scripts_dir/setup_test.sh",
				":scripts_dir/base_test.sh",
				":scripts_dir/vagrant_test.sh",
				":scripts_dir/cleanup_test.sh",
				":scripts_dir/zerodisk_test.sh",
			},
		},

	},
}

var testDefaults = &defaults{
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
		Description:      "Test Default Rancher template",
		MinPackerVersion: "0.4.0",
	},
	BuildInf: BuildInf{
		BaseURL:   "",
		BuildName: "",
		Name:      ":type-:release-:image-:arch",
	},
	build: build{
		BuilderType: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
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
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox-iso": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware-iso": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = out/rancher-packer.box",
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{
					"execute_command = :commands_src_dir/execute_test.command",
				},
/*
				Scripts: []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/base_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
					":scripts_dir/zerodisk_test.sh",
				},
*/
			},
		},
	},
}

var testSupportedUbuntu = &distro{
	BuildInf: BuildInf{BaseURL: "http://releases.ubuntu.com/"},
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
				Settings: []string{
					"boot_command = :commands_src_dir/boot_test.command",
					"shutdown_command = :commands_src_dir/shutdown_test.command",
				},
			},
			"virtualbox-iso": {
				VMSettings: []string{"memory=2048"},
			},
			"vmware-iso": {
				VMSettings: []string{"memsize=2048"},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"output = out/:type-:arch-:version-:image-packer.box",
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{
					"execute_command = :commands_src_dir/execute_test.command",
				},
/*
				Scripts: []string{
					"scripts/setup_test.sh",
					"scripts/base_test.sh",
					"scripts/vagrant_test.sh",
					"scripts/cleanup_test.sh",
					"scripts/zerodisk_test.sh",
				},
*/
			},
		},
	},
}

var testSupportedCentOS = &distro{
	BuildInf: BuildInf{BaseURL: ""},
	IODirInf: IODirInf{
		OutDir: "out/centos",
		SrcDir: "src/centos",
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
	PackerInf: PackerInf{MinPackerVersion: "", Description: "Test supported distribution template"},
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands/",
		HTTPDir:        "http/",
		HTTPSrcDir:     ":src_dir/http/",
		OutDir:         "../test_files/out/:type/",
		ScriptsDir:     "scripts/",
		ScriptsSrcDir:  ":src_dir/scripts/",
		SrcDir:         "../test_files/src/:type/",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
		BaseURL:   "http://releases.ubuntu.com/",
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
		BuilderType: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
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
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox-iso": {
				VMSettings: []string{
					"cpus=1",
					"memory=2048",
				},
			},
			"vmware-iso": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=2048",
					"numvcpus=1",
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = out/:type-:arch-:version-:image-packer.box",
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{
					"execute_command = :commands_src_dir/execute_test.command",
				},
/*
				Scripts: []string{
					"scripts/setup_test.sh",
					"scripts/base_test.sh",
					"scripts/vagrant_test.sh",
					"scripts/cleanup_test.sh",
					"scripts/zerodisk_test.sh",
				},
*/
			},
		},
	},
}

var testDistroDefaultCentOS = rawTemplate{PackerInf: PackerInf{MinPackerVersion: "", Description: "Test template config and Rancher options for CentOS"},
	IODirInf: IODirInf{CommandsSrcDir: ":src_dir/commands/", HTTPDir: "http/", HTTPSrcDir: ":src_dir/http/", OutDir: "out/centos", ScriptsDir: "scripts/", ScriptsSrcDir: ":src_dir/scripts/", SrcDir: "../test_files/src/centos"},
	BuildInf: BuildInf{Name: ":type-:release-:image-:arch", BuildName: "", BaseURL: "http://www.centos.org/pub/centos/"},
	date:     today,
	delim:    ":",
	Type:     "centos",
	Arch:     "x86_64",
	Image:    "minimal",
	Release:  "6",
	varVals:  map[string]string{},
	vars:     map[string]string{},
	build: build{
		BuilderType: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				Settings:   []string{"boot_command = :commands_src_dir/boot_test.command", "boot_wait = 5s", "disk_size = 20000", "http_directory = http", "iso_checksum_type = sha256", "shutdown_command = :commands_src_dir/shutdown_test.command", "ssh_password = vagrant", "ssh_port = 22", "ssh_username = vagrant", "ssh_wait_timeout = 240m"},
				VMSettings: []string{},
			},
			"virtualbox-iso": {
				Settings:   []string{},
				VMSettings: []string{"cpus=1", "memory=1024"},
			},
			"vmware-iso": {
				Settings:   []string{},
				VMSettings: []string{"cpuid.coresPerSocket=1", "memsize=1024", "numvcpus=1"},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{"keep_input_artifact = false", "output = out/rancher-packer.box"},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{"execute_command = :commands_src_dir/execute_test.command"},
/*
				Scripts:  []string{":scripts_dir/setup_test.sh", ":scripts_dir/base_test.sh", ":scripts_dir/vagrant_test.sh", ":scripts_dir/cleanup_test.sh", ":scripts_dir/zerodisk_test.sh"},
*/
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
		BuilderType: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				Settings: []string{
					"ssh_wait_timeout = 300m",
				},
			},
			"virtualbox-iso": {
				VMSettings: []string{
					"memory=4096",
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"output = :out_dir/packer.box",
				},
				Except: []string{
					"docker",
				},
				Only: []string{
					"virtualbox-iso",
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{
					"execute_command = :commands_src_dir/execute_test.command",
				},
/*
				Scripts: []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
				},
*/
				Except: []string{
					"docker",
				},
				Only: []string{
					"virtualbox-iso",
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
		BuilderType: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
				Settings: []string{
					"ssh_wait_timeout = 300m",
				},
			},
			"virtualbox-iso": {
				VMSettings: []string{
					"memory=4096",
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
		BuilderType: []string{
			"virtualbox-iso",
		},
/*		Provisioner: map[string]provisioner{
			"salt-masterless": {
				Settings: []string{
					"local_state_tree = ~/saltstates/centos6/salt",
					"skip_bootstrap = true",
				},
			},
		},
*/
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
		BuilderType: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
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
			"virtualbox-iso": {
				VMSettings: []string{
					"cpus=1",
					"memory=4096",
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/packer.box",
				},
				Except: []string{
					"docker",
				},
				Only: []string{
					"virtualbox-iso",
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{
					"execute_command = :commands_src_dir/execute_test.command",
				},
/*
				Scripts: []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
				},
*/
				Except: []string{
					"docker",
				},
				Only: []string{
					"virtualbox-iso",
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
		BuilderType: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
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
			"virtualbox-iso": {
				VMSettings: []string{
					"cpus=1",
					"memory=4096",
				},
			},
			"vmware-iso": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = out/someComposedBoxName.box",
				},
			},
		},
		Provisioners: map[string]provisioner{
			"shell": {
				Settings: []string{
					"execute_command = :commands_src_dir/execute_test.command",
				},
/*				Scripts: []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/base_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
					":scripts_dir/zerodisk_test.sh",
				},
*/
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
		BuilderType: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
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
			"virtualbox-iso": {
				VMSettings: []string{
					"cpus=1",
					"memory=4096",
				},
			},
			"vmware-iso": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = out/someComposedBoxName.box",
				},
			},
		},
		Provisioners: map[string]provisionerer{
/*
			"salt-masterless": saltProvisioner{
				provisioner: {
					Settings: []string {
						"local_state_tree = ~/saltstates/centos6/salt",
						"skip_bootstrap = true",
					},
				},
			},
*/
/*
			"shell": &shellProvisioner{

				provisioner: provisioner{
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
				},

				Scripts: []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/base_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
					":scripts_dir/zerodisk_test.sh",
				},
			},
*/
		},
	},
}

var testSupported, testSupportedNoBaseURL supported
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
	testSupportedNoBaseURL.Distro = map[string]*distro{}
	for k, v := range testSupported.Distro {
		v.BaseURL = ""
		testSupportedNoBaseURL.Distro[k] = v
	}
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
