package ranchr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcloud/tar"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDirWalk(t *testing.T) {
	Convey("Testing Archive.DirWalk...", t, func() {
		tst := Archive{}
		Convey("Given an empty path, calling DirWalk", func() {
			err := tst.DirWalk("")
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Files slice should be empty", func() {
				So(tst.Files, ShouldBeEmpty)
			})
		})

		Convey("Given an invalid path, calling DirWalk", func() {
			err := tst.DirWalk("invalid/path/to/archive")
			Convey("Should result in an error", func() {
				So(err.Error(), ShouldEqual, "Unable to create a list of directory contents because the received path, invalid/path/to/archive, does not exist")
			})
		})

		Convey("Given a valid path, calling DirWalk", func() {
			err := tst.DirWalk("../test_files/src/ubuntu/scripts/")
			Convey("Should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in a slice of file paths", func() {
				So(len(tst.Files), ShouldEqual, 6)
			})
		})
	})
}

func TestAddFilename(t *testing.T) {
	Convey("Given an Archive and a filename", t, func() {
		tst := Archive{}
		Convey("Adding a path to the Files slice", func() {
			err := tst.addFilename("", "../test_files/src/ubuntu/scripts/test_file.sh", nil, nil)
			Convey("Should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in the file being added to the list", func() {
				So(tst.Files[0].p, ShouldEqual, "../test_files/src/ubuntu/scripts/test_file.sh")
				So(tst.Files[0].info, ShouldBeNil)
			})
		})
		Convey("Adding an invalid path to the Files slice", func() {
			err := tst.addFilename("", "../test_files/src/ubuntu/scripts/dne_test_file.sh", nil, nil)
			Convey("Should result in an error", func() {
				So(err.Error(), ShouldEqual, "The passed filename, ../test_files/src/ubuntu/scripts/dne_test_file.sh, does not exist. Nothing added.")
				So(tst.Files, ShouldBeEmpty)
			})
		})
	})
}

func TestAddFile(t *testing.T) {
	Convey("Given a tar writer and a filename", t, func() {
		tst := Archive{}
		Convey("Given a target archive location at ../test_files/out/test.tar", func() {
			filename := "../test_files/out/test1.tar"
			Convey("Given a create for target archive", func() {
				testFile, err := os.Create(filename)
				Convey("The file should be created", func() {
					So(err, ShouldBeNil)
					defer testFile.Close()
				})
				Convey("Given a new tar writer for the target archive file", func() {
					tW := tar.NewWriter(testFile)
					defer tW.Close()
					err := tst.addFile(tW, "../test_files/src/ubuntu/scripts/test_file.sh")
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
			})
		})
	})
}

func TestPriorBuild(t *testing.T) {
	Convey("Given an Archive", t, func() {
		tst := Archive{}
		Convey("Given a src that does not exist", func() {
			filename := "../test_files/out/test2.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			err := tst.priorBuild("../test_fils/", "gzip")
			Convey("A request to back up a non-existant directory should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			tW.Close()
		})

		Convey("back up a directory: empty directory", func() {
			tst := Archive{}
			filename := "../test_files/out/test3.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			err := tst.priorBuild("../test_files/empty/", "gzip")
			Convey("Should not result in an error", func() {
				So(err, ShouldBeNil)
			})

			Convey("Should not result in a gzip archive", func() {
				var tarFiles []string
				files, _ := ioutil.ReadDir("../test_files/conf/")
				for _, file := range files {
					if filepath.Ext(file.Name()) == ".gz" {
						tarFiles = append(tarFiles, file.Name())
					}
				}
				So(tarFiles, ShouldBeNil)
			})
			tW.Close()
		})

		Convey("back up a directory", func() {
			tst := Archive{}
			filename := "../test_files/out/test4.tar"
			testFile, _ := os.Create(filename)
			tW := tar.NewWriter(testFile)
			if err := tst.priorBuild("../test_files/src/ubuntu/scripts", "gzip"); err == nil {
				Convey("A directory was archived.", func() {
					So(tW, ShouldNotBeNil)
				})
			}
			tW.Close()
		})

	})
}

/*
	// Remove any tarballs that may be created
	files, _ := ioutil.ReadDir("../test_files/out")
	var ext string
	for _, file := range files {
		if !file.IsDir() {
			ext = filepath.Ext(file.Name())
			switch ext {
			case ".gz":
				fallthrough
			case ".tar":
				os.Remove("../test_files/out/" + file.Name())
			}
		}
	}
}
*/

func TestFormattedNow(t *testing.T) {
	Convey("Given a call to formattedNow()", t, func() {
		fmtDateTime := formattedNow()
		// This doesn't feel right, as it just replicated the actual function
		// but I don't know how else to generate the needed dynamic value
		// and doing it this way will at least detect changes to the formula.
		dateTime := time.Now().Local().Format("2006-01-02T150405Z0700")
		Convey("The string should be now in an ISO 8601 like format format w local timezone(Z). The :s have been stripped out of the time though.", func() {
			So(fmtDateTime, ShouldEqual, dateTime)
		})
	})
}
