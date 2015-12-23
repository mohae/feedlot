package app

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func init() {
	// Psuedo-random is fine here
	rand.Seed(time.Now().UTC().UnixNano())
}

type ReleaseError struct {
	name      string
	operation string
	slug      string
}

func (r ReleaseError) Error() string {
	return fmt.Sprintf("%s %s: %s", r.name, r.operation, r.slug)
}

func emptyPageErr(name, operation string) error {
	return ReleaseError{name: name, operation: operation, slug: "page empty"}
}

func checksumNotFoundErr(name, operation string) error {
	return ReleaseError{name: name, operation: operation, slug: "checksum not found on page"}
}

func checksumNotSetErr(name string) error {
	return ReleaseError{name: name, operation: "setISOChecksum", slug: "checksum not set"}
}

func noArchErr(name string) error {
	return ReleaseError{name: name, operation: "SetISOInfo", slug: "arch not set"}
}

func noFullVersionErr(name string) error {
	return ReleaseError{name: name, operation: "SetISOInfo", slug: "full version not set"}
}

func noMajorVersionErr(name string) error {
	return ReleaseError{name: name, operation: "SetISOInfo", slug: "major version not set"}
}

func noMinorVersionErr(name string) error {
	return ReleaseError{name: name, operation: "SetISOInfo", slug: "minor version not set"}
}

func noReleaseErr(name string) error {
	return ReleaseError{name: name, operation: "SetISOInfo", slug: "release not set"}
}

func setVersionInfoErr(name string, err error) error {
	return ReleaseError{name: name, operation: "SetVersionInfo", slug: err.Error()}
}

func unsupportedReleaseErr(d Distro, name string) error {
	return fmt.Errorf("%s %s: unsupported release", d, name)
}

func osTypeBuilderErr(name, typ string) error {
	return ReleaseError{name: name, operation: "getOSType", slug: fmt.Sprintf("%s is not supported by this distro", typ)}
}

// Iso image information
type iso struct {
	// The baseURL for download url formation. Usage of this is distro
	// specific.
	BaseURL string
	// The url for the checksum page
	ReleaseURL string
	// The actual checksum for the ISO file that this struct represents.
	Checksum string
	// The type of the Checksum.
	ChecksumType string
	// Name of the ISO.
	Name string
}

func (i iso) imageURL() string {
	return fmt.Sprintf("%s%s", i.ReleaseURL, i.Name)
}

type releaser interface {
	SetISOInfo() error
	setISOChecksum() error
	setReleaseURL()
	setVersionInfo() error
}

// Release information. Usage of Release and ReleaseFull, along with what
// constitutes valid values, are distro dependent.
type release struct {
	iso
	Arch         string // iso architecture
	Distro       string // iso distro
	Image        string // iso image
	Release      string // release is the string used for the build
	MajorVersion string // iso major version
	MinorVersion string // iso minor version
	FixVersion   string // iso fix version, if applicable
	FullVersion  string // the full version number. See CentOS for example
}

// centos wrapper to release.
type centos struct {
	release
	mirrorURL string
	region    string
	country   string
	sponsor   string
}

// setMirrorURL gets a mirror url.  If region or country is set, the mirror
// list is filtered before obtaining the mirror url.  IF there is more than
// 1 url in the possible mirror list; the mirror url is psuedo-randomly
// selected.
//
// Mirrors only have  the latest release for a given version.
func (r *centos) setMirrorURL() error {
	// get the mirror list
	resp, err := http.Get("https://www.centos.org/download/full-mirrorlist.csv")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	text, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	lines := bytes.Split(text, []byte("\n"))
	for _, line := range lines {
		if len(line) == 0 {
			break
		}
		// replace \" with nothing as the csv Reader chokes on it intermitently
		// we don't actually care about the quoting here
		line = bytes.Replace(line, []byte(`\"`), []byte(""), -1)
		buf.Write(line)
		buf.WriteRune('\n')
	}
	rdr := csv.NewReader(&buf)
	rdr.LazyQuotes = true
	rdr.TrimLeadingSpace = true
	records, err := rdr.ReadAll()
	if err != nil {
		return err
	}
	// remove any records that don't have a http mirror link
	filtered := excludeRecords("", 4, records)
	// filter on region
	filtered = filterRecords(r.region, 0, filtered)
	// if sponsor is specified; filter on sponsor
	if len(r.sponsor) > 0 {
		// if OSUOSL make it Oregon State Univiersity
		if strings.ToUpper(r.sponsor) == "OSUOSL" {
			r.sponsor = "Oregon State University"
		}
		filtered = filterRecords(r.sponsor, 2, filtered)
		if len(filtered) == 0 {
			return ReleaseError{name: CentOS.String(), operation: fmt.Sprintf("filter mirror: region: %q, sponsor: %q", r.region, r.sponsor), slug: "no matches found"}
		}
		goto PICK
	}
	// filter on country
	filtered = filterRecords(r.country, 1, filtered)
	// it's an error state if everything is filtered out
	if len(filtered) == 0 {
		return ReleaseError{name: CentOS.String(), operation: fmt.Sprintf("filter mirror: region: %q, country: %q", r.region, r.country), slug: "no matches found"}
	}
PICK:
	// get a random mirror url
	tmpURL := filtered[rand.Intn(len(filtered))][4]
	r.mirrorURL = fmt.Sprintf("%s%s/isos/%s/", appendSlash(tmpURL), r.Release, r.Arch)
	return nil
}

// setVersionInfo makes sure that the version info is properly set. Pre=CentOS
// 7, the version is in the form of a point release number. With CentOS 7,
// a monthstamp was added. This func determines the current minor version and
// full version for the release.
//
// The point version may not exist. It is only populated if it is found in the
// iso name string.
func (r *centos) setVersionInfo() error {
	if r.Release == "" {
		return noReleaseErr(CentOS.String())
	}
	if !strings.HasPrefix(r.Release, "6") && !strings.HasPrefix(r.Release, "7") {
		return unsupportedReleaseErr(CentOS, r.Release)
	}
	// If the BaseURL isn't set, find a mirror to use
	if r.BaseURL == "" {
		err := r.setMirrorURL()
		if err != nil {
			return err
		}
	}
	if strings.HasPrefix(r.Release, "6") {
		err := r.setVersion6Info()
		return err
	}
	if strings.HasPrefix(r.Release, "7") {
		err := r.setVersion7Info()
		return err
	}
	return nil
}

func (r *centos) setVersion6Info() error {
	// ensure that the image is all lowercase
	r.Image = strings.ToLower(r.Image)
	tokens, err := tokensFromURL(r.mirrorURL)
	if err != nil {
		return setVersionInfoErr(r.mirrorURL, fmt.Errorf("could not determine version information: tokenizing release page: %s", err))
	}
	// get the tokens that are links
	links := inlineElementsFromTokens("a", "href", tokens)
	if len(links) == 0 || links == nil {
		return setVersionInfoErr(r.mirrorURL, fmt.Errorf("could not determine version information: extract release page links failed"))
	}
	// get the links that start with CentOS-, these have the full version number
	// and split it into it's parts
	links = extractLinksHasPrefix(links, []string{"CentOS-"})
	if len(links) == 0 {
		return setVersionInfoErr(r.mirrorURL, errors.New("could not determine version information: no CentOS iso links found"))
	}
	parts := strings.Split(links[0], "-")
	// parts should be 5 elements: e.g. CentOS-6.7-x86_64-minimal.iso
	if len(parts) < 4 {
		return setVersionInfoErr(r.mirrorURL, errors.New("could not determine version information: parse of CentOS iso links failed"))
	}
	r.FullVersion = parts[1]
	parts = strings.Split(parts[1], ".")
	if len(parts) < 2 {
		return setVersionInfoErr(r.mirrorURL, errors.New("could not determine version information: could not parse version info from CentOS iso links"))
	}
	r.MajorVersion = parts[0]
	r.MinorVersion = parts[1]
	return nil
}

func (r *centos) setVersion7Info() error {
	// the image should start with a cap. If NetInst ends up being supported
	// this will need to be revisited
	r.Image = fmt.Sprintf("%s%s", strings.ToUpper(r.Image[:1]), r.Image[1:])
	// get the page from the url
	tokens, err := tokensFromURL(r.mirrorURL)
	if err != nil {
		return setVersionInfoErr(r.mirrorURL, fmt.Errorf("could not determine version information: tokenizing release page: %s", err))
	}
	links := inlineElementsFromTokens("a", "href", tokens)
	if len(links) == 0 || links == nil {
		return setVersionInfoErr(r.mirrorURL, fmt.Errorf("could not determine version information: extract release page links failed"))
	}
	links = extractLinksHasPrefix(links, []string{"CentOS-"})
	if len(links) == 0 {
		return setVersionInfoErr(r.mirrorURL, fmt.Errorf("could not determine version information: no CentOS iso links found"))
	}
	// extract the monthstamp and fix number this may or may not include a fix number
	parts := strings.Split(links[0], "-")
	r.MajorVersion = parts[1]
	if len(parts) < 5 {
		return setVersionInfoErr(r.mirrorURL, errors.New("could not determine version information: parse of CentOS iso links failed"))
	}
	tmp := strings.Split(parts[4], ".")
	r.MinorVersion = tmp[0]
	return nil
}

// Sets the ISO information for a Packer template.
func (r *centos) SetISOInfo() error {
	if r.Arch == "" {
		return noArchErr(CentOS.String())
	}
	r.setISOName()
	r.setReleaseURL()
	return r.setISOChecksum()
}

// setISOName() sets the name of the iso for the release specified.
func (r *centos) setISOName() {
	if r.MajorVersion == "6" {
		r.setISOName6()
		return
	}
	r.setISOName7()
}

func (r *centos) setISOName6() {
	var buff bytes.Buffer
	buff.WriteString("CentOS-")
	buff.WriteString(r.FullVersion)
	buff.WriteByte('-')
	buff.WriteString(r.Arch)
	buff.WriteByte('-')
	buff.WriteString(strings.ToLower(r.Image))
	buff.WriteString(".iso")
	r.Name = buff.String()
}

func (r *centos) setISOName7() {
	var buff bytes.Buffer
	buff.WriteString("CentOS-")
	buff.WriteString(r.MajorVersion)
	buff.WriteByte('-')
	buff.WriteString(r.Arch)
	buff.WriteByte('-')
	buff.WriteString(r.Image)
	buff.WriteByte('-')
	buff.WriteString(r.MinorVersion)
	buff.WriteString(".iso")
	r.Name = buff.String()
}

// setISOChecksum finds the URL for the checksum page for the current mirror,
// retrieves the page, and finds the checksum for the release ISO.
func (r *centos) setISOChecksum() error {
	if r.ChecksumType == "" {
		return checksumNotSetErr(fmt.Sprintf("%s %s", CentOS.String(), r.Release))
	}
	url := r.checksumURL()
	page, err := bodyStringFromURL(url)
	if err != nil {
		return err
	}
	// Now that we have a page...we need to find the checksum and set it
	err = r.findISOChecksum(page)
	if err != nil {
		return err
	}
	return nil
}

func (r *centos) findISOChecksum(page string) error {
	if page == "" {
		return emptyPageErr(r.Name, "findISOChecksum")
	}
	pos := strings.Index(page, r.Name)
	if pos < 0 {
		return checksumNotFoundErr(r.Name, "findISOChecksum")
	}
	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "\n")
	// The checksum we want is the last element in the array
	r.Checksum = strings.TrimSpace(tmpSl[len(tmpSl)-1])
	return nil
}

func (r *centos) checksumURL() string {
	return fmt.Sprintf("%s%ssum.txt", r.mirrorURL, strings.ToLower(r.ChecksumType))
}

func (r *centos) setReleaseURL() {
	r.ReleaseURL = r.BaseURL
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (r *centos) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso", "vmware-vmx":
		switch r.Arch {
		case "x86_64":
			return "centos-64", nil
		case "x386":
			return "centos-32", nil
		}
	case "virtualbox-iso", "virtualbox-ovf":
		switch r.Arch {
		case "x86_64":
			return "RedHat_64", nil
		case "x386":
			return "RedHat_32", nil
		}
	}
	// Shouldn't get here unless the buildType passed is an unsupported one.
	return "", osTypeBuilderErr(CentOS.String(), buildType)
}

// An Debian specific wrapper to release
type debian struct {
	release
}

// setVersionInfo set the release information for debian. In Rancher, debian
// releases are specified as just the major version, so r.Release will have the
// same value as r.MajorVersion. Images use major.minor.fix numbering system.
func (r *debian) setVersionInfo() error {
	if r.Release == "" {
		return noReleaseErr(Debian.String())
	}
	// to find the current release number, get the index of debian-cd
	tokens, err := tokensFromURL(r.BaseURL)
	if err != nil {
		return setVersionInfoErr(r.BaseURL, err)
	}
	hrefs := inlineElementsFromTokens("a", "href", tokens)
	if len(hrefs) == 0 || hrefs == nil {
		return setVersionInfoErr(r.BaseURL, fmt.Errorf("could not tokenize release page: %s", err))
	}
	for _, href := range hrefs {
		if strings.HasPrefix(href, r.Release) {
			parts := strings.Split(href, "-")
			r.FullVersion = parts[0]
			nums := strings.Split(parts[0], ".")
			if len(nums) != 3 {
				return setVersionInfoErr(r.Release, fmt.Errorf("unable to parse release number into its parts"))
			}
			r.MajorVersion = nums[0]
			r.MinorVersion = nums[1]
			r.FixVersion = strings.TrimSuffix(nums[2], "/")
			break
		}
	}
	if r.FullVersion == "" {
		return setVersionInfoErr(r.Release, fmt.Errorf("could not set the current release number"))
	}
	r.setReleaseURL()
	return nil
}

// setReleaseURL set's the
func (r *debian) setReleaseURL() {
	var buff bytes.Buffer
	buff.WriteString(r.BaseURL)
	buff.WriteString(r.FullVersion)
	buff.WriteByte('/')
	buff.WriteString(r.Arch)
	buff.WriteString("/iso-cd/")
	r.ReleaseURL = buff.String()
}

func (r *debian) checksumURL() string {
	return fmt.Sprintf("%s%sSUMS", r.ReleaseURL, strings.ToUpper(r.ChecksumType))
}

// Sets the ISO information for a Packer template.
func (r *debian) SetISOInfo() error {
	if r.Arch == "" {
		return noArchErr(Debian.String())
	}
	r.setISOName()
	r.setReleaseURL()
	return r.setISOChecksum()
}

// setISOName() sets the name of the iso for the release specified.
func (r *debian) setISOName() {
	var buff bytes.Buffer
	buff.WriteString("debian-")
	buff.WriteString(r.FullVersion)
	buff.WriteString("-")
	buff.WriteString(r.Arch)
	buff.WriteString("-")
	buff.WriteString(r.Image)
	buff.WriteString(".iso")
	r.Name = buff.String()
	return
}

// setISOChecksum: Set the checksum value for the iso.
func (r *debian) setISOChecksum() error {
	if r.ChecksumType == "" {
		return checksumNotSetErr(fmt.Sprintf("%s %s", Debian.String(), r.Release))
	}
	page, err := bodyStringFromURL(r.checksumURL())
	if err != nil {
		return ReleaseError{name: r.Name, operation: "setISOChecksum", slug: err.Error()}
	}
	// Now that we have a page...we need to find the checksum and set it
	return r.findISOChecksum(page)
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
func (r *debian) findISOChecksum(page string) error {
	if page == "" {
		return emptyPageErr(r.Name, "findISOChecksum")
	}
	pos := strings.Index(page, r.Name)
	if pos < 0 {
		return checksumNotFoundErr(r.Name, "findISOChecksum")
	}
	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "\n")
	// The checksum we want is the last element in the array
	r.Checksum = strings.TrimSpace(tmpSl[len(tmpSl)-1])
	return nil
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (r *debian) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso", "vmware-vmx":
		switch r.Arch {
		case "amd64":
			return "debian-64", nil
		case "i386":
			return "debian-32", nil
		}
	case "virtualbox-iso", "vmware-ovf":
		switch r.Arch {
		case "amd64":
			return "Debian_64", nil
		case "i386":
			return "Debian_32", nil
		}
	}
	// Shouldn't get here unless the buildType passed is an unsupported one.
	return "", osTypeBuilderErr(Debian.String(), buildType)
}

// getReleaseVersion() get's the directory info so that the current version
// of the release can be extracted. This is abstracted out from
// d.getReleaseInfo() so that r.setReleaseInfo() can be tested. This method is
// not tested by the tests.
//
// Note: This method assumes that the baseurl will resolve to a directory
// listing that provide the information necessary to extract the current
// release: e.g. http://cdimage.debian.org/debian-cd/. If a custom url is being
// used, like for a mirror, either make sure that the releaseFull is set or
// that the url resolves to a page from which the current version can be
// extracted.
func (r *debian) getReleaseVersion() error {
	// if FullVersion is set, nothing to do
	if r.FullVersion != "" {
		return nil
	}
	p, err := bodyStringFromURL(r.BaseURL)
	if err != nil {
		return err
	}
	err = r.setReleaseInfo(p)
	if err != nil {
		return err
	}
	return err
}

// Since only the release is specified, the current version needs to be
// determined. For Debian, rancher can only grab the latest release as that is
// all the Debian makes available on their cdimage site.
func (r *debian) setReleaseInfo(s string) error {
	// look for the first line that starts with debian-(release)
	pos := strings.Index(s, fmt.Sprintf("a href=\"%s", r.Release))
	if pos < 0 {
		return fmt.Errorf("version search string 'a href =\"%s not found", r.Release)
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
	r.FullVersion = s[:5]
	return nil
}

// An Ubuntu specific wrapper to release
type ubuntu struct {
	release
}

// setVersionInfo set's the release's Version fields. Ubuntu uses a
// major.minor.seq version number, but the release number, major.minor is
// usually sufficient to get the current release as it is an alias of the full
// version number in Ubuntu release URLs.
func (r *ubuntu) setVersionInfo() error {
	if r.Release == "" {
		return noReleaseErr(Ubuntu.String())
	}
	// get the major version from the release
	parts := strings.Split(r.Release, ".")
	if len(parts) != 2 {
		return setVersionInfoErr(Ubuntu.String(), fmt.Errorf("cannot parse %q into version info", r.Release))
	}
	r.MajorVersion = parts[0]
	r.MinorVersion = parts[1]
	// Get the page for the release and extract the full version number from the
	// title. LTS support versions also have a fix number, this will ensure that
	// the correct one is obtained.
	r.setReleaseURL()
	tokens, err := tokensFromURL(r.ReleaseURL)
	if err != nil {
		return setVersionInfoErr(Ubuntu.String(), err)
	}
	elements := elementsFromTokens("title", tokens)
	if len(elements) == 0 {
		return setVersionInfoErr(Ubuntu.String(), fmt.Errorf("cannot find title on %s", r.ReleaseURL))
	}
	// get the full version from the title:
	parts = strings.Split(elements[0], " ")
	if len(parts) < 5 {
		// it's not lts
		r.FullVersion = r.Release
		return nil
	}
	// For lts, the version is part 2 of the title
	r.FullVersion = parts[1]
	return nil
}

// SetISOInfo set the ISO URL and ISO checksum information.
func (r *ubuntu) SetISOInfo() error {
	if r.Arch == "" {
		return noArchErr(Ubuntu.String())
	}
	// Set the ISO name.
	r.setISOName()
	// set the ISO URL
	r.setReleaseURL()
	// Set the Checksum information for this ISO.
	return r.setISOChecksum()
}

// setISOName() sets the name of the iso for the release specified.
func (r *ubuntu) setISOName() {
	var buff bytes.Buffer
	buff.WriteString("ubuntu-")
	buff.WriteString(r.FullVersion)
	buff.WriteString("-")
	buff.WriteString(r.Image)
	buff.WriteString("-")
	buff.WriteString(r.Arch)
	buff.WriteString(".iso")
	r.Name = buff.String()
	return
}

func (r *ubuntu) setReleaseURL() {
	var buff bytes.Buffer
	buff.WriteString(appendSlash(r.BaseURL))
	buff.WriteString(r.Release)
	buff.WriteByte('/')
	r.ReleaseURL = buff.String()
}

// setISOChecksum: Set the checksum value for the iso. Most of the actual work
// is done in findISOChecksum for testability reasons.
func (r *ubuntu) setISOChecksum() error {
	if r.ChecksumType == "" {
		return checksumNotSetErr(fmt.Sprintf("%s %s", Ubuntu.String(), r.Release))
	}
	page, err := bodyStringFromURL(r.checksumURL())
	if err != nil {
		return ReleaseError{name: r.Name, operation: "setISOChecksum", slug: err.Error()}
	}
	return r.findISOChecksum(page)
}

func (r *ubuntu) findISOChecksum(page string) error {
	// Now that we have a page...we need to find the checksum and set it
	if page == "" {
		return emptyPageErr(r.Name, "findISOChecksum")
	}
	pos := strings.Index(page, r.Name)
	if pos <= 0 {
		return checksumNotFoundErr(r.Name, "findISOChecksum")
	}
	tmpRel := page[:pos]
	tmpSl := strings.Split(tmpRel, "-")
	// the slice should contain 4 elements, unless Ubuntu has changed their naming
	// pattern .
	if len(tmpSl) < 4 {
		return checksumNotFoundErr(r.Name, "findISOChecksum")
	}
	pos = strings.Index(page, r.Name)
	if pos < 0 {
		return checksumNotFoundErr(r.Name, "findISOChecksum")
	}
	// Safety check...should never occur, but sanity check it anyways.
	if len(page) < pos-2 {
		return checksumNotFoundErr(r.Name, "findISOChecksum")
	}
	// Get the checksum string. If the substring request goes beyond the
	// variable boundary, be safe and make the request equal to the length
	// of the string.
	if pos-66 < 1 {
		r.Checksum = page[:pos-2]
	} else {
		r.Checksum = page[pos-66 : pos-2]
	}
	return nil
}

func (r *ubuntu) checksumURL() string {
	return fmt.Sprintf("%s%sSUMS", r.ReleaseURL, strings.ToUpper(r.ChecksumType))
}

// getOSType returns the OSType string for the provided builder. The OS Type
// varies by distro, arch, and builder.
func (r *ubuntu) getOSType(buildType string) (string, error) {
	switch buildType {
	case "vmware-iso", "vmware-vmx":
		switch r.Arch {
		case "amd64":
			return "ubuntu-64", nil
		case "i386":
			return "ubuntu-32", nil
		}
	case "virtualbox-iso", "vmware-ovf":
		switch r.Arch {
		case "amd64":
			return "Ubuntu_64", nil
		case "i386":
			return "Ubuntu_32", nil
		}
	}
	return "", osTypeBuilderErr(Ubuntu.String(), buildType)
}

// bodyStringFromURL returns the response body for the passed url as a string.
func bodyStringFromURL(url string) (string, error) {
	// Get the URL resource
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	// Close the response body--its idiomatic to defer it right away
	defer res.Body.Close()
	// Read the response body into page
	page, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if len(page) == 0 {
		return "", emptyPageErr(url, "bodyStringFromURL")
	}
	return string(page), nil
}

// tokensFromURL returns a slice of tokens from the specified url, or an error.
func tokensFromURL(url string) ([]html.Token, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// probably unnecessary, but we know that, unless there was a problem, the
	// number of tokens will always be > than 64
	tokens := make([]html.Token, 0, 64)
	tizer := html.NewTokenizer(resp.Body)
	for {
		typ := tizer.Next()
		if typ == html.ErrorToken {
			return tokens, nil
		}
		tokens = append(tokens, tizer.Token())
	}
}

// elementsFromTokens will return the contents of the specified html tag as
// a list. This function assumes that an element has its begin tag, content,
// and end tag separated by a newline; each is its own token. It does not
// handle inline elements,  elements whose begin tag, content, and end tag are
// all part of the same line, or token.
func elementsFromTokens(name string, tokens []html.Token) []string {
	var content []string
	for i, token := range tokens {
		if token.Type == html.StartTagToken && token.DataAtom.String() == name {
			// the next token is the content of the token
			content = append(content, tokens[i+1].Data)
		}
	}
	return content
}

func inlineElementsFromTokens(element, attrVal string, tokens []html.Token) []string {
	var found []string
	for _, token := range tokens {
		if token.Type == html.StartTagToken && token.DataAtom.String() == element {
			// if we aren't filtering by the attribute, just grab the value
			for _, attr := range token.Attr {
				if attrVal != "" {
					if attr.Key != attrVal {
						continue
					}
					found = append(found, attr.Val)
				}
			}
		}
	}
	return found
}

// filterLinksHasPrefix filters out links that start with the prefix and
// returns the remaining links
func filterLinksHasPrefix(links, prefixes []string) []string {
	var filtered []string
	for _, link := range links {
		for _, prefix := range prefixes {
			if strings.HasPrefix(link, prefix) {
				goto nextLink
			}
		}
		filtered = append(filtered, link)
	nextLink:
	}
	return filtered
}

// extractLinksHasPrefix returns the links that start with the prefix
func extractLinksHasPrefix(links, prefixes []string) []string {
	var extracted []string
	for _, link := range links {
		for _, prefix := range prefixes {
			if strings.HasPrefix(link, prefix) {
				extracted = append(extracted, link)
			}
		}
	}
	return extracted
}

// filterRecords filters the received records by comparing the value to match
// with the value in  the field n
func filterRecords(v string, n int, records [][]string) [][]string {
	// if the string is empty, no filtering is done
	if len(v) == 0 {
		return records
	}
	var filtered [][]string
	for _, record := range records {
		if record[n] == v {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// excludeRecords excludes records whose field, n, matches v.  The remaining
// records are returned
func excludeRecords(v string, n int, records [][]string) [][]string {
	var filtered [][]string
	for _, record := range records {
		if record[n] == v {
			continue
		}
		filtered = append(filtered, record)
	}
	return filtered
}
