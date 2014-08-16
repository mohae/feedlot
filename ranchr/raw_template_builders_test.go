// raw_template_builders_test.go: tests for builders.
package ranchr

import (
	"testing"

	_ "github.com/mohae/deepcopy"
	. "github.com/smartystreets/goconvey/convey"
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
	Type:    "ubuntu",
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
			"shell",
		},
		Provisioners: map[string]*provisioner{
			"shell": {
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
	Type:    "centos",
	Arch:    "x86_64",
	Image:   "minimal",
	Release: "6",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
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
			"salt-masterless": {
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
			"disk_size = 2000",
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

func TestCreateBuilderVirtualboxISO(t *testing.T) {
	Convey("Given an Ubuntu based raw template with a VirtualboxISO builder", t, func() {
		var settings map[string]interface{}
		var err error
		Convey("Calling createBuilderVirtualBoxISO", func() {
			settings, _, err = testBuilderUbuntu.createBuilderVirtualBoxISO()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in a map[string]interface with...", func() {
				So(settings["boot_wait"], ShouldEqual, "5s")	
				So(settings["disk_size"], ShouldEqual, "20000")	
				So(settings["http_directory"], ShouldEqual, "http")	
				So(settings["iso_checksum_type"], ShouldEqual, "sha256")	
				So(settings["shutdown_command"], ShouldEqual, "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'")	
				So(settings["ssh_password"], ShouldEqual, "vagrant")	
				So(settings["ssh_port"], ShouldEqual, "22")	
				So(settings["ssh_username"], ShouldEqual, "vagrant")	
				So(settings["type"], ShouldEqual, "virtualbox-iso")
				So(MarshalJSONToString.Get(settings["vboxmanage"]), ShouldEqual, "[[\"modifyvm\",\"{{.Name}}\",\"--cpus\",\"1\"],[\"modifyvm\",\"{{.Name}}\",\"--memory\",\"4096\"]]")	
			})
		})
	})

	Convey("Given a centos based raw template with a VirtualboxISO builder", t, func() {
		var settings map[string]interface{}
		var err error
		Convey("Calling createBuilderVirtualBoxISO", func() {
			settings, _, err = testBuilderCentOS.createBuilderVirtualBoxISO()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in a map[string]interface with...", func() {
				So(settings["boot_wait"], ShouldEqual, "5s")	
				So(settings["disk_size"], ShouldEqual, "20000")	
				So(settings["http_directory"], ShouldEqual, "http")	
				So(settings["iso_checksum_type"], ShouldEqual, "sha256")	
				So(settings["shutdown_command"], ShouldEqual, "echo 'vagrant'|sudo -S shutdown -t5 -h now")	
				So(settings["ssh_password"], ShouldEqual, "vagrant")	
				So(settings["ssh_port"], ShouldEqual, "22")	
				So(settings["ssh_username"], ShouldEqual, "vagrant")	
				So(settings["type"], ShouldEqual, "virtualbox-iso")
				So(MarshalJSONToString.Get(settings["vboxmanage"]), ShouldEqual, "[[\"modifyvm\",\"{{.Name}}\",\"--cpus\",\"1\"],[\"modifyvm\",\"{{.Name}}\",\"--memory\",\"4096\"]]")	
			})
		})
	})

}

func TestCreateBuilderVMWareISO(t *testing.T) {
	Convey("Given a raw template with a VMWareISO builder", t, func() {
		var settings map[string]interface{}
		var err error
		Convey("Calling createBuilderVMWareISO", func() {
			settings, _, err = testBuilderUbuntu.createBuilderVMWareISO()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in a map[string]interface with...", func() {
				So(settings["boot_wait"], ShouldEqual, "5s")	
				So(settings["disk_size"], ShouldEqual, "20000")	
				So(settings["http_directory"], ShouldEqual, "http")	
				So(settings["iso_checksum_type"], ShouldEqual, "sha256")	
				So(settings["shutdown_command"], ShouldEqual, "echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'")	
				So(settings["ssh_password"], ShouldEqual, "vagrant")	
				So(settings["ssh_port"], ShouldEqual, "22")	
				So(settings["ssh_username"], ShouldEqual, "vagrant")	
				So(settings["type"], ShouldEqual, "vmware-iso")
				So(MarshalJSONToString.Get(settings["vmx_data"]), ShouldEqual, "{\"cpus\":\"1\",\"memory\":\"4096\"}")	
			})
		})
	})
}

func TestRawTemplateUpdatebuilders(t *testing.T) {
	Convey("Given a rawTemplate", t, func() {
		Convey("updating builders with a nil map", func() {
			testBuilderUbuntu.updateBuilders(nil)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.Get(testBuilderUbuntu.Builders), ShouldEqual, MarshalJSONToString.Get(builderOrig))
			})
		})
		Convey("updating builders with another builder", func() {
			testBuilderUbuntu.updateBuilders(builderNew)
			Convey("Should result in an updated builder", func() {
				So(MarshalJSONToString.Get(testBuilderUbuntu.Builders), ShouldEqual, MarshalJSONToString.Get(builderMerged))
			})
		})
	})
}

func TestRawTemplateUpdateBuildercommon(t *testing.T) {
	Convey("Given a rawTemplate", t, func() {
		Convey("updating the common builder", func() {
			testBuilderUbuntu.updateBuilderCommon(builderNew["common"])
			Convey("Should result in an common builder", func() {
				So(MarshalJSONToString.Get(testBuilderUbuntu.Builders["common"]), ShouldEqual, MarshalJSONToString.Get(builderMerged["common"]))
			})
		})
	})
}

func TestRawTemplateBuildersSettingsToMap(t *testing.T) {
	Convey("Given a builder and a raw template", t, func() {
		Convey("Converting the Settings slice to a map", func() {
			settings := vbB.settingsToMap(rawTpl)
			Convey("Should result in a map containing", func() {
				So(settings["boot_wait"], ShouldEqual, "5s")
				So(settings["disk_size"], ShouldEqual, "2000")
				So(settings["ssh_port"], ShouldEqual, "22")
				So(settings["ssh_username"], ShouldEqual, "vagrant")
			})
		})
	})
}

func TestDeepCopyMapStringPBuilder( t *testing.T) {
	Convey("Given a builder", t , func() {
		Convey("Doing a deep copy on it", func() {
			copy := DeepCopyMapStringPBuilder(testDistroDefaults.Templates["ubuntu"].Builders)
			Convey("Should result in a copy", func() {
				So(MarshalJSONToString.GetIndented(copy["common"]), ShouldEqual, MarshalJSONToString.GetIndented(testDistroDefaults.Templates["ubuntu"].Builders["common"])) 
			})
		})
	})
}
