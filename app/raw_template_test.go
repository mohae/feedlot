package app

import (
	"strings"
	"testing"
	"time"
)

var testRawTpl = newRawTemplate()

var updatedBuilders = map[string]builder{
	"common": {
		templateSection{
			Type: "common",
			Settings: []string{
				"ssh_wait_timeout = 300m",
			},
		},
	},
	"virtualbox-iso": {
		templateSection{
			Type:     "virtualbox-iso",
			Settings: []string{},
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"memory=4096",
				},
			},
		},
	},
}

var comparePostProcessors = map[string]postProcessor{
	"vagrant": {
		templateSection{
			Type: "vagrant",
			Settings: []string{
				"output = :out_dir/packer.box",
			},
			Arrays: map[string]interface{}{
				"except": []string{
					"docker",
				},
				"only": []string{
					"virtualbox-iso",
				},
			},
		},
	},
	"vagrant-cloud": {
		templateSection{
			Type: "vagrant-cloud",
			Settings: []string{
				"access_token = getAValidTokenFrom-VagrantCloud.com",
				"box_tag = foo/bar/baz",
				"no_release = false",
				"version = 1.0.2",
			},
			Arrays: map[string]interface{}{},
		},
	},
}

var compareProvisioners = map[string]provisioner{
	"shell": {
		templateSection{
			Type: "shell",
			Settings: []string{
				"execute_command = execute_test.command",
			},
			Arrays: map[string]interface{}{
				"scripts": []string{
					"setup_test.sh",
					"vagrant_test.sh",
					"cleanup_test.sh",
				},
				"except": []string{
					"docker",
				},
				"only": []string{
					"virtualbox-iso",
				},
			},
		},
	},
}

var testBuildNewTPL = &rawTemplate{
	PackerInf: PackerInf{
		Description: "Test build new template",
	},
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "12.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{
			"virtualbox-iso",
		},
		Builders: map[string]builder{
			"common": {
				templateSection{
					Type: "common",
					Settings: []string{
						"ssh_wait_timeout = 300m",
					},
				},
			},
			"virtualbox-iso": {
				templateSection{
					Type: "virtualbox-iso",
					Arrays: map[string]interface{}{
						"vm_settings": []string{
							"memory=4096",
						},
					},
				},
			},
		},
		PostProcessorIDs: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
					Type: "vagrant",
					Settings: []string{
						"output = :out_dir/packer.box",
					},
					Arrays: map[string]interface{}{
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
			"vagrant-cloud": {
				templateSection{
					Type: "vagrant-cloud",
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar/baz",
						"no_release = false",
						"version = 1.0.2",
					},
				},
			},
		},
		ProvisionerIDs: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
			"shell": {
				templateSection{
					Type: "shell",
					Settings: []string{
						"execute_command = execute_test.command",
					},
					Arrays: map[string]interface{}{
						"scripts": []string{
							"setup_test.sh",
							"vagrant_test.sh",
							"cleanup_test.sh",
						},
						"except": []string{
							"docker",
						},
						"only": []string{
							"virtualbox-iso",
						},
					},
				},
			},
		},
	},
}

var expecteNewTemplateBuildInf = BuildInf{
	Name:      "",
	BuildName: "",
	BaseURL:   "",
	Region:    &region,
	Country:   &country,
}

var testRawTemplateBuilderOnly = &rawTemplate{
	PackerInf: PackerInf{MinPackerVersion: "0.4.0", Description: "Test supported distribution template"},
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/out/:distro/:build_name",
		PackerOutputDir:   "packer_boxes/:distro/:build_name",
		SourceDir:         "../test_files/src/:distro",
	},
	BuildInf: BuildInf{
		Name:      ":build_name",
		BuildName: "",
		BaseURL:   "http://releases.ubuntu.org/",
	},
	date:    today,
	delim:   ":",
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "12.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build:   build{},
}

var testRawTemplateWOSection = &rawTemplate{
	PackerInf: PackerInf{MinPackerVersion: "0.4.0", Description: "Test supported distribution template"},
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/out/:distro/:build_name",
		PackerOutputDir:   "packer_boxes/:distro/:build_name",
		SourceDir:         "../test_files/src/:distro",
	},
	BuildInf: BuildInf{
		Name:      ":build_name",
		BuildName: "",
		BaseURL:   "http://releases.ubuntu.org/",
	},
	date:    today,
	delim:   ":",
	Distro:  "ubuntu",
	Arch:    "amd64",
	Image:   "server",
	Release: "12.04",
	varVals: map[string]string{},
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs:       []string{"amazon-ebs"},
		Builders:         map[string]builder{},
		PostProcessorIDs: []string{"compress"},
		PostProcessors:   map[string]postProcessor{},
		ProvisionerIDs:   []string{"ansible-local"},
		Provisioners:     map[string]provisioner{},
	},
}

func TestRequiredSettingError(t *testing.T) {
	err := &RequiredSettingError{AnsibleLocal.String(), "test_setting"}
	if err.Error() != "ansible-local.test_setting: required setting" {
		t.Errorf("Expected \"ansible-local.test_setting: required setting\", got %q", err)
	}
}

func TestNewRawTemplate(t *testing.T) {
	rawTpl := newRawTemplate()
	if MarshalJSONToString.Get(rawTpl) != MarshalJSONToString.Get(testRawTpl) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testRawTpl), MarshalJSONToString.Get(rawTpl))
	}
}

func TestReplaceVariables(t *testing.T) {
	r := newRawTemplate()
	r.varVals = map[string]string{
		":arch":            "amd64",
		":command_src_dir": "commands",
		":image":           "server",
		":name":            ":distro-:release:-:image-:arch",
		":out_dir":         "../test_files/out/:distro",
		":release":         "14.04",
		":src_dir":         "../test_files/src/:distro",
		":distro":          "ubuntu",
	}
	r.delim = ":"
	s := r.replaceVariables("../test_files/src/:distro")
	if s != "../test_files/src/ubuntu" {
		t.Errorf("Expected \"../test_files/src/ubuntu\", got %q", s)
	}
	s = r.replaceVariables("../test_files/src/:distro/command")
	if s != "../test_files/src/ubuntu/command" {
		t.Errorf("Expected \"../test_files/src/ubuntu/command\", got %q", s)
	}
	s = r.replaceVariables("http")
	if s != "http" {
		t.Errorf("Expected \"http\", got %q", s)
	}
	s = r.replaceVariables("../test_files/out/:distro")
	if s != "../test_files/out/ubuntu" {
		t.Errorf("Expected \"../test_files/out/ubuntu\", got %q", s)
	}
}

func TestSetDefaults(t *testing.T) {
	r := newRawTemplate()
	r.setDefaults(testSupportedCentOS)
	if r.Arch == "" {
		t.Error("expected Arch to not be empty. it was")
	}
	if r.Image == "" {
		t.Error("expected Image to not be empty, it was")
	}
	if r.Release == "" {
		t.Error("expected Release to not be empty, it was")
	}
	if MarshalJSONToString.Get(r.IODirInf) != MarshalJSONToString.Get(testSupportedCentOS.IODirInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.IODirInf), MarshalJSONToString.Get(r.IODirInf))
	}
	if MarshalJSONToString.Get(r.PackerInf) != MarshalJSONToString.Get(testSupportedCentOS.PackerInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.PackerInf), MarshalJSONToString.Get(r.PackerInf))
	}
	if MarshalJSONToString.Get(r.BuildInf) != MarshalJSONToString.Get(testSupportedCentOS.BuildInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.BuildInf), MarshalJSONToString.Get(r.BuildInf))
	}
	if MarshalJSONToString.Get(r.BuilderIDs) != MarshalJSONToString.Get(testSupportedCentOS.BuilderIDs) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.BuilderIDs), MarshalJSONToString.Get(r.BuilderIDs))
	}
	if MarshalJSONToString.Get(r.PostProcessorIDs) != MarshalJSONToString.Get(testSupportedCentOS.PostProcessorIDs) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.PostProcessorIDs), MarshalJSONToString.Get(r.PostProcessorIDs))
	}
	if MarshalJSONToString.Get(r.ProvisionerIDs) != MarshalJSONToString.Get(testSupportedCentOS.ProvisionerIDs) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.ProvisionerIDs), MarshalJSONToString.Get(r.ProvisionerIDs))
	}
	if r.Builders != nil {
		t.Errorf("Expected builders to be nil, got %q", MarshalJSONToString.Get(r.Builders))
	}
	if r.PostProcessors != nil {
		t.Errorf("Expected postprocessors to be nil, got %q", MarshalJSONToString.Get(r.PostProcessors))
	}
	if r.Provisioners != nil {
		t.Errorf("Expected provisioners to be nil, got %q", MarshalJSONToString.Get(r.Provisioners))
	}
}

func TestRawTemplateUpdateBuildSettings(t *testing.T) {
	r := newRawTemplate()
	r.setDefaults(testSupportedCentOS)
	err := r.updateBuildSettings(testBuildNewTPL)
	if err != nil {
		t.Errorf("got %q want nil", err)
		return
	}
	if r.Arch != testBuildNewTPL.Arch {
		t.Errorf("expected Arch to be %q, got %q", testBuildNewTPL.Arch, r.Arch)
	}
	if r.Image != testBuildNewTPL.Image {
		t.Errorf("expected Image to be %q, got %q", testBuildNewTPL.Image, r.Image)
	}
	if r.Release != testBuildNewTPL.Release {
		t.Errorf("expected Release to be %q, got %q", testBuildNewTPL.Release, r.Release)
	}
	if MarshalJSONToString.Get(r.IODirInf) != MarshalJSONToString.Get(testSupportedCentOS.IODirInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.IODirInf), MarshalJSONToString.Get(r.IODirInf))
	}
	if MarshalJSONToString.Get(r.PackerInf) != MarshalJSONToString.Get(testBuildNewTPL.PackerInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.PackerInf), MarshalJSONToString.Get(r.PackerInf))
	}
	if MarshalJSONToString.Get(r.BuildInf) != MarshalJSONToString.Get(testSupportedCentOS.BuildInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.BuildInf), MarshalJSONToString.Get(r.BuildInf))
	}
	if MarshalJSONToString.Get(r.BuilderIDs) != MarshalJSONToString.Get(testBuildNewTPL.BuilderIDs) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.BuilderIDs), MarshalJSONToString.Get(r.BuilderIDs))
	}
	if MarshalJSONToString.Get(r.PostProcessorIDs) != MarshalJSONToString.Get(testBuildNewTPL.PostProcessorIDs) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.PostProcessorIDs), MarshalJSONToString.Get(r.PostProcessorIDs))
	}
	if MarshalJSONToString.Get(r.ProvisionerIDs) != MarshalJSONToString.Get(testBuildNewTPL.ProvisionerIDs) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.ProvisionerIDs), MarshalJSONToString.Get(r.ProvisionerIDs))
	}
	if MarshalJSONToString.Get(r.Builders) != MarshalJSONToString.Get(updatedBuilders) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(updatedBuilders), MarshalJSONToString.Get(r.Builders))
	}
	if MarshalJSONToString.Get(r.PostProcessors) != MarshalJSONToString.Get(comparePostProcessors) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(comparePostProcessors), MarshalJSONToString.Get(r.PostProcessors))
	}
	if MarshalJSONToString.Get(r.Provisioners) != MarshalJSONToString.Get(compareProvisioners) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(compareProvisioners), MarshalJSONToString.Get(r.Provisioners))
	}
}

func TestMergeVariables(t *testing.T) {
	r := testDistroDefaults.Templates[Ubuntu]
	r.mergeVariables()
	if r.TemplateOutputDir != "../test_files/out/ubuntu/" {
		t.Errorf("Expected \"../test_files/out/ubuntu/\", got %q", r.TemplateOutputDir)
	}
	if r.PackerOutputDir != "packer_boxes/ubuntu/" {
		t.Errorf("Expected \"packer_boxes/ubuntu/\", got %q", r.PackerOutputDir)
	}
	if r.SourceDir != "../test_files/src/ubuntu" {
		t.Errorf("Expected \"../test_files/src/ubuntu/\", got %q", r.SourceDir)
	}
}

func TestPackerInf(t *testing.T) {
	oldPackerInf := PackerInf{MinPackerVersion: "0.40", Description: "test info"}
	newPackerInf := PackerInf{}
	oldPackerInf.update(newPackerInf)
	if oldPackerInf.MinPackerVersion != "0.40" {
		t.Errorf("Expected \"0.40\", got %q", oldPackerInf.MinPackerVersion)
	}
	if oldPackerInf.Description != "test info" {
		t.Errorf("Expected \"test info\", got %q", oldPackerInf.Description)
	}

	oldPackerInf = PackerInf{MinPackerVersion: "0.40", Description: "test info"}
	newPackerInf = PackerInf{MinPackerVersion: "0.50"}
	oldPackerInf.update(newPackerInf)
	if oldPackerInf.MinPackerVersion != "0.50" {
		t.Errorf("Expected \"0.50\", got %q", oldPackerInf.MinPackerVersion)
	}
	if oldPackerInf.Description != "test info" {
		t.Errorf("Expected \"test info\", got %q", oldPackerInf.Description)
	}

	oldPackerInf = PackerInf{MinPackerVersion: "0.40", Description: "test info"}
	newPackerInf = PackerInf{Description: "new test info"}
	oldPackerInf.update(newPackerInf)
	if oldPackerInf.MinPackerVersion != "0.40" {
		t.Errorf("Expected \"0.40\", got %q", oldPackerInf.MinPackerVersion)
	}
	if oldPackerInf.Description != "new test info" {
		t.Errorf("Expected \"new test info\", got %q", oldPackerInf.Description)
	}

	oldPackerInf = PackerInf{MinPackerVersion: "0.40", Description: "test info"}
	newPackerInf = PackerInf{MinPackerVersion: "0.5.1", Description: "updated"}
	oldPackerInf.update(newPackerInf)
	if oldPackerInf.MinPackerVersion != "0.5.1" {
		t.Errorf("Expected \"0.5.1\", got %q", oldPackerInf.MinPackerVersion)
	}
	if oldPackerInf.Description != "updated" {
		t.Errorf("Expected \"updated\", got %q", oldPackerInf.Description)
	}
}

func TestBuildInf(t *testing.T) {
	oldBuildInf := BuildInf{Name: "old Name", BuildName: "old BuildName"}
	newBuildInf := BuildInf{}
	oldBuildInf.update(newBuildInf)
	if oldBuildInf.Name != "old Name" {
		t.Errorf("Expected \"old Name\", got %q", oldBuildInf.Name)
	}
	if oldBuildInf.BuildName != "old BuildName" {
		t.Errorf("Expected \"old BuildName\", got %q", oldBuildInf.BuildName)
		t.Errorf("Expected \"old BuildName\", got %q", oldBuildInf.BuildName)
	}

	newBuildInf.Name = "new Name"
	oldBuildInf.update(newBuildInf)
	if oldBuildInf.Name != "new Name" {
		t.Errorf("Expected \"new Name\", got %q", oldBuildInf.Name)
	}
	if oldBuildInf.BuildName != "old BuildName" {
		t.Errorf("Expected \"old BuildName\", got %q", oldBuildInf.BuildName)
	}

	newBuildInf.BuildName = "new BuildName"
	oldBuildInf.update(newBuildInf)
	if oldBuildInf.Name != "new Name" {
		t.Errorf("Expected \"new Name\", got %q", oldBuildInf.Name)
	}
	if oldBuildInf.BuildName != "new BuildName" {
		t.Errorf("Expected \"new BuildName\", got %q", oldBuildInf.BuildName)
	}
}

func TestRawTemplateSetBaseVarVals(t *testing.T) {
	now := time.Now()
	splitDate := strings.Split(now.String(), " ")
	tests := []struct {
		Distro    string
		Release   string
		Arch      string
		Image     string
		BuildName string
	}{
		{"ubuntu", "14.04", "amd64", "server", "14.04-test"},
		{"centos", "7", "x86_64", "minimal", "7-test"},
	}

	r := newRawTemplate()
	r.delim = ":"
	for i, test := range tests {
		r.Distro = test.Distro
		r.Release = test.Release
		r.Arch = test.Arch
		r.Image = test.Image
		r.BuildName = test.BuildName
		// make the map empty
		r.varVals = map[string]string{}
		r.setBaseVarVals()
		tmp, ok := r.varVals[":distro"]
		if !ok {
			t.Errorf("%d: expected :distro to be in map, it wasn't", i)
		} else {
			if tmp != test.Distro {
				t.Errorf("%d: expected :distro to be %q, got %q", i, test.Distro, tmp)
			}
		}
		tmp, ok = r.varVals[":release"]
		if !ok {
			t.Errorf("%d: expected :release to be in map, it wasn't", i)
		} else {
			if tmp != test.Release {
				t.Errorf("%d: expected :release to be %q, got %q", i, test.Release, tmp)
			}
		}
		tmp, ok = r.varVals[":arch"]
		if !ok {
			t.Errorf("%d: expected :arch to be in map, it wasn't", i)
		} else {
			if tmp != test.Arch {
				t.Errorf("%d: expected :arch to be %q, got %q", i, test.Arch, tmp)
			}
		}
		tmp, ok = r.varVals[":image"]
		if !ok {
			t.Errorf("%d: expected :image to be in map, it wasn't", i)
		} else {
			if tmp != test.Image {
				t.Errorf("%d: expected :image to be %q, got %q", i, test.Image, tmp)
			}
		}
		tmp, ok = r.varVals[":date"]
		if !ok {
			t.Errorf("%d: expected :date to be in map, it wasn't", i)
		} else {
			if tmp != splitDate[0] {
				t.Errorf("%d: expected :date to be %q, got %q", i, splitDate[0], tmp)
			}
		}
		tmp, ok = r.varVals[":build_name"]
		if !ok {
			t.Errorf("%d: expected :build_name to be in map, it wasn't", i)
		} else {
			if tmp != test.BuildName {
				t.Errorf("%d: expected :build_name to be %q, got %q", i, test.BuildName, tmp)
			}
		}
	}
}

func TestRawTemplateMergeString(t *testing.T) {
	tests := []struct {
		value    string
		dflt     string
		expected string
	}{
		{"", "", ""},
		{"", "src", "src"},
		{"dir", "src", "dir"},
		{"dir/", "src", "dir/"},
		{"dir", "", "dir"},
		{"dir/", "", "dir/"},
	}
	r := newRawTemplate()
	for i, test := range tests {
		v := r.mergeString(test.value, test.dflt)
		if v != test.expected {
			t.Errorf("mergeString %d: expected %q, got %q", i, test.expected, v)
		}
	}
}

func TestFindSource(t *testing.T) {
	tests := []struct {
		p           string
		component   string
		isDir       bool
		src         string
		expectedErr string
	}{
		{"", "", false, "", "cannot find source: no path received"},
		{"something", "", false, "", "find something: file does not exist"},
		{"http/preseed.cfg", "", false, "../test_files/src/ubuntu/http/preseed.cfg", ""},
		{"cookbook1", "chef-solo", true, "../test_files/src/chef-solo/cookbook1/", ""},
		{"14.04_ubuntu_build.txt", "", false, "../test_files/src/ubuntu/14.04/ubuntu_build/14.04_ubuntu_build.txt", ""},
		{"1404_ubuntu_build.txt", "", false, "../test_files/src/ubuntu/1404/ubuntu_build/1404_ubuntu_build.txt", ""},
		{"14_ubuntu_build.txt", "", false, "../test_files/src/ubuntu/14/ubuntu_build/14_ubuntu_build.txt", ""},
		{"ubuntu_build_text.txt", "", false, "../test_files/src/ubuntu/ubuntu_build/ubuntu_build_text.txt", ""},
		{"ubuntu_build.txt", "", false, "../test_files/src/ubuntu_build/ubuntu_build.txt", ""},
		{"14.04_amd64_build_text.txt", "", false, "../test_files/src/ubuntu/14.04/amd64/14.04_amd64_build_text.txt", ""},
		{"1404_amd64_build_text.txt", "", false, "../test_files/src/ubuntu/1404/amd64/1404_amd64_build_text.txt", ""},
		{"14_amd64_build_text.txt", "", false, "../test_files/src/ubuntu/14/amd64/14_amd64_build_text.txt", ""},
		{"14.04_text.txt", "", false, "../test_files/src/ubuntu/14.04/14.04_text.txt", ""},
		{"1404_text.txt", "", false, "../test_files/src/ubuntu/1404/1404_text.txt", ""},
		{"14_text.txt", "", false, "../test_files/src/ubuntu/14/14_text.txt", ""},
		{"amd64_text.txt", "", false, "../test_files/src/ubuntu/amd64/amd64_text.txt", ""},
		{"ubuntu_text.txt", "", false, "../test_files/src/ubuntu/ubuntu_text.txt", ""},
		{"chef.cfg", "", false, "", "find chef.cfg: file does not exist"},
		{"minion", "salt", false, "", "find minion: file does not exist"},
		{"master", "salt-masterless", false, "", "find master: file does not exist"},
		{"chef.cfg", "chef-solo", false, "../test_files/src/chef-solo/chef.cfg", ""},
		{"chef.cfg", "chef-client", false, "../test_files/src/chef-client/chef.cfg", ""},
		{"chef.cfg", "chef", false, "../test_files/src/chef/chef.cfg", ""},
		{"commands", "shell", true, "../test_files/src/ubuntu/14/commands/", ""},
		{"ubuntu_build.txt", "", false, "../test_files/src/ubuntu_build/ubuntu_build.txt", ""}}
	r := newRawTemplate()
	r.Distro = "ubuntu"
	r.Arch = "amd64"
	r.Release = "14.04"
	r.Image = "server"
	r.SourceDir = "../test_files/src"
	r.BuildName = "ubuntu_build"
	for i, test := range tests {
		src, err := r.findSource(test.p, test.component, test.isDir)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestFindSource %d: expected %q got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("TestFindSource %d: expected %q, got no error", i, test.expectedErr)
			continue
		}
		if test.src != src {
			t.Errorf("TestFindSource %d: expected %q, got %q", i, test.src, src)
		}
	}
}

func TestFindCommandFile(t *testing.T) {
	tests := []struct {
		component   string
		p           string
		src         string
		expectedErr string
	}{
		{"", "", "", "the passed command filename was empty"},
		{"", "test.command", "", "find commands/test.command: file does not exist"},
		{"", "execute.command", "../test_files/src/commands/execute.command", ""},
		{"shell", "execute_test.command", "../test_files/src/shell/commands/execute_test.command", ""},
		{"chef-solo", "execute.command", "../test_files/src/chef-solo/commands/execute.command", ""},
		{"chef-solo", "chef.command", "../test_files/src/chef/commands/chef.command", ""},
		{"shell", "ubuntu.command", "../test_files/src/ubuntu/commands/ubuntu.command", ""},
		{"shell", "ubuntu-14.command", "../test_files/src/ubuntu/14/commands/ubuntu-14.command", ""},
	}
	r := newRawTemplate()
	r.Distro = "ubuntu"
	r.Arch = "amd64"
	r.Release = "14.04"
	r.Image = "server"
	r.SourceDir = "../test_files/src"
	r.BuildName = "ubuntu_build"
	for i, test := range tests {
		src, err := r.findCommandFile(test.p, test.component)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestFindCommandFile %d: expected %q got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("TestFindCommandFile %d: expected %q, got no error", i, test.expectedErr)
			continue
		}
		if test.src != src {
			t.Errorf("TestFindCommandFile %d: expected %q, got %q", i, test.src, src)
		}
	}
}

func TestCommandsFromFile(t *testing.T) {
	tests := []struct {
		component   string
		p           string
		expected    []string
		expectedErr string
	}{
		{"", "", []string{}, "the passed command filename was empty"},
		{"", "test.command", []string{}, "find commands/test.command: file does not exist"},
		{"shell", "execute.command", []string{"echo 'vagrant'|sudo -S sh '{{.Path}}'"}, ""},
		{"shell", "boot.command", []string{"<esc><wait>", "<esc><wait>", "<enter><wait>"}, ""},
	}
	r := newRawTemplate()
	r.Distro = "ubuntu"
	r.Arch = "amd64"
	r.Release = "14.04"
	r.Image = "server"
	r.SourceDir = "../test_files/src"
	r.BuildName = "ubuntu_build"
	for i, test := range tests {
		commands, err := r.commandsFromFile(test.p, test.component)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestCommandsFromFile %d: expected %q got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("TestCommandsFromFile %d: expected %q, got no error", i, test.expectedErr)
			continue
		}
		if len(commands) != len(test.expected) {
			t.Errorf("TestCommandsFromFile %d: expected commands slice to have a len of %d got %d", i, len(test.expected), len(commands))
			continue
		}
		for i, v := range commands {
			if v != test.expected[i] {
				t.Errorf("TestCommandsFromFile %d: expected commands slice to be %v, got %v", i, test.expected, commands)
				break
			}
		}
	}

}

func TestBuildOutPath(t *testing.T) {
	tests := []struct {
		includeComponent bool
		component        string
		path             string
		expected         string
	}{
		{false, "", "", "out"},
		{true, "", "", "out"},
		{false, "vagrant", "", "out"},
		{true, "vagrant", "", "out/vagrant"},
		{false, "", "file.txt", "out/file.txt"},
		{false, "", "path/to/file.txt", "out/path/to/file.txt"},
		{false, "shell", "file.txt", "out/file.txt"},
		{false, "shell", "path/to/file.txt", "out/path/to/file.txt"},
		{true, "", "file.txt", "out/file.txt"},
		{true, "", "path/to/file.txt", "out/path/to/file.txt"},
		{true, "shell", "file.txt", "out/shell/file.txt"},
		{true, "shell", "path/to/file.txt", "out/shell/path/to/file.txt"},
	}
	r := newRawTemplate()
	r.TemplateOutputDir = "out"
	for i, test := range tests {
		r.IncludeComponentString = &test.includeComponent
		p := r.buildOutPath(test.component, test.path)
		if p != test.expected {
			t.Errorf("TestBuildOutPath %d: expected %q, got %q", i, test.expected, p)
		}
	}
}

func TestBuildTemplateResourcePath(t *testing.T) {
	tests := []struct {
		includeComponent bool
		isDir            bool
		component        string
		path             string
		expected         string
	}{
		{false, false, "", "", ""},
		{true, false, "", "", ""},
		{false, false, "vagrant", "", ""},
		{true, false, "vagrant", "", "vagrant"},
		{false, false, "", "file.txt", "file.txt"},
		{false, false, "", "path/to/file.txt", "path/to/file.txt"},
		{false, false, "shell", "file.txt", "file.txt"},
		{false, false, "shell", "path/to/file.txt", "path/to/file.txt"},
		{true, false, "", "file.txt", "file.txt"},
		{true, false, "", "path/to/file.txt", "path/to/file.txt"},
		{true, false, "shell", "file.txt", "shell/file.txt"},
		{true, false, "shell", "path/to/file.txt", "shell/path/to/file.txt"},
		{false, true, "", "source/", "source/"},
		{true, true, "", "source", "source/"},
		{false, true, "file", "source/", "source/"},
		{true, true, "file", "source/", "file/source/"},
	}
	r := newRawTemplate()
	r.TemplateOutputDir = "out"
	for i, test := range tests {
		r.IncludeComponentString = &test.includeComponent
		p := r.buildTemplateResourcePath(test.component, test.path, test.isDir)
		if p != test.expected {
			t.Errorf("TestBuildTemplateResourcePath %d: expected %q, got %q", i, test.expected, p)
		}
	}

}
