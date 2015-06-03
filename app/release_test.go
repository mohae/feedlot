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
	_ "net/url"
	"strings"
	"testing"
)

func newTestCentOS() centOS {
	c := centOS{release{Release: "6", Image: "minimal", Arch: "x86_64"}}
	c.ChecksumType = "sha256"
	return c
}

func TestCentOSisoRedirectURL(t *testing.T) {
	c := newTestCentOS()
	redirect := c.isoRedirectURL()
	if redirect != "http://isoredirect.centos.org/centos/6/isos/x86_64/" {
		t.Errorf("TestCentOSisoRedirectURL expected \"http://isoredirect.centos.org/centos/6/isos/x86_64/\", got %q", redirect)
	}
}

func TestCentOSsetReleaseInfo(t *testing.T) {
	c := newTestCentOS()
	err := c.setReleaseInfo()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if !strings.HasPrefix(c.Release, "6") {
			t.Errorf("Expected %q to start with \"6\"", c.Release)
		}
	}
}

func TestCentOSsetReleaseNumber(t *testing.T) {
	c := newTestCentOS()
	err := c.setReleaseNumber()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if !strings.HasPrefix(c.ReleaseFull, "6.") {
			t.Errorf("Expected %q to start with \"6.\"", c.ReleaseFull)
		}
	}
}

func TestCentOSGetOSType(t *testing.T) {
	c := centOS{release{Arch: "x86_64"}}
	buildType := "vmware-iso"
	res, err := c.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if res != "centos-64" {
			t.Errorf("Expected \"centos-64\", got %q", res)
		}
	}

	buildType = "virtualbox-iso"
	res, err = c.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if res != "RedHat_64" {
			t.Errorf("Expected \"RedHat_64\", got %q", res)
		}
	}

	c = centOS{release{Arch: "x386"}}
	buildType = "vmware-iso"
	res, err = c.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if res != "centos-32" {
			t.Errorf("Expected \"centos-32\", got %q", res)
		}
	}

	buildType = "virtualbox-iso"
	res, err = c.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if res != "RedHat_32" {
			t.Errorf("Expected \"RedHat_32\", got %q", res)
		}
	}

	buildType = "voodoo"
	res, err = c.getOSType(buildType)
	if err == nil {
		t.Error("Expected error to not be nil, it was")
	} else {
		if err.Error() != " does not support the voodoo builder" {
			t.Errorf("Expected \" does not support the voodoo builder\", got %q", err.Error())
		}
	}
}

func TestCentOSSetISOChecksum(t *testing.T) {
	c := newTestCentOS()
	c.setReleaseInfo()
	//	c.SetISOInfo()
	err := c.setISOChecksum()
	if err == nil {
		t.Error("Expected error to not be nil, it was")
	} else {
		if err.Error() != "Get sha256sum.txt: unsupported protocol scheme \"\"" {
			t.Errorf("expected \"Get sha256sum.txt: unsupported protocol scheme \"\"\", got %q", err.Error())
		}
	}

	c.Arch = ""
	err = c.setISOChecksum()
	if err == nil {
		t.Error("Expected error to not be nil, it was")
	} else {
		if err.Error() != "Get sha256sum.txt: unsupported protocol scheme \"\"" {
			t.Errorf("Expected \"Get sha256sum.txt: unsupported protocol scheme \"\"\", got %q", err.Error())
		}
	}

	c = newTestCentOS()
	c.Arch = "x86_64"
	c.setReleaseInfo()
	c.setISOName()
	c.SetISOInfo()
	err = c.setISOChecksum()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if c.Checksum != "5458f357e8a55e3a866dd856896c7e0ac88e7f9220a3dd74c58a3b0acede8e4d" {
			t.Errorf("Expected \"5458f357e8a55e3a866dd856896c7e0ac88e7f9220a3dd74c58a3b0acede8e4d\", got %q", c.Checksum)
		}
	}
}

func TestCentOSChecksumURL(t *testing.T) {
	c := newTestCentOS()
	c.setReleaseInfo()
	c.setISOURL()
	checksumURL := c.checksumURL()
	if !strings.HasPrefix(checksumURL, "http://") {
		t.Errorf("Expected %q to start with \"http://\", it did not.", checksumURL)
	}
	expected := c.ReleaseFull + "/isos/" + c.Arch + "/sha256sum.txt"
	if !strings.HasSuffix(checksumURL, expected) {
		t.Errorf("Expected %q to end with %q, it did not.", checksumURL, expected)
	}
}

func TestCentOSSetISOName(t *testing.T) {
	c := newTestCentOS()
	c.setReleaseInfo()
	c.setISOName()
	expected := "CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
	if c.Name != expected {
		t.Errorf("Expected %q, got %q", expected, c.Name)
	}
}

/*
func TestCentOSSetISOInfo(t *testing.T) {
	tests := []struct {
		arch             string
		release          string
		expectedURL      string
		expectedChecksum string
		expectedErr      string
	}{
		{"", "", "", "", "centos SetISOInfo error: empty arch"},
		{"x86_64", "", "", "", "centos SetISOInfo error: empty release"},
		{"x86_64", "6", "", "", ""},
	}
	c := newTestCentOS()
	for i, test := range tests {
		c.Arch = test.arch
		c.Release = test.release
		err := c.SetISOInfo()
		if err != nil {
			if test.expectedErr != err.Error() {
				t.Errorf("TestCentOSSetISOInfo %d: expected %q, got %q", i, test.expectedErr, err.Error())
				continue
			}
		}
		if test.expectedURL != c.isoURL {
			t.Errorf("TestCentOSSetISOInfo %d: expected iso url to be %q, got %q", i, test.expectedURL, c.isoURL)
		}
		if test.expectedChecksum != c.Checksum {
			t.Errorf("TestCentOSSetISOInfo %d: expected checksum to be %q, got %q", i, test.expectedChecksum, c.Checksum)
		}
	}
}
*/

func TestCentOSsetISOURL(t *testing.T) {
	c := newTestCentOS()
	c.ReleaseFull = "6.6"
	c.BaseURL = "http://mirror.cogentco.com/pub/linux/centos/"
	c.setISOName()
	err := c.setISOURL()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if !strings.HasPrefix(c.isoURL, "http://") {
			t.Errorf("Expected %q to have a prefix of \"http://\", it didn't", c.isoURL)
		}
		expected := c.BaseURL + c.ReleaseFull + "/isos/" + c.Arch + "/CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
		if c.isoURL != expected {
			t.Errorf("Expected %q, got %q", expected, c.isoURL)
		}
	}

	c.BaseURL = "http://example.com/"
	err = c.setISOURL()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		expected := "http://example.com/" + c.ReleaseFull + "/isos/" + c.Arch + "/CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
		if c.isoURL != expected {
			t.Errorf("Expected %q, got %q", expected, c.isoURL)
		}
	}
}

func TestCentOSrandomISOURL(t *testing.T) {
	c := newTestCentOS()
	c.setReleaseInfo()
	c.setISOName()
	randomURL, err := c.randomISOURL()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if !strings.HasPrefix(randomURL, "http://") {
			t.Errorf("Expected URL to start with \"http://\", got %q", randomURL)
		}
		expected := c.ReleaseFull + "/isos/" + c.Arch + "/CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
		if !strings.HasSuffix(randomURL, expected) {
			t.Errorf("Expected %q, got %q", expected, randomURL)
		}
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
		{"6", "x86", "", "", "SetReleaseInfo: setReleaseNumber http://mirrorlist.centos.org/?release=6&arch=x86&repo=os: Bad arch - not in list - x86_64 alpha s390x ia64 s390 i386 "},
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
			"", "SetReleaseInfo: setReleaseNumber http://mirrorlist.centos.org/?release=6&arch=x86&repo=os: Bad arch - not in list - x86_64 alpha s390x ia64 s390 i386 "},
	}
	for i, test := range tests {
		c := newTestCentOS()
		c.Arch = test.arch
		c.Release = test.release
		err := c.setReleaseInfo()
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestCentOSFindISOChecksum %d: sexpected %q, got %q", i, test.expectedErr, err.Error())
			}
			continue
		}
		c.setISOName()
		checksum, err := c.findISOChecksum(test.page)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestCentOSFindISOChecksum %d: expected %q, got %q", i, test.expectedErr, err.Error())
			}
			continue
		}
		if test.expected != checksum {
			t.Errorf("TestCentOSFindISOChecksum %d: expected %q, got %q", i, test.expected, checksum)
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
	d.ReleaseFull = ""
	err := d.setReleaseInfo(page)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	expected := "7.8.0"
	if d.ReleaseFull != expected {
		t.Errorf("Expected %q, got %q", expected, d.ReleaseFull)
	}
}

func newTestDebian() debian {
	d := debian{release{Release: "7", ReleaseFull: "7.8.0", Image: "netinst", Arch: "amd64"}}
	d.ChecksumType = "sha256"
	return d
}

func TestDebianFindISOChecksum(t *testing.T) {
	d := newTestDebian()
	s, err := d.findISOChecksum("")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != " findISOChecksum error: page was empty" {
			t.Errorf(" findISOChecksum error: page was empty", err.Error())
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
	s, err = d.findISOChecksum(checksumPage)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "debian-7.8.0-amd64-whatever.iso findISOChecksum error: checksum not found on page" {
			t.Errorf("TestDebianFindISOChecksum: expected \"debian-7.8.0-amd64-whatever.iso findISOChecksum error: checksum not found on page\", got %q", err.Error())
		}
	}

	d.ReleaseFull = "7.8.0"
	d.Name = "debian-7.8.0-amd64-netinst.iso"
	d.Image = "netinst"
	d.Arch = "amd64"
	s, err = d.findISOChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if s != "e39c36d6adc0fd86c6edb0e03e22919086c883b37ca194d063b8e3e8f6ff6a3a" {
			t.Errorf("Expected \"e39c36d6adc0fd86c6edb0e03e22919086c883b37ca194d063b8e3e8f6ff6a3a\", got %q", s)
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

	d.setISOURL()
	expected = "http://cdimage.debian.org/debian-cd/7.8.0/amd64/iso-cd/" + d.Name
	if d.isoURL != expected {
		t.Errorf("Expected %q, got %q", expected, d.isoURL)
	}

}

func TestDebianGetOSType(t *testing.T) {
	d := debian{release{Arch: "amd64"}}
	buildType := "vmware-iso"
	res, err := d.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "debian-64" {
			t.Errorf("Expected \"debian-64\", got %q", res)
		}
	}

	buildType = "virtualbox-iso"
	res, err = d.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "Debian_64" {
			t.Errorf("Expected \"Debian_64\", got %q", res)
		}
	}

	d = debian{release{Arch: "i386"}}
	buildType = "vmware-iso"
	res, err = d.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "debian-32" {
			t.Errorf("Expected \"debian-32\", got %q", res)
		}
	}

	buildType = "virtualbox-iso"
	res, err = d.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "Debian_32" {
			t.Errorf("Expected \"Debian_32\", got %q", res)
		}
	}

	buildType = "voodoo"
	res, err = d.getOSType(buildType)
	if err == nil {
		t.Error("Expected an error, received nil")
	} else {
		if err.Error() != " does not support the voodoo builder" {
			t.Errorf("Expected \" does not support the voodoo builder\", got %q", err.Error())
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
	s, err := u.findISOChecksum("")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != " findISOChecksum error: page was empty" {
			t.Errorf("TestUbuntuFindISOChecksum: expected \" findISOChecksum error: page was empty\", got %q", err.Error())
		}
	}

	checksumPage := `cab6b0458601520242eb0337ccc9797bf20ad08bf5b23926f354198928191da5 *ubuntu-14.04-desktop-amd64.iso
207a53944d5e8bbb278f4e1d8797491bfbb759c2ebd4a162f41e1383bde38ab2 *ubuntu-14.04-desktop-i386.iso
ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388 *ubuntu-14.04-server-amd64.iso
85c738fefe7c9ff683f927c23f5aa82864866c2391aeb376abfec2dfc08ea873 *ubuntu-14.04-server-i386.iso
bc3b20ad00f19d0169206af0df5a4186c61ed08812262c55dbca3b7b1f1c4a0b *wubi.exe`
	u.Name = "ubuntu-14.03-whatever.iso"
	u.Image = "whatever"
	s, err = u.findISOChecksum(checksumPage)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "ubuntu-14.04-whatever-amd64.iso findISOChecksum error: checksum not found on page" {
			t.Errorf("TestUbuntuFindISOChecksum: expected \"ubuntu-14.04-whatever-amd64.iso findISOChecksum error: checksum not found on page\", got %q", err.Error())
		}
	}

	u.Release = "14.04"
	u.Image = "server"
	u.Arch = "amd64"
	s, err = u.findISOChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if s != "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388" {
			t.Errorf("Expected \"ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388\", got %q", s)
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
	s, err = u.findISOChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if s != "3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54" {
			t.Errorf("Expected \"3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54\", got %q", s)
		}
	}
}

func TestUbuntuGetOSType(t *testing.T) {
	u := ubuntu{release{Arch: "amd64"}}
	buildType := "vmware-iso"
	res, err := u.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "ubuntu-64" {
			t.Errorf("Expected \"3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54\", got %q", res)
		}
	}

	buildType = "virtualbox-iso"
	res, err = u.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "Ubuntu_64" {
			t.Errorf("Expected \"Ubuntu_64\", got %q", res)
		}
	}

	u = ubuntu{release{Arch: "i386"}}
	buildType = "vmware-iso"
	res, err = u.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "ubuntu-32" {
			t.Errorf("Expected \"ubuntu-32\", got %q", res)
		}
	}

	buildType = "virtualbox-iso"
	res, err = u.getOSType(buildType)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if res != "Ubuntu_32" {
			t.Errorf("Expected \"Ubuntu_32\", got %q", res)
		}
	}

	buildType = "voodoo"
	res, err = u.getOSType(buildType)
	if err == nil {
		t.Error("Expected an error, received nil")
	} else {
		if err.Error() != " does not support the voodoo builder" {
			t.Errorf("Expected \" does not support the voodoo builder\", got %q", err.Error())
		}
	}
}

func TestNoArchError(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", " SetISOInfo error: empty arch"},
		{"CentOS", "CentOS SetISOInfo error: empty arch"},
		{"CentOS", "CentOS SetISOInfo error: empty arch"},
	}
	for i, test := range tests {
		err := NoArchError(test.name)
		if err.Error() != test.expected {
			t.Errorf("TestNoArchError %d: expected %q, got %q", i, test.expected, err.Error())
		}
	}
}

func TestNoReleaseError(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"", " SetISOInfo error: empty release"},
		{"CentOS", "CentOS SetISOInfo error: empty release"},
		{"CentOS", "CentOS SetISOInfo error: empty release"},
	}
	for i, test := range tests {
		err := NoReleaseError(test.name)
		if err.Error() != test.expected {
			t.Errorf("TestNoReleaseError %d: expected %q, got %q", i, test.expected, err.Error())
		}
	}
}

func TestEmptyPageError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		expected  string
	}{
		{"", "", "  error: page was empty"},
		{"", "test", " test error: page was empty"},
		{"CentOS-6.6-x86_64-minimal.iso", "", "CentOS-6.6-x86_64-minimal.iso  error: page was empty"},
		{"CentOS-6.6-x86_64-minimal.iso", "find something", "CentOS-6.6-x86_64-minimal.iso find something error: page was empty"},
		{"CentOS-6.5-whatever.iso", "", "CentOS-6.5-whatever.iso  error: page was empty"},
		{"CentOS-6.5-whatever.iso", "test", "CentOS-6.5-whatever.iso test error: page was empty"},
	}
	for i, test := range tests {
		err := EmptyPageError(test.name, test.operation)
		if err.Error() != test.expected {
			t.Errorf("TestEmptyPageError %d: expected %q, got %q", i, test.expected, err.Error())
		}
	}
}

func TestChecksumNotFoundError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		expected  string
	}{
		{"", "", "  error: checksum not found on page"},
		{"", "test", " test error: checksum not found on page"},
		{"CentOS-6.6-x86_64-minimal.iso", "", "CentOS-6.6-x86_64-minimal.iso  error: checksum not found on page"},
		{"CentOS-6.6-x86_64-minimal.iso", "checksum search", "CentOS-6.6-x86_64-minimal.iso checksum search error: checksum not found on page"},
		{"CentOS-6.5-whatever.iso", "", "CentOS-6.5-whatever.iso  error: checksum not found on page"},
		{"CentOS-6.5-whatever.iso", "checksum parse", "CentOS-6.5-whatever.iso checksum parse error: checksum not found on page"},
	}
	for i, test := range tests {
		err := ChecksumNotFoundError(test.name, test.operation)
		if err.Error() != test.expected {
			t.Errorf("TestChecksumNotFoundError %d: expected %q, got %q", i, test.expected, err.Error())
		}
	}
}
