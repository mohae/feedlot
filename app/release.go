package app

import (
	"bytes"
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
	Name      string
	Operation string
	Problem   string
}

func (r ReleaseError) Error() string {
	return fmt.Sprintf("%s %s error: %s", r.Name, r.Operation, r.Problem)
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
	isoredirectURL string
}

// isoRedirectURL returns the currect url for the desired version and architecture.
func (r *centos) setISORedirectURL() {
	var buff bytes.Buffer
	buff.WriteString("http://isoredirect.centos.org/centos/")
	buff.WriteString(r.Release)
	buff.WriteString("/isos/")
	buff.WriteString(r.Arch)
	buff.WriteString("/")
	r.isoredirectURL = buff.String()
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
	r.setISORedirectURL()
	tokens, err := tokensFromURL(r.isoredirectURL)
	if err != nil {
		return setVersionInfoErr(r.isoredirectURL, fmt.Errorf("could not tokenize release page: %s", err))
	}
	links := inlineElementsFromTokens("a", "href", tokens)
	if len(links) == 0 || links == nil {
		return setVersionInfoErr(r.isoredirectURL, fmt.Errorf("could not extract links from release page"))
	}
	// filter out non-http mirror links
	links = filterLinksHasPrefix(links, []string{"http://www.", "ftp:", "http://wiki", "http://bugs", "http://bittorrent", "/"})
	if len(links) == 0 {
		return setVersionInfoErr(r.isoredirectURL, fmt.Errorf("filter of links from release page failed"))
	}
	r.BaseURL = strings.TrimSpace(links[rand.Intn(len(links)-1)])
	if strings.HasPrefix(r.Release, "6") {
		err = r.setVersion6Info()
		return err
	}
	if strings.HasPrefix(r.Release, "7") {
		err = r.setVersion7Info()
		return err
	}
	return unsupportedReleaseErr(CentOS, r.Release)
}

func (r *centos) setVersion6Info() error {
	parts := strings.Split(r.BaseURL, "/")
	if len(parts) < 7 {
		return ReleaseError{Name: CentOS.String(), Operation: "setVersion7Info", Problem: fmt.Sprintf("could not determine the current release of version %s", r.Release)}
	}
	// go through each part until we get to the version
	for _, part := range parts {
		if strings.HasPrefix(part, r.Release) {
			r.FullVersion = part
			break
		}
	}
	if r.FullVersion == "" {
		return ReleaseError{Name: CentOS.String(), Operation: "setVersion7Info", Problem: fmt.Sprintf("could not find the current point release for %s", r.Release)}
	}
	nums := strings.Split(r.FullVersion, ".")
	r.MajorVersion = nums[0]
	if len(nums) > 1 {
		r.MinorVersion = nums[1]
	}
	return nil
}

func (r *centos) setVersion7Info() error {
	// get the page from the url
	tokens, err := tokensFromURL(r.BaseURL)
	if err != nil {
		return setVersionInfoErr(r.isoredirectURL, fmt.Errorf("could not tokenize release page: %s", err))
	}
	links := inlineElementsFromTokens("a", "href", tokens)
	if len(links) == 0 || links == nil {
		return setVersionInfoErr(r.isoredirectURL, fmt.Errorf("could not extract links from release page"))
	}
	links = extractLinksHasPrefix(links, []string{fmt.Sprintf("CentOS-7-%s-%s", r.Arch, r.Image)})
	if len(links) == 0 {
		return setVersionInfoErr(r.isoredirectURL, fmt.Errorf("extract of links from release page failed"))
	}
	// extract the monthstamp and fix number this may or may not include a fix number
	parts := strings.Split(links[0], "-")
	r.MajorVersion = parts[1]
	if len(parts) > 5 {
		monthstamp := parts[4]
		tmp := strings.Split(parts[5], ".")
		r.MinorVersion = fmt.Sprintf("%s-%s", monthstamp, tmp[0])
	} else {
		tmp := strings.Split(parts[4], ".")
		r.MinorVersion = tmp[0]
	}
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
	buff.WriteString(r.Image)
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
	return fmt.Sprintf("%s%ssum.txt", r.BaseURL, strings.ToLower(r.ChecksumType))
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
		return ReleaseError{Name: r.Name, Operation: "setISOChecksum", Problem: err.Error()}
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
		return ReleaseError{Name: r.Name, Operation: "setISOChecksum", Problem: err.Error()}
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

// filterLinksHasPrefix filters out links that start with the exclude pattern
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

// extractLinksHasPrefix filters out links that start with the exclude pattern
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
