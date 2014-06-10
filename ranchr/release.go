package ranchr

import (
	"errors"
	_ "fmt"
	"io/ioutil"
	"net/http"
	_ "os"
	"strings"
)

type isoer interface {
	SetChecksum() string
	SetURL() string
	SetFilename() string
}

// Iso information. The Baseurl and ChecksumType are set during creation
type iso struct {
	BaseURL      string
	Checksum     string
	ChecksumType string
	Filename     string
	URL          string
}

// Release information. Values are set during creation
type release struct {
	iso
	Arch        string
	Distro      string
	Image       string
	Release     string
	ReleaseFull string
}

// This is a generic implementation of isoer. Makes things easier for me, though
// there is probably a better
type genericISOer struct {
	release
}

func (r *genericISOer) SetChecksum() string {
	return ""
}

func (r *genericISOer) SetURL() string {
	return ""
}

func (r genericISOer) Interface() interface{} {
	return nil
}

type ubuntu struct {
	release
}

// Sets the ISO information for a Packer template. If any error occurs, the
// error is saved to the setting variable. This will be reflected in the
// resulting Packer template, which will render it unusable until it is fixed.
func (u *ubuntu) SetISOInfo() error {
	u.SetFilename()
	if err := u.SetChecksum(); err != nil {
		return err
	}
	u.SetURL()

	return nil
}

func (u *ubuntu) SetChecksum() error {
	// Don't check for ReleaseFull existence since Release will also resolve for Ubuntu dl directories.
	// if the last character of the base url isn't a /, add it
	var page string
	var err error

	if page, err = getStringFromURL(u.BaseURL + u.Release + "/" + strings.ToUpper(u.ChecksumType) + "SUMS"); err != nil {
		logger.Error(err.Error())
		return err
	}

	// Now that we have a page...we need to find the checksum and set it
	if u.Checksum, err = u.findChecksum(page); err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (u *ubuntu) SetURL() {
	// Its ok to use Release in the directory path because Release will resolve correctly, at the directory level, for Ubuntu.
	u.URL = u.BaseURL + u.Release + "/" + u.Filename

	return
}

// findChecksum(s string, isoName string) finds the line in the incoming string with the isoName requested, strips out the checksum and returns it
// This is for releases.ubuntu.com checksums which are in a plain text file with each line representing n iso image and checksum pair, each line is in the format of:
//      checksumText image.isoname
// Notes: \n separate lines
//      since this is plain text processing we don't worry about runes
func (u *ubuntu) findChecksum(s string) (string, error) {
	pos := strings.Index(s, u.Filename)
	if pos <= 0 {
		// if it wasn't found, there's a chance that there's an extension on the release number
		// e.g. 12.04.4 instead of 12.04. This usually affects the LTS versions, I think.
		// For this look for a line  that contains .iso.
		// Substring the release string and explode it on '-'. Update isoName
		pos = strings.Index(s, ".iso")
		if pos < 0 {
			err := errors.New("Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
			logger.Error(err.Error())
			return "", err
		}
		tmpRel := s[:pos]
		tmpSl := strings.Split(tmpRel, "-")
		if len(tmpSl) < 3 {
			err := errors.New("Unable to parse release information on the Ubuntu checksum page.")
			logger.Error(err.Error())
			return "", err
		}

		u.ReleaseFull = tmpSl[1]
		u.SetFilename()

		pos = strings.Index(s, u.Filename)
		if pos < 0 {
			err := errors.New("Unable to retrieve checksum while looking for the release string on the Ubuntu checksums page.")
			logger.Error(err.Error())
			return "", err
		}
	}

	if pos - 66 < 0 {
		u.Checksum = s[:pos-2]
	} else {
		u.Checksum = s[pos-66 : pos-2]
	}

	return u.Checksum, nil
}

func (u *ubuntu) SetFilename() {
	if u.ReleaseFull == "" {
		u.Filename = "ubuntu-" + u.Release + "-" + u.Image + "-" + u.Arch + ".iso"
	} else {
		u.Filename = "ubuntu-" + u.ReleaseFull + "-" + u.Image + "-" + u.Arch + ".iso"
	}

	return
}

type CentOS struct {
	release
}

func getStringFromURL(url string) (string, error) {
	// Get the URL resource
	res, err := http.Get(url)
	if err != nil {
		logger.Critical(err)
		return "", err
	}

	// Close the response body--its idiomatic to defer it right away
	defer res.Body.Close()

	// Read the resoponse body into page
	page, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logger.Critical(err)
		return "", err
	}
	//convert the page to a string and return it
	return string(page), nil

}
