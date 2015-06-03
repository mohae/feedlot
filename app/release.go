package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"code.google.com/p/go.net/html"
)

func init() {
	// Psuedo-random is fine here
	rand.Seed(time.Now().UTC().UnixNano())
}

type ReleaseError struct {
	Name      string
	Operation string
	Problem   string
}

func (r ReleaseError) Error() string {
	return fmt.Sprintf("%s %s error: %s", r.Name, r.Operation, r.Problem)
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

type releaser interface {
	SetISOInfo() error
	setISOChecksum() error
	setISOURL() error
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

// centOS wrapper to release.
type centOS struct {
	release
}

// isoRedirectURL returns the currect url for the desired version and architecture.
func (c *centOS) isoRedirectURL() string {
	var buff bytes.Buffer
	buff.WriteString("http://isoredirect.centos.org/centos/")
	buff.WriteString(c.Release)
	buff.WriteString("/isos/")
	buff.WriteString(c.Arch)
	buff.WriteString("/")
	return buff.String()
}

// Sets the ISO information for a Packer template.
func (c *centOS) SetISOInfo() error {
	if c.Arch == "" {
		return NoArchError(CentOS.String())
	}
	if c.Release == "" {
		return NoReleaseError(CentOS.String())
	}
	// Make sure that the version and release are set, Release and FullRelease
	// respectively. Make sure they are both set properly.
	err := c.setReleaseInfo()
	if err != nil {
		return err
	}
	c.setISOName()
	err = c.setISOURL()
	if err != nil {
		return err
	}
	// Set the Checksum information for the ISO image.
	err = c.setISOChecksum()
	return err
}

// setReleaseInfo makes sure that both c.Release and c.ReleaseFull are properly
// set. The release number set in the file may be either the release or the
// version.
// For CentOS, the Release is an int, e.g. 6 or 7 while the ReleaseFull is
// the current release version, e.g. 6,6. When only the Release number is
// specified, Rancher will determine what the current version of the release
// is and use that as the ReleaseFull.
func (c *centOS) setReleaseInfo() error {
	version := strings.Split(c.Release, ".")
	// If this was a release string, it will have two parts.
	if len(version) > 1 {
		c.ReleaseFull = c.Release // Set release full with the release number
		c.Release = version[0]
		return nil
	}
	// Otherwise, figure out what the current release is. The mirrorlist
	// will give us that information.
	err := c.setReleaseNumber()
	if err != nil {
		return fmt.Errorf("SetReleaseInfo: %s", err.Error())
	}
	return nil
}

// releaseNumber checks the mirrorlist for the current version and arch. It
// extracts the release number, using it to set ReleaseFull.
func (c *centOS) setReleaseNumber() error {
	var page string
	var err error
	var buff bytes.Buffer
	buff.WriteString("http://mirrorlist.centos.org/?release=")
	buff.WriteString(c.Release)
	buff.WriteString("&arch=")
	buff.WriteString(c.Arch)
	buff.WriteString("&repo=os")
	mirrorURL := buff.String()
	page, err = getStringFromURL(mirrorURL)
	if err != nil {
		return err
	}
	// Could just parse the string, but breaking it up is simpler.
	lines := strings.Split(page, "\n")
	// Each line is an URL, split the first one to make it easier to get the version.
	urlParts := strings.Split(lines[0], "/")
	// if there aren't sufficient parts, generate an error from the first message
	if len(urlParts) < 4 {
		return fmt.Errorf("setReleaseNumber %s: %s", mirrorURL, urlParts[0])
	}
	// The release is 3rd from last.
	c.ReleaseFull = urlParts[len(urlParts)-4]
	return nil
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (c *centOS) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso", "vmware-vmx":
		switch c.Arch {
		case "x86_64":
			return "centos-64", nil
		case "x386":
			return "centos-32", nil
		}
	case "virtualbox-iso", "virtualbox-ovf":
		switch c.Arch {
		case "x86_64":
			return "RedHat_64", nil
		case "x386":
			return "RedHat_32", nil
		}
	}
	// Shouldn't get here unless the buildType passed is an unsupported one.
	return "", fmt.Errorf("%s does not support the %s builder", c.Distro, buildType)
}

// setISOChecksum finds the URL for the checksum page for the current mirror,
// retrieves the page, and finds the checksum for the release ISO.
func (c *centOS) setISOChecksum() error {
	if c.ChecksumType == "" {
		return fmt.Errorf("checksum type not set")
	}
	url := c.checksumURL()
	page, err := getStringFromURL(url)
	if err != nil {
		return err
	}
	// Now that we have a page...we need to find the checksum and set it
	c.Checksum, err = c.findISOChecksum(page)
	if err != nil {
		return err
	}
	return nil
}

// checksumURL returns the url of the checksum page for the ISO.
func (c *centOS) checksumURL() string {
	// The base url is the same as the ISO's so strip the name and add the checksum page.
	var buff bytes.Buffer
	buff.WriteString(trimSuffix(c.isoURL, c.Name))
	buff.WriteString(strings.ToLower(c.ChecksumType))
	buff.WriteString("sum.txt")
	return buff.String()
}

// setISOURL sets the url of the ISO. If the BaseURL is set, that is used. If
// it isn't set, a isoredirect url for the ISO will be randomly selected and
// used.
func (c *centOS) setISOURL() error {
	if c.BaseURL != "" {
		var buff bytes.Buffer
		buff.WriteString(c.BaseURL)
		if !strings.HasSuffix(c.BaseURL, "/") {
			buff.WriteString("/")
		}
		buff.WriteString(c.ReleaseFull)
		buff.WriteString("/isos/")
		buff.WriteString(c.Arch)
		buff.WriteString("/")
		buff.WriteString(c.Name)
		c.isoURL = buff.String()
		return nil
	}
	var err error
	c.isoURL, err = c.randomISOURL()
	if err != nil {
		return ReleaseError{Name: c.Name, Operation: "setISOURL", Problem: err.Error()}
	}
	return nil
}

// randomISOURL gets a random url for the current ISO.
func (c *centOS) randomISOURL() (string, error) {
	redirectURL := c.isoRedirectURL()
	page, err := getStringFromURL(redirectURL)
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(strings.NewReader(page))
	if err != nil {
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
	if len(isoURLs) < 1 {
		return "", fmt.Errorf("no valid iso URLs were found")
	}
	// Randomly choose from the slice.
	url := trimSuffix(isoURLs[rand.Intn(len(isoURLs)-1)], "\n") + c.Name
	return url, nil
}

// findISOChecksum finds the checksum in the passed page string for the current
// ISO image. This is for releases.ubuntu.com checksums which are in a plain
// text file with each line representing an iso image and checksum pair, each
// line is in the format of:
//     checksumText image.isoname
//
// Notes:
//   * \n separate lines
//   * since this is plain text processing we don't worry about runes
func (c *centOS) findISOChecksum(page string) (string, error) {
	if page == "" {
		return "", EmptyPageError(c.Name, "findISOChecksum")
	}
	pos := strings.Index(page, c.Name)
	if pos < 0 {
		return "", ChecksumNotFoundError(c.Name, "findISOChecksum")
	}
	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "\n")
	// The checksum we want is the last element in the array
	checksum := strings.TrimSpace(tmpSl[len(tmpSl)-1])
	return checksum, nil
}

// Set the name of the ISO.
func (c *centOS) setISOName() {
	var buff bytes.Buffer
	buff.WriteString("CentOS-")
	buff.WriteString(c.ReleaseFull)
	buff.WriteString("-")
	buff.WriteString(c.Arch)
	buff.WriteString("-")
	buff.WriteString(c.Image)
	buff.WriteString(".iso")
	c.Name = buff.String()
	return
}

// An Debian specific wrapper to release
type debian struct {
	release
}

// Sets the ISO information for a Packer template.
func (d *debian) SetISOInfo() error {
	if d.Arch == "" {
		return NoArchError(Debian.String())
	}
	if d.Release == "" {
		return NoReleaseError(Debian.String())
	}
	// Make sure that the version and release are set, Release and FullRelease
	// respectively. Make sure they are both set properly.
	//err := d.setReleaseInfo()
	//if err != nil {
	//	jww.ERROR.Println(err)
	//	return err
	//}
	// set ReleaseFull. if needed
	err := d.getReleaseVersion()
	if err != nil {
		return ReleaseError{Name: d.Name, Operation: "SetISOInfo", Problem: err.Error()}
	}
	d.setISOName()
	d.setISOURL()
	// Set the Checksum information for the ISO image.
	err = d.setISOChecksum()
	return err
}

// setISOChecksum: Set the checksum value for the iso.
func (d *debian) setISOChecksum() error {
	// Don't check for ReleaseFull existence since Release will also resolve
	// for Debian dl directories.
	var page string
	var err error
	page, err = getStringFromURL(appendSlash(d.BaseURL) + appendSlash(d.ReleaseFull) + "amd64/iso-cd/" + strings.ToUpper(d.ChecksumType) + "SUMS")
	if err != nil {
		return ReleaseError{Name: d.Name, Operation: "setISOChecksum", Problem: err.Error()}
	}
	// Now that we have a page...we need to find the checksum and set it
	d.Checksum, err = d.findISOChecksum(page)
	return err
}

func (d *debian) setISOURL() error {
	// If the base isn't set, use cdimage.debian.org
	if d.BaseURL == "" {
		d.BaseURL = "http://cdimage.debian.org/debian-cd/"
	}
	// Its ok to use Release in the directory path because Release will resolve
	// correctly, at the directory level, for Debian.
	d.isoURL = appendSlash(d.BaseURL) + appendSlash(d.ReleaseFull) + appendSlash(d.Arch) + appendSlash("iso-cd") + d.Name
	// This never errors so return nil...error is needed for other implementations
	return nil
}

// findISOChecksum finds the checksum in the passed page string for the current
// ISO image. This is for cdimage.debian.org/debian-cd/ checksums which are in
// a plain text file with each line representing an iso image and checksum pair,
// each line is in the format of:
//      checksumText image.isoname
//
// Notes:
//   * \n separate lines
//   * since this is plain text processing we don't worry about runes
func (d *debian) findISOChecksum(page string) (string, error) {
	if page == "" {
		return "", EmptyPageError(d.Name, "findISOChecksum")
	}
	pos := strings.Index(page, d.Name)
	if pos < 0 {
		return "", ChecksumNotFoundError(d.Name, "findISOChecksum")
	}
	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "\n")
	// The checksum we want is the last element in the array
	checksum := strings.TrimSpace(tmpSl[len(tmpSl)-1])
	return checksum, nil
}

// setISOName() sets the name of the iso for the release specified.
func (d *debian) setISOName() {
	// ReleaseFull is set on LTS, otherwise just set it equal to the Release.
	if d.ReleaseFull == "" {
		d.ReleaseFull = d.Release
	}
	var buff bytes.Buffer
	buff.WriteString("debian-")
	buff.WriteString(d.ReleaseFull)
	buff.WriteString("-")
	buff.WriteString(d.Arch)
	buff.WriteString("-")
	buff.WriteString(d.Image)
	buff.WriteString(".iso")
	d.Name = buff.String()
	return
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (d *debian) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso", "vmware-vmx":
		switch d.Arch {
		case "amd64":
			return "debian-64", nil
		case "i386":
			return "debian-32", nil
		}
	case "virtualbox-iso", "vmware-ovf":
		switch d.Arch {
		case "amd64":
			return "Debian_64", nil
		case "i386":
			return "Debian_32", nil
		}
	}
	// Shouldn't get here unless the buildType passed is an unsupported one.
	err := fmt.Errorf("%s does not support the %s builder", d.Distro, buildType)
	return "", err
}

// getReleaseVersion() get's the directory info so that the current version
// of the release can be extracted. This is abstracted out from
// d.getReleaseInfo() so that d.setReleaseInfo() can be tested. This method is
// not tested by the tests.
//
// Note: This method assumes that the baseurl will resolve to a directory
// listing that provide the information necessary to extract the current
// release: e.g. http://cdimage.debian.org/debian-cd/. If a custom url is being
// used, like for a mirror, either make sure that the releaseFull is set or
// that the url resolves to a page from which the current version can be
// extracted.
func (d *debian) getReleaseVersion() error {
	// if ReleaseFull is set, nothing to do
	if d.ReleaseFull != "" {
		return nil
	}
	p, err := getStringFromURL(d.BaseURL)
	if err != nil {
		return err
	}
	err = d.setReleaseInfo(p)
	if err != nil {
		return err
	}
	return err
}

// Since only the release is specified, the current version needs to be
// determined. For Debian, rancher can only grab the latest release as that is
// all the Debian makes available on their cdimage site.
func (d *debian) setReleaseInfo(s string) error {
	// look for the first line that starts with debian-(release)
	pos := strings.Index(s, fmt.Sprintf("a href=\"%s", d.Release))
	if pos < 0 {
		return fmt.Errorf("version search string 'a href =\"%s not found", d.Release)
	}
	// remove everything before that
	s = s[pos+8:]
	// find the next .iso, we only care about in between
	pos = strings.Index(s, "\"")
	if pos > 0 {
		s = s[:pos]
	}
	// take the next 5 chars as the release full, e.g. 7.8.0
	if len(s) < 5 {
		return fmt.Errorf("expected version string to be 5 chars: got %s", s)
	}
	d.ReleaseFull = s[:5]
	return nil
}

// An Ubuntu specific wrapper to release
type ubuntu struct {
	release
}

// SetISOInfo sets the ISO information for a Packer template.
func (u *ubuntu) SetISOInfo() error {
	if u.Arch == "" {
		return NoArchError(Ubuntu.String())
	}
	if u.Release == "" {
		return NoReleaseError(Ubuntu.String())
	}
	// Set the ISO name.
	u.setISOName()
	// Set the Checksum information for this ISO.
	err := u.setISOChecksum()
	if err != nil {
		return err
	}
	// Set the URL for the ISO image.
	u.setISOURL()
	return nil
}

// setISOChecksum: Set the checksum value for the iso.
func (u *ubuntu) setISOChecksum() error {
	// Don't check for ReleaseFull existence since Release will also resolve
	// for Ubuntu dl directories.
	var page string
	var err error
	page, err = getStringFromURL(appendSlash(u.BaseURL) + appendSlash(u.Release) + strings.ToUpper(u.ChecksumType) + "SUMS")
	if err != nil {
		return ReleaseError{Name: u.Name, Operation: "setISOChecksum", Problem: err.Error()}
	}
	// Now that we have a page...we need to find the checksum and set it
	u.Checksum, err = u.findISOChecksum(page)
	if err != nil {
		return err
	}
	return nil
}

func (u *ubuntu) setISOURL() error {
	// Its ok to use Release in the directory path because Release will resolve
	// correctly, at the directory level, for Ubuntu.
	u.isoURL = appendSlash(u.BaseURL) + appendSlash(u.Release) + u.Name
	// This never errors so return nil...error is needed for other
	// implementations of the interface.
	return nil
}

// findISOChecksum finds the checksum in the passed page string for the current
// ISO image. This is for releases.ubuntu.com checksums which are in a plain
// text file with each line representing an iso image and checksum pair, each
// line is in the format of:
//     checksumText image.isoname
//
// Notes:
//   * \n separate lines
//   * since this is plain text processing we don't worry about runes
//   * Ubuntu LTS images can have an additional release number, which is
//     incremented each release. Because of this, a second search is performed
//     if the first one fails to find a match.
func (u *ubuntu) findISOChecksum(page string) (string, error) {
	if page == "" {
		return "", EmptyPageError(u.Name, "findISOChecksum")
	}
	pos := strings.Index(page, u.Name)
	if pos <= 0 {
		// if it wasn't found, there's a chance that there's an extension on the release number
		// e.g. 12.04.4 instead of 12.04. This should only be true for LTS releases.
		// For this look for a line  that contains .iso.
		// Substring the release string and explode it on '-'. Update isoName
		pos = strings.Index(page, ".iso")
		if pos < 0 {
			return "", ReleaseError{Name: u.Name, Operation: "findISOChecksum", Problem: fmt.Sprintf("iso information not found on checksum page")}
		}
		tmpRel := page[:pos]
		tmpSl := strings.Split(tmpRel, "-")
		// 3 is just an arbitrarily small number as there should always
		// be more than 3 elements in the split slice.
		if len(tmpSl) < 3 {
			return "", ReleaseError{Name: u.Name, Operation: "findISOChecksum", Problem: fmt.Sprintf("cannot parse checksum page")}
		}
		u.ReleaseFull = tmpSl[1]
		u.setISOName()
		pos = strings.Index(page, u.Name)
		if pos < 0 {
			return "", ChecksumNotFoundError(u.Name, "findISOChecksum")
		}
	}
	// Safety check...should never occur, but sanity check it anyways.
	if len(page) < pos-2 {
		return "", ChecksumNotFoundError(u.Name, "findISOChecksum")
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

// setISOName() sets the name of the iso for the release specified.
func (u *ubuntu) setISOName() {
	// ReleaseFull is set on LTS, otherwise just set it equal to the Release.
	if u.ReleaseFull == "" {
		u.ReleaseFull = u.Release
	}
	var buff bytes.Buffer
	buff.WriteString("ubuntu-")
	buff.WriteString(u.ReleaseFull)
	buff.WriteString("-")
	buff.WriteString(u.Image)
	buff.WriteString("-")
	buff.WriteString(u.Arch)
	buff.WriteString(".iso")
	u.Name = buff.String()
	fmt.Printf("ubuntu iso: %s\n", u.Name)
	return
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (u *ubuntu) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso", "vmware-vmx":
		switch u.Arch {
		case "amd64":
			return "ubuntu-64", nil
		case "i386":
			return "ubuntu-32", nil
		}
	case "virtualbox-iso", "vmware-ovf":
		switch u.Arch {
		case "amd64":
			return "Ubuntu_64", nil
		case "i386":
			return "Ubuntu_32", nil
		}
	}
	// Shouldn't get here unless the buildType passed is an unsupported one.
	err := fmt.Errorf("%s does not support the %s builder", u.Distro, buildType)
	return "", err
}

// getStringFromURL returns the response body for the passed url as a string.
func getStringFromURL(url string) (string, error) {
	// Get the URL resource
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// Close the response body--its idiomatic to defer it right away
	defer res.Body.Close()
	// Read the resoponse body into page
	page, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	p := string(page)
	if len(p) == 0 {
		return "", EmptyPageError(url, "getSTringFromURL")
	}
	//convert the page to a string and return it
	return p, nil
}

func NoArchError(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "empty arch"}
}

func NoReleaseError(name string) error {
	return ReleaseError{Name: name, Operation: "SetISOInfo", Problem: "empty release"}
}

func EmptyPageError(name, operation string) error {
	return ReleaseError{Name: name, Operation: operation, Problem: "page was empty"}
}

func ChecksumNotFoundError(name, operation string) error {
	return ReleaseError{Name: name, Operation: operation, Problem: "checksum not found on page"}
}
