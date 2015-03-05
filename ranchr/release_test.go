// Some of these tests query the actual URLs; as such they may fail if the
// remote server is not available or if the destination url is no longer
// valid.
//
// The risk of tests failing due to remote not being availabe was deemed
// acceptable. If the destination url is no longer valid, or the assumptions
// made in the code lead to error, this is fine as the code should be updated
// to reflect the changes in remote.
package ranchr

import (
	_ "net/url"
	"strings"
	"testing"
)

func newTestUbuntu() ubuntu {
	u := ubuntu{release{Release: "14.04", Image: "server", Arch: "amd64"}}
	u.ChecksumType = "sha256"
	u.BaseURL = "http://releases.ubuntu.com/"
	return u
}

func TestUbuntuFindChecksum(t *testing.T) {
	u := newTestUbuntu()
	s, err := u.findChecksum("")
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "page to parse was empty; unable to process request for " {
			t.Errorf("Expected \"page to parse was empty; unable to process request for \", got %q", err.Error())
		}
	}

	checksumPage := `cab6b0458601520242eb0337ccc9797bf20ad08bf5b23926f354198928191da5 *ubuntu-14.04-desktop-amd64.iso
207a53944d5e8bbb278f4e1d8797491bfbb759c2ebd4a162f41e1383bde38ab2 *ubuntu-14.04-desktop-i386.iso
ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388 *ubuntu-14.04-server-amd64.iso
85c738fefe7c9ff683f927c23f5aa82864866c2391aeb376abfec2dfc08ea873 *ubuntu-14.04-server-i386.iso
bc3b20ad00f19d0169206af0df5a4186c61ed08812262c55dbca3b7b1f1c4a0b *wubi.exe`
	u.Name = "ubuntu-14.03-whatever.iso"
	u.Image = "whatever"
	s, err = u.findChecksum(checksumPage)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "unable to find ubuntu-14.04-whatever-amd64.iso's checksum" {
			t.Errorf("Expected \"unable to find ubuntu-14.04-whatever-amd64.iso's checksum\", got %q", err.Error())
		}
	}

	u.Release = "14.04"
	u.Image = "server"
	u.Arch = "amd64"
	s, err = u.findChecksum(checksumPage)
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
	s, err = u.findChecksum(checksumPage)
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

func newTestCentOS() centOS {
	c := centOS{release{Release: "6", Image: "minimal", Arch: "x86_64"}}
	c.ChecksumType = "sha256"
	return c
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

func TestCentOSSetISOName(t *testing.T) {
	c := newTestCentOS()
	c.setReleaseInfo()
	c.setISOName()
	expected := "CentOS-" + c.ReleaseFull + "-" + c.Arch + "-" + c.Image + ".iso"
	if c.Name != expected {
		t.Errorf("Expected %q, got %q", expected, c.Name)
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

func TestCentOSSetChecksum(t *testing.T) {
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

func TestCentOSFindChecksum(t *testing.T) {
	var err error
	var s string
	c := newTestCentOS()
	c.setReleaseInfo()
	c.setISOName()
	c.setISOURL()
	s, err = c.findChecksum("")
	if err == nil {
		t.Error("Expected error to not be nil, it was")
	} else {
		if err.Error() != "the string passed to centOS.findChecksum(s string) was empty; unable to process request" {
			t.Errorf("Expected \"the string passed to centOS.findChecksum(s string) was empty; unable to process request\", got %q", err.Error())
		}
	}

	checksumPage := `c796ab378319393f47b29acd8ceaf21e1f48439570657945226db61702a4a2a1  CentOS-6.5-x86_64-bin-DVD1.iso
afd2fc37e1597c64b3c3464083c0022f436757085d9916350fb8310467123f77  CentOS-6.5-x86_64-bin-DVD2.iso
58b40b26415133ed2af8e2f53b73b5f2aa013723742ce17671b5bb1880a20a99  CentOS-6.5-x86_64-LiveCD.iso
e3efa9a6ca6f58ac4be0a6cdb09cc4f19125040124e1c162bc5cfef26a8926f0  CentOS-6.5-x86_64-LiveDVD.iso
f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0  CentOS-6.5-x86_64-minimal.iso
d8aaf698408c0c01843446da4a20b1ac03d27f87aad3b3b7b7f42c6163be83b9  CentOS-6.5-x86_64-netinstall.iso`
	c.Name = "CentOS-6.5-whatever.iso"
	c.Image = "whatever"
	s, err = c.findChecksum(checksumPage)
	if err == nil {
		t.Error("Expected error to not be nil, it was")
	} else {
		if err.Error() != "unable to find ISO information while looking for the release string on the CentOS checksums page" {
			t.Errorf("Expected \"unable to find ISO information while looking for the release string on the CentOS checksums page\", got %q", err.Error())
		}
	}

	checksumPage = `c796ab378319393f47b29acd8ceaf21e1f48439570657945226db61702a4a2a1  CentOS-6.5-x86_64-bin-DVD1.iso
afd2fc37e1597c64b3c3464083c0022f436757085d9916350fb8310467123f77  CentOS-6.5-x86_64-bin-DVD2.iso
58b40b26415133ed2af8e2f53b73b5f2aa013723742ce17671b5bb1880a20a99  CentOS-6.5-x86_64-LiveCD.iso
e3efa9a6ca6f58ac4be0a6cdb09cc4f19125040124e1c162bc5cfef26a8926f0  CentOS-6.5-x86_64-LiveDVD.iso
f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0  CentOS-6.5-x86_64-minimal.iso
d8aaf698408c0c01843446da4a20b1ac03d27f87aad3b3b7b7f42c6163be83b9  CentOS-6.5-x86_64-netinstall.iso`
	c.Release = "6"
	c.ReleaseFull = "6.5"
	c.Image = "minimal"
	c.Name = "CentOS-6.5-x86_64-minimal.iso"
	s, err = c.findChecksum(checksumPage)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if s != "f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0" {
			t.Errorf("Expected \"f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0\", got %q", s)
		}
	}
}

func TestCentOSsetISOName(t *testing.T) {
	c := newTestCentOS()
	c.setReleaseInfo()
	c.setISOName()
	if !strings.HasPrefix(c.Name, "CentOS-") {
		t.Errorf("Expected %q to start with \"CentOS-\", it didn't", c.Name)
	}
	if !strings.HasSuffix(c.Name, ".iso") {
		t.Errorf("Expected %q to end with \".iso\", it didn't", c.Name)
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
		if err.Error() != "page to parse was empty; unable to process request for " {
			t.Errorf("Expected \"page to parse was empty; unable to process request for \", got %q", err.Error())
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
		if err.Error() != "unable to find debian-7.8.0-amd64-whatever.iso's checksum" {
			t.Errorf("Expected \"unable to find debian-7.8.0-amd64-whatever.iso's checksum\", got %q", err.Error())
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
	expected = "cdimage.debian.org/debian-cd/7.8.0/amd64/iso-cd/" + d.Name
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
