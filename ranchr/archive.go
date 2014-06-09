package ranchr

import (
	"archive/tar"
	_ "bytes"
	"compress/gzip"
	"errors"
	_ "fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Archive struct {
	OutDir string
	Name   string
	Type   string
	Files  []string
}

func (a *Archive) SrcWalk(src string) error {
	// If the directory exists, create a tarball out of it.
	return filepath.Walk(src, a.addFilename)
}

func (a *Archive) addFilename(path string, f os.FileInfo, err error) error {
	// Add a file to the slice of files for which an archive will be created.
	a.Files = append(a.Files, path)
	return nil
}

func (a *Archive) addFile(tW *tar.Writer, filename string) error {
	// Add the passed file, if it exists, to the archive, otherwise error.
	// This preserves mode and modification.
	// TODO prserve ownership
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		return err
	}

	tHdr := new(tar.Header)
	tHdr.Name = filename
	tHdr.Size = stat.Size()
	tHdr.Mode = int64(stat.Mode())
	tHdr.ModTime = stat.ModTime()

	// Write the file header to the tarball.
	if err := tW.WriteHeader(tHdr); err != nil {
		return err
	}

	// Add the file to the tarball.
	if _, err := io.Copy(tW, file); err != nil {
		return err
	}

	return nil
}

func (a *Archive) priorBuild(src string, t string) error {
	if err := a.SrcWalk(src); err != nil {
		logger.Error("Walk of directory '" + src + "' failed: " + err.Error())
		return err
	}

	if len(a.Files) < 0 {
		// This isn't a real error, just log it and return a non-error state.
		err := errors.New("No prior builds to archive.")
		logger.Info(err.Error())
		return nil
	}

	// Append the date and time in RFC3339 format. This is done with seconds resolution
	// to minimize chance of collision, how remote that may be.
	// TODO make it 8601 compliant (RFC3339 + Z)
	fName := path.Base(a.Files[0]) + time.Now().Local().Format(time.RFC3339) + ".tar.gz"

	// Create the new archive file.
	tBall, err := os.Create(fName)
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
		if err := a.addFile(tW, file); err != nil {
			logger.Critical(err.Error())
			return err
		}
	}

	/*
		switch t {
		case "gzip", "z", "gunzip":
			if err := a.gzipToFile(fName); err != nil {
				logger.Error(err.Error())
				return err
			}
		default:
			err := errors.New(t + " not a supported compression algorithm.")
			logger.Error(err.Error())
			return err
		}
	*/
	return nil
}

//lzip lzma lzop

/*
func (a *Archive) gzipToFile(fName string) error {
	// Archives all the files as a gzip archive using the passed name.
	fName += ".tar.gz"

	fmt.Println(fName)


	var wB bytes.Buffer

	w := gzip.NewWriter(&wB)
	wr, err := file.Create(fName)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	// Add each file
	var rB []byte

	for i, file := range a.Files {
		if rB, err = ioutil.ReadFile(file); err != nil {
			logger.Error(err.Error())
			return err
		}
		w.Write(rB)
		logger.Info(file + " was added to " + fName)
	}

	err = wr.Close()
	if err != nil {
		logger.Error(err.Error())
		return err
	}


/*
	if err := a.SrcWalk(i.OutDir); err != nil {
		logger.Warn("Archive of " + i.OutDir + " encountered an error. " + err.Error())
	}
	for i, d := range a.Files {
		fmt.Printf("%+v:%+v\n", i, d)
	}


	return nil
}
*/
