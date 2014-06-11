package ranchr

import (
	_"archive/tar"
	_ "bytes"
	"compress/gzip"
	"errors"
	_"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/dotcloud/tar"
)

type Archive struct {
	OutDir string
	Name   string
	Type   string
	directory 
}


type directory struct {
	// This is just a struct to attach SrcWalk to. Makes keeping track of the
	// children easier
	Files []string
}

func (d *directory) SrcWalk(src string) error {
	// If the directory exists, create a tarball out of it.
	return filepath.Walk(src, d.addFilename)
}

func (d *directory) addFilename(path string, f os.FileInfo, err error) error {
	// Add a file to the slice of files for which an archive will be created.
	d.Files = append(d.Files, path)
	return nil
}

func (a *Archive) addFile(tW *tar.Writer, filename string) error {
	// Add the passed file, if it exists, to the archive, otherwise error.
	// This preserves mode and modification.
	// TODO prserve ownership
	file, err := os.Open(filename)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	defer file.Close()

	var fileStat os.FileInfo
	if fileStat, err = file.Stat(); err != nil {
		logger.Error(err.Error())
		return err
	}

	switch fileMode := fileStat.Mode(); {
	case fileMode.IsDir():
		logger.Info("Is Directory: ", filename)
		return nil
	}

	tarHeader := new(tar.Header)
	tarHeader.Name = filename
	tarHeader.Size = fileStat.Size()
	tarHeader.Mode = int64(fileStat.Mode())
	tarHeader.ModTime = fileStat.ModTime()

	// Write the file header to the tarball.
	if err := tW.WriteHeader(tarHeader); err != nil {
		logger.Error(err.Error())
		return err
	}

	// Add the file to the tarball.
	if _, err := io.Copy(tW, file); err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (a *Archive) priorBuild(src string, t string) error {
	// see if src exists, if it doesn't then don't do anything
	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			logger.Info("tarball of prior build run not needed because " + src + " does not exist")
			return nil
		}
		logger.Error(err.Error())
		return err
	}

	logger.Info("Creating tarball from " + src + " using ", t)
	if err := a.SrcWalk(src); err != nil {
		logger.Error("Walk of directory '" + src + "' failed: " + err.Error())
		return err
	}

	if len(a.Files) <= 0 {
		// This isn't a real error, just log it and return a non-error state.
		err := errors.New("No prior builds to archive.")
		logger.Error(err.Error())
		return nil
	}

//	if len(a.Files) = 1 {
		
	// Get the current date and time in RFC3339 format with custom formatting.
	nowF := formattedNow()
	tarBallName := path.Dir(a.Files[0]) + "-" + nowF + ".tar.gz"
	logger.Info(tarBallName)
	// Create the new archive file.
	tBall, err := os.Create(tarBallName)
	if err != nil {
		logger.Critical(err.Error())
		return err
	}

	defer tBall.Close()

	// Create the gzip writer.
	gw := gzip.NewWriter(tBall)
	defer gw.Close()

	// Create the tar writer.
	tW := tar.NewWriter(gw)
	defer tW.Close()

	// Go through each file in the path and add it to the archive
	for _, file := range a.Files {
		// 
		if err := a.addFile(tW, file); err != nil {
			logger.Critical(err.Error())
			return err
		}
	}

	return nil
}

func formattedNow() string {
	// Time in RFC3339 format with :s replaced with _s. This is done 
	// with seconds resolution to minimize chance of collision, how 
	// remote that may be.
	// TODO make it 8601 compliant (RFC3339 + Z)
	return strings.Replace(time.Now().Local().Format(time.RFC3339), ":", "_", -1)
}
