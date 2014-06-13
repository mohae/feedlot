package ranchr

import (
	_ "fmt"
	"os"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)
type test struct {
	Name string
	VarValue string
	ExpectedErrS string
}
type SupportedTest struct {
	test
	Expected Supported
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
	// make sure the test data is set
	setCommonTestData()
	var tmpEnv string

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

func TestDefaults(t *testing.T) {
	tmpEnvDefaultsFile := os.Getenv(EnvDefaultsFile)
	Convey("Given a defaults struct", t, func() {
		Convey("Given an empty default file environment setting", func() {
			d := defaults{}
			os.Setenv(EnvDefaultsFile, "")
			Convey("A load should result in an error", func() {
				err := d.Load()
				So(err.Error(), ShouldEqual, "could not retrieve the default Settings file because the RANCHER_DEFAULTS_FILE ENV variable was not set. Either set it or check your rancher.cfg setting")
			})
		})
		Convey("Given a valid defaults configuration file", func() {
			d := defaults{}
			os.Setenv(EnvDefaultsFile, "../test_files/conf/defaults_test.toml")
			Convey("A load should not error and result in data loaded", func() {
				err := d.Load()
				So(err, ShouldBeNil)
				So(d, ShouldResemble, testDefaults)
			})
		})
	})
	_ = os.Setenv(EnvDefaultsFile, tmpEnvDefaultsFile)
}

func TestSupported(t *testing.T) {
	tmpEnv := os.Getenv(EnvSupportedFile)
	Convey("Given a Supported struct", t, func() {
		Convey("Given an empty supported file environment setting", func() {
			s := Supported{}
			os.Setenv(EnvSupportedFile, "")
			Convey("A load should result in an error", func() {
				err := s.Load()
				So(err.Error(), ShouldEqual, "could not retrieve the Supported information because the RANCHER_SUPPORTED_FILE Env variable was not set. Either set it or check your rancher.cfg setting")
			})
		})
		Convey("Given a valid defaults configuration file", func() {
			s := Supported{}
			os.Setenv(EnvSupportedFile, "../test_files/conf/supported_test.toml")
			Convey("A load should not error and result in data loaded", func() {
				err := s.Load()
				So(err, ShouldBeNil)
				So(s, ShouldResemble, testSupported)
			})
		})
	})

	_ = os.Setenv(EnvSupportedFile, tmpEnv)
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
