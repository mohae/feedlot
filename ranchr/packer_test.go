package ranchr

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Testcreate(t *testing.T) {
	Convey("Given a PackerTemplate", t, func() {
		p := packerTemplate{}
		Convey("Given an IODirInf, BuildInf, and scripts slice", func() {
			var scripts []string
			b := BuildInf{BuildName: "test build"}
			Convey("Calling TemplateToJSON without IODirInf.HTTPDir set", func() {
				i := IODirInf{}
				err := p.create(i, b, scripts)
				So(err.Error(), ShouldEqual, "ranchr.create: HTTPDir directory for "+b.BuildName+" not set")
			})
			Convey("Calling TemplateToJSON without IODirInf.HTTPSrcDir set", func() {
				i := IODirInf{HTTPDir: "http"}
				err := p.create(i, b, scripts)
				So(err.Error(), ShouldEqual, "ranchr.create: HTTPSrcDir directory for "+b.BuildName+" not set")
			})
			Convey("Calling TemplateToJSON without IODirInf.OutDir set", func() {
				i := IODirInf{HTTPDir: "http", HTTPSrcDir: "../test_files/http/"}
				err := p.create(i, b, scripts)
				So(err.Error(), ShouldEqual, "ranchr.create: output directory for "+b.BuildName+" not set")
			})
			Convey("Calling TemplateToJSON without IODirInf.SrcDir set", func() {
				i := IODirInf{HTTPDir: "http", HTTPSrcDir: "../test_files/http/", OutDir: "../test_files/out/"}
				err := p.create(i, b, scripts)
				So(err.Error(), ShouldEqual, "ranchr.create: SrcDir directory for "+b.BuildName+" not set")
			})

			Convey("Calling TemplateToJSON without IODirInf.ScriptsDir set", func() {
				i := IODirInf{HTTPDir: "http", HTTPSrcDir: "../test_files/http/", OutDir: "../test_files/out/", SrcDir: "../test_files/"}
				err := p.create(i, b, scripts)
				So(err.Error(), ShouldEqual, "ranchr.create: ScriptsDir directory for "+b.BuildName+" not set")
			})

			Convey("Calling TemplateToJSON without IODirInf.ScriptsSrcDir set", func() {
				i := IODirInf{HTTPDir: "http", HTTPSrcDir: "../test_files/http/", OutDir: "../test_files/out/", SrcDir: "../test_files/", ScriptsDir: "scripts"}

				err := p.create(i, b, scripts)
				So(err.Error(), ShouldEqual, "ranchr.create: ScriptsSrcDir directory for "+b.BuildName+" not set")
			})

			Convey("Calling TemplateToJSON with IODirInf and an output directory with contents", func() {
				i := IODirInf{HTTPDir: "http", HTTPSrcDir: "../test_files/src/ubuntu/http/", OutDir: "../test_files/out/build/", SrcDir: "../test_files/src/ubuntu/", ScriptsDir: "scripts", ScriptsSrcDir: "../test_files/src/ubuntu/scripts/"}
				Scripts := []string{"cleanup_test.sh", "setup_test.sh", "test_file.sh"}
				err := p.create(i, b, Scripts)
				So(err, ShouldBeNil)
			})

			Convey("Calling TemplateToJSON with IODirInf; requesting shell scripts that do not exist", func() {
				i := IODirInf{HTTPDir: "http", HTTPSrcDir: "../test_files/http/", OutDir: "../test_files/out/build/", SrcDir: "../test_files/", ScriptsDir: "scripts", ScriptsSrcDir: "../test_files/scripts/"}
				Scripts := []string{"cleanup_test.sh", "setup_test.sh", "not_there.sh", "missing.sh", "test_file.sh"}
				err := p.create(i, b, Scripts)
				So(err.Error(), ShouldEqual, "open ../test_files/scripts/test_file.sh: no such file or directory")
			})

		})
	})
}
