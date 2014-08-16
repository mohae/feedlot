package ranchr

import (
	"testing"
	_ "time"

	. "github.com/smartystreets/goconvey/convey"
)

var updatedBuilders =   map[string]*builder{
	"common": {
		templateSection{
			Settings: []string{
				"ssh_wait_timeout = 300m",
			},
		},
	},
	"virtualbox-iso": {
		templateSection{
			Settings: []string{},
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"memory=4096",
				},
			},
		},
	},
}


func init() {
	setCommonTestData()
}

func TestNewRawTemplate(t *testing.T) {
	Convey("Testing NewRawTemplate", t, func() {
		Convey("Given a request for a newRawTemplate()", func() {
			rawTpl := newRawTemplate()
			Convey("The raw template should equal--we don't test the date because it is always changeing", func() {
				So(rawTpl, ShouldResemble, testRawTemplate)
			})
		})
	})
}

func TestCreatePackerTemplate(t *testing.T) {
	Convey("Given a template", t, func() {
		r := &rawTemplate{}
		r = testDistroDefaults.Templates["ubuntu"]
		Convey("Calling rawTemplate.CreatePackerTemplate() should result in", func() {
			var pTpl packerTemplate
			var err error
			pTpl, err = r.createPackerTemplate()
			So(err, ShouldBeNil)
			So(pTpl, ShouldNotResemble, r)
		})
	})
}

func TestCreateBuilders(t *testing.T) {
	Convey("Given a template", t, func() {
		r := &rawTemplate{}
		r = testDistroDefaults.Templates["ubuntu"]
		var bldrs []interface{}
		var vars map[string]interface{}
		var err error
		Convey("Given a call to createBuilders", func() {
			Convey("Given a valid Builder Type", func() {
				//first merge the variables so that create builders will work
				r.mergeVariables()
				bldrs, vars, err = r.createBuilders()
				So(err, ShouldBeNil)
				So(vars, ShouldBeNil)
				So(bldrs, ShouldNotResemble, r)
			})
			Convey("Given an unsupported Builder Types", func() {
				r.BuilderTypes[0] = "unsupported"
				bldrs, vars, err = r.createBuilders()
				So(err.Error(), ShouldEqual, "The requested builder, 'unsupported', is not supported by Rancher")
				So(vars, ShouldBeNil)
				So(bldrs, ShouldBeNil)
			})
			Convey("Given no Builder Types", func() {
				r.BuilderTypes = nil
				bldrs, vars, err = r.createBuilders()
				So(err.Error(), ShouldEqual, "rawTemplate.createBuilders: no builder types were configured, unable to create builders")
				So(vars, ShouldBeNil)
				So(bldrs, ShouldBeNil)
			})

		})
	})
}

func TestReplaceVariables(t *testing.T) {
	Convey("Given a new raw template", t, func() {
		r := newRawTemplate()
		Convey("Given a Variable:Value map", func() {
			r.varVals = map[string]string{
				":arch":            "amd64",
				":command_src_dir": ":src_dir/commands",
				":http_dir":        "http",
				":http_src_dir":    ":src_dir/http",
				":image":           "server",
				":name":            ":type-:release:-:image-:arch",
				":out_dir":         "../test_files/out/:type",
				":release":         "14.04",
				":scripts_dir":     "scripts",
				":scripts_src_dir": ":src_dir/scripts",
				":src_dir":         "../test_files/src/:type",
				":type":            "ubuntu",
			}
			r.delim = ":"
			Convey("Given a string to perform replacement on", func() {
				s := r.replaceVariables(":src_dir/command")
				So(s, ShouldEqual, "../test_files/src/ubuntu/command")
			})
			Convey("Given a string without any delimiters", func() {
				s := r.replaceVariables("http")
				So(s, ShouldEqual, "http")
			})
			Convey("Given another string", func() {
				s := r.replaceVariables("../test_files/out/:type")
				So(s, ShouldEqual, "../test_files/out/ubuntu")
			})
		})
	})
}

func TestRawTemplateVariableSection(t *testing.T) {
	Convey("Given a RawTemplate variable section", t, func() {
		r := newRawTemplate()
		Convey("rawTemplate.variable section", func() {
			res, err := r.variableSection()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("the variable section should be empty", func() {
				So(res, ShouldResemble, map[string]interface{}{})
			})
		})
	})
}

func TestRawTemplateSetDefaults(t *testing.T) {
	Convey("Given a raw template", t, func() {
		r := newRawTemplate()
		Convey("setting the distro defaults", func() {
			r.setDefaults(testSupported.Distro["centos"])
			Convey("should result in setting that match the defaults", func() {
				So(r.IODirInf, ShouldResemble, testSupported.Distro["centos"].IODirInf)
				So(r.PackerInf, ShouldResemble, testSupported.Distro["centos"].PackerInf)
				So(r.BuildInf, ShouldResemble, testSupported.Distro["centos"].BuildInf)
				So(r.BuilderTypes, ShouldResemble, testSupported.Distro["centos"].BuilderTypes)
				So(r.PostProcessorTypes, ShouldResemble, testSupported.Distro["centos"].PostProcessorTypes)
				So(r.ProvisionerTypes, ShouldResemble, testSupported.Distro["centos"].ProvisionerTypes)
				So(r.Builders, ShouldResemble, testSupported.Distro["centos"].Builders)
				So(r.PostProcessors, ShouldResemble, testSupported.Distro["centos"].PostProcessors)
				So(r.Provisioners, ShouldResemble, testSupported.Distro["centos"].Provisioners)
			})
		})
	})
}

func TestRawTemplateUpdateBuildSettings(t *testing.T) {
	Convey("Given a raw template with defaults set", t, func() {
		r := newRawTemplate()
		r.setDefaults(testSupported.Distro["centos"])
		Convey("updating the build settings", func() {
			r.updateBuildSettings(testBuilds.Build["test1"])
			Convey("should result in updated build settings", func() {
				So(r.IODirInf, ShouldResemble, testSupported.Distro["centos"].IODirInf)
				So(r.PackerInf, ShouldResemble, testBuilds.Build["test1"].PackerInf)
				So(r.BuildInf, ShouldResemble, testSupported.Distro["centos"].BuildInf)
				So(r.BuilderTypes, ShouldResemble, testBuilds.Build["test1"].BuilderTypes)
				So(r.PostProcessorTypes, ShouldResemble, testBuilds.Build["test1"].PostProcessorTypes)
				So(r.ProvisionerTypes, ShouldResemble, testBuilds.Build["test1"].ProvisionerTypes)
				So(MarshalJSONToString.Get(r.Builders), ShouldResemble, MarshalJSONToString.Get(updatedBuilders))
				So(MarshalJSONToString.Get(r.PostProcessors), ShouldResemble, MarshalJSONToString.Get(testBuilds.Build["test1"].PostProcessors))
				So(MarshalJSONToString.Get(r.Provisioners), ShouldResemble, MarshalJSONToString.Get(testBuilds.Build["test1"].Provisioners))
			})
		})
	})
}

func TestRawTemplatescriptNames(t *testing.T) {
	Convey("Given a raw template with a shell provisioner", t, func() {
		r := testDistroDefaults.Templates["ubuntu"]
		Convey("getting the script names", func() {
			scripts := r.ScriptNames()
			Convey("Scripts should not be nil", func() {
				So(scripts, ShouldNotBeNil)
			})
			Convey("should result in a slice of script names", func() {
				So(scripts, ShouldContain, "setup_test.sh")
				So(scripts, ShouldContain, "vagrant_test.sh")
				So(scripts, ShouldContain, "sudoers_test.sh")
				So(scripts, ShouldContain, "cleanup_test.sh")
				So(scripts, ShouldNotContain, "setup.sh")
			})
		})
	})
}

func TestMergeVariables(t *testing.T) {
	Convey("Given a raw template", t, func() {
		r := testDistroDefaults.Templates["ubuntu"]
		Convey("Merging the variables", func() {
			r.mergeVariables()
			So(r.CommandsSrcDir, ShouldEqual, "../test_files/src/ubuntu/commands")
			So(r.HTTPDir, ShouldEqual, "http")
			So(r.HTTPSrcDir, ShouldEqual, "../test_files/src/ubuntu/http")
			So(r.OutDir, ShouldEqual, "../test_files/out/ubuntu")
			So(r.ScriptsDir, ShouldEqual, "scripts")
			So(r.ScriptsSrcDir, ShouldEqual, "../test_files/src/ubuntu/scripts")
			So(r.SrcDir, ShouldEqual, "../test_files/src/ubuntu")
		})
	})
}

func TestIODirInf(t *testing.T) {

	Convey("Given a IODirInf", t, func() {

		Convey("Given an empty new IODirInf", func() {
			oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
			newIODirInf := IODirInf{}

			oldIODirInf.update(newIODirInf)
			So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
			So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
			So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
			So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
			So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
			So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
			So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
		})

		Convey("Given a populated new IODirInf", func() {
			oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
			newIODirInf := IODirInf{CommandsSrcDir: "new CommandsSrcDir", HTTPDir: "new HTTPDir", HTTPSrcDir: "new HTTPSrcDir", OutDir: "new OutDir", ScriptsDir: "new ScriptsDir", ScriptsSrcDir: "new ScriptsSrcDir", SrcDir: "new SrcDir"}

			oldIODirInf.update(newIODirInf)
			So(oldIODirInf.CommandsSrcDir, ShouldEqual, "new CommandsSrcDir/")
			So(oldIODirInf.HTTPDir, ShouldEqual, "new HTTPDir/")
			So(oldIODirInf.HTTPSrcDir, ShouldEqual, "new HTTPSrcDir/")
			So(oldIODirInf.OutDir, ShouldEqual, "new OutDir/")
			So(oldIODirInf.ScriptsDir, ShouldEqual, "new ScriptsDir/")
			So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "new ScriptsSrcDir/")
			So(oldIODirInf.SrcDir, ShouldEqual, "new SrcDir/")
		})

		Convey("Given only one changed value, only that value should change", func() {
			Convey("Given an updates CommandsSrcDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{CommandsSrcDir: "CommandsSrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "CommandsSrcDir/")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates HTTPDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{HTTPDir: "HTTPDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "HTTPDir/")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})
			Convey("Given an updates HTTPSrcDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{HTTPSrcDir: "HTTPSrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "HTTPSrcDir/")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates OutDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{OutDir: "OutDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "OutDir/")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates ScriptsDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{ScriptsDir: "ScriptsDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "ScriptsDir/")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates ScriptsSrcDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{ScriptsSrcDir: "ScriptsSrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "ScriptsSrcDir/")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates SrcDir, only that value should change", func() {
				oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{SrcDir: "SrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "SrcDir/")
			})
		})
	})
}

func TestPackerInf(t *testing.T) {
	Convey("Given a PackerInf", t, func() {

		Convey("Given an empty new PackerInf", func() {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{}

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.40")
			So(oldPackerInf.Description, ShouldEqual, "test info")
		})

		Convey("Given a new MinPackerVersion", func() {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{MinPackerVersion: "0.50"}

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.50")
			So(oldPackerInf.Description, ShouldEqual, "test info")
		})

		Convey("Given a new description", func() {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{Description: "new test info"}

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.40")
			So(oldPackerInf.Description, ShouldEqual, "new test info")
		})

		Convey("Given a new MinPackerVersion and BuildName", func() {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{MinPackerVersion: "0.5.1", Description: "updated"}

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.5.1")
			So(oldPackerInf.Description, ShouldEqual, "updated")
		})
	})
}

func TestBuildInf(t *testing.T) {
	Convey("Given a BuildInf", t, func() {
		oldBuildInf := BuildInf{Name: "old Name", BuildName: "old BuildName"}
		newBuildInf := BuildInf{}

		Convey("Given a new BuildInf", func() {
			Convey("Given an empty new Name", func() {
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "old Name")
				So(oldBuildInf.BuildName, ShouldEqual, "old BuildName")
			})

			Convey("Given an empty new BuildName", func() {
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "old Name")
				So(oldBuildInf.BuildName, ShouldEqual, "old BuildName")
			})

			Convey("Given a new Name", func() {
				newBuildInf.Name = "new Name"
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "new Name")
				So(oldBuildInf.BuildName, ShouldEqual, "old BuildName")
			})

			Convey("Given a new BuildName", func() {
				newBuildInf.Name = "old Name"
				newBuildInf.BuildName = "new BuildName"
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "old Name")
				So(oldBuildInf.BuildName, ShouldEqual, "new BuildName")
			})

			Convey("Given a new Name and BuildName", func() {
				newBuildInf.Name = "Name"
				newBuildInf.BuildName = "BuildName"
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "Name")
				So(oldBuildInf.BuildName, ShouldEqual, "BuildName")
			})
		})
	})
}
