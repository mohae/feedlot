package ranchr

import (
	_"errors"
	_"fmt"
	"io/ioutil"
	_"os"
	"log"
	"net/http"
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

func (u *ubuntu) SetChecksum() string {
	// Don't check for ReleaseFull existence since Release will also resolve for Ubuntu dl directories.
		// if the last character of the base url isn't a /, add it
	if !strings.HasSuffix(u.BaseURL, "/") {
		u.BaseURL = u.BaseURL + "/"
	}

	page := getStringFromURL(u.BaseURL + "/" + u.Release + "/" + strings.ToUpper(u.ChecksumType) + "SUMS")

	// Now that we have a page...we need to find the checksum and set it
	checksum := u.findChecksum(page)

	return checksum
}

func (u *ubuntu) SetURL() string {
	// Check for ReleaseFull existence because it matters for the filename when it is a LTS version, e.g. 12.04.4 vs 12.04
	if u.ReleaseFull != "" {
		u.Filename = "ubuntu-" + u.ReleaseFull + "-" + u.Image + "-" + u.Arch + ".iso"
	} else {
		u.Filename = "ubuntu-" + u.Release + "-" + u.Image + "-" + u.Arch + ".iso"
	}

	// Its ok to use Release in the directory path because Release will resolve correctly, at the directory level, for Ubuntu.
	u.URL = u.BaseURL + "/" + u.Release + "/" + u.Filename

	return u.URL
}

// findChecksum(s string, isoName string) finds the line in the incoming string with the isoName requested, strips out the checksum and returns it
// This is for releases.ubuntu.com checksums which are in a plain text file with each line representing n iso image and checksum pair, each line is in the format of:
//      checksumText image.isoname
// Notes: \n separate lines
//      since this is plain text processing we don't worry about runes
func (u *ubuntu) findChecksum (s string) string {
	pos := strings.Index(s, u.Filename)
	if pos <= 0 {
		// if it wasn't found, there's a chance that there's an extension on the release number
		// e.g. 12.04.4 instead of 12.04. This usually affects the LTS versions, I think.
		// For this look for a line  that contains .iso.
		// Substring the release string and explode it on '-'. Update isoName
		pos = strings.Index(s, ".iso")
		if pos < 0 {
			Log.Error("Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
			return ""
		}
		tmpRel := s[:pos]
		
		tmpSl := strings.Split(tmpRel, "-")
		if len(tmpSl) < 3 {
			Log.Error("Unable to parse release information on the Ubuntu checksum page.")
			return ""
		}

		u.Release = tmpSl[1]
		
		_ = u.SetISOFilename()
		pos = strings.Index(s, u.Filename)
		if pos < 0 {
			Log.Error("Unable to retrieve checksum while looking for the release string on the Ubuntu checksums page.")
			return ""
		}
	}

	return s[pos-66 : pos-2]
}

func (u *ubuntu) SetISOFilename() string {
//	fmt.Println("\n\nUBUNTU\t%+v\n\n", ubuntu)
	u.Filename = "ubuntu-" + u.Release + "-" + u.Image + "-" + u.Arch + ".iso"
	return u.Filename
}


type CentOS struct {
	release
}

func getStringFromURL(url string) string {
	// Get the URL resource
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}

	// Close the response body--its idiomatic to defer it right away
	defer res.Body.Close()

	// Read the resoponse body into page
	page, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	//convert the page to a string and return it
	return string(page)

}
