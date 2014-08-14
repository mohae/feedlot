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

var prOrig = map[string]*provisioner{
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

var prNew = map[string]*provisioner{
	templateSection{
		Settings: []string{
			"type = shell",
		},
		Arrays: map[string]interface{}{
			"only": []string{
				"vmware-iso",
			},
			"override": map[string]interface{}{
				"vmware-iso": map[string]interface{}{
					"scripts": []string{
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/vmware.sh",
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

var prMerged = map[string]*provisioner{
	templateSection{
		Settings: []string{
			"execute_command= echo 'vagrant' | sudo -S sh '{{.Path}}'",
			"type = shell",
		},
		Arrays: map[string]interface{}{
			"only": []string{
				"vmware-iso",
			},
			"override": map[string]interface{}{
				"vmware-iso": map[string]interface{}{
					"scripts": []string{
						"scripts/base.sh",
						"scripts/vagrant.sh",
						"scripts/vmware.sh",
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

func TestRawTemplateUpdateProvisioners(t *testing.T) {
	Convey("Given a template", t, func() {
		Convey("Updating Provisioners with nil", func() {
			testDistroDefaults.Templates["centos"].updateProvisioners(nil)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.Get(testDistroDefaults.Templates["centos"].Provisioners), ShouldEqual, MarshalJSONToString.Get(prOrig))
			})
		})
		Convey("Updating Provisioners with new values", func() {
			testDistroDefaults.Templates["centos"].updateProvisioners(prNew)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.Get(testDistroDefaults.Templates["centos"].Provisioners), ShouldEqual, MarshalJSONToString.Get(prMerged))
			})
		})
	
	})
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

func TestRawTemplateCreateProvisioners(t *testing.T) {
	Convey("Given a template", t, func() {
		var prov interface{}
		var err error
		Convey("Creating Provisioners", func() {
			prov, _, err = testDistroDefaults.Templates["centos"].createProvisioners()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in Provisioners", func() {
				So(MarshalJSONToString.Get(prov), ShouldEqual, "[{\"compression_level\":\"8\",\"include\":[\"include1\",\"include2\"],\"keep_input_artifact\":\"true\",\"only\":[\"virtualbox-iso\"],\"output\":\"out/rancher-packer.box\",\"type\":\"vagrant\"},null]")
			})
		})
	})
}


