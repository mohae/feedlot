package ranchr

import (
	_ "fmt"
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type test struct {
	Name         string
	VarValue     string
	ExpectedErrS string
}

type defaultsTest struct {
	test
	Expected defaults
}

var testDefaultsCases = []defaultsTest{
	{
		test: test{
			Name:         "Defaults: Empty Filename",
			VarValue:     "",
			ExpectedErrS: "could not retrieve the default Settings file because the " + EnvDefaultsFile + " ENV variable was not set. Either set it or check your rancher.cfg setting",
		},
		Expected: defaults{},
	},
	{
		test: test{
			Name:         "Defaults: Load defaults_test.",
			VarValue:     "../test_files/conf/defaults_test.toml",
			ExpectedErrS: "",
		},
		Expected: defaults{
			IODirInf: IODirInf{
				OutDir:      "out/:type/:build_name",
				ScriptsDir:  ":src_dir/scripts",
				SrcDir:      "src/:type",
				ScriptsSrcDir:      "",
				CommandsSrcDir: "",
			},
			PackerInf: PackerInf{
				MinPackerVersion: "",
				Description:      "Test Default Rancher template",
			},
			BuildInf: BuildInf{
				Name:      ":type-:release-:image-:arch",
				BuildName: "",
			},
			build: build{
				BuilderType: []string{
					"virtualbox-iso",
					"vmware-iso",
				},
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
				PostProcessors: map[string]postProcessors{
					"vagrant": {
						Settings: []string{
							"keep_input_artifact = false",
							"output = :out_dir/someComposedBoxName.box",
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
		},
	},
}

type SupportedTest struct {
	test
	Expected Supported
}

var testSupportedCases = []SupportedTest{
	{
		test: test{
			Name:         "Supported: Empty Filename",
			VarValue:     "",
			ExpectedErrS: "could not retrieve the Supported information because the " + EnvSupportedFile + " Env variable was not set. Either set it or check your rancher.cfg setting",
		},
		Expected: Supported{},
	},
	{
		test: test{
			Name:         "Supported: Load supported_test.toml",
			VarValue:     "../test_files/conf/supported_test.toml",
			ExpectedErrS: "",
		},
		Expected: Supported{
			Distro: map[string]distro{
				"ubuntu": {
					BuildInf: BuildInf{},
					IODirInf: IODirInf{},
					PackerInf: PackerInf{
						MinPackerVersion: "",
						Description:      "Test supported distribution template",
					},
					BaseURL: "http://releases.ubuntu.com/",
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
						"version = 12.04",
						"image = server",
						"arch = amd64",
					},
					build: build{
						Builders: map[string]builder{
							"common": {
								Settings: []string{
									"boot_command = :commands_dir/boot.command",
									"shutdown_command = :commands_dir/shutdown.command",
								},
							},
							"virtualbox-iso": {
								VMSettings: []string{"memory=2048"},
							},
							"vmware-iso": {
								VMSettings: []string{"memory=2048"},
							},
						},
						PostProcessors: map[string]postProcessors{
							"vagrant": {
								Settings: []string{
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
				},
				"centos": {
					BuildInf: BuildInf{},
					IODirInf: IODirInf{
						OutDir: "out/centos",
						SrcDir: "src/centos",
					},
					PackerInf: PackerInf{
						MinPackerVersion: "",
						Description:      "Test template config and Rancher options for CentOS",
					},
					BaseURL: "http://www.centos.org/pub/centos/",
					Arch: []string{
						"i386",
						"x86_64",
					},
					Image: []string{
						"minimal",
						"netinstall",
					},
					Release: []string{
						"5.10",
						"6.5",
					},
					DefImage: []string{
						"version = 6.5",
						"image = minimal",
						"arch = x86_64",
					},
				},
			},
		},
	},
}

type BuildsTest struct {
	test
	Expected Builds
}

var testBuildsCases = []BuildsTest{
	{
		test: test{
			Name:         "Builds: Empty Filename",
			VarValue:     "",
			ExpectedErrS: "could not retrieve the Builds configurations because the " + EnvBuildsFile + "Env variable was not set. Either set it or check your rancher.cfg setting",
		},
		Expected: Builds{},
	},
	{
		test: test{
			Name:         "Builds: Load builds_test.",
			VarValue:     "../test_files/conf/builds_test.toml",
			ExpectedErrS: "",
		},
		Expected: Builds{
			Build: map[string]RawTemplate{
				"test1": {
					PackerInf: PackerInf{
						MinPackerVersion: "",
						Description:      "Test build template",
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
						PostProcessors: map[string]postProcessors{
							"vagrant": {
								Settings: []string{
									"output = out_dir/packer.box",
								},
							},
						},
						Provisioners: map[string]provisioners{
							"shell": {
								Settings: []string{
									"execute_command = execute.command",
								},
								Scripts: []string{
									":scripts_dir/ubuntu/setup.sh",
									":scripts_dir/ubuntu/vagrant.sh",
									":scripts_dir/ubuntu/cleanup.sh",
								},
							},
						},
					},
				},
			},
		},
	},
}

type buildListsTest struct {
	test
	Expected buildLists
}

var testBuildListsCases = []buildListsTest{
	{
		test: test{
			Name:         "BuildLists: Empty Filename",
			VarValue:     "",
			ExpectedErrS: "could not retrieve the BuildLists file because the " + EnvBuildListsFile + " Env variable was not set. Either set it or check your rancher.cfg setting",
		},
		Expected: buildLists{},
	},
	{
		test: test{
			Name:         "BuildLists: Load build_lists_test.",
			VarValue:     "../test_files/conf/build_lists_test.toml",
			ExpectedErrS: "",
		},
		Expected: buildLists{
			List: map[string]list{
				"testlist-1": {
					Builds: []string{
						"test1",
						"test2",
					},
				},
				"testlist-2": {
					Builds: []string{
						"test1",
						"test2",
						"test3",
						"test4",
					},
				},
			},
		},
	},
}

func TestMain(t *testing.T) {

	var tmpEnv string

	tmpEnv = os.Getenv(EnvDefaultsFile)
	dflt := defaults{}
	for _, test := range testDefaultsCases {
		_ = os.Setenv(EnvDefaultsFile, test.VarValue)
		if err := dflt.Load(); err != nil {
			if err.Error() != test.ExpectedErrS {
				t.Errorf(test.Name, "error:", err.Error())
			} else {
				t.Logf(test.Name, "OK")
			}

		} else {
			if !reflect.DeepEqual(dflt, test.Expected) {
				t.Error(test.Name, "Expected:", test.Expected, "Got:", dflt)
			} else {
				t.Logf(test.Name, "OK")
			}
		}
	}

	_ = os.Setenv(EnvDefaultsFile, tmpEnv)

	tmpEnv = os.Getenv(EnvSupportedFile)
	sd := Supported{}
	for _, test := range testSupportedCases {
		_ = os.Setenv(EnvSupportedFile, test.VarValue)
		if err := sd.Load(); err != nil {
			if err.Error() != test.ExpectedErrS {
				t.Errorf(test.Name, "error:", err.Error())
			} else {
				t.Logf(test.Name, "OK")
			}
		} else {
			if !reflect.DeepEqual(sd, test.Expected) {
				t.Error(test.Name, "Expected:", test.Expected, "Got:", sd)
			} else {
				t.Logf(test.Name, "OK")
			}
		}
	}
	_ = os.Setenv(EnvSupportedFile, tmpEnv)

	tmpEnv = os.Getenv(EnvBuildsFile)

	b := Builds{}
	for _, test := range testBuildsCases {
		_ = os.Setenv(EnvBuildsFile, test.VarValue)
		if err := b.Load(); err != nil {
			if err.Error() != test.ExpectedErrS {
				t.Errorf(test.Name+" error: ", err.Error())
			} else {
				t.Logf(test.Name, "OK")
			}
		} else {
			if !reflect.DeepEqual(b, test.Expected) {
				t.Error(test.Name, "Expected:", test.Expected, "Got:", b)
			} else {
				t.Logf(test.Name, "OK")
			}
		}
	}

	_ = os.Setenv(EnvBuildsFile, tmpEnv)

	tmpEnv = os.Getenv(EnvBuildListsFile)
	bl := buildLists{}
	for _, test := range testBuildListsCases {
		_ = os.Setenv(EnvBuildListsFile, test.VarValue)
		if err := bl.Load(); err != nil {
			if err.Error() != test.ExpectedErrS {
				t.Errorf(test.Name+" error: ", err.Error())
			} else {
				t.Logf(test.Name, "OK")
			}
		} else {
			if !reflect.DeepEqual(bl, test.Expected) {
				t.Error(test.Name, "Expected:", test.Expected, "Got:", bl)
			} else {
				t.Logf(test.Name, "OK")
			}
		}
	}
	_ = os.Setenv(EnvBuildListsFile, tmpEnv)
}

func TestBuilderStuff(t *testing.T) {
	Convey("Given a builder, or two", t, func() {
		b := builder{}
		b.Settings = []string{"key1=value1", "key2=value2", "key3=value3"}
		b.VMSettings = []string{"VMkey1=VMvalue1"}
		newSettings := []string{"key4=value4", "key2=value22"}
		newVMSettings := []string{"VMkey1=VMvalue11", "VMkey2=VMvalue2"}
		Convey("Given two settings slices", func() {
			b.mergeSettings(newSettings)
			Convey("They should be merged", func() {
				So(b.Settings, ShouldContain, "key1=value1")
				So(b.Settings, ShouldContain, "key2=value22")
				So(b.Settings, ShouldContain, "key3=value3")
				So(b.Settings, ShouldContain, "key4=value4")
				So(b.Settings, ShouldNotContain, "key2=value2")
			})
		})

		Convey("Given two vm settings slices", func() {
			b.mergeVMSettings(newVMSettings)
			Convey("They should be merged", func() {
				So(b.VMSettings, ShouldContain, "VMkey1=VMvalue11")
				So(b.VMSettings, ShouldContain, "VMkey2=VMvalue2")
				So(b.VMSettings, ShouldNotContain, "VMkey1=VMvalue1")
			})
		})

		Convey("Given a builder settings", func() {
			rawTpl := &RawTemplate{}
			res := b.settingsToMap(rawTpl)
			Convey("They should be turned into a map[string]interface", func() {
				So(res, ShouldResemble, map[string]interface{}{"key1":"value1", "key2":"value2", "key3":"value3"})
			})
		})
	})

	Convey("Given a postProcessor or two", t, func() {
		pp := postProcessors{}
		pp.Settings = []string{"key1=value1", "key2=value2"}
		rawTpl := &RawTemplate{}

		Convey("transform settings to map should result in", func() {
			res := pp.settingsToMap("vagrant", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"vagrant", "key1":"value1", "key2":"value2"})
			})
		})
	})

	Convey("Given a provisioner or two", t, func() {
		p := provisioners{}
		p.Settings = []string{"key1=value1", "key2=value2"}
		rawTpl := &RawTemplate{}
	
		Convey("transform settingns map should result in", func() {
			res := p.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"shell", "key1":"value1", "key2":"value2"})
			})
		})

		Convey("transform settings map with a command file name embedded should result in", func() {
			p := provisioners{}
			p.Settings = []string{"key1=value1", "execute_command=../test_files/commands/execute.command"}
			res := p.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"shell", "key1":"value1",
"execute_command":"Error: open ../test_files/commands/execute.command: no such file or directory"})
			})
		})
	})
}
