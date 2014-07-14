package ranchr

import (
	_ "net/url"
	_ "strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestUbuntu() ubuntu {
	u := ubuntu{release{Release: "14.04", Image: "server", Arch: "amd64"}}
	u.ChecksumType = "sha256"
	u.BaseURL = "http://releases.ubuntu.com/"
	return u
}

func TestUbuntuSetISOInfo(t *testing.T) {
	Convey("Given a new release object for Ubuntu", t, func() {
		u := newTestUbuntu()
		Convey("Set ISO info", func() {
			err := u.SetISOInfo()
			Convey("Should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("The URL should be", func() {
				So(u.isoURL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
			})
			Convey("The Checksum should be", func() {
				So(u.Checksum, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
			})
			Convey("The Name should be", func() {
				So(u.Name, ShouldEqual, "ubuntu-14.04-server-amd64.iso")
			})
		})
		Convey("Set ISO info, error", func() {
			u.Release = ""
			err := u.SetISOInfo()
			Convey("Attempt to set ISO information with bad values. The error should be", func() {
				So(err.Error(), ShouldEqual, "Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
			})
		})
	})
}

func TestUbuntuSetChecksum(t *testing.T) {
	Convey("Given an Ubuntu struct", t, func() {
		u := newTestUbuntu()
		Convey("Given the ubuntu 14.04 information, check checksum retrieval is working.", func() {
			err := u.setChecksum()
			Convey("Setting the Checksum should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("The set checksum should be ", func() {
				So(u.Checksum, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
			})
		})
		Convey("Check SetChecksum results with an error on getting url", func() {
			u.ChecksumType = "ABC"
			u.BaseURL = "http://releasea.ubuntu.com"
			err := u.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldEqual, "Get http://releasea.ubuntu.com/14.04/ABCSUMS: dial tcp: lookup releasea.ubuntu.com: no such host")
			})
		})
		Convey("Calling checksum with an invalid filename but valid settings for an iso", func() {
			u.Name = "aslk"
			err := u.setChecksum()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in a valid checksum as a result of the extra processing necessary to handle incrimental releases, e.g. 12.04.4 vs 12.04, which rebuilds the filename, resulting in a valid filename as long as the other settings are valid", func() {
				So(u.Checksum, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
			})
		})
		Convey("Check SetChecksum results with an error on parsing url get results", func() {
			// use at least 1 invalid setting
			u.Image = "random"
			err := u.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldEqual, "Unable to retrieve checksum while looking for ubuntu-14.04-random-amd64.iso on the Ubuntu checksums page.")
			})
		})
	})
}

func TestUbuntuSetURL(t *testing.T) {
	Convey("Given an Ubuntu struct", t, func() {
		u := newTestUbuntu()
		u.setName()
		Convey("Setting the URL", func() {
			u.setISOURL()
			So(u.isoURL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
		})
	})
}

func TestUbuntuFindChecksum(t *testing.T) {
	Convey("Given an Ubuntu struct", t, func() {
		var err error
		var s string
		u := newTestUbuntu()
		Convey("Finding the checksum using an empty string to be searched", func() {
			s, err = u.findChecksum("")
			So(err.Error(), ShouldEqual, "the string passed to ubuntu.findChecksum(isoName string) was empty; unable to process request")
			So(s, ShouldEqual, "")
		})
		Convey("And a results checksums page for the target iso", func() {
			checksumPage := `cab6b0458601520242eb0337ccc9797bf20ad08bf5b23926f354198928191da5 *ubuntu-14.04-desktop-amd64.iso
207a53944d5e8bbb278f4e1d8797491bfbb759c2ebd4a162f41e1383bde38ab2 *ubuntu-14.04-desktop-i386.iso
ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388 *ubuntu-14.04-server-amd64.iso
85c738fefe7c9ff683f927c23f5aa82864866c2391aeb376abfec2dfc08ea873 *ubuntu-14.04-server-i386.iso
bc3b20ad00f19d0169206af0df5a4186c61ed08812262c55dbca3b7b1f1c4a0b *wubi.exe`
			Convey("Finding a an invalid release string", func() {
				u.Name = "ubuntu-14.03-whatever.iso"
				u.Image = "whatever"
				s, err = u.findChecksum(checksumPage)
				So(err.Error(), ShouldEqual, "Unable to retrieve checksum while looking for ubuntu-14.04-whatever-amd64.iso on the Ubuntu checksums page.")
				So(s, ShouldEqual, "")
			})
		})
		Convey("And a results checksums page for the target iso", func() {
			checksumPage := `cab6b0458601520242eb0337ccc9797bf20ad08bf5b23926f354198928191da5 *ubuntu-14.04-desktop-amd64.iso
207a53944d5e8bbb278f4e1d8797491bfbb759c2ebd4a162f41e1383bde38ab2 *ubuntu-14.04-desktop-i386.iso
ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388 *ubuntu-14.04-server-amd64.iso
85c738fefe7c9ff683f927c23f5aa82864866c2391aeb376abfec2dfc08ea873 *ubuntu-14.04-server-i386.iso
bc3b20ad00f19d0169206af0df5a4186c61ed08812262c55dbca3b7b1f1c4a0b *wubi.exe`
			Convey("Finding a valid release string", func() {
				u.Release = "14.03"
				s, err = u.findChecksum(checksumPage)
				So(err, ShouldBeNil)
				So(s, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
			})
		})
		Convey("And a results checksums page for ubuntu 12.04.4", func() {
			checksumPage := `3a80be812e7767e1e78b4e1aeab4becd11a677910877250076096fe3b470ba22 *ubuntu-12.04.4-alternate-amd64.iso
bf48efb08cb1962bebaabeb81e4a0a00c6fd7dff5ff50927d0ba84923d0deed4 *ubuntu-12.04.4-alternate-i386.iso
fa28d4b4821d6e8c5e5543f8d9f5ed8176400e078fe9177fa2774214b7296c84 *ubuntu-12.04.4-desktop-amd64.iso
c0ba532d8fadaa3334023f96925b93804e859dba2b4c4e4cda335bd1ebe43064 *ubuntu-12.04.4-desktop-i386.iso
3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54 *ubuntu-12.04.4-server-amd64.iso
fbe7f159337551cc5ce9f0ff72acefef567f3dcd30750425287588c554978501 *ubuntu-12.04.4-server-i386.iso
2d92c01bcfcd0911c5e4256250647c30e3351f3190515f972f83027c4260e7e5 *ubuntu-12.04.4-wubi-amd64.tar.xz
2bde9af4e9f8a6ce493955188a624cb3ffdf46294805605d9d51a1ee62997dcf *ubuntu-12.04.4-wubi-i386.tar.xz
819f42fdd7cc431b6fd7fa5bae022b0a8c55a0f430eb3681e4750c4f1eceaf91 *wubi.exe`
			Convey("Finding a valid release using an updated image number string", func() {
				u.Release = "12.04"
				s, err = u.findChecksum(checksumPage)
				So(err, ShouldBeNil)
				So(s, ShouldEqual, "3aeb42816253355394897ae80d99a9ba56217c0e98e05294b51f0f5b13bceb54")
			})
		})
	})
}

func TestUbuntuGetOSType(t *testing.T) {
	Convey("Given an ubuntu struct getting the OS Type from a string", t, func() {
		Convey("Given amd64 architecture", func() {
			u := ubuntu{release{Arch: "amd64"}}
			Convey("Given a vmware builder", func() {
				buildType := "vmware-iso"
				Convey("Calling ubuntu.getOSType should result in", func() {
					res, err := u.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "ubuntu-64")
				})

			})
			Convey("Given a virtualbox builder", func() {
				buildType := "virtualbox-iso"
				Convey("Calling ubuntu.getOSType should result in", func() {
					res, err := u.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "Ubuntu_64")
				})
			})
		})
		Convey("Given i386 architecture", func() {
			u := ubuntu{release{Arch: "i386"}}
			Convey("Given a vmware builder", func() {
				buildType := "vmware-iso"
				Convey("Calling ubuntu.getOSType should result in", func() {
					res, err := u.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "ubuntu-32")
				})

			})
			Convey("Given a virtualbox builder", func() {
				buildType := "virtualbox-iso"
				Convey("Calling ubuntu.getOSType should result in", func() {
					res, err := u.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "Ubuntu_32")
				})
			})
			Convey("Given an invalid builder", func() {
				buildType := "voodoo"
				Convey("Calling ubuntu.getOSType should result in", func() {
					res, err := u.getOSType(buildType)
					So(err.Error(), ShouldEqual, "ubuntu.getOSType: the builder 'voodoo' is not supported")
					So(res, ShouldEqual, "")
				})
			})
		})
	})
}

func newTestCentOS() centOS {
	c := centOS{release{Release: "6", Image: "minimal", Arch: "x86_64"}}
	c.ChecksumType = "sha256"
	return c
}

func TestCentOSISORedirectURL(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		Convey("calling isoRedirectURL", func() {
			redirectURL := c.isoRedirectURL()
			So(redirectURL, ShouldEqual, "http://isoredirect.centos.org/centos/"+c.Release+"/isos/"+c.Arch+"/")
		})
	})
}

func TestSetISOInfo(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		Convey("calling SetISOInfo", func() {
			err := c.SetISOInfo()
			So(err, ShouldBeNil)
			So(c.ReleaseFull, ShouldStartWith, "6")
			So(c.Name, ShouldNotEqual, "")
			So(c.isoURL, ShouldStartWith, "http://")
			So(c.isoURL, ShouldEndWith, ".iso")
		})
		Convey("calling SetISOInfo with empty Arch", func() {
			c.Arch = ""
			err := c.SetISOInfo()
			So(err.Error(), ShouldEqual, "Arch was empty, unable to continue")
		})
		Convey("calling SetISOInfo with empty Arch", func() {
			c.Release = ""
			err := c.SetISOInfo()
			So(err.Error(), ShouldEqual, "Release was empty, unable to continue")
		})
		Convey("calling SetISOInfo with an invalid baseURL", func() {
			c.BaseURL = "http://example.com/"
			err := c.SetISOInfo()
			So(err.Error(), ShouldEqual, "Unable to find ISO information while looking for the release string on the CentOS checksums page.")
		})
		Convey("calling SetISOInfo with an empty checksumType", func() {
			c.ChecksumType = ""
			err := c.SetISOInfo()
			So(err.Error(), ShouldEqual, "Checksum Type not set")
		})
	})
}

func TestCentOSsetReleaseNumber(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		Convey("calling setReleseNumber", func() {
			err := c.setReleaseNumber()
			So(err, ShouldBeNil)
			// Only test for the first part, as the sub-version number may change
			So(c.ReleaseFull, ShouldStartWith, "6.")
		})
	})
}

func TestCentOSsetReleaseInfo(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		Convey("calling setReleseInfo", func() {
			err := c.setReleaseInfo()
			So(err, ShouldBeNil)
			// Release should always be a whole number
			So(c.Release, ShouldEqual, "6")
		})
	})
}

func TestCentOSSetName(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		c.setReleaseInfo()
		Convey("Setting the name", func() {
			c.setName()
			So(c.Name, ShouldEqual, "CentOS-"+c.ReleaseFull+"-"+c.Arch+"-"+c.Image+".iso")
		})
	})
}

func TestCentOSGetOSType(t *testing.T) {
	Convey("Given an centOS struct getting the OS Type from a string", t, func() {
		Convey("Given x86_64 architecture", func() {
			c := centOS{release{Arch: "x86_64"}}
			Convey("Given a vmware builder", func() {
				buildType := "vmware-iso"
				Convey("Calling centOS.getOSType should result in", func() {
					res, err := c.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "centos-64")
				})

			})
			Convey("Given a virtualbox builder", func() {
				buildType := "virtualbox-iso"
				Convey("Calling centOS.getOSType should result in", func() {
					res, err := c.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "RedHat_64")
				})
			})
		})
		Convey("Given x386 architecture", func() {
			c := centOS{release{Arch: "x386"}}
			Convey("Given a vmware builder", func() {
				buildType := "vmware-iso"
				Convey("Calling centOS.getOSType should result in", func() {
					res, err := c.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "centos-32")
				})

			})
			Convey("Given a virtualbox builder", func() {
				buildType := "virtualbox-iso"
				Convey("Calling centOS.getOSType should result in", func() {
					res, err := c.getOSType(buildType)
					So(err, ShouldBeNil)
					So(res, ShouldEqual, "RedHat_32")
				})
			})
			Convey("Given an invalid builder", func() {
				buildType := "voodoo"
				Convey("Calling centOS.getOSType should result in", func() {
					res, err := c.getOSType(buildType)
					So(err.Error(), ShouldEqual, "centOS.getOSType: the builder 'voodoo' is not supported")
					So(res, ShouldEqual, "")
				})
			})
		})
	})
}

func TestCentOSSetChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		c := newTestCentOS()
		c.setReleaseInfo()
		Convey("Given the CentOS 6 information, check checksum retrieval is working.", func() {
			Convey("setting the Checksum without the baseURL set", func() {
				err := c.setChecksum()
				So(err.Error(), ShouldEqual, "Get sha256sum.txt: unsupported protocol scheme \"\"")
			})
			Convey("Setting the Checksum without the checksumType set", func() {
				c.ChecksumType = ""
				err := c.setChecksum()
				So(err.Error(), ShouldEqual, "Checksum Type not set")
			})
			Convey("Setting the checksum with the baseurl set  ", func() {
				c.ChecksumType = "sha256"
				c.setISOURL()
				err := c.setChecksum()
				So(err, ShouldBeNil)
				// TODO investigate this test further
				//				So(c.Checksum, ShouldEqual, "f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0")
			})
		})
		Convey("Check SetChecksum results with an error on getting url", func() {
			c.ChecksumType = "ABC"
			c.BaseURL = "http://example.com/notaurl/" + c.ReleaseFull + "/isos/" + c.Arch + "/"
			err := c.setChecksum()
			Convey("The error should be ", func() {
				//				So(err.Error(), ShouldEqual, "Get " + c.BaseURL + "abcsum.txt: dial tcp: lookup adfarfawer.com: no such host")
				So(err, ShouldNotBeNil)
			})
		})
		Convey("Check SetChecksum results with an error on parsing url get results", func() {
			c.Name = "aslk"
			err := c.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldNotEqual, "ssz")
			})
		})
	})
}

func TestCentOSChecksumURL(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		c.setReleaseInfo()
		c.setISOURL()
		Convey("Setting the ISO URL", func() {
			checksumURL := c.checksumURL()
			So(checksumURL, ShouldStartWith, "http://")
			So(checksumURL, ShouldEndWith, c.ReleaseFull+"/isos/"+c.Arch+"/sha256sum.txt")
		})
	})
}

func TestCentOSsetISOURL(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		c.setReleaseInfo()
		c.setName()
		Convey("Setting the ISO URL", func() {
			err := c.setISOURL()
			So(err, ShouldBeNil)
			So(c.isoURL, ShouldStartWith, "http://")
			So(c.isoURL, ShouldEndWith, c.ReleaseFull+"/isos/"+c.Arch+"/CentOS-"+c.ReleaseFull+"-"+c.Arch+"-"+c.Image+".iso")
		})
		Convey("Setting the isoURL without an empty BaseURL", func() {
			c.BaseURL = "http://example.com/"
			err := c.setISOURL()
			So(err, ShouldBeNil)
			So(c.isoURL, ShouldEqual, "http://example.com/CentOS-"+c.ReleaseFull+"-"+c.Arch+"-"+c.Image+".iso")
		})
	})
}

func TestCentOSrandomISOURL(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		c.setReleaseInfo()
		c.setName()
		Convey("Setting the ISO URL", func() {
			randomURL, err := c.randomISOURL()
			So(err, ShouldBeNil)
			So(randomURL, ShouldStartWith, "http://")
			So(randomURL, ShouldEndWith, c.ReleaseFull+"/isos/"+c.Arch+"/CentOS-"+c.ReleaseFull+"-"+c.Arch+"-"+c.Image+".iso")
		})
	})
}

func TestCentOSFindChecksum(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		var err error
		var s string
		c := newTestCentOS()
		c.setReleaseInfo()
		c.setName()
		c.setISOURL()
		Convey("Finding the checksum using an empty string to be searched", func() {
			s, err = c.findChecksum("")
			So(err.Error(), ShouldEqual, "the string passed to centOS.findChecksum(s string) was empty; unable to process request")
			So(s, ShouldEqual, "")
		})
		Convey("And a results checksums page for the target iso", func() {
			checksumPage := `c796ab378319393f47b29acd8ceaf21e1f48439570657945226db61702a4a2a1  CentOS-6.5-x86_64-bin-DVD1.iso
afd2fc37e1597c64b3c3464083c0022f436757085d9916350fb8310467123f77  CentOS-6.5-x86_64-bin-DVD2.iso
58b40b26415133ed2af8e2f53b73b5f2aa013723742ce17671b5bb1880a20a99  CentOS-6.5-x86_64-LiveCD.iso
e3efa9a6ca6f58ac4be0a6cdb09cc4f19125040124e1c162bc5cfef26a8926f0  CentOS-6.5-x86_64-LiveDVD.iso
f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0  CentOS-6.5-x86_64-minimal.iso
d8aaf698408c0c01843446da4a20b1ac03d27f87aad3b3b7b7f42c6163be83b9  CentOS-6.5-x86_64-netinstall.iso`
			Convey("Finding a an invalid release string", func() {
				c.Name = "CentOS-6.5-whatever.iso"
				c.Image = "whatever"
				s, err = c.findChecksum(checksumPage)
				So(err.Error(), ShouldEqual, "Unable to find ISO information while looking for the release string on the CentOS checksums page.")
				So(s, ShouldEqual, "")
			})
		})
		Convey("And a results checksums page for the target iso", func() {
			checksumPage := `c796ab378319393f47b29acd8ceaf21e1f48439570657945226db61702a4a2a1  CentOS-6.5-x86_64-bin-DVD1.iso
afd2fc37e1597c64b3c3464083c0022f436757085d9916350fb8310467123f77  CentOS-6.5-x86_64-bin-DVD2.iso
58b40b26415133ed2af8e2f53b73b5f2aa013723742ce17671b5bb1880a20a99  CentOS-6.5-x86_64-LiveCD.iso
e3efa9a6ca6f58ac4be0a6cdb09cc4f19125040124e1c162bc5cfef26a8926f0  CentOS-6.5-x86_64-LiveDVD.iso
f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0  CentOS-6.5-x86_64-minimal.iso
d8aaf698408c0c01843446da4a20b1ac03d27f87aad3b3b7b7f42c6163be83b9  CentOS-6.5-x86_64-netinstall.iso`
			Convey("Finding a valid release string", func() {
				c.Release = "6"
				c.ReleaseFull = "6.5"
				c.Image = "minimal"
				c.Name = "CentOS-6.5-x86_64-minimal.iso"
				s, err = c.findChecksum(checksumPage)
				So(err, ShouldBeNil)
				So(s, ShouldEqual, "f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0")
			})
		})
	})
}

func TestCentOSsetName(t *testing.T) {
	Convey("Given a CentOS struct", t, func() {
		c := newTestCentOS()
		c.setReleaseInfo()
		Convey("calling setName", func() {
			c.setName()
			So(c.Name, ShouldStartWith, "CentOS-")
			So(c.Name, ShouldEndWith, ".iso")
		})
	})
}
