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

var ppOrig = map[string]*postProcessor{
	"vagrant": {
		templateSection{
			Settings: []string{
				"compression_level = 9",
				"keep_input_artifact = false",
				"output = out/rancher-packer.box",
			},
			Arrays: map[string]interface{}{
				"include": []string{
					"include1",
					"include2",
				},
				"only": []string{
					"virtualbox-iso",
				},
			},
		},
	},
	"vagrant-cloud": {
		templateSection{
			Settings: []string{
				"access_token = getAValidTokenFrom-VagrantCloud.com",
				"box_tag = foo/bar",
				"no_release = true",
				"version = 1.0.1",
			},
//			Arrays: map[string]interface{}{},
		},
	},
}

var ppNew = map[string]*postProcessor{
	"vagrant": {
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
					"vmware-iso": map[string]interface{}{
						"output": "overridden-vmware.box",
					},
				},
			},
		},
	},
}

var ppMerged = map[string]*postProcessor{
	"vagrant": {
		templateSection{
			Settings: []string{
				"compression_level=8",
				"keep_input_artifact=true",
				"output = out/rancher-packer.box",
			},
			Arrays: map[string]interface{}{
				"include": []string{
					"include1",
					"include2",
				},
				"only": []string{
					"virtualbox-iso",
				},
				"override": map[string]interface{}{
					"virtualbox": map[string]interface{}{
						"output": "overridden-virtualbox.box",
					},
					"vmware-iso": map[string]interface{}{
						"output": "overridden-vmware.box",
					},
				},
			},
		},
	},
	"vagrant-cloud": {
		templateSection{
			Settings: []string{
				"access_token = getAValidTokenFrom-VagrantCloud.com",
				"box_tag = foo/bar",
				"no_release = true",
				"version = 1.0.1",
			},
		},
	},
}

func TestRawTemplateUpdatePostProcessors(t *testing.T) {
	Convey("Given a template", t, func() {
		Convey("Updating PostProcessores with nil", func() {
			testDistroDefaults.Templates["centos"].updatePostProcessors(nil)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.Get(testDistroDefaults.Templates["centos"].PostProcessors), ShouldEqual, MarshalJSONToString.Get(ppOrig))
			})
		})
		Convey("Updateing PostProcessors with new values", func() {
			testDistroDefaults.Templates["centos"].updatePostProcessors(ppNew)
			Convey("Should result in no changes", func() {
				So(MarshalJSONToString.Get(testDistroDefaults.Templates["centos"].PostProcessors), ShouldEqual, MarshalJSONToString.Get(ppMerged))
			})
		})
	
	})
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

func TestRawTemplateCreatePostProcessors(t *testing.T) {
	Convey("Given a template", t, func() {
		var pp interface{}
		var err error
		Convey("Creating PostProcessors", func() {
			pp, _, err = testDistroDefaults.Templates["centos"].createPostProcessors()
			Convey("Should not error", func() {
				So(err, ShouldBeNil)
			})
			Convey("Should result in postProcessors", func() {
				So(MarshalJSONToString.Get(pp), ShouldEqual, "[{\"compression_level\":\"8\",\"include\":[\"include1\",\"include2\"],\"keep_input_artifact\":\"true\",\"only\":[\"virtualbox-iso\"],\"output\":\"out/rancher-packer.box\",\"type\":\"vagrant\"},null]")
			})
		})
	})
}

func TestDeepCopyMapStringPPostProcessor(t *testing.T) {
	Convey("Given a map[string]*postProcessor", t, func() {
		Convey("Doing a deep copy of it", func() {
			copy := DeepCopyMapStringPPostProcessor(ppOrig)
			Convey("Should result in a copy", func() {
				So(MarshalJSONToString.Get(copy), ShouldEqual,MarshalJSONToString.Get(ppOrig))
			})
		})
	})
}
