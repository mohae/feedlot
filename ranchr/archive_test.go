package ranchr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcloud/tar"
	"github.com/stretchr/testify/assert"
)

func TestDirWalk(t *testing.T) {
	tst := Archive{}
	err := tst.DirWalk("")
	assert.Nil(t, err)
	assert.Empty(t, tst.Files)

	err = tst.DirWalk("invalid/path/to/archive")
	if assert.NotNil(t, err) {
		assert.Equal(t, "invalid/path/to/archive does not exist", err.Error())
	}

	err = tst.DirWalk("../test_files/src/ubuntu/scripts/")
	assert.Nil(t, err)
	assert.Equal(t, 6, len(tst.Files))
}

func TestAddFilename(t *testing.T) {
	tst := Archive{}
	err := tst.addFilename("", "../test_files/src/ubuntu/scripts/test_file.sh", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, "../test_files/src/ubuntu/scripts/test_file.sh", tst.Files[0].p)
	assert.Nil(t, tst.Files[0].info)

	tst = Archive{}
	err = tst.addFilename("", "../test_files/src/ubuntu/scripts/dne_test_file.sh", nil, nil)
	if assert.NotNil(t, err) {
		assert.Equal(t, "../test_files/src/ubuntu/scripts/dne_test_file.sh does not exist.", err.Error())
		assert.Empty(t, tst.Files)
	}
}

func TestAddFile(t *testing.T) {
	tst := Archive{}
	filename := "../test_files/out/test1.tar"
	testFile, err := os.Create(filename)
	if assert.Nil(t, err) {
		defer testFile.Close()
		tW := tar.NewWriter(testFile)
		defer tW.Close()
		err := tst.addFile(tW, "../test_files/src/ubuntu/scripts/test_file.sh")
		if assert.Nil(t, err) {
			err := tst.addFile(tW, "../test_files/doesntExist")
			if assert.NotNil(t, err) {
				assert.Equal(t, "open ../test_files/doesntExist: no such file or directory", err.Error())
			}
		}
	}
}

func TestPriorBuild(t *testing.T) {
	tst := Archive{}
	filename := "../test_files/out/test2.tar"
	testFile, _ := os.Create(filename)
	tW := tar.NewWriter(testFile)
	err := tst.priorBuild("../test_fils/", "gzip")
	assert.Nil(t, err)
	tW.Close()

	tst = Archive{}
	filename = "../test_files/out/test3.tar"
	testFile, _ = os.Create(filename)
	tW = tar.NewWriter(testFile)
	err = tst.priorBuild("../test_files/empty/", "gzip")
	assert.Nil(t, err)

	var tarFiles []string
	files, _ := ioutil.ReadDir("../test_files/conf/")
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".gz" {
			tarFiles = append(tarFiles, file.Name())
		}
	}
	assert.Nil(t, tarFiles)
	tW.Close()

	tst = Archive{}
	filename = "../test_files/out/test4.tar"
	testFile, _ = os.Create(filename)
	tW = tar.NewWriter(testFile)
	if err := tst.priorBuild("../test_files/src/ubuntu/scripts", "gzip"); err == nil {
		assert.NotNil(t, tW)
	}
	tW.Close()

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

func TestFormattedNow(t *testing.T) {
	fmtDateTime := formattedNow()
	// This doesn't feel right, as it just replicated the actual function
	// but I don't know how else to generate the needed dynamic value
	// and doing it this way will at least detect changes to the formula.
	dateTime := time.Now().Local().Format("2006-01-02T150405Z0700")
	assert.Equal(t, fmtDateTime, dateTime)
}
