package ranchr

import (
	"testing"
	_"time"

	. "github.com/smartystreets/goconvey/convey"
)



func TestRawTemplate(t *testing.T) {
	Convey("Testing RawTemplate", t, func() {
		Convey("Given a request for a newRawTemplate()", func() {
			rawTpl := newRawTemplate()
			Convey("The raw template should equal--we don't test the date because it is always changeing", func() {
				So(rawTpl, ShouldResemble, testRawTemplate)
			})
		})
	})
}

func TestIODirInf(t *testing.T) {

	Convey("Given a IODirInf", t, func() {

		Convey("Given an empty new IODirInf", func () {
			oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
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

		Convey("Given a populated new IODirInf", func () {
			oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
			newIODirInf := IODirInf{CommandsSrcDir:"new CommandsSrcDir", HTTPDir: "new HTTPDir", HTTPSrcDir: "new HTTPSrcDir", OutDir:"new OutDir", ScriptsDir:"new ScriptsDir", ScriptsSrcDir: "new ScriptsSrcDir", SrcDir: "new SrcDir"}

			oldIODirInf.update(newIODirInf)
			So(oldIODirInf.CommandsSrcDir, ShouldEqual, "new CommandsSrcDir")
			So(oldIODirInf.HTTPDir, ShouldEqual, "new HTTPDir")
			So(oldIODirInf.HTTPSrcDir, ShouldEqual, "new HTTPSrcDir")
			So(oldIODirInf.OutDir, ShouldEqual, "new OutDir")
			So(oldIODirInf.ScriptsDir, ShouldEqual, "new ScriptsDir")
			So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "new ScriptsSrcDir")
			So(oldIODirInf.SrcDir, ShouldEqual, "new SrcDir")
		})

		Convey("Given only one changed value, only that value should change", func () {
			Convey("Given an updates CommandsSrcDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{CommandsSrcDir :"CommandsSrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates HTTPDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{HTTPDir:"HTTPDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})
			Convey("Given an updates HTTPSrcDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{HTTPSrcDir: "HTTPSrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})
		
			Convey("Given an updates OutDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{OutDir: "OutDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "old ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates ScriptsDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{ScriptsDir: "ScriptsDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsDir, ShouldEqual, "ScriptsDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates ScriptsSrcDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{ScriptsSrcDir: "ScriptsSrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "old SrcDir")
			})

			Convey("Given an updates SrcDir, only that value should change", func () {
				oldIODirInf := IODirInf{CommandsSrcDir:"old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir:"old OutDir", ScriptsDir:"old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
				newIODirInf := IODirInf{SrcDir: "SrcDir"}

				oldIODirInf.update(newIODirInf)
				So(oldIODirInf.CommandsSrcDir, ShouldEqual, "old CommandsSrcDir")
				So(oldIODirInf.HTTPDir, ShouldEqual, "old HTTPDir")
				So(oldIODirInf.HTTPSrcDir, ShouldEqual, "old HTTPSrcDir")
				So(oldIODirInf.OutDir, ShouldEqual, "old OutDir")
				So(oldIODirInf.ScriptsSrcDir, ShouldEqual, "old ScriptsSrcDir")
				So(oldIODirInf.SrcDir, ShouldEqual, "SrcDir")
			})
		})
	})
}

func TestPackerInf(t *testing.T) {
	Convey("Given a PackerInf", t, func() {

		Convey("Given an empty new PackerInf", func () {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{}			

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.40")
			So(oldPackerInf.Description, ShouldEqual, "test info")
		})


		Convey("Given a new MinPackerVersion", func () {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{MinPackerVersion:"0.50"}			

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.50")
			So(oldPackerInf.Description, ShouldEqual, "test info")
		})

		Convey("Given a new description", func () {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{Description:"new test info"}			

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.40")
			So(oldPackerInf.Description, ShouldEqual, "new test info")
		})	

		Convey("Given a new MinPackerVersion and BuildName", func () {
			oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
			newPackerInf := PackerInf{MinPackerVersion:"0.5.1", Description:"updated"}			

			oldPackerInf.update(newPackerInf)
			So(oldPackerInf.MinPackerVersion, ShouldEqual, "0.5.1")
			So(oldPackerInf.Description, ShouldEqual, "updated")
		})
	})
}

func TestBuildInf(t *testing.T) {
	Convey("Given a BuildInf", t, func() {
		oldBuildInf := BuildInf{Name:"old Name", BuildName: "old BuildName"}
		newBuildInf := BuildInf{}

		Convey("Given a new BuildInf", func() {
			Convey("Given an empty new Name", func () {
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "old Name")
				So(oldBuildInf.BuildName, ShouldEqual, "old BuildName")
			})

			Convey("Given an empty new BuildName", func () {
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "old Name")
				So(oldBuildInf.BuildName, ShouldEqual, "old BuildName")
			})

			Convey("Given a new Name", func () {
				newBuildInf.Name = "new Name"
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "new Name")
				So(oldBuildInf.BuildName, ShouldEqual, "old BuildName")
			})

			Convey("Given a new BuildName", func () {
				newBuildInf.Name = "old Name"
				newBuildInf.BuildName = "new BuildName"
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "old Name")
				So(oldBuildInf.BuildName, ShouldEqual, "new BuildName")
			})	

			Convey("Given a new Name and BuildName", func () {
				newBuildInf.Name = "Name"
				newBuildInf.BuildName = "BuildName"
				oldBuildInf.update(newBuildInf)
				So(oldBuildInf.Name, ShouldEqual, "Name")
				So(oldBuildInf.BuildName, ShouldEqual, "BuildName")
			})
		})
	})
}

