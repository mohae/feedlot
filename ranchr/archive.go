package ranchr

import (
	_"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	_"strings"
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
	Files []file
}

type file struct {
	p string
	info os.FileInfo
}

func (d *directory) DirWalk(dirPath string) error {
	// If the directory exists, create a list of its contents.
	fullPath, err := filepath.Abs(dirPath)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	callback := func(p string, fi os.FileInfo, err error) error {
		return d.addFilename(fullPath, p, fi, err)
	}
	
	return filepath.Walk(fullPath, callback)
}

func (d *directory) addFilename(root string, p string, fi os.FileInfo, err error) error {
	// Add a file to the slice of files for which an archive will be created.
	if !fi.IsDir() {
		return nil
	}
	rel, err := filepath.Rel(root, p)
	if err != nil {
		logger.Error(err.Error())
		return err
	}
	if rel == "." {
		logger.Debug("Don't add the relative root")
		return nil
	}
	d.Files = append(d.Files, file{p: rel, info: fi})
	logger.Tracef("relative: %v\tabs: %v", rel, p)
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
		logger.Trace("Is Directory: ", filename)
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

func (a *Archive) priorBuild(p string, t string) error {
	// Ff the src directory exists, an archive is created
	// and the directory is deleted.
	logger.Debugf("t: %v\nsrc: %v", t, p)
	// see if src exists, if it doesn't then don't do anything
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			logger.Trace("processing of prior build run not needed because " + p + " does not exist")
			return nil
		}
		logger.Error(err.Error())
		return err
	}
	if err := a.archivePriorBuild(p, t); err != nil {
		logger.Error(err.Error())
		return err
	}
	if err := a.deletePriorBuild(p); err != nil {
		logger.Error(err.Error())
		return err
	}
	return nil
}

func (a *Archive) archivePriorBuild(p string, t string) error {
	logger.Debug("Creating tarball from " + p + " using ", t)
	// SrcWalk, as written will always return nil
	if err := a.DirWalk(p); err != nil {
		logger.Trace(err.Error())
		return err
	}

	logger.Debugf("The archive files were: %v", a.Files)
	if len(a.Files) <= 1 {
		// This isn't a real error, just log it and return a non-error state.
		err := errors.New("No prior builds to archive.")
		logger.Error(err.Error())
		return nil
	}
	// Get the current date and time in RFC3339 format with custom formatting.
	nowF := formattedNow()
	relPath := path.Dir(p)
	tarBallName := relPath + a.Name + "-" + nowF + ".tar.gz"
	logger.Debugf("The files within %v will be archived and saved as %v.", p, tarBallName)
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
	var cnt int	
	for _, file := range a.Files {
		// 
		if err := a.addFile(tW, appendSlash(relPath) + file.p); err != nil {
			logger.Critical(err.Error())
			return err
		}
		cnt++
	}
	logger.Tracef("Exiting priorBuild. %v files were added to the archive.", cnt)
	return nil
}

func (a *Archive) deletePriorBuild(p string) error {
	//delete the contents of the passed directory
	return deleteDirContent(p)
}

func formattedNow() string {
	// Time in ISO 8601 like format. The difference being the : have been
	// removed from the time.
	return time.Now().Local().Format("2006-01-02T150405Z0700")
}
