package ranchr

import (
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/dotcloud/tar"
)

func TestArchive(t *testing.T) {
	Convey("Testing Archive", t, func() {
		Convey("return the current date time, local, in rfc3339 format", func() {
			fmtDateTime := formattedNow()
			
			Convey("The string should be now in rfc3339 format w local timezone. The : are replaced with _. Only test for not empty since this value changes.", func() {
				So(fmtDateTime, ShouldNotEqual, "")
			})
		})


		Convey("Get a slice of paths within a directory", func() {
			tst := Archive{}
			err := tst.SrcWalk("../test_files/scripts")
			if err == nil {
				Convey("The files within the walked path should be:", func() {
					So(tst.Files, ShouldNotBeEmpty)
				})
			}
		})

		Convey("add a path to Files slice", func() {
			tst := Archive{}
			tst.addFilename("test/file", nil, nil)
			Convey("The path slice should have 'test/file'", func() {
				So(tst.Files, ShouldContain, "test/file")
			})
		})

		Convey("add a file to a tar.Writer", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.addFile(tW, "../test_files/scripts/test_file.sh"); err == nil {
				Convey("The file was added", func() {
					So(tW, ShouldNotBeNil)
				})
			}
	
			Convey("add a file that doesn't exist", func() {
				if err := tst.addFile(tW, "../test_files/doesntExist"); err == nil {
					Convey(err, ShouldEqual, "z")
				}
			})

			Convey("add a file with permission issues", func() {
				if err := tst.addFile(tW, "../test_files/no_permissions.txt"); err == nil {
					Convey(err, ShouldEqual, "z")
				}
			})

			tW.Close()
			os.Remove(filename)
		})

		Convey("back up a directory: src does not exist", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.priorBuild("../test_fils/", "gzip"); err != nil {
				Convey("The prior build was archived.", func() {
					So(err, ShouldEqual, "ADFSA")
				})
			}
			tW.Close()
			os.Remove(filename)
		})

		Convey("back up a directory: empty directory", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.priorBuild("../test_files/empty/", "gzip"); err != nil {
				Convey("The prior build was archived.", func() {
					So(err, ShouldEqual, "ADFSA")
				})
			}
			tW.Close()
			os.Remove(filename)
		})

		Convey("back up a directory: directory doesn't exist", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.priorBuild("../test_files/3empty2/", "gzip"); err != nil {
				Convey("The prior build was archived.", func() {
					So(err, ShouldEqual, "ADFSA")
				})
			}
			tW.Close()
			os.Remove(filename)
		})

		Convey("back up a directory", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.priorBuild("../test_files/", "gzip"); err == nil {
				Convey("The prior build was archived.", func() {
					So(tW, ShouldNotBeNil)
				})
			}
			tW.Close()
			os.Remove(filename)
		})
					
	})	
}
