// raw_template_builders_test.go: tests for builders.
package ranchr

import (
	"testing"
)

var testBuilderUbuntu = &rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: "../test_files/ubuntu/src/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     "../test_files/ubuntu/src/http",
		OutDir:         "../test_files/ubuntu/out/ubuntu",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  "../test_files/src/ubuntu/scripts",
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
						"boot_command = ../test_files/src/ubuntu/commands/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = ../test_files/src/ubuntu/commands/shutdown_test.command",
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
			"shell-scripts",
		},
		Provisioners: map[string]*provisioner{
			"shell-scripts": {
				templateSection{
					Settings: []string{
						"execute_command = ../test_files/src/ubuntu/commands/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"../test_files/src/ubuntu/scripts/setup_test.sh",
							"../test_files/src/ubuntu/scripts/base_test.sh",
							"../test_files/src/ubuntu/scripts/vagrant_test.sh",
							"../test_files/src/ubuntu/scripts/cleanup_test.sh",
							"../test_files/src/ubuntu/scripts/zerodisk_test.sh",
						},
					},
				},
			},
		},
	},
}

var testBuilderCentOS = &rawTemplate{
	IODirInf: IODirInf{
		CommandsSrcDir: "../test_files/centos/src/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     "../test_files/centos/src/http",
		OutDir:         "../test_files/out/centos",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  "../test_files/centos/src/scripts",
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
						"boot_command = ../test_files/src/centos/commands/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = ../test_files/src/centos/commands/shutdown_test.command",
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
			"shell-scripts",
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
			"shell-scripts": {
				templateSection{
					Settings: []string{
						"execute_command = ../test_files/src/centos/commands/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"../test_files/centos/src/scripts/setup_test.sh",
							"../test_files/centos/src/scripts/base_test.sh",
							"../test_files/centos/src/scripts/vagrant_test.sh",
							"../test_files/centos/src/scripts/cleanup_test.sh",
							"../test_files/centos/src/scripts/zerodisk_test.sh",
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
				"boot_command = ../test_files/src/ubuntu/commands/boot_test.command",
				"boot_wait = 5s",
				"disk_size = 20000",
				"http_directory = http",
				"iso_checksum_type = sha256",
				"shutdown_command = ../test_files/src/ubuntu/commands/shutdown_test.command",
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
				"boot_command = ../test_files/src/ubuntu/commands/boot_test.command",
				"boot_wait = 15s",
				"disk_size = 20000",
				"http_directory = http",
				"iso_checksum_type = sha256",
				"shutdown_command = ../test_files/src/ubuntu/commands/shutdown_test.command",
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
				"boot_command = ../test_files/src/ubuntu/commands/boot_test.command",
				"boot_wait = 15s",
				"disk_size = 20000",
				"http_directory = http",
				"iso_checksum_type = sha256",
				"shutdown_command = ../test_files/src/ubuntu/commands/shutdown_test.command",
				"ssh_password = vagrant",
				"ssh_port = 22",
				"ssh_username = vagrant",
				"ssh_wait_timeout = 240m",
			},
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

func TestCreateBuilderVirtualbox(t *testing.T) {
	var settings map[string]interface{}
	var err error
	settings, _, err = testBuilderUbuntu.createVirtualBoxISO()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	if settings["boot_wait"] != "5s" {
		t.Errorf("Expected \"5s\", got %q", settings["boot_wait"])
	}
	if settings["disk_size"] != 20000 {
		t.Errorf("Expected 20000, got %d", settings["disk_size"])
	}
	if settings["http_directory"] != "http" {
		t.Errorf("Expected \"http\", got %q", settings["http_directory"])
	}
	if settings["iso_checksum_type"] != "sha256" {
		t.Errorf("Expected \"sha256\", got %q", settings["iso_checksum_type"])
	}
	if settings["shutdown_command"] != "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'" {
		t.Errorf("Expected \"echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'\", got %q", settings["shutdown_command"])
	}
	if settings["ssh_password"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_password"])
	}
	if settings["ssh_port"] != 22 {
		t.Errorf("Expected 22, got %q", settings["ssh_port"])
	}
	if settings["ssh_username"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_username"])
	}
	if settings["type"] != "virtualbox-iso" {
		t.Errorf("Expected \"virtualbox-iso\", got %q", settings["type"])
	}
	if MarshalJSONToString.Get(settings["vboxmanage"]) != "[[\"modifyvm\",\"{{.Name}}\",\"--cpus\",\"1\"],[\"modifyvm\",\"{{.Name}}\",\"--memory\",\"4096\"]]" {
		t.Errorf("Expected \"[[\"modifyvm\",\"{{.Name}}\",\"--cpus\",\"1\"],[\"modifyvm\",\"{{.Name}}\",\"--memory\",\"4096\"]]\", got %q", MarshalJSONToString.Get(settings["vboxmanage"]))
	}

	settings, _, err = testBuilderCentOS.createVirtualBoxOVF()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	if settings["shutdown_command"] != "echo 'vagrant'|sudo -S shutdown -t5 -h now" {
		t.Errorf("Expected \"echo 'vagrant'|sudo -S shutdown -t5 -h now\", got %q", settings["shutdown_command"])
	}
	if settings["ssh_password"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_password"])
	}
	if settings["ssh_port"] != 22 {
		t.Errorf("Expected 22, got %d", settings["ssh_port"])
	}
	if settings["ssh_username"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_username"])
	}
	if settings["type"] != "virtualbox-ovf" {
		t.Errorf("Expected \"virtualbox-ovf\", got %q", settings["type"])
	}
	if MarshalJSONToString.Get(settings["vboxmanage"]) != "[[\"modifyvm\",\"{{.Name}}\",\"--cpus\",\"1\"],[\"modifyvm\",\"{{.Name}}\",\"--memory\",\"4096\"]]" {
		t.Errorf("Expected \"[[\"modifyvm\",\"{{.Name}}\",\"--cpus\",\"1\"],[\"modifyvm\",\"{{.Name}}\",\"--memory\",\"4096\"]]\", got %q", MarshalJSONToString.Get(settings["vboxmanage"]))
	}
}

func TestCreateBuilderVMWare(t *testing.T) {
	var settings map[string]interface{}
	var err error
	settings, _, err = testBuilderUbuntu.createVMWareISO()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	if settings["boot_wait"] != "5s" {
		t.Errorf("Expected \"5s\", got %q", settings["boot_wait"])
	}
	if settings["disk_size"] != 20000 {
		t.Errorf("Expected 20000, got %d", settings["disk_size"])
	}
	if settings["http_directory"] != "http" {
		t.Errorf("Expected \"http\", got %q", settings["http_directory"])
	}
	if settings["iso_checksum_type"] != "sha256" {
		t.Errorf("Expected \"sha256\", got %q", settings["iso_checksum_type"])
	}
	if settings["shutdown_command"] != "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'" {
		t.Errorf("Expected \"echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'\", got %q", settings["shutdown_command"])
	}
	if settings["ssh_password"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_password"])
	}
	if settings["ssh_port"] != 22 {
		t.Errorf("Expected 22, got %d", settings["ssh_port"])
	}
	if settings["ssh_username"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_username"])
	}
	if settings["type"] != "vmware-iso" {
		t.Errorf("Expected \"vmware-iso\", got %q", settings["type"])
	}
	expected := map[string]string{"cpuid.coresPerSocket": "1", "memsize": "1024", "numvcpus": "1"}
	if MarshalJSONToString.Get(settings["vmx_data"]) != MarshalJSONToString.Get(expected) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(settings["vmx_data"]))
	}

	settings, _, err = testBuilderCentOS.createVMWareVMX()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	if settings["shutdown_command"] != "echo 'vagrant'|sudo -S shutdown -t5 -h now" {
		t.Errorf("Expected \"echo 'vagrant'|sudo -S shutdown -t5 -h now\", got %q", settings["shutdown_command"])
	}
	if settings["ssh_password"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_password"])
	}
	if settings["ssh_port"] != 22 {
		t.Errorf("Expected 22, got %d", settings["ssh_port"])
	}
	if settings["ssh_username"] != "vagrant" {
		t.Errorf("Expected \"vagrant\", got %q", settings["ssh_username"])
	}
	if settings["type"] != "vmware-vmx" {
		t.Errorf("Expected \"vmware-vmx\", got %q", settings["type"])
	}

	vmx := settings["vmx_data"].(map[string]string)
	cpus, ok := vmx["numvcpus"]
	if !ok {
		t.Error("Expected the \"numvcpus\" entry to exist in vmx_data map, not found")
	} else {
		if vmx["numvcpus"] != "1" {
			t.Errorf("Expected \"1\", got %q", cpus)
		}
	}
	mem, ok := vmx["memsize"]
	if !ok {
		t.Error("Expected the \"memsize\" entry to exist in vmx_data map, not found")
	} else {
		if mem != "1024" {
			t.Errorf("Expected \"1024\", got %q", mem)
		}
	}
}

func TestRawTemplateUpdatebuilders(t *testing.T) {
	testBuilderUbuntu.updateBuilders(nil)
	if MarshalJSONToString.Get(testBuilderUbuntu.Builders) != MarshalJSONToString.Get(builderOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(builderOrig), MarshalJSONToString.Get(testBuilderUbuntu.Builders))
	}

	testBuilderUbuntu.updateBuilders(builderNew)
	if MarshalJSONToString.Get(testBuilderUbuntu.Builders) != MarshalJSONToString.Get(builderMerged) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(builderMerged), MarshalJSONToString.Get(testBuilderUbuntu.Builders))
	}
}

func TestRawTemplateUpdateBuildercommon(t *testing.T) {
	testBuilderUbuntu.updateCommon(builderNew["common"])
	if MarshalJSONToString.Get(testBuilderUbuntu.Builders["common"]) != MarshalJSONToString.Get(builderMerged["common"]) {
		t.Errorf("expected %q, got %q", MarshalJSONToString.Get(builderMerged["common"]), MarshalJSONToString.Get(testBuilderUbuntu.Builders["common"]))
	}
}

func TestRawTemplateBuildersSettingsToMap(t *testing.T) {
	settings := vbB.settingsToMap(rawTpl)
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

func TestDeepCopyMapStringPBuilder(t *testing.T) {
	cpy := DeepCopyMapStringPBuilder(testDistroDefaults.Templates[Ubuntu].Builders)
	if MarshalJSONToString.Get(cpy["common"]) != MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu].Builders["common"]) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu].Builders["common"]), MarshalJSONToString.Get(cpy["common"]))
	}
}
