package ranchr

import (
	_ "fmt"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func init() {
	setCommonTestData()
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
			Convey("Merging them should result in", func() {
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
		newSettings := []string{"key1=value1", "key2=value22", "key3=value3"}
		Convey("Given two settings slices", func() {
			pp.mergeSettings(newSettings)
			Convey("Merging them should result in", func() {
				So(pp.Settings, ShouldContain, "key1=value1")
				So(pp.Settings, ShouldContain, "key2=value22")
				So(pp.Settings, ShouldContain, "key3=value3")
				So(pp.Settings, ShouldNotContain, "key2=value2")
			})
		})

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
		newSettings := []string{"key1=value1", "key2=value22", "key3=value3"}
		Convey("Given two settings slices", func() {
			p.mergeSettings(newSettings)
			Convey("Merging them should result in", func() {
				So(p.Settings, ShouldContain, "key1=value1")
				So(p.Settings, ShouldContain, "key2=value22")
				So(p.Settings, ShouldContain, "key3=value3")
				So(p.Settings, ShouldNotContain, "key2=value2")
			})
		})

		Convey("transform settingns map should result in", func() {
			res := p.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"shell", "key1":"value1", "key2":"value2"})
			})
		})

		Convey("transform settings map with an invalid command file name embedded should result in", func() {
			p := provisioners{}
			p.Settings = []string{"key1=value1", "execute_command=../test_files/commands/execute.command"}
			res := p.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"shell", "key1":"value1",
"execute_command":"Error: open ../test_files/commands/execute.command: no such file or directory"})
			})
		})

		Convey("transform settings map with an invalid command file name embedded should result in", func() {
			p := provisioners{}
			p.Settings = []string{"key1=value1", "execute_command=../test_files/src/ubuntu/commands/execute_test.command"}
			res := p.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"shell", "key1":"value1",
"execute_command":"\"echo 'vagrant'|sudo -S sh '{{.Path}}'\""})
			})
		})

		Convey("given a slice with new script names, ", func() {
			p := provisioners{}
			p.Scripts = []string{"script1", "script2"}
			script := []string{"script3", "script4"}
			p.setScripts(script)
			Convey("Should result in the slice being replaced", func() {
				So(p.Scripts, ShouldResemble, []string{"script3", "script4"})
			})
		})
	})
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

func TestBuildsStuff(t *testing.T) {
	Convey("Given a Builds struct", t, func() {	
		b := Builds{}
		tmpEnv := os.Getenv(EnvBuildsFile)
		Convey("Given a filename that doesn't exist", func() {
				os.Setenv(EnvBuildsFile, "../test_files/notthere.toml")
				err := b.Load()
				Convey("A load should result in an error", func() {			
					So(err.Error(), ShouldEqual, "open ../test_files/notthere.toml: no such file or directory")
				})
		})
		Convey("Given a build filename", func() {
			Convey("A load should result in", func() {			
				os.Setenv(EnvBuildsFile, "../test_files/conf/builds_test.toml")
				err := b.Load()
				So(err, ShouldBeNil)
				So(b, ShouldResemble, testBuilds)
			})		
		})
		Convey("Given an empty build filename", func() {
			Convey("A load should result in", func() {			
				os.Setenv(EnvBuildsFile, "")
				err := b.Load()
				So(err.Error(), ShouldEqual, "could not retrieve the Builds configurations because the " + EnvBuildsFile + "Env variable was not set. Either set it or check your rancher.cfg setting")
			})		
		})
		os.Setenv(EnvBuildsFile, tmpEnv)
	})
}

func TestBuildListsStuff(t *testing.T) {
	Convey("Given a buildLists struct", t, func() {	
		b := buildLists{}
		tmpEnv := os.Getenv(EnvBuildListsFile)
		Convey("Given a filename that doesn't exist", func() {
			os.Setenv(EnvBuildListsFile, "../test_files/notthere.toml")
			err := b.Load()
			Convey("A load should result in an error", func() {			
				So(err.Error(), ShouldEqual, "open ../test_files/notthere.toml: no such file or directory")
			})
		})
		Convey("Given a BuildLists name", func() {
			os.Setenv(EnvBuildListsFile, "../test_files/conf/build_lists_test.toml")
			err := b.Load()
			Convey("A load should successfully load the file", func() {			
				So(err, ShouldBeNil)
				So(b, ShouldResemble, buildLists{map[string]list{"testlist-1":{Builds: []string{"test1", "test2"}}, "testlist-2":{Builds: []string{"test1", "test2", "test3", "test4"}}}})
			})
		})
		Convey("Given an empty filename", func() {
			os.Setenv(EnvBuildListsFile, "")
			err := b.Load()
			Convey("A load should result in an error", func() {			
				So(err.Error(), ShouldEqual, "could not retrieve the BuildLists file because the " + EnvBuildListsFile + " Env variable was not set. Either set it or check your rancher.cfg setting")
			})
		})

		os.Setenv(EnvBuildListsFile, tmpEnv)
	})
}
