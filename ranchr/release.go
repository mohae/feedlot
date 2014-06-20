package ranchr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	_ "os"
	"strings"
	"time"
)

func init() {
	// Psuedo-random is fine here
	rand.Seed(time.Now().UTC().UnixNano())
}

type isoer interface {
	SetChecksum() string
	SetURL() string
	SetName() string
}

// Iso information. The Baseurl and ChecksumType are set during creation
type iso struct {
	BaseURL      string
	Checksum     string
	ChecksumType string
	Name	     string
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

type ubuntu struct {
	release
}

// Sets the ISO information for a Packer template. If any error occurs, the
// error is saved to the setting variable. This will be reflected in the
// resulting Packer template, which will render it unusable until it is fixed.
func (u *ubuntu) SetISOInfo() error {
	u.setName()
	if err := u.setChecksum(); err != nil {
		return err
	}
	u.setURL()

	return nil
}

func (u *ubuntu) setChecksum() error {
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

func (u *ubuntu) setURL() {
	// Its ok to use Release in the directory path because Release will resolve correctly, at the directory level, for Ubuntu.
	u.URL = u.BaseURL + u.Release + "/" + u.Name

	return
}

// findChecksum(s string, isoName string) finds the line in the incoming string with the isoName requested, strips out the checksum and returns it
// This is for releases.ubuntu.com checksums which are in a plain text file with each line representing n iso image and checksum pair, each line is in the format of:
//      checksumText image.isoname
// Notes: \n separate lines
//      since this is plain text processing we don't worry about runes
func (u *ubuntu) findChecksum(isoName string) (string, error) {
	if isoName == "" {
		err := errors.New("the string passed to ubuntu.findChecksum(isoName string) was empty; unable to process request")
		logger.Error(err.Error())
		return "", err
	}
	pos := strings.Index(isoName, u.Name)
	if pos <= 0 {
		// if it wasn't found, there's a chance that there's an extension on the release number
		// e.g. 12.04.4 instead of 12.04. This usually affects the LTS versions, I think.
		// For this look for a line  that contains .iso.
		// Substring the release string and explode it on '-'. Update isoName
		pos = strings.Index(isoName, ".iso")
		if pos < 0 {
			err := errors.New("Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
			logger.Error(err.Error())
			return "", err
		}
		tmpRel := isoName[:pos]
		tmpSl := strings.Split(tmpRel, "-")
		if len(tmpSl) < 3 {
			err := errors.New("Unable to parse release information on the Ubuntu checksum page.")
			logger.Error(err.Error())
			return "", err
		}

		u.ReleaseFull = tmpSl[1]
		u.setName()

		pos = strings.Index(isoName, u.Name)
		if pos < 0 {
			err := errors.New("Unable to retrieve checksum while looking for " + u.Name + " on the Ubuntu checksums page.")
			logger.Error(err.Error())
			return "", err
		}
	}

	if len(isoName) < pos - 2 {
		err := errors.New("Unable to retrieve checksum information for " + u.Name)
		logger.Error(err.Error())
		return "", err
	}

	if pos - 66 < 1 {
		u.Checksum = isoName[:pos-2]
	} else {
		u.Checksum = isoName[pos-66 : pos-2]
	}

	return u.Checksum, nil
}

func (u *ubuntu) setName() {
	if u.ReleaseFull == "" {
		u.ReleaseFull = u.Release
	}
	u.Name = "ubuntu-" + u.ReleaseFull + "-" + u.Image + "-" + u.Arch + ".iso"

	return
}

func (u *ubuntu) getOSType(buildType string) string {
	// Get the OSType string for the provided builder
	// OS Type varies by distro and bit and builder.
	switch buildType {
	case "vmware-iso":
		switch u.Arch {
		case "amd64":
			return "ubuntu-64"
		case "i386":
			return "ubuntu-32"
		default:
			return "linux"
		}
	case "virtualbox-iso":
		switch u.Arch {
		case "amd64":
			return "Ubuntu_64"
		case "i386":
			return "Ubuntu_32"
		default:
			return "linux"
		}
	default:
		return "linux"
	}
	return ""
}

type centOS struct {
	release
}

// For CentOS, only the whole version is needed. The latest version will be pulled from the iso name
var centOSMirrorListURL = "http://mirrorlist.centos.org/?release=%s&arch=%s&repo=os"
// Sets the ISO information for a Packer template. If any error occurs, the
// error is saved to the setting variable. This will be reflected in the
// resulting Packer template, which will render it unusable until it is fixed.
func (c *centOS) SetISOInfo() error {
	if err := c.setBaseURL(); err != nil {
		return err
	}
	if err := c.setChecksum(); err != nil {
		return err
	}
	c.setURL()
	return nil
}

func (c *centOS) getOSType(buildType string) string {
	return buildType
	// Get the OSType string for the provided builder
	// OS Type varies by distro and bit and builder.
	switch buildType {
	case "vmware-iso":
		switch c.Arch {
		case "x86_64":
			return "centos-64"
		case "x386":
			return "centos-32"
		default:
			return "linux"
		}
	case "virtualbox-iso":
		switch c.Arch {
		case "x86_64":
			return "RedHat_64"
		case "x386":
			return "RedHat_32"
		default:
			return "linux"
		}
	default:
		return "linux"
	}
	return ""
}

func (c *centOS) setBaseURL() error {
	// Uses the release and arch information to get the list of mirrors.
	// Picks a mirror and uses that as the baseURL. This is only done if 
	// the baseURL is empty. If there is a custom value, it is assumed that
	// it represents the mirror that Rancher should use.
	var err error
	if c.BaseURL != "" {
		logger.Debug("No changes made; centOS.BaseURL was already set to " + c.BaseURL)
		return nil
	}
	if c.Arch == "" {
		err = errors.New("Unable to set BaseURL information for CentOS because the Arch was not set.")
		logger.Critical(err.Error())
		return err
	}
	if c.Release == "" {
		err = errors.New("Unable to set BaseURL information for CentOS because the Release was not set.")
		logger.Critical(err.Error())
		return err
	}
	// We only care about the version, not the release. The version release will
	// always bring up the currently supported version.
	version := strings.Split(c.Release, ".")
	mirrorURL := fmt.Sprintf(centOSMirrorListURL, version[0], c.Arch)
	var page string
	if page, err = getStringFromURL(mirrorURL); err != nil {
		logger.Error(err.Error())
		return err
	}	
	mirrors := strings.Split(page, "\n")
	mirrorCount := len(mirrors)
	if mirrorCount <= 1 {
		err = errors.New("Encountered unexpected results while processing mirror results from " + mirrorURL)
		logger.Error(err.Error())
		return err
	}
	// Use the URL provided by the list as the starting point
	c.BaseURL = mirrors[rand.Intn(mirrorCount - 1)]
	// But that url is not actually usable for our needs so get the URL structure
	// and modify it so that it works for isos.
	baseURL, err := url.Parse(c.BaseURL)
	if err != nil {
		return err
	}
	pathParts := strings.Split(baseURL.Path, "/")
	l := len(pathParts)
	path := ""
	for i := 0; i < l - 4; i++ {
		path += pathParts[i] + "/"
	}
	if path != "" {
		baseURL.Path = trimSuffix(path, "/")
	}
	// Set the ReleaseFull 
	c.ReleaseFull = pathParts[l-4]
	c.BaseURL = baseURL.String()
	c.setName()
	return nil
}

func (c *centOS) setChecksum() error {
	// Don't check for ReleaseFull existence since Release will also resolve for Ubuntu dl directories.
	// if the last character of the base url isn't a /, add it
	var page string
	var err error
	if page, err = getStringFromURL(c.BaseURL + "/" + c.ReleaseFull + "/isos/" + c.Arch + "/" + strings.ToLower(c.ChecksumType) + "sum.txt"); err != nil {
		logger.Error(err.Error())
		return err
	}
	// Now that we have a page...we need to find the checksum and set it
	if c.Checksum, err = c.findChecksum(page); err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}

func (c *centOS) setURL() {
	// The release needs to be set to the current release, not the major version, this only
	// applies if the number is a whole number.
	releaseParts := strings.Split(c.Release, ".")
	if len(releaseParts) == 1 {
		// This is just the version, replace it with a release
		urlParts := strings.Split(c.BaseURL, "/")
		// The version is the 3rd from last
		c.Release = urlParts[len(urlParts) - 1]
	}
	c.URL = c.BaseURL + "/" + c.ReleaseFull + "/isos/" + c.Arch + "/" + c.Name
	return
}

func (c *centOS) findChecksum(page string) (string, error) {
	// Finds the line in the incoming string with the isoName requested,
	// strips out the checksum and returns it. This is for CentOS checksums
	// which are in plaintext.
	//      checksumText  image.isoname
	// Notes: \n separate lines and two space separate the checksum and image name
	//      since this is plain text processing we don't worry about runes
	if page == "" {
		err := errors.New("the string passed to centOS.findChecksum(s string) was empty; unable to process request")
		logger.Error(err.Error())
		return "", err
	}
	pos := strings.Index(page, c.Name)
	if pos < 0 {
		err := errors.New("Unable to find ISO information while looking for the release string on the CentOS checksums page.")
		logger.Error(err.Error())
		return "", err
	}
	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "\n")
	// The checksum we want is the last element in the array
	checksum := strings.TrimSpace(tmpSl[len(tmpSl)-1])
	return checksum, nil
}

func (c *centOS) setName() {
	if c.ReleaseFull == "" {
		c.ReleaseFull = c.Release
	}
	c.Name = "CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
	return
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
