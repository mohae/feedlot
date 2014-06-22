package ranchr

import (
	"compress/gzip"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/dotcloud/tar"
)

// Hold information about an archive.
type Archive struct {
	// Path to the target directory for the archive output.
	OutDir string

	// Name of the archive.
	Name   string

	// Compression type to be used.
	Type   string

	// List of files to add to the archive.
	directory
}

// Container for files to add to an archive.
type directory struct {
	// A slice of file structs.
	Files []file
}

// Basic information about a file
type file struct {
	// The file's path
	p    string

	// The file's FileInfo
	info os.FileInfo
}

// Walk the passed path, making a list of all the files that are children of
// the path.
func (d *directory) DirWalk(dirPath string) error {
	// If the directory exists, create a list of its contents.
	fullPath, err := filepath.Abs(dirPath)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	// Set up the call back function.
	callback := func(p string, fi os.FileInfo, err error) error {
		return d.addFilename(fullPath, p, fi, err)
	}

	// Walk the tree.
	return filepath.Walk(fullPath, callback)
}

// Add the current file information to the file slice.
func (d *directory) addFilename(root string, p string, fi os.FileInfo, err error) error {
	// Add a file to the slice of files for which an archive will be created.
	logger.Tracef("BEGIN:  root: %v, path: %v, fi: %+v", root, p, fi)

	// Get the relative information.
	rel, err := filepath.Rel(root, p)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if rel == "." {
		logger.Trace("Don't add the relative root")
		return nil
	}

	// Add the file information.
	d.Files = append(d.Files, file{p: rel, info: fi})
	logger.Tracef("END relative: %v\tabs: %v", rel, p)
	return nil
}

func (a *Archive) addFile(tW *tar.Writer, filename string) error {
	// Add the passed file, if it exists, to the archive, otherwise error.
	// This preserves mode and modification.
	// TODO preserve ownership
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

	// Don't add directories--is this a good idea?
	fileMode := fileStat.Mode()
	if fileMode.IsDir() {
		logger.Trace("Is Directory: ", filename)
		return nil
	}

	// Create the tar header stuff.
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

// priorBuild handles archiving prior build artifacts, if it exists, and then
// deleting those artifacts. This prevents any stale elements from persisting
// to the new build.
func (a *Archive) priorBuild(p string, t string) error {
	logger.Debugf("t: %v\nsrc: %v", t, p)

	// See if src exists, if it doesn't then don't do anything
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			logger.Trace("processing of prior build run not needed because " + p + " does not exist")
			return nil
		}
		logger.Error(err.Error())
		return err
	}

	// Archive the old artifacts.
	if err := a.archivePriorBuild(p, t); err != nil {
		logger.Error(err.Error())
		return err
	}

	// Delete the old artifacts.
	if err := a.deletePriorBuild(p); err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (a *Archive) archivePriorBuild(p string, t string) error {
	logger.Trace("Creating tarball from "+p+" using ", t)

	// Get a list of directory contents
	if err := a.DirWalk(p); err != nil {
		logger.Error(err.Error())
		return err
	}

	if len(a.Files) <= 1 {
		// This isn't a real error, just log it and return a non-error state.
		logger.Debug("No prior builds to archive.")
		return nil
	}

	// Get the current date and time in a slightly modifie ISO 8601 format:
	// the colons are stripped from the time.
	nowF := formattedNow()
	
	// Get the relative path so that it can be added to the tarball name.
	relPath := path.Dir(p)
	// The tarball's name is the directory name + current time + extensions.
	tarBallName := relPath + a.Name + "-" + nowF + ".tar.gz"
	logger.Tracef("The files within %v will be archived and saved as %v.", p, tarBallName)

	// Create the new archive file.
	tBall, err := os.Create(tarBallName)
	if err != nil {
		logger.Critical(err.Error())
		return err
	}
	// Close the file with error handling
	defer func() {
		if cerr := tBall.Close(); cerr != nil && err == nil {
			logger.Error(cerr.Error())
			err = cerr
		}
	}()

	// The tarball gets compressed with gzip
	gw := gzip.NewWriter(tBall)
	defer gw.Close()

	// Create the tar writer.
	tW := tar.NewWriter(gw)
	defer tW.Close()

	// Go through each file in the path and add it to the archive
	var i	int
	var f 	file
	for i, f = range a.Files {
		if err := a.addFile(tW, appendSlash(relPath) + f.p); err != nil {
			logger.Critical(err.Error())
			return err
		}
	}

	logger.Tracef("Exiting priorBuild. %v files were added to the archive.", i)

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
