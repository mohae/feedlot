// raw_template_post_processors_test.go: tests for post_processors.
package app

import (
	"fmt"
	"testing"
)

var testPostProcessorTemplate = &rawTemplate{
	PackerInf: PackerInf{
		Description: "Test build template #1",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "1204",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"output = :out_dir/packer.box",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection{
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

var testPostProcessorsAllTemplate = &rawTemplate{
	PackerInf: PackerInf{
		Description: "Test build template #1",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "1204",
	IODirInf: IODirInf{
		OutDir: "../test_files/out/:build_name",
		SrcDir: "../test_files/src",
	},
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"compress",
			"vagrant",
		},
		PostProcessors: map[string]postProcessor{
			"compress": {
				templateSection{
					Settings: []string{
						"output = foo.tar.gz",
					},
				},
			},
			"docker-import": {
				templateSection{
					Settings: []string{
						"repository = mitchellh/packer",
						"tag = 0.7",
					},
				},
			},
			"docker-push": {
				templateSection{
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
				templateSection{
					Settings: []string{
						"path = save/path",
					},
				},
			},
			"docker-tag": {
				templateSection{
					Settings: []string{
						"repository = mitchellh/packer",
						"tag = 0.7",
					},
				},
			},
			"vagrant": {
				templateSection{
					Settings: []string{
						"compression_level = 6",
						"keep_input_artifact = false",
						"output = :out_dir/packer.box",
						"vagrantfile_template = template/path",
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
				templateSection{
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
				templateSection{
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
		ProvisionerTypes: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection{
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

var ppOrig = map[string]postProcessor{
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

var ppNew = map[string]postProcessor{
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

var ppMerged = map[string]postProcessor{
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

func TestCreatePostProcessors(t *testing.T) {
	_, err := testRawTemplateBuilderOnly.createPostProcessors()
	if err == nil {
		t.Error("Expected error \"configuration not found\", got nil")
	} else {
		if err.Error() != "configuration not found" {
			t.Errorf("Expected \"configuration not found\", got %q", err.Error())
		}
	}

	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", Compress.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", Compress.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", Compress.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "docker-import"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", DockerImport.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", DockerImport.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", DockerImport.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "docker-push"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", DockerPush.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", DockerPush.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", DockerPush.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "docker-save"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", DockerSave.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", DockerSave.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", DockerSave.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "docker-tag"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", DockerTag.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", DockerTag.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", DockerTag.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "vagrant"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", Vagrant.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", Vagrant.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", Vagrant.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "vagrant-cloud"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", VagrantCloud.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", VagrantCloud.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", VagrantCloud.String(), err.Error())
		}
	}

	testRawTemplateWOSection.build.PostProcessorTypes[0] = "vsphere"
	_, err = testRawTemplateWOSection.createPostProcessors()
	if err == nil {
		t.Errorf("Expected error \"%s post-processor error: configuration not found\", got nil", VSphere.String())
	} else {
		if err.Error() != fmt.Sprintf("%s post-processor error: configuration not found", VSphere.String()) {
			t.Errorf("Expected error \"%s post-processor error: configuration not found\", got %q", VSphere.String(), err.Error())
		}
	}

}

func TestRawTemplateUpdatePostProcessors(t *testing.T) {
	err := testDistroDefaults.Templates[CentOS].updatePostProcessors(nil)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err.Error())
	}
	if MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors) != MarshalJSONToString.Get(ppOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(ppOrig), MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors))
	}

	err = testDistroDefaults.Templates[CentOS].updatePostProcessors(ppNew)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err.Error())
	}
	if MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors) != MarshalJSONToString.Get(ppMerged) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(ppMerged), MarshalJSONToString.Get(testDistroDefaults.Templates[CentOS].PostProcessors))
	}
}

func TestPostProcessorsSettingsToMap(t *testing.T) {
	res := pp.settingsToMap("vagrant", testRawTpl)
	if MarshalJSONToString.Get(res) != MarshalJSONToString.Get(map[string]interface{}{"type": "vagrant", "compression_level": "8", "keep_input_artifact": true}) {
		t.Errorf("expected %q, got %q", MarshalJSONToString.Get(map[string]interface{}{"type": "vagrant", "compression_level": "8", "keep_input_artifact": true}), MarshalJSONToString.Get(res))
	}
}

func TestCompressPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"output": "foo.tar.gz",
		"type":   "compress",
	}

	pp, err := testPostProcessorsAllTemplate.createCompress()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
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

	pp, err := testPostProcessorsAllTemplate.createDockerPush()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
		}
	}
}

func TestDockerSavePostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"path": "save/path",
		"type": "docker-save",
	}

	pp, err := testPostProcessorsAllTemplate.createDockerSave()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
		}
	}
}

func TestDockerTagPostProcessor(t *testing.T) {
	expected := map[string]interface{}{
		"repository": "mitchellh/packer",
		"tag":        "0.7",
		"type":       "docker-tag",
	}

	pp, err := testPostProcessorsAllTemplate.createDockerTag()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
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
		"vagrantfile_template": "template/path",
	}

	pp, err := testPostProcessorsAllTemplate.createVagrant()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
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

	pp, err := testPostProcessorsAllTemplate.createVagrantCloud()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
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

	pp, err := testPostProcessorsAllTemplate.createVSphere()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if MarshalJSONToString.Get(expected) != MarshalJSONToString.Get(pp) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(expected), MarshalJSONToString.Get(pp))
		}
	}
}

func TestDeepCopyMapStringPostProcessor(t *testing.T) {
	cpy := DeepCopyMapStringPostProcessor(ppOrig)
	if MarshalJSONToString.Get(cpy) != MarshalJSONToString.Get(ppOrig) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(ppOrig), MarshalJSONToString.Get(cpy))
	}
}
