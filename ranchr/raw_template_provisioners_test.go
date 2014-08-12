// raw_template_provisioners_test.go: tests for provisioners.
package ranchr

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var pr = &provisioner{
	templateSection{
		Settings: []string{
			"execute_command= echo 'vagrant' | sudo -S sh '{{.Path}}'",
			"type = shell",
		},
		Arrays: map[string]interface{}{
			"override": map[string]interface{}{
				"virtualbox-iso": map[string]interface{}{
					"scripts": []string{
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/virtualbox.sh",
						"scripts/cleanup.sh",
					},
				},
			},
			"scripts": []string{
				"scripts/base.sh",
				"scripts/vagrant.sh",
				"scripts/cleanup.sh",
			},
		},
	},
}

func TestProvisionersSettingsToMap(t *testing.T) {
	Convey("Given a provisioner and a raw template", t, func() {
		Convey("transform settingns map should result in", func() {
			res := pr.settingsToMap("shell", rawTpl)
			Convey("Should result in a map[string]interface{}", func() {
				So(res, ShouldResemble, map[string]interface{}{"type":"shell","execute_command": "echo 'vagrant' | sudo -S sh '{{.Path}}'"})
			})
		})
	})
}

