// raw_template_post_processors_test.go: tests for post_processors.
package app

import (
	"reflect"
	"testing"
)

var testPostProcessorsAllTemplate = &RawTemplate{
	PackerInf: PackerInf{
		Description: "Test build template #1",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "12.04",
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/out/:build_name",
		PackerOutputDir:   "packer_boxes/:build_name",
		SourceDir:         "../test_files/src",
	},
	VarVals: map[string]string{},
	Dirs:    map[string]string{},
	Files:   map[string]string{},
	Build: Build{
		BuilderIDs: []string{
			"virtualbox-iso",
		},
		Builders: map[string]BuilderC{
			"common": {
				TemplateSection{
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				TemplateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
						},
					},
				},
			},
		},
		PostProcessorIDs: []string{
			"compress",
			"vagrant",
		},
		PostProcessors: map[string]PostProcessorC{
			"atlas": {
				TemplateSection{
					Settings: []string{
						"artifact=hashicorp/foobar",
						"artifact_type=aws.ami",
						"token={{user `atlas_token`}}",
					},
					Arrays: map[string]interface{}{
						"metadata": map[string]string{
							"created_at": "{{timestamp}}",
						},
					},
				},
			},
			"compress": {
				TemplateSection{
					Settings: []string{
						"output = foo.tar.gz",
					},
				},
			},
			"docker-import": {
				TemplateSection{
					Settings: []string{
						"repository = mitchellh/packer",
						"tag = 0.7",
					},
				},
			},
			"docker-push": {
				TemplateSection{
					Settings: []string{
						"login = false",
						"login_email = email@test.com",
						"login_username = username",
						"login_password = password",
						"login_server = server.test.com",
					},
				},
			},
			"docker-save": {
				TemplateSection{
					Settings: []string{
						"path = save/path",
					},
				},
			},
			"docker-tag": {
				TemplateSection{
					Settings: []string{
						"repository = mitchellh/packer",
						"tag = 0.7",
					},
				},
			},
			"vagrant": {
				TemplateSection{
					Settings: []string{
						"compression_level = 6",
						"keep_input_artifact = false",
						"output = :out_dir/packer.box",
						"vagrantfile_template = template/VagrantFile.template",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"include": []string{
							"include/include1",
							"include/include2",
						},
					},
				},
			},
			"vagrant-cloud": {
				TemplateSection{
					Settings: []string{
						"access_token = vagrant-cloud-token",
						"box_download_url=download.example.com/box",
						"box_tag=hashicorp/precise64",
						"no_release=true",
						"vagrant_cloud_url=https://vagrantcloud.com/api/v1",
						"version=0.0.1",
						"version_description=initial",
					},
				},
			},
			"vsphere": {
				TemplateSection{
					Settings: []string{
						"cluster=target-cluster",
						"datacenter=target-datacenter",
						"datastore=vm-datastore",
						"disk_mode=thick",
						"host=vsphere-host",
						"insecure=false",
						"password=password",
						"resource_pool=rpool",
						"username=username",
						"vm_folder=vm-folder",
						"vm_name=packervm",
						"vm_network=vm-network",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
		},
		Provisioners: map[string]ProvisionerC{
			"shell": {
				TemplateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"cleanup_test.sh",
						},
					},
				},
			},
		},
	},
}

var pp = &PostProcessorC{
	TemplateSection{
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

var ppOrig = map[string]PostProcessorC{
	"vagrant": {
		TemplateSection{
			Settings: []string{
				"compression_level = 9",
				"keep_input_artifact = false",
				"output = out/feedlot-packer.box",
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
		TemplateSection{
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

var ppNew = map[string]PostProcessorC{
	"vagrant": {
		TemplateSection{
			Settings: []string{
				"compression_level=8",
				"keep_input_artifact=true",
			},
			Arrays: map[string]interface{}{
				"only": []string{
					"digitalocean",
				},
				"except": []string{
					"googlecompute",
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
}

var ppMerged = map[string]PostProcessorC{
	"vagrant": {
		TemplateSection{
			Settings: []string{
				"compression_level=8",
				"keep_input_artifact=true",
				"output = out/feedlot-packer.box",
			},
			Arrays: map[string]interface{}{
				"except": []string{
					"googlecompute",
				},
				"include": []string{
					"include1",
					"include2",
				},
				"only": []string{
					"digitalocean",
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
		TemplateSection{
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
	tpl, ok := testDistroDefaults.Templates[CentOS]
	if !ok {
		t.Error("expected \"CentOS\" template to exist. It didn't")
		return
	}
	err := tpl.updatePostProcessors(nil)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
		return
	}
	if !reflect.DeepEqual(testDistroDefaults.Templates[CentOS].PostProcessors, ppOrig) {
		t.Errorf("Expected %#v, got %#v", ppOrig, testDistroDefaults.Templates[CentOS].PostProcessors)
	}

	tpl, ok = testDistroDefaults.Templates[CentOS]
	if !ok {
		t.Error("expected \"CentOS\" template to exist. It didn't")
		return
	}
	err = tpl.updatePostProcessors(ppNew)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
	if !reflect.DeepEqual(testDistroDefaults.Templates[CentOS].PostProcessors, ppMerged) {
		t.Errorf("Expected %#v, got %#v", ppMerged, testDistroDefaults.Templates[CentOS].PostProcessors)
	}
}

func TestPostProcessorsSettingsToMap(t *testing.T) {
	expected := map[string]interface{}{"type": "vagrant", "compression_level": "8", "keep_input_artifact": true}
	res := pp.settingsToMap("vagrant", testRawTpl)
	if !reflect.DeepEqual(res, expected) {
		t.Errorf("expected %#v, got %#v", expected, res)
	}
}

func TestAtlasPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"artifact":      "hashicorp/foobar",
		"artifact_type": "aws.ami",
		"metadata": map[string]string{
			"created_at": "{{timestamp}}",
		},
		"token": "{{user `atlas_token`}}",
		"type":  "atlas",
	}
	pp, err := testPostProcessorsAllTemplate.createAtlas("atlas")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestCompressPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"output": "foo.tar.gz",
		"type":   "compress",
	}

	pp, err := testPostProcessorsAllTemplate.createCompress("compress")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestDockerPushPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"login":          false,
		"login_email":    "email@test.com",
		"login_username": "username",
		"login_password": "password",
		"login_server":   "server.test.com",
		"type":           "docker-push",
	}

	pp, err := testPostProcessorsAllTemplate.createDockerPush("docker-push")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestDockerSavePostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"path": "save/path",
		"type": "docker-save",
	}

	pp, err := testPostProcessorsAllTemplate.createDockerSave("docker-save")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestDockerTagPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"repository": "mitchellh/packer",
		"tag":        "0.7",
		"type":       "docker-tag",
	}

	pp, err := testPostProcessorsAllTemplate.createDockerTag("docker-tag")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestVagrantPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"compression_level": 6,
		"except": []string{
			"docker",
		},
		"include": []string{
			"include/include1",
			"include/include2",
		},
		"keep_input_artifact":  false,
		"output":               ":out_dir/packer.box",
		"type":                 "vagrant",
		"vagrantfile_template": "template/VagrantFile.template",
	}

	pp, err := testPostProcessorsAllTemplate.createVagrant("vagrant")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestVagrantCloudPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"access_token":        "vagrant-cloud-token",
		"box_download_url":    "download.example.com/box",
		"box_tag":             "hashicorp/precise64",
		"no_release":          "true",
		"type":                "vagrant-cloud",
		"vagrant_cloud_url":   "https://vagrantcloud.com/api/v1",
		"version":             "0.0.1",
		"version_description": "initial",
	}

	pp, err := testPostProcessorsAllTemplate.createVagrantCloud("vagrant-cloud")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestVagrantVSphereProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"cluster":       "target-cluster",
		"datacenter":    "target-datacenter",
		"datastore":     "vm-datastore",
		"disk_mode":     "thick",
		"host":          "vsphere-host",
		"insecure":      false,
		"password":      "password",
		"resource_pool": "rpool",
		"type":          "vsphere",
		"username":      "username",
		"vm_folder":     "vm-folder",
		"vm_name":       "packervm",
		"vm_network":    "vm-network",
	}

	pp, err := testPostProcessorsAllTemplate.createVSphere("vsphere")
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if !reflect.DeepEqual(expected, pp) {
			t.Errorf("Expected %#v, got %#v", expected, pp)
		}
	}
}

func TestDeepCopyMapStringPostProcessor(t *testing.T) {
	cpy := DeepCopyMapStringPostProcessorC(ppOrig)
	pp := map[string]PostProcessorC{}
	for k, v := range cpy {
		pp[k] = v.(PostProcessorC)
	}
	msg, ok := EvalPostProcessors(pp, ppOrig)
	if !ok {
		t.Error(msg)
	}
}
