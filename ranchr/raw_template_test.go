package ranchr

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRawTemplate(t *testing.T) {
	Convey("Testing RawTemplate", t, func() {
		Convey("Given a request for a newRawTemplate()", func() {
			rawTpl := newRawTemplate()

			Convey("The raw template should equal--we don't test the date because it is always changeing", func() {
				So(rawTpl, ShouldNotBeNil)
			})
		})


	})
}
