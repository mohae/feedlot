// Some of these tests query the actual URLs; as such they may fail if the
// remote server is not available or if the destination url is no longer
// valid.
//
// The risk of tests failing due to remote not being availabe was deemed
// acceptable. If the destination url is no longer valid, or the assumptions
// made in the code lead to error, this is fine as the code should be updated
// to reflect the changes in remote.
package app

import (
	"fmt"
	"strings"
	"testing"
)

func newTestCentOS() centos {
	c := centos{release: release{Release: "6", Image: "Minimal", Arch: "x86_64"}, region: "", country: "CA"}
	c.ChecksumType = "sha256"
	return c
}

func TestCentOSsetReleaseInfo(t *testing.T) {
	tests := []struct {
		region      string
		country     string
		sponsor     string
		release     string
		fullVersion string
		expMajor    string
		minor       string
		expError    string
	}{
		{"", "", "", "", "", "", "", "release not set"},
		{"", "IL", "", "6", "", "6", "", ""}, // minor is empty because it may chagne with a new release
		{"", "IL", "", "6", "6.6", "6", "6", ""},
		{"", "IL", "", "6", "6.6", "6", "6", ""},
		{"", "IL", "", "7", "", "7", "", ""},
		{"North America", "", "", "7", "", "7", "", ""},
		{"US", "", "Oregon State University", "7", "", "7", "", ""},
		{"", "", "osuosl", "7", "", "7", "", ""},
		{"", "", "OSUOSL", "7", "", "7", "", ""},
		{"", "", "Rackspace", "7", "", "7", "", ""},
		{"", "ZZ", "", "7", "", "7", "", "filter mirror: region: \"\", country: \"ZZ\": no matches found"},
		{"", "IL", "", "8", "", "", "", "CentOS 8: not supported"},
	}
	c := newTestCentOS()
	for i, test := range tests {
		c.Release = test.release
		c.region = test.region
		c.country = test.country
		c.sponsor = test.sponsor
		err := c.setVersionInfo()
		if err != nil {
			if err.Error() != test.expError {
				t.Errorf("%d: got %q; want %q", i, err, test.expError)
			}
			continue
		}
		if test.expError != "" {
			t.Errorf("%d: got no error, want %q", i, test.expError)
			continue
		}
		if c.MajorVersion != test.expMajor {
			t.Errorf("%d: got %s, want %s", i, c.MajorVersion, test.expMajor)
			continue
		}
		// minor should not be empty
		if len(c.MinorVersion) == 0 {
			t.Errorf("%d: minorVersion was empty, want a value", i)
			continue
		}
		if c.MinorVersion == test.minor {
			t.Errorf("%d: minor versions should not match; it should be the latest minor version, not %s", i, c.MinorVersion)
		}
	}
}

func TestCentOSGetOSType(t *testing.T) {
	tests := []struct {
		buildType Builder
		arch      string
		expected  string
		err       string
	}{
		{VMWareISO, "x86_64", "centos-64", ""},
		{VMWareISO, "x386", "centos-32", ""},
		{VMWareVMX, "x86_64", "centos-64", ""},
		{VMWareVMX, "x386", "centos-32", ""},
		{VirtualBoxISO, "x86_64", "RedHat_64", ""},
		{VirtualBoxISO, "x386", "RedHat_32", ""},
		{VirtualBoxOVF, "x86_64", "RedHat_64", ""},
		{VirtualBoxOVF, "x386", "RedHat_32", ""},
		{QEMU, "x86_64", "", ""},
		{QEMU, "x386", "", ""},
		{UnsupportedBuilder, "x86_64", "", fmt.Sprintf("CentOS %s: not supported", UnsupportedBuilder)},
		{UnsupportedBuilder, "x386", "", fmt.Sprintf("CentOS %s: not supported", UnsupportedBuilder)},
	}
	for i, test := range tests {
		c := centos{release{Arch: test.arch}, "", "", ""}
		res, err := c.getOSType(test.buildType)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("%d: got %q want %q", i, err, test.err)
			}
			continue
		}
		if test.err != "" {
			t.Errorf("%d: got no error, want %q", i, test.err)
			continue
		}
		if res != test.expected {
			t.Errorf("%d: got %s, want %s", i, res, test.expected)
		}
	}
}

func TestCentOSSetISOChecksum(t *testing.T) {
	c := newTestCentOS()
	c.setVersionInfo()
	c.ChecksumType = ""
	err := c.setISOChecksum()
	if err == nil {
		t.Errorf("Got nil, want %q", ErrChecksumTypeNotSet)
	} else {
		if err != ErrChecksumTypeNotSet {
			t.Errorf("got %q, want %q", err, ErrChecksumTypeNotSet)
		}
	}

	c = newTestCentOS()
	c.Arch = "x86_64"
	c.setVersionInfo()
	c.setISOName()
	c.SetISOInfo()
	err = c.setISOChecksum()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		// only check to see if its empty, checking the actual checksum will cause failure
		// if the iso image has been updated since last test update.
		if c.Checksum == "" {
			t.Error("Expected checksum to not be empty, it was")
		}
	}
}

func TestCentOSSetChecksumURL(t *testing.T) {
	c := newTestCentOS()
	c.setVersionInfo()
	c.SetISOInfo()
	checksumURL := c.checksumURL()
	if !strings.HasPrefix(checksumURL, "https://") && !strings.HasPrefix(checksumURL, "http://") {
		t.Errorf("Expected %q to start with either \"http://\" or \"https://\", it did not.", checksumURL)
	}
	expected := c.Release + "/isos/" + c.Arch + "/sha256sum.txt"
	if !strings.HasSuffix(checksumURL, expected) {
		t.Errorf("Expected %q to end with %q, it did not.", checksumURL, expected)
	}
}

func TestCentOSSetISOName(t *testing.T) {
	// 6
	c := newTestCentOS()
	c.Release = "6"
	c.setVersionInfo()
	c.setISOName()
	expected := fmt.Sprintf("CentOS-%s-%s-%s.iso", c.FullVersion, c.Arch, strings.ToLower(c.Image))
	if c.Name != expected {
		t.Errorf("Expected %q, got %q", expected, c.Name)
	}
	// 7
	c = newTestCentOS()
	c.Release = "7"
	c.setVersionInfo()
	c.setISOName()
	expected = fmt.Sprintf("CentOS-%s-%s-%s-%s.iso", c.MajorVersion, c.Arch, c.Image, c.MinorVersion)
	if c.Name != expected {
		t.Errorf("Expected %q, got %q", expected, c.Name)
	}
}

func TestCentOS6setISOURL(t *testing.T) {
	c := newTestCentOS()
	c.FullVersion = "6.6"
	c.MajorVersion = "6"
	c.BaseURL = "http://bay.uchicago.edu/centos/6.6/isos/x86_64/"
	c.ReleaseURL = c.BaseURL
	c.setISOName()
	url := c.imageURL()
	if !strings.HasPrefix(url, "http://") {
		t.Errorf("Expected %q to have a prefix of \"http://\", it didn't", url)
	}
	expected := fmt.Sprintf("%sCentOS-%s-%s-%s.iso", c.BaseURL, c.FullVersion, c.Arch, strings.ToLower(c.Image))
	if url != expected {
		t.Errorf("Expected %q, got %q", expected, url)
	}

	c.BaseURL = "http://example.com/"
	c.ReleaseURL = c.BaseURL
	url = c.imageURL()
	expected = fmt.Sprintf("%sCentOS-%s-%s-%s.iso", c.BaseURL, c.FullVersion, c.Arch, strings.ToLower(c.Image))
	if url != expected {
		t.Errorf("Expected %q, got %q", expected, url)
	}
}

func TestCentOSFindISOChecksum(t *testing.T) {
	tests := []struct {
		release     string
		arch        string
		page        string
		expected    string
		expectedErr string
	}{
		{"6", "x86", "", "", "page empty"},
		{"6", "x86_64",
			`a63241b0f767afa1f9f7e59e6f0f00d6b8d19ed85936a7934222c03a92e61bf3  CentOS-6.6-x86_64-bin-DVD1.iso
89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597  CentOS-6.6-x86_64-bin-DVD2.iso
5458f357e8a55e3a866dd856896c7e0ac88e7f9220a3dd74c58a3b0acede8e4d  CentOS-6.6-x86_64-minimal.iso
ad8f6de098503174c7609d172679fa0dd276f4b669708933d9c4927bd3fe1017  CentOS-6.6-x86_64-netinstall.iso`,
			"5458f357e8a55e3a866dd856896c7e0ac88e7f9220a3dd74c58a3b0acede8e4d", ""},
		{"6", "x86",
			`a63241b0f767afa1f9f7e59e6f0f00d6b8d19ed85936a7934222c03a92e61bf3  CentOS-6.6-x86_64-bin-DVD1.iso
89dac78769b26f8facf98ce85020a605b7601fec1946b0597e22ced5498b3597  CentOS-6.6-x86_64-bin-DVD2.iso
5458f357e8a55e3a866dd856896c7e0ac88e7f9220a3dd74c58a3b0acede8e4d  CentOS-6.6-x86_64-minimal.iso
ad8f6de098503174c7609d172679fa0dd276f4b669708933d9c4927bd3fe1017  CentOS-6.6-x86_64-netinstall.iso`,
			"", "checksum not found"},
	}
	for i, test := range tests {
		c := newTestCentOS()
		c.Arch = test.arch
		c.Release = test.release
		c.FullVersion = "6.6"
		c.MinorVersion = "6"
		c.MajorVersion = "6"
		c.BaseURL = "http://bay.uchicago.edu/centos/6.6/isos/x86_64/"
		c.ReleaseURL = c.BaseURL
		c.setISOName()
		err := c.findISOChecksum(test.page)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestCentOSFindISOChecksum %d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expected != c.Checksum {
			t.Errorf("TestCentOSFindISOChecksum %d: expected %q, got %q", i, test.expected, c.Checksum)
		}
	}
}

func TestDebianSetReleaseInfo(t *testing.T) {
	page := `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 3.2 Final//EN">
<html>
 <head>
  <title>Index of /debian-cd</title>
 </head>
 <body>
<h1>Index of /debian-cd</h1>
<pre><img src="/icons/blank.gif" alt="Icon "> <a href="?C=N;O=D">Name</a>                    <a href="?C=M;O=A">Last modified</a>      <a href="?C=S;O=A">Size</a>  <hr><img src="/icons/back.gif" alt="[PARENTDIR]"> <a href="/">Parent Directory</a>                             -
<img src="/icons/folder.gif" alt="[DIR]"> <a href="7.8.0-live/">7.8.0-live/</a>             2015-01-14 05:00    -
<img src="/icons/folder.gif" alt="[DIR]"> <a href="7.8.0/">7.8.0/</a>                  2015-01-12 03:07    -
<img src="/icons/folder.gif" alt="[DIR]"> <a href="current-live/">current-live/</a>           2015-01-14 05:00    -
<img src="/icons/folder.gif" alt="[DIR]"> <a href="current/">current/</a>                2015-01-12 03:07    -
<img src="/icons/folder.gif" alt="[DIR]"> <a href="project/">project/</a>                2005-05-23 18:50    -
<img src="/icons/compressed.gif" alt="[   ]"> <a href="ls-lR.gz">ls-lR.gz</a>                2015-03-05 03:12   39K
<hr></pre>
<address>Apache/2.4.10 (Unix) Server at cdimage.debian.org Port 80</address>
</body></html>

`
	d := newTestDebian()
	d.FullVersion = ""
	err := d.setReleaseInfo(page)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	}
	expected := "7.8.0"
	if d.FullVersion != expected {
		t.Errorf("Expected %q, got %q", expected, d.FullVersion)
	}
}

func newTestDebian() debian {
	d := debian{release{Release: "7", FullVersion: "7.8.0", Image: "netinst", Arch: "amd64"}}
	d.ChecksumType = "sha256"
	return d
}

func TestDebianFindISOChecksum(t *testing.T) {
	d := newTestDebian()
	err := d.findISOChecksum("")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err != ErrPageEmpty {
			t.Errorf("expected \"%q\", got %q ", ErrPageEmpty, err)
		}
	}

	checksumPage := `
6489ad85505b02a73e1049ba70137f25d5eab0d0b25701b15cedb840753a53a3  debian-7.8.0-amd64-CD-9.iso
72763957bcd206882ca29ce524f9ca1940d1e5bce5ed15ccc6c346b35feb4a41  debian-7.8.0-amd64-kde-CD-1.iso
1c8164b42e27a55657ab971b5e32ca2045dda6c49e3484279de63167c6aada31  debian-7.8.0-amd64-lxde-CD-1.iso
e39c36d6adc0fd86c6edb0e03e22919086c883b37ca194d063b8e3e8f6ff6a3a  debian-7.8.0-amd64-netinst.iso
fb1c1c30815da3e7189d44b6699cf9114b16e44ea139f0cd4df5f1dde3659272  debian-7.8.0-amd64-xfce-CD-1.iso
23a0d89e337c96f3e26a1ab5d49392d3129fbcb3d982b9d58a797039b01f1e7f  debian-update-7.8.0-amd64-CD-1.iso
`
	d.Name = "debian-7.8.0-amd64-whatever.iso"
	d.Image = "whatever"
	err = d.findISOChecksum(checksumPage)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err != ErrChecksumTypeNotSet {
			t.Errorf("got %q; want %q", err, ErrChecksumTypeNotSet)
		}
	}

	d.FullVersion = "7.8.0"
	d.Name = "debian-7.8.0-amd64-netinst.iso"
	d.Image = "netinst"
	d.Arch = "amd64"
	err = d.findISOChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected no error, got %q", err)
	} else {
		if d.Checksum != "e39c36d6adc0fd86c6edb0e03e22919086c883b37ca194d063b8e3e8f6ff6a3a" {
			t.Errorf("Expected \"e39c36d6adc0fd86c6edb0e03e22919086c883b37ca194d063b8e3e8f6ff6a3a\", got %q", d.Checksum)
		}
	}
}

func TestDebianSetISO(t *testing.T) {
	d := newTestDebian()
	d.setISOName()
	expected := "debian-7.8.0-amd64-netinst.iso"
	if d.Name != expected {
		t.Errorf("Expected %q, got %q", expected, d.Name)
	}
	d.ReleaseURL = "http://cdimage.debian.org/debian-cd/7.8.0/amd64/iso-cd/"
	url := d.imageURL()
	expected = "http://cdimage.debian.org/debian-cd/7.8.0/amd64/iso-cd/" + d.Name
	if url != expected {
		t.Errorf("Expected %q, got %q", expected, url)
	}

}

func TestDebianGetOSType(t *testing.T) {
	tests := []struct {
		buildType Builder
		arch      string
		expected  string
		err       string
	}{
		{VMWareISO, "amd64", "debian-64", ""},
		{VMWareISO, "i386", "debian-32", ""},
		{VMWareVMX, "amd64", "debian-64", ""},
		{VMWareVMX, "i386", "debian-32", ""},
		{VirtualBoxISO, "amd64", "Debian_64", ""},
		{VirtualBoxISO, "i386", "Debian_32", ""},
		{VirtualBoxOVF, "amd64", "Debian_64", ""},
		{VirtualBoxOVF, "i386", "Debian_32", ""},
		{QEMU, "amd64", "", ""},
		{QEMU, "i386", "", ""},
		{UnsupportedBuilder, "amd64", "", fmt.Sprintf("Debian %s: not supported", UnsupportedBuilder)},
		{UnsupportedBuilder, "i386", "", fmt.Sprintf("Debian %s: not supported", UnsupportedBuilder)},
	}
	for i, test := range tests {
		d := debian{release{Arch: test.arch}}
		res, err := d.getOSType(test.buildType)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("%d: got %q want %q", i, err, test.err)
			}
			continue
		}
		if test.err != "" {
			t.Errorf("%d: got no error, want %q", i, test.err)
			continue
		}
		if res != test.expected {
			t.Errorf("%d: got %s, want %s", i, res, test.expected)
		}
	}
}

func newTestUbuntu() ubuntu {
	u := ubuntu{release{Release: "14.04", Image: "server", Arch: "amd64"}}
	u.ChecksumType = "sha256"
	u.BaseURL = "http://releases.ubuntu.com/"
	return u
}

func TestUbuntuFindISOChecksum(t *testing.T) {
	u := newTestUbuntu()
	err := u.findISOChecksum("")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err != ErrPageEmpty {
			t.Errorf("got %q; want %q", err, ErrPageEmpty)
		}
	}

	checksumPage := `cab6b0458601520242eb0337ccc9797bf20ad08bf5b23926f354198928191da5 *ubuntu-14.04-desktop-amd64.iso
207a53944d5e8bbb278f4e1d8797491bfbb759c2ebd4a162f41e1383bde38ab2 *ubuntu-14.04-desktop-i386.iso
ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388 *ubuntu-14.04-server-amd64.iso
85c738fefe7c9ff683f927c23f5aa82864866c2391aeb376abfec2dfc08ea873 *ubuntu-14.04-server-i386.iso
bc3b20ad00f19d0169206af0df5a4186c61ed08812262c55dbca3b7b1f1c4a0b *wubi.exe`
	u.Name = "ubuntu-14.04-whatever.iso"
	u.Image = "whatever"
	err = u.findISOChecksum(checksumPage)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err != ErrChecksumNotFound {
			t.Errorf("TestUbuntuFindISOChecksum: got %q, want %q", err, ErrChecksumNotFound)
		}
	}

	u.Release = "14.04"
	u.Image = "server"
	u.Arch = "amd64"
	u.Name = "ubuntu-14.04-server-amd64.iso"
	err = u.findISOChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected no error, got %q", err)
	} else {
		if u.Checksum != "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388" {
			t.Errorf("Expected \"ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388\", got %q", u.Checksum)
		}
	}

	checksumPage = `3a80be812e7767e1e78b4e1aeab4becd11a677910877250076096fe3b470ba22 *ubuntu-12.04.4-alternate-amd64.iso
bf48efb08cb1962bebaabeb81e4a0a00c6fd7dff5ff50927d0ba84923d0deed4 *ubuntu-12.04.4-alternate-i386.iso
fa28d4b4821d6e8c5e5543f8d9f5ed8176400e078fe9177fa2774214b7296c84 *ubuntu-12.04.4-desktop-amd64.iso
c0ba532d8fadaa3334023f96925b93804e859dba2b4c4e4cda335bd1ebe43064 *ubuntu-12.04.4-desktop-i386.iso
3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54 *ubuntu-12.04.4-server-amd64.iso
fbe7f159337551cc5ce9f0ff72acefef567f3dcd30750425287588c554978501 *ubuntu-12.04.4-server-i386.iso
2d92c01bcfcd0911c5e4256250647c30e3351f3190515f972f83027c4260e7e5 *ubuntu-12.04.4-wubi-amd64.tar.xz
2bde9af4e9f8a6ce493955188a624cb3ffdf46294805605d9d51a1ee62997dcf *ubuntu-12.04.4-wubi-i386.tar.xz
819f42fdd7cc431b6fd7fa5bae022b0a8c55a0f430eb3681e4750c4f1eceaf91 *wubi.exe`

	u.Release = "12.04"
	u.Name = "ubuntu-12.04.4-server-amd64.iso"
	err = u.findISOChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected no error, got %q", err)
	} else {
		if u.Checksum != "3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54" {
			t.Errorf("Expected \"3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54\", got %q", u.Checksum)
		}
	}
}

func TestUbuntuGetOSType(t *testing.T) {
	tests := []struct {
		buildType Builder
		arch      string
		expected  string
		err       string
	}{
		{VMWareISO, "amd64", "ubuntu-64", ""},
		{VMWareISO, "i386", "ubuntu-32", ""},
		{VMWareVMX, "amd64", "ubuntu-64", ""},
		{VMWareVMX, "i386", "ubuntu-32", ""},
		{VirtualBoxISO, "amd64", "Ubuntu_64", ""},
		{VirtualBoxISO, "i386", "Ubuntu_32", ""},
		{VirtualBoxOVF, "amd64", "Ubuntu_64", ""},
		{VirtualBoxOVF, "i386", "Ubuntu_32", ""},
		{QEMU, "amd64", "", ""},
		{QEMU, "i386", "", ""},
		{UnsupportedBuilder, "amd64", "", fmt.Sprintf("Ubuntu %s: not supported", UnsupportedBuilder)},
		{UnsupportedBuilder, "i386", "", fmt.Sprintf("Ubuntu %s: not supported", UnsupportedBuilder)},
	}
	for i, test := range tests {
		u := ubuntu{release{Arch: test.arch}}
		res, err := u.getOSType(test.buildType)
		if err != nil {
			if err.Error() != test.err {
				t.Errorf("%d: got %q want %q", i, err, test.err)
			}
			continue
		}
		if test.err != "" {
			t.Errorf("%d: got no error, want %q", i, test.err)
			continue
		}
		if res != test.expected {
			t.Errorf("%d: got %s, want %s", i, res, test.expected)
		}
	}
}

var records = [][]string{
	[]string{"EU", "Ireland", "Strencom", "http://www.strencom.net/", "http://mirror.strencom.net/centos/", "", ""},
	[]string{"EU", "Italy", "    GARR/CILEA", "http://mirror.garr.it", "http://ct.mirror.garr.it/mirrors/CentOS/", "ftp://ct.mirror.garr.it/mirrors/CentOS/", "rsync://ct.mirror.garr.it/CentOS/ "},
	[]string{"US", "", "Oregon State University", "http://osuosl.org/", "http://ftp.osuosl.org/pub/centos/", "ftp://ftp.osuosl.org/pub/centos/", "rsync://ftp.osuosl.org/centos/"},
	[]string{"US", "", "Rackspace", "http://www.rackspace.com/", "http://mirror.rackspace.com/CentOS/", "", ""},
	[]string{"US", "", "Rackspace", "http://www.rackspace.com/", "http://mirror.rackspace.com/CentOS/", "", ""},
	[]string{"US", "CT", "Connecticut Education Network", "http://www.ct.gov/cen", "http://mirror.net.cen.ct.gov/centos/", "", ""},
	[]string{"US", "CT", "Connecticut Education Network", "http://www.ct.gov/cen", "http://mirror.net.cen.ct.gov/centos/", "", ""},
	[]string{"US", "DC", "ServInt", "http://www.servint.com/", "http://centos.servint.com/", "", ""},
	[]string{"US", "DC", "ServInt", "http://www.servint.com/", "http://centos.servint.com/", "", ""},
}

func TestFilterRecords(t *testing.T) {
	tests := []struct {
		value    string
		index    int
		expected [][]string
	}{
		{"", 0, [][]string{}},
		{"", 1, [][]string{}},
		{
			"EU", 0, [][]string{
				[]string{"EU", "Ireland", "Strencom", "http://www.strencom.net/", "http://mirror.strencom.net/centos/", "", ""},
				[]string{"EU", "Italy", "    GARR/CILEA", "http://mirror.garr.it", "http://ct.mirror.garr.it/mirrors/CentOS/", "ftp://ct.mirror.garr.it/mirrors/CentOS/", "rsync://ct.mirror.garr.it/CentOS/ "},
			},
		},
		{
			"Italy", 1, [][]string{
				[]string{"EU", "Italy", "    GARR/CILEA", "http://mirror.garr.it", "http://ct.mirror.garr.it/mirrors/CentOS/", "ftp://ct.mirror.garr.it/mirrors/CentOS/", "rsync://ct.mirror.garr.it/CentOS/ "},
			},
		},
		{
			"DC", 1, [][]string{
				[]string{"US", "DC", "ServInt", "http://www.servint.com/", "http://centos.servint.com/", "", ""},
				[]string{"US", "DC", "ServInt", "http://www.servint.com/", "http://centos.servint.com/", "", ""},
			},
		},
		{
			"Rackspace", 2, [][]string{
				[]string{"US", "", "Rackspace", "http://www.rackspace.com/", "http://mirror.rackspace.com/CentOS/", "", ""},
				[]string{"US", "", "Rackspace", "http://www.rackspace.com/", "http://mirror.rackspace.com/CentOS/", "", ""},
			},
		},
		{
			"Oregon State University", 2, [][]string{
				[]string{"US", "", "Oregon State University", "http://osuosl.org/", "http://ftp.osuosl.org/pub/centos/", "ftp://ftp.osuosl.org/pub/centos/", "rsync://ftp.osuosl.org/centos/"},
			},
		},
	}
	for i, test := range tests {
		filtered := filterRecords(test.value, test.index, records)
		if i < 2 {
			if MarshalJSONToString.Get(filtered) != MarshalJSONToString.Get(records) {
				t.Errorf("got %v; expected %v", filtered, records)
			}
			continue
		}

	}
}
