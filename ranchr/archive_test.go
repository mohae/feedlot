package ranchr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/dotcloud/tar"
)

func TestArchive(t *testing.T) {
	Convey("Testing Archive", t, func() {

		tmpDir := os.TempDir()
		_ = tmpDir

		Convey("return the current date time, local, in rfc3339 format", func() {
			fmtDateTime := formattedNow()
			// This doesn't feel right, as it just replicated the actual function
			// but I don't know how else to generate the needed dynamic value
			// and doing it this way will at least detect changes to the formula.
			dateTime := time.Now().Local().Format("2006-01-02T150405Z0700")
			Convey("The string should be now in an ISO 8601 like format format w local timezone(Z). The :s have been stripped out of the time though.", func() {
				So(fmtDateTime, ShouldEqual, dateTime)
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

		Convey("Given a target archive location at ../test_files/test.tar", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			Convey("Given a create for target archive", func() {
				testFile, err := os.Create(filename)
				Convey("The file should be created", func() {
					So(err, ShouldBeNil)
					defer testFile.Close()
				})
				Convey("Given a new tar writer for the target archive file", func() {
					tW := tar.NewWriter(testFile)
					defer tW.Close()
					err := tst.addFile(tW, "../test_files/scripts/test_file.sh")
					Convey("Adding a file to the archive", func() {
						Convey("Should result in no error", func() {
							So(err, ShouldBeNil)
						})
					})

					err1 := tst.addFile(tW, "../test_files/doesntExist")					
					Convey("Adding a file that doesn't exist to the archive", func() {
						Convey("Should result in an error", func() {
							So(err1.Error(), ShouldEqual, "open ../test_files/doesntExist: no such file or directory")
						})
					})
				})	
				os.Remove(filename)	
			})
		})

		Convey("back up a directory: src does not exist", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			err := tst.priorBuild("../test_fils/", "gzip")
			Convey("A request to back up a non-existant directory should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			tW.Close()
			os.Remove(filename)
		})


		Convey("back up a directory: empty directory", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			err := tst.priorBuild("../test_files/empty/", "gzip")
			Convey("Should not result in an error", func() {
					So(err, ShouldBeNil)
			})

			Convey("Should not result in a gzip archive", func() {
				var tarFiles []string
				files, _ := ioutil.ReadDir("../test_files/")
				for _, file := range files {
					if filepath.Ext(file.Name()) == ".gz" {
						tarFiles = append(tarFiles, file.Name())
					}
				}
				So(tarFiles, ShouldBeNil)
			})
			tW.Close()
			os.Remove(filename)
		})

		Convey("back up a directory", func() {
			tst := Archive{}
			filename := "../test_files/test.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.priorBuild("../test_files/scripts", "gzip"); err == nil {
				Convey("The prior build was archived.", func() {
					So(tW, ShouldNotBeNil)
				})
			}
			tW.Close()
			os.Remove(filename)
		})
					
	})	

	// Remove any tarballs that may be created
	var tarFiles []string
	files, _ := ioutil.ReadDir("../test_files/")
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".gz" {
			tarFiles = append(tarFiles, file.Name())
			// and remove it because it shouldn't be there
			_ = os.Remove(file.Name())
		}
	}
}
