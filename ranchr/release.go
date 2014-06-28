package ranchr

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.net/html"
	jww "github.com/spf13/jwalterweatherman"
)

func init() {
	// Psuedo-random is fine here
	rand.Seed(time.Now().UTC().UnixNano())
}

// Distro Iso information
type iso struct {
	// The BaseURL for download url formation. Usage of this is distro
	// specific.
	BaseURL string

	// The actual checksum for the ISO file that this struct represents.
	Checksum string

	// The type of the Checksum.
	ChecksumType string

	// Name of the ISO.
	Name string

	// The URL of the ISO
	isoURL string
}

// Release information. Usage of Release and ReleaseFull, along with what
// constitutes valid values, are distro dependent.
type release struct {
	iso
	Arch        string
	Distro      string
	Image       string
	Release     string
	ReleaseFull string
}

// An Ubuntu specific wrapper to release
type ubuntu struct {
	release
}

// Sets the ISO information for a Packer template.
func (u *ubuntu) SetISOInfo() error {
	// Set the ISO name.
	u.setName()

	// Set the Checksum information for this ISO.
	if err := u.setChecksum(); err != nil {
		jww.ERROR.Println(err)
		return err
	}

	// Set the URL for the ISO image.
	u.setISOURL()

	return nil
}

// Set the checksum value for the iso.
func (u *ubuntu) setChecksum() error {
	// Don't check for ReleaseFull existence since Release will also resolve
	// for Ubuntu dl directories.
	var page string
	var err error

	if page, err = getStringFromURL(appendSlash(u.BaseURL) + appendSlash(u.Release) + strings.ToUpper(u.ChecksumType) + "SUMS"); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	// Now that we have a page...we need to find the checksum and set it
	if u.Checksum, err = u.findChecksum(page); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	return nil
}

func (u *ubuntu) setISOURL() {
	// Its ok to use Release in the directory path because Release will resolve
	// correctly, at the directory level, for Ubuntu.
	u.isoURL = appendSlash(u.BaseURL) + appendSlash(u.Release) + u.Name

	return
}

// findChecksum finds the checksum in the passed page string for the current
// ISO image. This is for releases.ubuntu.com checksums which are in a plain
// text file with each line representing an iso image and checksum pair, each
// line is in the format of:
//      checksumText image.isoname
//
// Notes:
//	\n separate lines
//      since this is plain text processing we don't worry about runes
//      Ubuntu LTS images can have an additional release number, which is
//  	incremented each release. Because of this, a second search is performed
//	if the first one fails to find a match.
func (u *ubuntu) findChecksum(page string) (string, error) {
	if page == "" {
		err := errors.New("the string passed to ubuntu.findChecksum(isoName string) was empty; unable to process request")
		jww.ERROR.Println(err.Error())
		return "", err
	}

	pos := strings.Index(page, u.Name)
	if pos <= 0 {
		// if it wasn't found, there's a chance that there's an extension on the release number
		// e.g. 12.04.4 instead of 12.04. This should only be true for LTS releases.
		// For this look for a line  that contains .iso.
		// Substring the release string and explode it on '-'. Update isoName
		pos = strings.Index(page, ".iso")

		if pos < 0 {
			err := errors.New("Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
			jww.ERROR.Println(err.Error())
			return "", err
		}

		tmpRel := page[:pos]
		tmpSl := strings.Split(tmpRel, "-")

		// 3 is just an arbitrarily small number as there should always
		// be more than 3 elements in the split slice.
		if len(tmpSl) < 3 {
			err := errors.New("Unable to parse release information on the Ubuntu checksum page.")
			jww.ERROR.Println(err.Error())
			return "", err
		}

		u.ReleaseFull = tmpSl[1]
		u.setName()
		pos = strings.Index(page, u.Name)

		if pos < 0 {
			err := errors.New("Unable to retrieve checksum while looking for " + u.Name + " on the Ubuntu checksums page.")
			jww.ERROR.Println(err.Error())
			return "", err
		}
	}

	// Safety check...should never occur, but sanity check it anyways.
	if len(page) < pos-2 {
		err := errors.New("Unable to retrieve checksum information for " + u.Name)
		jww.ERROR.Println(err.Error())
		return "", err
	}

	// Get the checksum string. If the substring request goes beyond the
	// variable boundary, be safe and make the request equal to the length
	// of the string.
	if pos-66 < 1 {
		u.Checksum = page[:pos-2]
	} else {
		u.Checksum = page[pos-66 : pos-2]
	}

	return u.Checksum, nil
}

func (u *ubuntu) setName() {
	// ReleaseFull is set on LTS, otherwise just set it equal to the Release.
	if u.ReleaseFull == "" {
		u.ReleaseFull = u.Release
	}
	u.Name = "ubuntu-" + u.ReleaseFull + "-" + u.Image + "-" + u.Arch + ".iso"

	return
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (u *ubuntu) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso":

		switch u.Arch {
		case "amd64":
			return "ubuntu-64", nil
		case "i386":
			return "ubuntu-32", nil
		}

	case "virtualbox-iso":

		switch u.Arch {
		case "amd64":
			return "Ubuntu_64", nil
		case "i386":
			return "Ubuntu_32", nil
		}

	}

	// Shouldn't get here unless the buildType passed is an unsupported one.
	err := errors.New(fmt.Sprintf("release.getOSType: the builder '%s' is unsupported", buildType))
	return "", err
}

// centOS wrapper to release.
type centOS struct {
	release
}

// isoRedirectURL returns the currect url for the desired version and architecture.
func (c *centOS) isoRedirectURL() string {
	return fmt.Sprintf("http://isoredirect.centos.org/centos/%s/isos/%s/", c.Release, c.Arch)
}

// Sets the ISO information for a Packer template.
func (c *centOS) SetISOInfo() error {
	jww.TRACE.Printf("Current state: %+v", c)

	if c.Arch == "" {
		err := errors.New("Arch was empty, unable to continue")
		jww.ERROR.Println(err.Error())
		return err
	}

	if c.Release == "" {
		err := errors.New("Release was empty, unable to continue")
		jww.ERROR.Println(err.Error())
		return err
	}

	// Make sure that the version and release are set, Release and FullRelease
	// respectively. Make sure they are both set properly.
	err := c.setReleaseInfo()

	if err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	c.setName()
	err = c.setISOURL()

	if err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	// Set the Checksum information for the ISO image.
	if err := c.setChecksum(); err != nil {
		jww.ERROR.Printf("%+v", err)
		return err
	}

	return nil
}

// setReleaseInfo makes sure that both c.Release and c.ReleaseFull are properly
// set. The release number set in the file may be either the release or the
// version.
func (c *centOS) setReleaseInfo() error {
	version := strings.Split(c.Release, ".")
	// If this was a release string, it will have two parts.
	if len(version) > 1 {
		c.Release = version[0]
		return nil
	}

	// Otherwise, figure out what the current release is. The mirrorlist
	// will give us that information.
	err := c.setReleaseNumber()
	if err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	return nil
}

// releaseNumber checks the mirrorlist for the current version and arch. It
// extracts the release number, using it to set ReleaseFull.
func (c *centOS) setReleaseNumber() error {
	var page string
	var err error

	mirrorURL := fmt.Sprintf("http://mirrorlist.centos.org/?release=%s&arch=%s&repo=os", c.Release, c.Arch)

	if page, err = getStringFromURL(mirrorURL); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	// Could just parse the string, but breaking it up is simpler.
	lines := strings.Split(page, "\n")

	// Each line is an URL, split the first one to make it easier to get the version.
	urlParts := strings.Split(lines[0], "/")

	// The release is 3rd from last.
	c.ReleaseFull = urlParts[len(urlParts)-4]
	jww.TRACE.Println(c.ReleaseFull)

	return nil
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (c *centOS) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso":

		switch c.Arch {
		case "x86_64":
			return "centos-64", nil
		case "x386":
			return "centos-32", nil
		}

	case "virtualbox-iso":

		switch c.Arch {
		case "x86_64":
			return "RedHat_64", nil
		case "x386":
			return "RedHat_32", nil
		}

	}

	// Shouldn't get here unless the buildType passed is an unsupported one.
	err := errors.New(fmt.Sprintf("release.getOSType: the builder '%s' is unsupported", buildType))
	return "", err
}

// setChecksum finds the URL for the checksum page for the current mirror,
// retrieves the page, and finds the checksum for the release ISO.
func (c *centOS) setChecksum() error {
	var page string
	var err error
	url := c.checksumURL()
	jww.TRACE.Println("URL:", url)

	if page, err = getStringFromURL(url); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	jww.TRACE.Print(page)
	// Now that we have a page...we need to find the checksum and set it

	if c.Checksum, err = c.findChecksum(page); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	jww.TRACE.Println(c.Checksum)
	return nil
}

// checksumURL returns the url of the checksum page for the ISO.
func (c *centOS) checksumURL() string {
	// The base url is the same as the ISO's so strip the name and add the checksum page.
	url := trimSuffix(c.isoURL, c.Name) + strings.ToLower(c.ChecksumType) + "sum.txt"
	return url
}

// setISOURL sets the url of the ISO. If the BaseURL is set, that is used. If it
// isn't set, a isoredirect url for the ISO will be randomly selected and used.
func (c *centOS) setISOURL() error {
	if c.BaseURL != "" {
		c.isoURL = c.BaseURL + c.Name
		jww.TRACE.Println(c.isoURL)
		return nil
	}

	var err error
	c.isoURL, err = c.randomISOURL()

	if err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	return nil
}

// randomISOURL gets a random url for the current ISO.
func (c *centOS) randomISOURL() (string, error) {
	redirectURL := c.isoRedirectURL()
	jww.TRACE.Printf("redirectURL: %s", redirectURL)
	page, err := getStringFromURL(redirectURL)
	jww.TRACE.Print(page)

	if err != nil {
		jww.ERROR.Println(err.Error())
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(page))

	if err != nil {

		jww.ERROR.Println(err.Error())
		return "", err
	}

	var f func(*html.Node)
	var isoURLs []string

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					// Only add iso urls that aren't ftp, since we aren't supporting
					// checksum retrieval via ftp
					if strings.Contains(a.Val, c.Arch) && !strings.Contains(a.Val, "ftp://") {
						isoURLs = append(isoURLs, a.Val)
					}
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
		return
	}
	f(doc)

	// Randomly choose from the slice.
	url := trimSuffix(isoURLs[rand.Intn(len(isoURLs)-1)], "\n") + c.Name
	jww.TRACE.Println("ISO url: ", url)
	return url, nil
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
		jww.ERROR.Println(err.Error())
		return "", err
	}

	pos := strings.Index(page, c.Name)

	if pos < 0 {
		err := errors.New("Unable to find ISO information while looking for the release string on the CentOS checksums page.")
		jww.ERROR.Println(err.Error())
		return "", err
	}

	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "\n")
	// The checksum we want is the last element in the array
	checksum := strings.TrimSpace(tmpSl[len(tmpSl)-1])

	jww.TRACE.Println("Checksum: " + checksum)
	return checksum, nil
}

// Set the name of the ISO.
func (c *centOS) setName() {
	c.Name = "CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
	return
}

// Kind of like `wget`; return a string from the passed URL.
func getStringFromURL(url string) (string, error) {
	// Get the URL resource
	res, err := http.Get(url)

	if err != nil {
		jww.ERROR.Println(err)
		return "", err
	}

	// Close the response body--its idiomatic to defer it right away
	defer res.Body.Close()

	// Read the resoponse body into page
	page, err := ioutil.ReadAll(res.Body)

	if err != nil {
		jww.ERROR.Print(err)
		return "", err
	}

	//convert the page to a string and return it
	return string(page), nil
}
