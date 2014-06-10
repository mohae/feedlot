package ranchr

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSetISOStuff(t *testing.T) {
	Convey("Given a new release object for Ubuntu", t, func() {
		u := ubuntu{}
		u.Release = "14.04"
		u.Image = "server"
		u.Arch = "amd64"
		u.ChecksumType = "sha256"
		u.BaseURL = "http://releases.ubuntu.com/"

		Convey("Given a release of 14.04, image = server, arch = amd64", func() {
			u.SetFilename()

			Convey("The Filename should be 'ubuntu-14.04-server-amd64.iso'", func() {
				So(u.Filename, ShouldEqual, "ubuntu-14.04-server-amd64.iso")
			})
		})


		Convey("Given a release full of 14.04.4, image = server, arch = amd64", func() {
			time.Sleep(100 * time.Millisecond)
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
			u.SetFilename()
			u.SetURL()

			Convey("The result should be", func() {
				So(u.URL, ShouldEqual, "http://releases.ubuntu.com/14.04/ubuntu-14.04-server-amd64.iso")
			})
		})

		Convey("Given the ubuntu 14.04 information, check checksum retreival is working.", func() {
			u.SetFilename()
			u.SetURL()
			if err := u.SetChecksum(); err == nil {
				Convey("The set checksum should be ", func() {
					So(u.Checksum, ShouldEqual, "ababb88a492e08759fddcf4f05e5ccc58ec9d47fa37550d63931d0a5fa4f7388")
				})
			}
		})

		Convey("Check SetChecksum results with an error on getting url", func() {
			u.Release =  ""
			if err := u.SetChecksum(); err != nil {
				Convey("The error should be ", func() {
					So(err.Error(), ShouldEqual, "Unable to find ISO information while looking for the release string on the Ubuntu checksums page.")
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
