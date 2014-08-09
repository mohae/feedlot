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
		newSettings := []string{"key4=value4", "key2=value22"}

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
			b.Arrays = map[string]interface{}{}
			b.Arrays[VMSettings] = []string{"VMkey1=VMvalue1", "VMkey2=VMvalue2"}
			newVMSettings := []string{"VMkey1=VMvalue11", "VMkey3=VMvalue3"}
			mergedVMSettings := []string{"VMkey1=VMvalue11", "VMkey2=VMvalue2", "VMkey3=VMvalue3"}
			merged := b.mergeVMSettings(newVMSettings)
			Convey("They should be merged", func() {
				So(merged, ShouldResemble, mergedVMSettings)
				So(merged, ShouldContain, "VMkey1=VMvalue11")
				So(merged, ShouldContain, "VMkey2=VMvalue2")
				So(merged, ShouldNotContain, "VMkey1=VMvalue1")
			})
		})

		Convey("Given a builder settings", func() {
			rawTpl := &rawTemplate{}
			res := b.settingsToMap(rawTpl)
			Convey("They should be turned into a map[string]interface", func() {
				So(res, ShouldResemble, map[string]interface{}{"key1": "value1", "key2": "value2", "key3": "value3"})
			})
		})
	})
}

func TestPostProcessorStuff(t *testing.T) {
	Convey("Given a postProcessor or two", t, func() {
		pp := postProcessor{}
		pp.Settings = []string{"key1=value1", "key2=value2"}
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
	})
}

func TestProvisionerStuff(t *testing.T) {
	Convey("Given a provisioner or two", t, func() {
		p := provisioner{}
		p.Settings = []string{"key1=value1", "key2=value2"}
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

	})

}

func TestDefaults(t *testing.T) {
	tmpEnvDefaultsFile := os.Getenv(EnvDefaultsFile)
	Convey("Given a defaults struct", t, func() {
		Convey("Given an empty default file environment setting", func() {
			d := defaults{}
			os.Setenv(EnvDefaultsFile, "")
			Convey("A load should result in an error", func() {
				err := d.LoadOnce()
				So(err.Error(), ShouldEqual, "could not retrieve the default Settings because the RANCHER_DEFAULTS_FILE environment variable was not set. Either set it or check your rancher.cfg setting")
				So(d.MinPackerVersion, ShouldEqual, "")
			})
		})
		Convey("Given a valid defaults configuration file", func() {
			os.Setenv(EnvDefaultsFile, "../test_files/conf/defaults_test.toml")
			os.Setenv(EnvRancherFile, "../test_files/rancher.cfg")
			d := defaults{}
			err := d.LoadOnce()
			Convey("A load should not error and result in data loaded", func() {
				So(err, ShouldBeNil)
				So(d.IODirInf, ShouldResemble, testDefaults.IODirInf)
				So(d.PackerInf, ShouldResemble, testDefaults.PackerInf)
				So(d.BuildInf, ShouldResemble, testDefaults.BuildInf)
				So(d.build.BuilderTypes, ShouldResemble, testDefaults.build.BuilderTypes)
				So(MarshalJSONToString.Get(d.build.Builders[BuilderVirtualBoxISO]), ShouldEqual, MarshalJSONToString.Get(testDefaults.build.Builders[BuilderVirtualBoxISO]))
				So(d.build.PostProcessorTypes, ShouldResemble, testDefaults.build.PostProcessorTypes)
				So(MarshalJSONToString.Get(d.build.PostProcessors[PostProcessorVagrant]), ShouldEqual, MarshalJSONToString.Get(testDefaults.build.PostProcessors[PostProcessorVagrant]))
				So(MarshalJSONToString.Get(d.build.PostProcessors[PostProcessorVagrantCloud]), ShouldEqual, MarshalJSONToString.Get(testDefaults.build.PostProcessors[PostProcessorVagrantCloud]))
				So(d.build.ProvisionerTypes, ShouldResemble, testDefaults.build.ProvisionerTypes)
				So(MarshalJSONToString.Get(d.build.Provisioners[ProvisionerShell]), ShouldEqual, MarshalJSONToString.Get(testDefaults.build.Provisioners[ProvisionerShell]))
			})
		})
	})
	_ = os.Setenv(EnvDefaultsFile, tmpEnvDefaultsFile)
}

func TestSupported(t *testing.T) {
	tmpEnv := os.Getenv(EnvSupportedFile)

	Convey("Given a Supported struct", t, func() {
		Convey("Given an empty supported file environment setting", func() {
			s := supported{}
			os.Setenv(EnvSupportedFile, "")
			Convey("A load should result in an error", func() {
				err := s.LoadOnce()
				So(s.loaded, ShouldEqual, false)
				So(err.Error(), ShouldEqual, "could not retrieve the Supported information because the RANCHER_SUPPORTED_FILE environment variable was not set. Either set it or check your rancher.cfg setting")
			})
		})
		Convey("Given a valid defaults configuration file", func() {
			s := supported{}
			os.Setenv(EnvSupportedFile, "../test_files/conf/supported_test.toml")
			Convey("A load should not error and result in data loaded", func() {
				err := s.LoadOnce()
				So(err, ShouldBeNil)
				So(s.loaded, ShouldEqual, true)
				// Set this because, for some reason it isn't set in testing >.>
				testSupported.Distro["ubuntu"].BaseURL = "http://releases.ubuntu.com/"
				So(MarshalJSONToString.GetIndented(s.Distro["ubuntu"]), ShouldEqual, MarshalJSONToString.GetIndented(testSupported.Distro["ubuntu"]))
				So(MarshalJSONToString.GetIndented(s.Distro["centos"]), ShouldEqual, MarshalJSONToString.GetIndented(testSupported.Distro["centos"]))
			})
		})
	})

	_ = os.Setenv(EnvSupportedFile, tmpEnv)
}

func TestBuildsStuff(t *testing.T) {
	Convey("Given a Builds struct", t, func() {
		b := builds{}
		tmpEnv := os.Getenv(EnvBuildsFile)
		Convey("Given a filename that doesn't exist", func() {
			os.Setenv(EnvBuildsFile, "../test_files/notthere.toml")
			b.LoadOnce()
			Convey("A load should result in a log entry and the builds not being loaded", func() {
				So(b.loaded, ShouldEqual, false)
			})
		})
		Convey("Given a build filename", func() {
			Convey("A load should result in", func() {
				os.Setenv(EnvBuildsFile, "../test_files/conf/builds_test.toml")
				b.LoadOnce()
				So(b.loaded, ShouldEqual, true)
				So(MarshalJSONToString.GetIndented(b.Build["test1"]), ShouldEqual, MarshalJSONToString.GetIndented(testBuilds.Build["test1"]))
				So(MarshalJSONToString.GetIndented(b.Build["test2"]), ShouldEqual, MarshalJSONToString.GetIndented(testBuilds.Build["test2"]))
			})
		})
		Convey("Given an empty build filename", func() {
			Convey("A load should result in", func() {
				os.Setenv(EnvBuildsFile, "")
				b.LoadOnce()
				So(b.loaded, ShouldEqual, false)
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
				So(b, ShouldResemble, buildLists{map[string]list{"testlist-1": {Builds: []string{"test1", "test2"}}, "testlist-2": {Builds: []string{"test1", "test2", "test3", "test4"}}}})
			})
		})
		Convey("Given an empty filename", func() {
			os.Setenv(EnvBuildListsFile, "")
			err := b.Load()
			Convey("A load should result in an error", func() {
				So(err.Error(), ShouldEqual, "could not retrieve the BuildLists file because the "+EnvBuildListsFile+" environment variable was not set. Either set it or check your rancher.cfg setting")
			})
		})

		os.Setenv(EnvBuildListsFile, tmpEnv)
	})
}
