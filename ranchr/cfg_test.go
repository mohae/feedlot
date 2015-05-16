package ranchr

import (
	"os"
	"testing"
)

var testDefaults = &defaults{
	IODirInf: IODirInf{
		CommandsSrcDir: "commands",
		OutDir:         "../test_files/out/:distro/:build_name",
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
						"boot_command = boot_test.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = shutdown_test.command",
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
			"shell",
		},
		Provisioners: map[string]*provisioner{
			"shell": {
				templateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"sudoers_test.sh",
							"cleanup_test.sh",
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
						"boot_command = boot_test.command",
						"shutdown_command = shutdown_test.command",
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
			"shell",
			"file-uploads",
		},
		Provisioners: map[string]*provisioner{
			"shell": {
				templateSection{
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"base_test.sh",
							"vagrant_test.sh",
							"sudoers_test.sh",
							"cleanup_test.sh",
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
		"type":            "shell",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"override": map[string]interface{}{
			"virtualbox-iso": map[string]interface{}{
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"vmware.sh",
					"cleanup.sh",
				},
			},
		},
	}

	nw := map[string]interface{}{
		"type": "shell",
		"override": map[string]interface{}{
			"vmware-iso": map[string]interface{}{
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"vmware.sh",
					"cleanup.sh",
				},
			},
		},
	}

	newold := map[string]interface{}{
		"type":            "shell",
		"execute_command": "echo 'vagrant'|sudo -S sh '{{.Path}}'",
		"override": map[string]interface{}{
			"vmware-iso": map[string]interface{}{
				"scripts": []string{
					"base.sh",
					"vagrant.sh",
					"vmware.sh",
					"cleanup.sh",
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
		if err.Error() != "unable to retrieve the default settings: \"RANCHER_BUILDS_FILE\" was not set; check your \"rancher.toml\"" {
			t.Errorf("Expected \"unable to retrieve the default settings: \"RANCHER_BUILDS_FILE\" was not set; check your \"rancher.toml\"\", got %q", err.Error())
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
			t.Errorf("Expected \"could not retrieve the BuildLists file because the "+EnvBuildListsFile+" environment variable was not set. Either set it or check your rancher.toml setting\", got %q", err.Error())
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

func TestIODirInfUpdate(t *testing.T) {
	oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", OutDir: "old OutDir", SrcDir: "old SrcDir"}
	newIODirInf := IODirInf{}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", OutDir: "old OutDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{CommandsSrcDir: "new CommandsSrcDir", OutDir: "new OutDir", SrcDir: "new SrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "new CommandsSrcDir/" {
		t.Errorf("Expected \"new CommandsSrcDir/\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.OutDir != "new OutDir/" {
		t.Errorf("Expected \"new OutDir/\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.SrcDir != "new SrcDir/" {
		t.Errorf("Expected \"new SrcDir/\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", OutDir: "old OutDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{CommandsSrcDir: "CommandsSrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "CommandsSrcDir/" {
		t.Errorf("Expected \"CommandsSrcDir/\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", OutDir: "old OutDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{OutDir: "OutDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.OutDir != "OutDir/" {
		t.Errorf("Expected \"OutDir/\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}
}
