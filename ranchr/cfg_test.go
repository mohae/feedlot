package ranchr

import (
	"os"
	"reflect"
	"testing"
)

var testDefaults = &defaults{
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:distro/:build_name",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:distro",
	},
	PackerInf: PackerInf{
		Description:      "Test Default Rancher template",
		MinPackerVersion: "0.4.0",
	},
	BuildInf: BuildInf{
		BaseURL:   "",
		BuildName: "",
		Name:      ":build_name",
	},
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = :commands_src_dir/boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :commands_src_dir/shutdown_test.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Settings: []string{
						"virtualbox_version_file = .vbox_version",
					},
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpus=1",
							"memory=1024",
						},
					},
				},
			},
			"vmware-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"cpuid.coresPerSocket=1",
							"memsize=1024",
							"numvcpus=1",
						},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]*postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = out/rancher-packer.box",
					},
					Arrays: map[string]interface{}{
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
				},
			},
		},
		ProvisionerTypes: []string{
			"shell-scripts",
		},
		Provisioners: map[string]*provisioner{
			"shell-scripts": {
				templateSection{
					Settings: []string{
						"execute_command = :commands_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
		},
	},
	loaded: true,
}

var testSupportedUbuntu = &distro{
	BuildInf: BuildInf{
		BaseURL: "http://releases.ubuntu.com/",
	},
	IODirInf: IODirInf{},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test supported distribution template",
	},
	Arch: []string{
		"i386",
		"amd64",
	},
	Image: []string{
		"server",
	},
	Release: []string{
		"10.04",
		"12.04",
		"12.10",
		"13.04",
		"13.10",
	},
	DefImage: []string{
		"release = 12.04",
		"image = server",
		"arch = amd64",
	},
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_command = ../test_files/src/ubuntu/commands/boot_test.command",
						"shutdown_command = :command_src_dir/shutdown_test.command",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{"memory=2048"},
					},
				},
			},
			"vmware-iso": {
				templateSection{
					Arrays: map[string]interface{}{
						"vm_settings": []string{"memsize=2048"},
					},
				},
			},
		},
		PostProcessorTypes: []string{
			"vagrant",
		},
		PostProcessors: map[string]*postProcessor{
			"vagrant": {
				templateSection{
					Settings: []string{
						"output = out/:build_name-packer.box",
					},
				},
			},
		},
		ProvisionerTypes: []string{
			"shell-scripts",
			"file-uploads",
		},
		Provisioners: map[string]*provisioner{
			"shell-scripts": {
				templateSection{
					Settings: []string{
						"execute_command = :command_src_dir/execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/base_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/sudoers_test.sh",
							":scripts_dir/cleanup_test.sh",
						},
					},
				},
			},
			"file-uploads": {
				templateSection{
					Settings: []string{
						"source = source/dir",
						"destination = destination/dir",
					},
				},
			},
		},
	},
}

var testSupportedCentOS = &distro{
	BuildInf: BuildInf{BaseURL: ""},
	IODirInf: IODirInf{},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test template config and Rancher options for CentOS",
	},
	Arch: []string{
		"i386",
		"x86_64",
	},
	Image: []string{
		"minimal",
		"netinstall",
	},
	Release: []string{
		"5",
		"6",
	},
	DefImage: []string{
		"release = 6",
		"image = minimal",
		"arch = x86_64",
	},
}

func TestTemplateSectionMergeArrays(t *testing.T) {
	ts := &templateSection{}
	merged := ts.mergeArrays(nil, nil)
	if merged != nil {
		t.Errorf("Expected the merged array to be nil, was not nil: %#v", merged)
	}

	old := map[string]interface{}{
		"type":            "shell-scripts",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"override": map[string]interface{}{
			"virtualbox-iso": map[string]interface{}{
				"scripts": []string{
					"scripts/base.sh",
					"scripts/vagrant.sh",
					"scripts/vmware.sh",
					"scripts/cleanup.sh",
				},
			},
		},
	}

	nw := map[string]interface{}{
		"type": "shell-scripts",
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
	}

	newold := map[string]interface{}{
		"type":            "shell-scripts",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
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
	}

	merged = ts.mergeArrays(old, nil)
	if merged == nil {
		t.Errorf("Expected merged to be not nil, was nil")
	} else {
		if MarshalJSONToString.Get(merged) != MarshalJSONToString.Get(old) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(old), MarshalJSONToString.Get(merged))
		}
	}

	merged = ts.mergeArrays(nil, nw)
	if merged == nil {
		t.Errorf("Expected merged to be not nil, was nil")
	} else {
		if MarshalJSONToString.Get(merged) != MarshalJSONToString.Get(nw) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(nw), MarshalJSONToString.Get(merged))
		}
	}

	merged = ts.mergeArrays(old, nw)
	if merged == nil {
		t.Errorf("Expected merged to be not nil, was nil")
	} else {
		if MarshalJSONToString.Get(merged) != MarshalJSONToString.Get(newold) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(newold), MarshalJSONToString.Get(merged))
		}
	}
}

func TestBuilderMergeSettings(t *testing.T) {
	b := builder{}
	key1 := "key1=value1"
	key2 := "key2=value2"
	key3 := "key3=value3"

	b.Settings = []string{key1, key2, key3}
	b.mergeSettings(nil)
	if !stringSliceContains(b.Settings, key1) {
		t.Errorf("expected %s in slice: not found", key1)
	}
	if !stringSliceContains(b.Settings, key2) {
		t.Errorf("expected %s in slice: not found", key2)
	}
	if !stringSliceContains(b.Settings, key3) {
		t.Errorf("expected %s in slice: not found", key3)
	}

	key4 := "key4=value4"
	key2update := "key2=value22"
	newSettings := []string{key4, key2update}
	b.mergeSettings(newSettings)
	if !stringSliceContains(b.Settings, key1) {
		t.Errorf("expected %s in slice: not found", key1)
	}
	if !stringSliceContains(b.Settings, key2update) {
		t.Errorf("expected %s in slice: not found", key2update)
	}
	if !stringSliceContains(b.Settings, key3) {
		t.Errorf("expected %s in slice: not found", key3)
	}
	if !stringSliceContains(b.Settings, key3) {
		t.Errorf("expected %s in slice: not found", key4)
	}
	if stringSliceContains(b.Settings, key2) {
		t.Errorf("did not expect %s in slice: was found", key2)
	}
}

func TestMergeVMSettings(t *testing.T) {
	b := builder{}
	key1 := "VMkey1=VMvalue1"
	key2 := "VMkey2=VMvalue2"
	b.Arrays = map[string]interface{}{}
	b.Arrays[VMSettings] = []string{key1, key2}

	merged := b.mergeVMSettings(nil)
	if merged != nil {
		t.Errorf("Expected nil, got %v", merged)
	}

	key1update := "VMkey1=VMvalue11"
	key3 := "VMkey3=VMvalue3"
	newVMSettings := []string{key1update, key3}
	mergedVMSettings := []string{key1update, key2, key3}
	merged = b.mergeVMSettings(newVMSettings)
	if !reflect.DeepEqual(merged, mergedVMSettings) {
		t.Errorf("Expected %v, got %v", merged, mergedVMSettings)
	}
	if !stringSliceContains(merged, key1update) {
		t.Errorf("expected %s in slice: not found", key1update)
	}
	if !stringSliceContains(merged, key2) {
		t.Errorf("expected %s in slice: not found", key2)
	}
	if !stringSliceContains(merged, key3) {
		t.Errorf("expected %s in slice: not found", key3)
	}
	if stringSliceContains(merged, key1) {
		t.Errorf("did not expect %s in slice: was found", key1)
	}

	b.Settings = []string{key1, key2, key3}
	rawTpl := &rawTemplate{}
	res := b.settingsToMap(rawTpl)
	if len(res) != 3 {
		t.Errorf("Expected map to contain 3 elements, got %d", len(res))
	}

	v, ok := res["VMkey1"]
	if !ok {
		t.Errorf("Expected \"VMkey1\" to be in map, it isn't.")
	} else {
		if "VMvalue1" != v.(string) {
			t.Errorf("Expected \"VMkey1's\" to equal \"VMvalue1\", got %q", v.(string))
		}
	}

	v, ok = res["VMkey2"]
	if !ok {
		t.Errorf("Expected \"VMkey2\" to be in map, it isn't.")
	} else {
		if "VMvalue2" != v.(string) {
			t.Errorf("Expected \"VMkey2's\" to equal \"VMvalue2\", got %q", v.(string))
		}
	}

	v, ok = res["VMkey3"]
	if !ok {
		t.Errorf("Expected \"VMkey3\" to be in map, it isn't.")
	} else {
		if "VMvalue3" != v.(string) {
			t.Errorf("Expected \"VMkey3'\" to equal \"VMvalue3\", got %q", v.(string))
		}
	}
}

func TestPostProcessorMergeSettings(t *testing.T) {
	pp := postProcessor{}
	pp.Settings = []string{"key1=value1", "key2=value2"}
	pp.mergeSettings(nil)
	if !stringSliceContains(pp.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pp.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	newSettings := []string{"key1=value1", "key2=value22", "key3=value3"}
	pp.mergeSettings(newSettings)
	if !stringSliceContains(pp.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pp.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(pp.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
	if stringSliceContains(pp.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	post := postProcessor{}
	post.mergeSettings(newSettings)
	if !stringSliceContains(pp.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pp.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(pp.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
}

func TestProvisionerMergeSettings(t *testing.T) {
	p := provisioner{}
	p.Settings = []string{"key1=value1", "key2=value2"}
	p.mergeSettings(nil)
	if !stringSliceContains(p.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(p.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	newSettings := []string{"key1=value1", "key2=value22", "key3=value3"}
	p.mergeSettings(newSettings)
	if !stringSliceContains(p.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(p.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(p.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
	if stringSliceContains(p.Settings, "key2=value2") {
		t.Errorf("expected %s in slice: not found", "key2=value2")
	}

	pr := provisioner{}
	pr.mergeSettings(newSettings)
	if !stringSliceContains(pr.Settings, "key1=value1") {
		t.Errorf("expected %s in slice: not found", "key1=value1")
	}
	if !stringSliceContains(pr.Settings, "key2=value22") {
		t.Errorf("expected %s in slice: not found", "key2=value22")
	}
	if !stringSliceContains(pr.Settings, "key3=value3") {
		t.Errorf("expected %s in slice: not found", "key3=value3")
	}
}

func TestDefaults(t *testing.T) {
	tmpEnvDefaultsFile := os.Getenv(EnvDefaultsFile)
	d := defaults{}
	os.Setenv(EnvDefaultsFile, "")
	err := d.LoadOnce()
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "unable to retrieve the default settings: \"RANCHER_BUILDS_FILE\" was not set; check your \"rancher.cfg\"" {
			t.Errorf("Expected \"unable to retrieve the default settings: \"RANCHER_BUILDS_FILE\" was not set; check your \"rancher.cfg\"\", got %q", err.Error())
		} else {
			if d.MinPackerVersion != "" {
				t.Errorf("Expected \"\", got %q", d.MinPackerVersion)
			}
		}
	}
	os.Setenv(EnvDefaultsFile, tmpEnvDefaultsFile)
}

func TestSupported(t *testing.T) {
	tmpEnv := os.Getenv(EnvSupportedFile)
	s := supported{}
	os.Setenv(EnvSupportedFile, "")
	err := s.LoadOnce()
	if err == nil {
		t.Errorf("expected error, none occurred")
	} else {
		if err.Error() != "open : no such file or directory" {
			t.Errorf("expected \"open : no such file or directory\" got %q", err.Error())
		}
		if s.loaded {
			t.Errorf("expected Supported info not to be loaded, but it was")
		}
	}
	os.Setenv(EnvSupportedFile, tmpEnv)
}

func TestBuildsStuff(t *testing.T) {
	b := builds{}
	tmpEnv := os.Getenv(EnvBuildsFile)
	os.Setenv(EnvBuildsFile, "")
	b.LoadOnce()
	if b.loaded == true {
		t.Errorf("expected Build's loaded flag to be false, but it was")
	}

	os.Setenv(EnvBuildsFile, "../test_files/notthere.toml")
	b.LoadOnce()
	if b.loaded == true {
		t.Errorf("expected Build's loaded flag to be false, but it was")
	}

	os.Setenv(EnvBuildsFile, tmpEnv)
}

func TestBuildListsStuff(t *testing.T) {
	b := buildLists{}
	tmpEnv := os.Getenv(EnvBuildListsFile)

	os.Setenv(EnvBuildListsFile, "")
	err := b.Load()
	if err == nil {
		t.Error("Expected an error, but none received")
	} else {
		if err.Error() != EnvBuildListsFile+" not set, unable to retrieve the BuildLists file" {
			t.Errorf("Expected \"could not retrieve the BuildLists file because the "+EnvBuildListsFile+" environment variable was not set. Either set it or check your rancher.cfg setting\", got %q", err.Error())
		}
	}

	os.Setenv(EnvBuildListsFile, "../test_files/notthere.toml")
	err = b.Load()
	if err == nil {
		t.Error("Expected an error, but none received")
	} else {
		if err.Error() != "open ../test_files/notthere.toml: no such file or directory" {
			t.Errorf("Expected \"open ../test_files/notthere.toml: no such file or directory\", got %q", err.Error())
		}
	}

	os.Setenv(EnvBuildListsFile, tmpEnv)
}
