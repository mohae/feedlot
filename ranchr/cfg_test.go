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

func TestTemplateSectionMergeArrays(t *testing.T) {
	Convey("Given a templateSection", t, func() {
		t := &templateSection{}
		Convey("Merging two nil Array elements", func() {
			merged := t.mergeArrays(nil, nil)
			Convey("Should result in nil", func() {
				So(merged, ShouldBeNil)
			})
		})
		old := map[string]interface{}{
			"type": "shell",
			"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
			"override": map[string]interface{}{
				"virtualbox-iso": map[string]interface{}{
					"scripts": []string{
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/virtualbox.sh",
						"scripts/cleanup.sh",
					},
				},
			},
		}

		new := map[string]interface{}{
			"type": "shell",
			"override": map[string]interface{}{
				"vmware-iso": map[string]interface{}{
					"scripts": []string{
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/virtualbox.sh",
						"scripts/cleanup.sh",
					},
				},
			},
		}

		newold := map[string]interface{}{
			"type": "shell",
			"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
			"override": map[string]interface{}{
				"vmware-iso": map[string]interface{}{
					"scripts": []string{
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/virtualbox.sh",
						"scripts/cleanup.sh",
					},
				},
			},
		}

		Convey("Merging an existing Array element with nil", func() {
			merged := t.mergeArrays(old, nil) 
			Convey("Should not result in nil", func() {
				So(merged, ShouldNotBeNil)
			})
			Convey("Should resemble old", func() {
				So(MarshalJSONToString.Get(merged), ShouldEqual, MarshalJSONToString.Get(old))
			})
		})
		
		Convey("Merging an existing nil Array element with new values", func() {
			merged := t.mergeArrays(nil, new) 
			Convey("Should not result in nil", func() {
				So(merged, ShouldNotBeNil)
			})
			Convey("Should resemble new", func() {
				So(MarshalJSONToString.Get(merged), ShouldEqual, MarshalJSONToString.Get(new))

			})
		})
		Convey("Merging an existing nil Array element with new values", func() {
			merged := t.mergeArrays(old, new) 
			Convey("Should not result in nil", func() {
				So(merged, ShouldNotBeNil)
			})
			Convey("Should result in the Arrays being merged", func() {
				So(MarshalJSONToString.Get(merged), ShouldEqual, MarshalJSONToString.Get(newold))
			})
		})

	})
}

func TesMergeSettings(t *testing.T) {
	Convey("Given a builder with settings", t, func() {
		b := builder{}
		b.Settings = []string{"key1=value1", "key2=value2", "key3=value3"}
		Convey("Merging a nil slice", func() {
			b.mergeSettings(nil)
			Convey("Should result in no changes", func() {
				So(b.Settings, ShouldContain, "key1=value1")
				So(b.Settings, ShouldContain, "key2=value2")
				So(b.Settings, ShouldContain, "key3=value3")
			})
		})

		Convey("Given a slice of new settings", func() {
			newSettings := []string{"key4=value4", "key2=value22"}
			Convey("Merging the new slice", func() {
				b.mergeSettings(newSettings)
				Convey("Should result in", func() {
					So(b.Settings, ShouldContain, "key1=value1")
					So(b.Settings, ShouldContain, "key2=value22")
					So(b.Settings, ShouldContain, "key3=value3")
					So(b.Settings, ShouldContain, "key4=value4")
					So(b.Settings, ShouldNotContain, "key2=value2")
				})
			})
		})
	})
}

func TestMergeVMSettings(t *testing.T) {
	Convey("Given a builder with vm settings", t, func() {
		b := builder{}
		b.Arrays = map[string]interface{}{}
		b.Arrays[VMSettings] = []string{"VMkey1=VMvalue1", "VMkey2=VMvalue2"}
		Convey("merging the slice with a nil slice", func() {
			merged := b.mergeVMSettings(nil)
			Convey("should result in nil", func() {
				So(merged, ShouldBeNil)
			})
		})

		Convey("merging the slice with a populated slice", func() {
			newVMSettings := []string{"VMkey1=VMvalue11", "VMkey3=VMvalue3"}
			mergedVMSettings := []string{"VMkey1=VMvalue11", "VMkey2=VMvalue2", "VMkey3=VMvalue3"}
			merged := b.mergeVMSettings(newVMSettings)
			Convey("Shouuld result in a merged slice", func() {
				So(merged, ShouldResemble, mergedVMSettings)
				So(merged, ShouldContain, "VMkey1=VMvalue11")
				So(merged, ShouldContain, "VMkey2=VMvalue2")
				So(merged, ShouldNotContain, "VMkey1=VMvalue1")
			})
		})

		Convey("Given a builder settings", func() {
			b.Settings = []string{"key1=value1", "key2=value2", "key3=value3"}
			rawTpl := &rawTemplate{}
			Convey("Making a map from the settings", func() {
				res := b.settingsToMap(rawTpl)
				Convey("Should result in a map[string]interface of the values", func() {
					So(res, ShouldResemble, map[string]interface{}{"key1": "value1", "key2": "value2", "key3": "value3"})
				})
			})
		})
	})
}

func TestPostProcessorMergeSettings(t *testing.T) {
	Convey("Given a postProcessor or two", t, func() {
		pp := postProcessor{}
		pp.Settings = []string{"key1=value1", "key2=value2"}
		Convey("Merging a nil slice", func() {
			pp.mergeSettings(nil)
			Convey("Should result in no change", func() {
				So(pp.Settings, ShouldContain, "key1=value1")
				So(pp.Settings, ShouldContain, "key2=value2")
			})	
		})
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
		Convey("Given a postProcessor with a nil settings", func() {
			post := postProcessor{}
			Convey("Merging a slice", func() {
				post.mergeSettings(newSettings)
				Convey("Should result in post.Settings == new", func() {
					So(post.Settings, ShouldContain, "key1=value1")
					So(post.Settings, ShouldContain, "key2=value22")
					So(post.Settings, ShouldContain, "key3=value3")
				})
			})
		})
	})
}

func TestProvisionerMergeSettings(t *testing.T) {
	Convey("Given a provisioner or two", t, func() {
		p := provisioner{}
		p.Settings = []string{"key1=value1", "key2=value2"}
		Convey("Merging a nil slice", func() {
			p.mergeSettings(nil)
			Convey("Should result in no change", func() {
				So(p.Settings, ShouldContain, "key1=value1")
				So(p.Settings, ShouldContain, "key2=value2")
			})	
		})
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
		Convey("Given a provisioner with a nil settings", func() {
			pr := provisioner{}
			Convey("Merging a slice", func() {
				pr.mergeSettings(newSettings)
				Convey("Should result in pr.Settings == new", func() {
					So(pr.Settings, ShouldContain, "key1=value1")
					So(pr.Settings, ShouldContain, "key2=value22")
					So(pr.Settings, ShouldContain, "key3=value3")
				})
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
