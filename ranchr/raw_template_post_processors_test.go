// raw_template_post_processors_test.go: tests for post_processors.
package ranchr

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var pp = &postProcessor{
	templateSection{
		Settings: []string{
			"compression_level=8",
			"keep_input_artifact=true",
		},
		Arrays: map[string]interface{}{
			"override": map[string]interface{}{
				"virtualbox": map[string]interface{}{
					"output": "overridden-virtualbox.box",
				},
			},
		},
	},
}

func TestPostProcessorsSettingsToMap(t *testing.T) {
	Convey("Given a postProcessor and a raw template", t, func() {
		Convey("transform settings to map should result in", func() {
			res := pp.settingsToMap("vagrant", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(MarshalJSONToString.Get(res), ShouldEqual, MarshalJSONToString.Get(map[string]interface{}{"type": "vagrant", "compression_level": "8", "keep_input_artifact": true}))
			})
		})
	})
}
