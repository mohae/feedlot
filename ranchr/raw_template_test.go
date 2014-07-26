package ranchr

import (
	"testing"
	_ "time"

	. "github.com/smartystreets/goconvey/convey"
)

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

// TODO ShouldNotResemble vs ShouldResemble DeepEqual issue
func TestCreateDistroTemplate(t *testing.T) {
	Convey("Given a distro default template", t, func() {
		Convey("A distro template should be created", func() {
			r := rawTemplate{}
			r.createDistroTemplate(testDistroDefaults["ubuntu"])
			So(r, ShouldNotResemble, testDistroDefaults["ubuntu"])
		})
	})
}

func TestCreatePackerTemplate(t *testing.T) {
	Convey("Given a template", t, func() {
		r := rawTemplate{}
		r = testDistroDefaults["ubuntu"]
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
		r := rawTemplate{}
		r = testDistroDefaults["ubuntu"]
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
				So(err.Error(), ShouldEqual, "the requested builder, 'unsupported', is not supported")
				So(vars, ShouldBeNil)
				So(bldrs, ShouldBeNil)
			})
			Convey("Given no Builder Types", func() {
				r.BuilderTypes = nil
				bldrs, vars, err = r.createBuilders()
				So(err.Error(), ShouldEqual, "no builder types were configured, unable to create builders")
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

/*
// TODO check shouldnotresemble...ShouldResemble would end up with diffs which were just out
// of order map elements, not the result that should occur.
func TestCommonVMSettings(t *testing.T) {
	Convey("Given a template", t, func() {
		r := rawTemplate{}
		r = testDistroDefaults["ubuntu"]
		Convey("Given an invalid type ", func() {
			r.Type = "unknown"
			old := []string{
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
			}
			new := []string{
				"ssh_port = 222",
				"ssh_wait_timeout = 300m",
			}

		Convey("merging the setting should result in", func() {
				var settings map[string]interface{}
				var vars []string
				var err error
				settings, vars, err = r.commonVMSettings("common", old, new)
				So(err.Error(), ShouldEqual, "open :commands_dir/boot.command: no such file or directory")
				So(settings, ShouldBeNil)
				So(vars, ShouldBeNil)
			})
		})
	})

}
*/

// TODO ShouldNotResemble vs ShouldResemble DeepEqual issue
func TestMergeBuildSettings(t *testing.T) {
	Convey("Testing merging 2 build settings", t, func() {
		Convey("Given an existing Build configuration", func() {
			r := rawTemplate{}
			r = testDistroDefaults["ubuntu"]
			Convey("Merging the build should result in an updated template", func() {
				r.mergeBuildSettings(testBuilds.Build["test1"])
				So(r, ShouldNotResemble, testMergedBuilds["test1"])
			})
		})
	})
}

func TestMergeDistroSettings(t *testing.T) {
	Convey("Given an existing distro setting", t, func() {
		r := testDistroDefaults["ubuntu"]
		Convey("And some new settings", func() {
			d := testSupported.Distro["ubuntu"]
			Convey("Merging the two", func() {
				r.mergeDistroSettings(d)
				So(r, ShouldResemble, testDistroDefaults["ubuntu"])
			})
			Convey("Merging the two with a builderType", func() {
				d.BuilderTypes = []string{"virtualbox-iso", "vmware-iso"}
				expected := testDistroDefaults["ubuntu"]
				expected.BuilderTypes = d.BuilderTypes
				r.mergeDistroSettings(d)
				So(r, ShouldResemble, expected)
			})
		})
	})
}

/*
func TestScriptNames(t *testing.T) {
	Convey("Testing getting a slice of script names from the shell provisioner", t, func() {
		Convey("Given a shell provisioner", func() {
			var scripts []string
			r := rawTemplate{}
			r.Provisioners = testShellProvisioners1
			Convey("Calling rawTemplate.ScriptNames() should return a slice", func() {
				scripts = r.ScriptNames()
				So(scripts, ShouldNotBeNil)
				So(scripts, ShouldContain, "base_test.sh")
				So(scripts, ShouldContain, "setup_test.sh")
				So(scripts, ShouldContain, "vagrant_test.sh")
				So(scripts, ShouldContain, "cleanup_test.sh")
				So(scripts, ShouldContain, "zerodisk_test.sh")
				So(scripts, ShouldNotContain, "")
				So(scripts, ShouldNotContain, "zsetup.sh")
			})
		})
	})
}
*/

func TestMergeVariables(t *testing.T) {
	Convey("Given a raw template", t, func() {
		r := testDistroDefaults["ubuntu"]
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
