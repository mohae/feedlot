package ranchr

import (
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestUbuntu() ubuntu {
	u := ubuntu{}
	u.Release = "14.04"
	u.Image = "server"
	u.Arch = "amd64"
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
				So(u.URL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
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
			err := u.SetISOInfo();
			Convey("Attempt to set ISO information with bad values. The error should be", func() {
				So(err.Error(), ShouldEqual, "Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
			})
		})		
	})
}

func TestUbuntuSetChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
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
			u.ChecksumType =  "ABC"
			u.BaseURL = "http://releasea.ubuntu.com"
			err := u.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldEqual, "Get http://releasea.ubuntu.com14.04/ABCSUMS: dial tcp: lookup releasea.ubuntu.com14.04: no such host")
			})
		})
		Convey("Calling checksum with an invalid filename but valid settings for an iso", func() {
			u.Name =  "aslk"
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
			u.Image =  "random"
			err := u.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldEqual, "Unable to retrieve checksum while looking for ubuntu-14.04-random-amd64.iso on the Ubuntu checksums page.")
			})
		})
	})
}

func TestUbuntuSetURL(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		u := newTestUbuntu()
		u.setName()
		Convey("Setting the URL", func() {
			u.setURL()
			So(u.URL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
		})
	})
}

func TestUbuntuFindChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
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

/*
getPAge testing?
		Convey("Given a release of 14.04, image = server, arch = amd64", func() {
			u := newTestUbuntu()

			Convey("The Name should be 'ubuntu-14.04-server-amd64.iso'", func() {
				So(u.Name, ShouldEqual, "ubuntu-14.04-server-amd64.iso")
			})S
		})


		Convey("Given a release full of 14.04.4, image = server, arch = amd64", func() {
			u := newTestUbuntu()
			u.ReleaseFull = "14.04.4"
			u.SetName()
			Convey("The Name should be ", func() {
				So(u.Name, ShouldEqual, "ubuntu-14.04.4-server-amd64.iso")
			})
		})

		Convey("Given a blank url", func() {
			_, err := getStringFromURL("");
	
			Convey("The result should be an error", func() {
				So(err.Error(), ShouldEqual, "Get : unsupported protocol scheme \"\"")
			})
		})

		Convey("Given a local url", func() {
			if res, err := getStringFromURL("localhost:6060"); err == nil {
				Convey("The result should be ", func() {
					So(res, ShouldEqual, "")
				})
			}
		})


		Convey("Given the base url, release, and filename is set", func() {
			u := newTestUbuntu()
			Convey("The result should be", func() {
				So(u.URL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
			})
		})
			
		Convey("Given a url", func() {
			if res, err := getStringFromURL("http://www.example.com"); err == nil {
				Convey("The result should be ", func() {
					So(res, ShouldEqual, 
`<!doctype html>
<html>
<head>
    <title>Example Domain</title>

    <meta charset="utf-8" />
    <meta http-equiv="Content-type" content="text/html; charset=utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <style type="text/css">
    body {
        background-color: #f0f0f2;
        margin: 0;
        padding: 0;
        font-family: "Open Sans", "Helvetica Neue", Helvetica, Arial, sans-serif;
        
    }
    div {
        width: 600px;
        margin: 5em auto;
        padding: 50px;
        background-color: #fff;
        border-radius: 1em;
    }
    a:link, a:visited {
        color: #38488f;
        text-decoration: none;
    }
    @media (max-width: 700px) {
        body {
            background-color: #fff;
        }
        div {
            width: auto;
            margin: 0 auto;
            border-radius: 0;
            padding: 1em;
        }
    }
    </style>    
</head>

<body>
<div>
    <h1>Example Domain</h1>
    <p>This domain is established to be used for illustrative examples in documents. You may use this
    domain in examples without prior coordination or asking for permission.</p>
    <p><a href="http://www.iana.org/domains/example">More information...</a></p>
</div>
</body>
</html>
`)
				})
			}
		})
	})
}
*/

func newTestCentOS() centOS {
	c := centOS{}
	c.Release = "6"
	c.Image = "minimal"
	c.Arch = "x86_64"
	c.ChecksumType = "sha256"
	return c
}

func TestCentOSSetURL(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		c := newTestCentOS()
		Convey("Setting the URL", func() {
			// Since the baseurl is picked at random, to respect the mirrolist
			// structure, just make sure it isn't empty.	
			err := c.setBaseURL()
			So(err, ShouldBeNil)
			So(c.BaseURL, ShouldNotEqual, "")
			// And release full should be set
			So(c.ReleaseFull, ShouldNotEqual, "")
			//Make sure the name is set
			Convey("Setting the iso name", func() {
				c.setName()
				So(c.Name, ShouldEqual, "CentOS-" + c.ReleaseFull + "-" + c.Arch + "-minimal.iso")
				Convey("Setting the BaseURL", func() {
					c.setURL()
					So(c.URL, ShouldNotEqual, "")
				})
			})
		})
	})
}

func TestCentOSFindChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		var err error
		var s string
		c := newTestCentOS()
		c.setBaseURL()
		c.setName()
		c.setURL()
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

func TestCentOSSetISOInfo(t *testing.T) {
	Convey("Given a new release object for CentOS", t, func() {
		c := newTestCentOS()
		Convey("Set ISO info", func() {	
			err := c.SetISOInfo()
			Convey("Should not result in an  error", func() {
				So(err, ShouldBeNil)
			})
			Convey("The URL should be", func() {
				So(c.URL, ShouldNotEqual, "")
				Convey("And it should be a valid url", func(){
					u, err := url.Parse(c.URL)
					So(err, ShouldBeNil)
					So(u.Scheme, ShouldEqual, "http")
					So(u.Host, ShouldNotEqual, "")
					parts :=  strings.Split(u.Path,"/")
					l := len(parts)
					So(parts[l-4], ShouldEqual, "6.5")
					So(parts[l-3], ShouldEqual, "isos")
					So(parts[l-2], ShouldEqual, "x86_64")
				})
			})
			Convey("The BaseURL should exist", func() {
				// Since this is random we just check that scheme, host, and path exist
				So(c.BaseURL, ShouldNotEqual, "")
				Convey("And it should be a valid url", func() {
					u, err := url.Parse(c.BaseURL)
					So(err, ShouldBeNil)
					So(u.Scheme, ShouldNotEqual, "")
					So(u.Scheme, ShouldNotEqual, "ftp")
					So(u.Host, ShouldNotEqual, "")
					So(u.Path, ShouldNotEqual,"")
				})
			})
			Convey("The Checksum should be", func() {
				So(c.Checksum, ShouldEqual, "f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0")
			})
			Convey("The Name should be", func() {
				So(c.Name, ShouldEqual, "CentOS-6.5-x86_64-minimal.iso")
			})
		})
		Convey("Set ISO info, error", func() {
			c.Release = ""
			err := c.SetISOInfo()
			Convey("Attempt to set ISO information with bad values. The error should be", func() {
				So(err.Error(), ShouldEqual, "Unable to set BaseURL information for CentOS because the Release was not set.")
			})
		})		
	})
}


func TestCentOSSetChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		c := newTestCentOS()
		c.setBaseURL()
		c.setURL()
		Convey("Given the CentOS 6 information, check checksum retrieval is working.", func() {
			err := c.setChecksum()
			Convey("Should not result in an error", func() {
				So(err, ShouldBeNil)
			})
			Convey("The set checksum should be ", func() {
				So(c.Checksum, ShouldEqual, "f9d84907d77df62017944cb23cab66305e94ee6ae6c1126415b81cc5e999bdd0")
			})
		})
		Convey("Check SetChecksum results with an error on getting url", func() {
			c.ChecksumType =  "ABC"
			c.BaseURL = "http://adfarfawer.com/notaurl"
			err := c.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldEqual, "Get http://adfarfawer.com/notaurl/" + c.ReleaseFull + "/isos/" + c.Arch + "/abcsum.txt: dial tcp: lookup adfarfawer.com: no such host")
			})
		})		
		Convey("Check SetChecksum results with an error on parsing url get results", func() {
			c.Name =  "aslk"
			err := c.setChecksum()
			Convey("The error should be ", func() {
				So(err.Error(), ShouldNotEqual, "ssz")
			})
		})
	})
}
