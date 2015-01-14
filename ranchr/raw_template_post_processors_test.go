// raw_template_post_processors_test.go: tests for post_processors.
package ranchr

import (
	"testing"
)

var pp = &postProcessor{
	templateSection{
		Settings: []string{
			"compression_level=8",
			"keep_input_artifact=true",
		},
		Arrays: map[string]interface{}{
			"override": map[string]interface{}{
				"virtualbox-iso": map[string]interface{}{
					"output": "overridden-virtualbox-iso.box",
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
			Arrays: map[string]interface{}{},
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
					"virtualbox-iso": map[string]interface{}{
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
					"virtualbox-iso": map[string]interface{}{
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
			Arrays: map[string]interface{}{},
		},
	},
}

func TestRawTemplateUpdatePostProcessors(t *testing.T) {
	testDistroDefaults.Templates[CentOS].updatePostProcessors(nil)
	if MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors) != MarshalJSONToString.Get(ppOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(ppOrig), MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors))
	}

	testDistroDefaults.Templates[CentOS].updatePostProcessors(ppNew)
	if MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors) != MarshalJSONToString.Get(ppMerged) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(ppMerged), MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors))
	}
}

func TestPostProcessorsSettingsToMap(t *testing.T) {
	res := pp.settingsToMap("vagrant", rawTpl)
	if MarshalJSONToString.Get(res) != MarshalJSONToString.Get(map[string]interface{}{"type": "vagrant", "compression_level": "8", "keep_input_artifact": true}) {
		t.Errorf("expected %q, got %q", MarshalJSONToString.Get(map[string]interface{}{"type": "vagrant", "compression_level": "8", "keep_input_artifact": true}), MarshalJSONToString.Get(res))
	}
}

func TestRawTemplateCreatePostProcessors(t *testing.T) {
	var pp interface{}
	var err error
	pp, _, err = testDistroDefaults.Templates[CentOS].createPostProcessors()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		expected := []interface{}{
			map[string]interface{}{
				"compression_level":   8,
				"include":             []string{"include1", "include2"},
				"keep_input_artifact": true,
				"only":                []string{"virtualbox-iso"},
				"output":              "out/rancher-packer.box",
				"override": map[string]interface{}{
					"virtualbox-iso": map[string]interface{}{"output": "overridden-virtualbox.box"},
					"vmware-iso":     map[string]interface{}{"output": "overridden-vmware.box"},
				},
				"type": "vagrant",
			},
			nil,
		}
		if MarshalJSONToString.Get(pp) != MarshalJSONToString.Get(expected) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
		}
	}
}

func TestDeepCopyMapStringPPostProcessor(t *testing.T) {
	cpy := DeepCopyMapStringPPostProcessor(ppOrig)
	if MarshalJSONToString.Get(cpy) != MarshalJSONToString.Get(ppOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(ppOrig), MarshalJSONToString.Get(cpy))
	}
}
