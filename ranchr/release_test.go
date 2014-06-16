package ranchr

import (
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
	u.SetFilename()
	u.SetURL()
	
	return u
}

func TestSetISOInfo(t *testing.T) {
	Convey("Given a new release object for Ubuntu", t, func() {
		u := newTestUbuntu()
		Convey("Set ISO info", func() {
			if err := u.SetISOInfo(); err == nil {
				Convey("The URL should be", func() {
					So(u.URL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
				})
				Convey("The Checksum should be", func() {
					So(u.Checksum, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
				})
				Convey("The Filename should be", func() {
					So(u.Filename, ShouldEqual, "ubuntu-14.04-server-amd64.iso")
				})
			}
			
		})

		Convey("Set ISO info, error", func() {
			u.Release = ""
			if err := u.SetISOInfo(); err != nil {
				Convey("Attempt to set ISO information with bad values. The error should be", func() {
					So(err.Error(), ShouldEqual, "Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
				})
			}
		})		
	})
}

func TestSetChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		u := newTestUbuntu()
		Convey("Given the ubuntu 14.04 information, check checksum retrieval is working.", func() {
			if err := u.SetChecksum(); err == nil {
				Convey("The set checksum should be ", func() {
					So(u.Checksum, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
				})
			}
		})

		Convey("Check SetChecksum results with an error on getting url", func() {
			u.ChecksumType =  "ABC"
			u.BaseURL = "http://releasea.ubuntu.com"
			if err := u.SetChecksum(); err != nil {
				Convey("The error should be ", func() {
					So(err.Error(), ShouldEqual, "Get http://releasea.ubuntu.com14.04/ABCSUMS: dial tcp: lookup releasea.ubuntu.com14.04: no such host")
				})
			}
			})
				
		Convey("Check SetChecksum results with an error on parsing url get results", func() {
			u.Filename =  "aslk"
			if err := u.SetChecksum(); err != nil {
				Convey("The error should be ", func() {
					So(err.Error(), ShouldNotEqual, "ssz")
				})
			}
		})
	})
}

func TestSetURL(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		u := newTestUbuntu()
		Convey("Setting the URL", func() {
			u.SetURL()
			So(u.URL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
		})
	})
}

func TestFindChecksum(t *testing.T) {
	Convey("Given a supported distro struct", t, func() {
		var err error
		var s string
		u := newTestUbuntu()
		Convey("Finding the checksum using an empty string to be searched", func() {
			s, err = u.findChecksum("")
			So(err.Error(), ShouldEqual, "the string passed to ubuntu.findChecksum(s string) was empty; unable to process request")
			So(s, ShouldEqual, "")
		})
		Convey("And a results checksums page for the target iso", func() {
			checksumPage := `cab6b0458601520242eb0337ccc9797bf20ad08bf5b23926f354198928191da5 *ubuntu-14.04-desktop-amd64.iso
207a53944d5e8bbb278f4e1d8797491bfbb759c2ebd4a162f41e1383bde38ab2 *ubuntu-14.04-desktop-i386.iso
ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388 *ubuntu-14.04-server-amd64.iso
85c738fefe7c9ff683f927c23f5aa82864866c2391aeb376abfec2dfc08ea873 *ubuntu-14.04-server-i386.iso
bc3b20ad00f19d0169206af0df5a4186c61ed08812262c55dbca3b7b1f1c4a0b *wubi.exe`
			Convey("Finding a an invalid release string", func() {
				u.Filename = "ubuntu-14.03-whatever.iso"
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
		Convey("Given a release of 14.04, image = server, arch = amd64", func() {
			u := newTestUbuntu()

			Convey("The Filename should be 'ubuntu-14.04-server-amd64.iso'", func() {
				So(u.Filename, ShouldEqual, "ubuntu-14.04-server-amd64.iso")
			})
		})


		Convey("Given a release full of 14.04.4, image = server, arch = amd64", func() {
			u := newTestUbuntu()
			u.ReleaseFull = "14.04.4"
			u.SetFilename()
			Convey("The Filename should be ", func() {
				So(u.Filename, ShouldEqual, "ubuntu-14.04.4-server-amd64.iso")
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
