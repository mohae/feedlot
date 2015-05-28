package ranchr

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	jww "github.com/spf13/jwalterweatherman"
)

var timeFormat = "2006-01-02T150405Z0700"

// Archive holds information about an archive.
type Archive struct {
	// Path to the target directory for the archive output.
	OutDir string
	// Name of the archive.
	Name string
	// Compression type to be used.
	Type string
	// List of files to add to the archive.
	directory
}

// directory is a container for files to add to an archive.
type directory struct {
	// A slice of file structs.
	Files []file
}

// file contains information about a file
type file struct {
	// The file's path
	p string
	// The file's FileInfo
	info os.FileInfo
}

// DirWalk walks the passed path, making a list of all the files that are
// children of the path.
func (d *directory) DirWalk(dirPath string) error {
	// If the directory exists, create a list of its contents.
	if dirPath == "" {
		// If nothing was passed, do nothing. This is not an error.
		// However archive.Files will be nil
		jww.WARN.Println("No path information was received.")
		return nil
	}
	// See if the path exists
	exists, err := pathExists(dirPath)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	if !exists {
		err = fmt.Errorf("%s does not exist", dirPath)
		jww.ERROR.Println(err)
		return err
	}
	fullPath, err := filepath.Abs(dirPath)
	if err != nil {
		jww.ERROR.Println(err)
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
	// See if the path exists
	var exists bool
	exists, err = pathExists(p)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	if !exists {
		err = fmt.Errorf("%s does not exist", p)
		jww.ERROR.Println(err)
		return err
	}
	// Get the relative information.
	rel, err := filepath.Rel(root, p)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	if rel == "." {
		return nil
	}
	// Add the file information.
	d.Files = append(d.Files, file{p: rel, info: fi})
	return nil
}

func (a *Archive) addFile(tW *tar.Writer, filename string) error {
	// Add the passed file, if it exists, to the archive, otherwise error.
	// This preserves mode and modification.
	// TODO check ownership/permissions
	file, err := os.Open(filename)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	defer file.Close()
	var fileStat os.FileInfo
	fileStat, err = file.Stat()
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// Don't add directories--they result in tar header errors.
	fileMode := fileStat.Mode()
	if fileMode.IsDir() {
		return nil
	}
	// Create the tar header stuff.
	tH := new(tar.Header)
	tH.Name = filename
	tH.Size = fileStat.Size()
	tH.Mode = int64(fileStat.Mode())
	tH.ModTime = fileStat.ModTime()
	// Write the file header to the tarball.
	err = tW.WriteHeader(tH)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// Add the file to the tarball.
	_, err = io.Copy(tW, file)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	return nil
}

// priorBuild handles archiving prior build artifacts, if it exists, and then
// deleting those artifacts. This prevents any stale elements from persisting
// to the new build.
func (a *Archive) priorBuild(p string, t string, wg *sync.WaitGroup) error {
	defer wg.Done()
	// See if src exists, if it doesn't then don't do anything
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			jww.TRACE.Printf("processing of prior build run not needed because %s does not exist", p)
			return nil
		}
		jww.ERROR.Println(err)
		return err
	}
	// Archive the old artifacts.
	err = a.archivePriorBuild(p, t)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// Delete the old artifacts.
	err = a.deletePriorBuild(p)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	return nil
}

func (a *Archive) archivePriorBuild(p string, t string) error {
	// Get a list of directory contents
	err := a.DirWalk(p)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	if len(a.Files) <= 1 {
		// This isn't a real error, just log it and return a non-error state.
		return nil
	}
	// Get the current date and time in a slightly modifie ISO 8601 format:
	// the colons are stripped from the time.
	nowF := formattedNow()
	// Get the relative path so that it can be added to the tarball name.
	relPath := path.Dir(p)
	// The tarball's name is the directory name + current time + extensions.
	tBName := fmt.Sprintf("%s%s-%s.tar.gz", relPath, a.Name, nowF)
	// Create the new archive file.
	tBall, err := os.Create(tBName)
	if err != nil {
		jww.CRITICAL.Println(err)
		return err
	}
	// Close the file with error handling
	defer func() {
		cerr := tBall.Close()
		if cerr != nil && err == nil {
			jww.ERROR.Print(cerr)
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
	for _, f := range a.Files {
		err := a.addFile(tW, filepath.Join(relPath, f.p))
		if err != nil {
			jww.CRITICAL.Println(err)
			return err
		}
	}
	return nil
}

func (a *Archive) deletePriorBuild(p string) error {
	//delete the contents of the passed directory
	return deleteDir(p)
}

func formattedNow() string {
	// Time in ISO 8601 like format. The difference being the : have been
	// removed from the time.
	return time.Now().Local().Format(timeFormat)
}
