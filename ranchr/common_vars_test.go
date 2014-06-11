package ranchr

import (
	_"testing"
	"time"
)

// Variables that are used in various tests, so they aren't scattered every-
// where. If a variable is only used locally, then it will not appear here--
// or that is the hope, but some of the various struct setup for GoConvey
// will be here too...which means that mostly the old table driven test data
// will remain in the same file.
// I know lack of locality, but I'm tired of 1000+ line tests with mostly var
// setup.
var testDistroDefaults map[string]RawTemplate

var today = time.Now().Local().Format("2006-01-02")

var testRawTemplate = newRawTemplate()

var testMergedUbuntu = RawTemplate{
	PackerInf: PackerInf{MinPackerVersion: "", Description: ""},
	IODirInf: IODirInf{
		OutDir:      "out/:type/:build_name/",
		ScriptsDir:  ":src_dir/scripts/",
		SrcDir:      "src/:type/",
		ScriptsSrcDir:   "",
		CommandsSrcDir: "",
		HTTPDir: "",
		HTTPSrcDir: "",
	},
	BuildInf: BuildInf{Name: ":type-:release-:image-:arch", BuildName: ""},
	date: today,
	BaseURL: "http://releases.ubuntu.com/",
	delim: ":",
	Type: "ubuntu",
	Arch: "amd64",
	Image: "server",
	Release: "",
	varVals: map[string]string{},
	vars: map[string]string{},
	build: build{
		BuilderType: []string{"virtualbox-iso", "vmware-iso" },
		Builders: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :commands_dir/shutdown.command",
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
		PostProcessors: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/:type-:arch-:version-:image-packer.box",
				},
			},
		},
		Provisioners: map[string]provisioners{
			"shell": {
				Settings: []string{
					"execute_command = :commands_dir/execute.command",
				},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
	},
}

var testMergedCentOS = RawTemplate{PackerInf: PackerInf{MinPackerVersion:"", Description:""},
	 IODirInf: IODirInf{CommandsSrcDir:"/", HTTPDir:"/", HTTPSrcDir:"/", OutDir:"out/centos", ScriptsDir:":src_dir/scripts/", ScriptsSrcDir:"/", SrcDir:"src/centos"},
	BuildInf: BuildInf{Name:":type-:release-:image-:arch", BuildName:""},
	date: today,
	BaseURL:"http://www.centos.org/pub/centos/",
	delim:":",
	Type:"centos",
	Arch:"x86_64",
	Image:"minimal",
	Release:"",
	varVals:map[string]string{},
	vars:map[string]string{},
	build: build{
		BuilderType:[]string{"virtualbox-iso", "vmware-iso"},
		Builders:map[string]builder{
			"common":{
				Settings: []string{"boot_command = :commands_dir/boot.command", "boot_wait = 5s", "disk_size = 20000", "http_directory = http", "iso_checksum_type = sha256", "shutdown_command = :commands_dir/shutdown.command", "ssh_password = vagrant", "ssh_port = 22", "ssh_username = vagrant", "ssh_wait_timeout = 240m"},
				VMSettings:[]string{},
			},
			"virtualbox-iso":{
				Settings:[]string{}, 
				VMSettings:[]string{"cpus=1", "memory=1024"},
			},
			"vmware-iso":{
				Settings:[]string{}, 
				VMSettings: []string{"cpuid.coresPerSocket=1", "memsize=1024", "numvcpus=1"},
			},
		},	
		PostProcessors:map[string]postProcessors{
			"vagrant":{
				Settings:[]string{"keep_input_artifact = false", "output = :out_dir/someComposedBoxName.box"},
			},
		},
		Provisioners:map[string]provisioners{
			"shell":{
				Settings:[]string{"execute_command = :commands_dir/execute.command"}, 
				Scripts:[]string{":scripts_dir/setup.sh", ":scripts_dir/base.sh", ":scripts_dir/vagrant.sh", ":scripts_dir/cleanup.sh", ":scripts_dir/zerodisk.sh"},
			},
		},
	},
}

var testDistroUbuntu = distro{PackerInf: PackerInf{MinPackerVersion: "", Description: ""},
	IODirInf: IODirInf{
		OutDir:      "out/:type/:build_name/",
		ScriptsDir:  ":src_dir/scripts/",
		SrcDir:      "src/:type/",
		ScriptsSrcDir:   "",
		CommandsSrcDir: "",
		HTTPDir: "",
		HTTPSrcDir: "",
	},
	BuildInf: BuildInf{Name: ":type-:release-:image-:arch", BuildName: ""},
	build: build{
		BuilderType: []string{"virtualbox-iso", "vmware-iso" },
		Builders: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :commands_dir/shutdown.command",
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
		PostProcessors: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/:type-:arch-:version-:image-packer.box",
				},
			},
		},
		Provisioners: map[string]provisioners{
			"shell": {
				Settings: []string{
					"execute_command = :commands_dir/execute.command",
				},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
	},
}

var testSupported = Supported{}

func setCommonTestData() {
	testSupported.Distro = make(map[string]distro)
	testSupported.Distro["ubuntu"] = testDistroUbuntu

	testDistroDefaults = make(map[string]RawTemplate)
	testDistroDefaults["ubuntu"] = testMergedUbuntu
	testDistroDefaults["centos"] = testMergedCentOS
	
	return
}


//BuildUbuntu
//BuildCentos
