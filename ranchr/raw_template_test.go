package ranchr

import (
	"strings"
	"testing"
	"time"
)

var testRawTpl = newRawTemplate()

var updatedBuilders = map[string]*builder{
	"common": {
		templateSection{
			Settings: []string{
				"ssh_wait_timeout = 300m",
			},
			Arrays: map[string]interface{}{},
		},
	},
	"virtualbox-iso": {
		templateSection{
			Settings: []string{},
			Arrays: map[string]interface{}{
				"vm_settings": []string{
					"memory=4096",
				},
			},
		},
	},
}

var comparePostProcessors = map[string]*postProcessor{
	"vagrant": {
		templateSection{
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

var compareProvisioners = map[string]*provisioner{
	"shell-scripts": {
		templateSection{
			Settings: []string{
				"execute_command = :commands_src_dir/execute_test.command",
			},
			Arrays: map[string]interface{}{
				"scripts": []string{
					":scripts_dir/setup_test.sh",
					":scripts_dir/vagrant_test.sh",
					":scripts_dir/cleanup_test.sh",
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
	Release: "1204",
	build: build{
		BuilderTypes: []string{
			"virtualbox-iso",
		},
		Builders: map[string]*builder{
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
			"vagrant-cloud",
		},
		PostProcessors: map[string]*postProcessor{
			"vagrant": {
				templateSection{
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
					Settings: []string{
						"access_token = getAValidTokenFrom-VagrantCloud.com",
						"box_tag = foo/bar/baz",
						"no_release = false",
						"version = 1.0.2",
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
						"scripts": []string{
							":scripts_dir/setup_test.sh",
							":scripts_dir/vagrant_test.sh",
							":scripts_dir/cleanup_test.sh",
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

var testRawTemplateBuilderOnly = &rawTemplate{
	PackerInf: PackerInf{MinPackerVersion: "0.4.0", Description: "Test supported distribution template"},
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:distro/:build_name",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:distro",
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
	vars:    map[string]string{},
	build:   build{},
}

var testRawTemplateWOSection = &rawTemplate{
	PackerInf: PackerInf{MinPackerVersion: "0.4.0", Description: "Test supported distribution template"},
	IODirInf: IODirInf{
		CommandsSrcDir: ":src_dir/commands",
		HTTPDir:        "http",
		HTTPSrcDir:     ":src_dir/http",
		OutDir:         "../test_files/out/:distro/:build_name",
		ScriptsDir:     "scripts",
		ScriptsSrcDir:  ":src_dir/scripts",
		SrcDir:         "../test_files/src/:distro",
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
	vars:    map[string]string{},
	build: build{
		BuilderTypes:       []string{"amazon-ebs"},
		Builders:           map[string]*builder{},
		PostProcessorTypes: []string{"compress"},
		PostProcessors:     map[string]*postProcessor{},
		ProvisionerTypes:   []string{"ansible-local"},
		Provisioners:       map[string]*provisioner{},
	},
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
		":command_src_dir": ":src_dir/commands",
		":http_dir":        "http",
		":http_src_dir":    ":src_dir/http",
		":image":           "server",
		":name":            ":distro-:release:-:image-:arch",
		":out_dir":         "../test_files/out/:distro",
		":release":         "14.04",
		":scripts_dir":     "scripts",
		":scripts_src_dir": ":src_dir/scripts",
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

func TestRawTemplateUpdateBuildSettings(t *testing.T) {
	r := newRawTemplate()
	r.setDefaults(testSupportedCentOS)
	r.updateBuildSettings(testBuildNewTPL)
	if MarshalJSONToString.Get(r.IODirInf) != MarshalJSONToString.Get(testSupportedCentOS.IODirInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.IODirInf), MarshalJSONToString.Get(r.IODirInf))
	}
	if MarshalJSONToString.Get(r.PackerInf) != MarshalJSONToString.Get(testBuildNewTPL.PackerInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.PackerInf), MarshalJSONToString.Get(r.PackerInf))
	}
	if MarshalJSONToString.Get(r.BuildInf) != MarshalJSONToString.Get(testSupportedCentOS.BuildInf) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testSupportedCentOS.BuildInf), MarshalJSONToString.Get(r.BuildInf))
	}
	if MarshalJSONToString.Get(r.BuilderTypes) != MarshalJSONToString.Get(testBuildNewTPL.BuilderTypes) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.BuilderTypes), MarshalJSONToString.Get(r.BuilderTypes))
	}
	if MarshalJSONToString.Get(r.PostProcessorTypes) != MarshalJSONToString.Get(testBuildNewTPL.PostProcessorTypes) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.PostProcessorTypes), MarshalJSONToString.Get(r.PostProcessorTypes))
	}
	if MarshalJSONToString.Get(r.ProvisionerTypes) != MarshalJSONToString.Get(testBuildNewTPL.ProvisionerTypes) {
		t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testBuildNewTPL.ProvisionerTypes), MarshalJSONToString.Get(r.ProvisionerTypes))
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

/*
func TestRawTemplateScriptNames(t *testing.T) {
	r := testDistroDefaults.Templates[Ubuntu]
	scripts := r.ScriptNames()
	if scripts == nil {
		t.Error("Expected scripts to not be nil, it was")
	} else {
		if !stringSliceContains(scripts, "setup_test.sh") {
			t.Errorf("Expected slice to contain \"setup_test.sh\", not found")
		}
		if !stringSliceContains(scripts, "vagrant_test.sh") {
			t.Errorf("Expected slice to contain \"vagrant_test.sh\", not found")
		}
		if !stringSliceContains(scripts, "sudoers_test.sh") {
			t.Errorf("Expected slice to contain \"sudoers_test.sh\", not found")
		}
		if !stringSliceContains(scripts, "cleanup_test.sh") {
			t.Errorf("Expected slice to contain \"cleanup_test.sh\", not found")
		}
	}
}
*/
func TestMergeVariables(t *testing.T) {
	r := testDistroDefaults.Templates[Ubuntu]
	r.mergeVariables()
	if r.CommandsSrcDir != "../test_files/src/ubuntu/commands" {
		t.Errorf("Expected \"../test_files/src/ubuntu/commands\", got %q", r.CommandsSrcDir)
	}
	if r.HTTPDir != "http" {
		t.Errorf("Expected \"http\", got %q", r.HTTPDir)
	}
	if r.HTTPSrcDir != "../test_files/src/ubuntu/http" {
		t.Errorf("Expected \"../test_files/src/ubuntu/http\", got %q", r.HTTPSrcDir)
	}
	if r.OutDir != "../test_files/out/ubuntu" {
		t.Errorf("Expected \"../test_files/out/ubuntu\", got %q", r.OutDir)
	}
	if r.ScriptsDir != "scripts" {
		t.Errorf("Expected \"scripts\", got %q", r.ScriptsDir)
	}
	if r.ScriptsSrcDir != "../test_files/src/ubuntu/scripts" {
		t.Errorf("Expected \"../test_files/src/ubuntu/scripts\", got %q", r.ScriptsSrcDir)
	}
	if r.SrcDir != "../test_files/src/ubuntu" {
		t.Errorf("Expected \"../test_files/src/ubuntu\", got %q", r.SrcDir)
	}
}

func TestIODirInf(t *testing.T) {
	oldIODirInf := IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf := IODirInf{}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "old HTTPDir" {
		t.Errorf("Expected \"old HTTPDir\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "old HTTPSrcDir" {
		t.Errorf("Expected \"old HTTPSrcDir\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "old ScriptsDir" {
		t.Errorf("Expected \"old ScriptsDir\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "old ScriptsSrcDir" {
		t.Errorf("Expected \"old ScriptsSrcDir\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{CommandsSrcDir: "new CommandsSrcDir", HTTPDir: "new HTTPDir", HTTPSrcDir: "new HTTPSrcDir", OutDir: "new OutDir", ScriptsDir: "new ScriptsDir", ScriptsSrcDir: "new ScriptsSrcDir", SrcDir: "new SrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "new CommandsSrcDir/" {
		t.Errorf("Expected \"new CommandsSrcDir/\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "new HTTPDir/" {
		t.Errorf("Expected \"new HTTPDir/\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "new HTTPSrcDir/" {
		t.Errorf("Expected \"new HTTPSrcDir/\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "new OutDir/" {
		t.Errorf("Expected \"new OutDir/\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "new ScriptsDir/" {
		t.Errorf("Expected \"new ScriptsDir/\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "new ScriptsSrcDir/" {
		t.Errorf("Expected \"new ScriptsSrcDir/\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "new SrcDir/" {
		t.Errorf("Expected \"new SrcDir/\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{CommandsSrcDir: "CommandsSrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "CommandsSrcDir/" {
		t.Errorf("Expected \"CommandsSrcDir/\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "old HTTPDir" {
		t.Errorf("Expected \"old HTTPDir\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "old HTTPSrcDir" {
		t.Errorf("Expected \"old HTTPSrcDir\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "old ScriptsDir" {
		t.Errorf("Expected \"old ScriptsDir\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "old ScriptsSrcDir" {
		t.Errorf("Expected \"old ScriptsSrcDir\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{HTTPDir: "HTTPDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "HTTPDir/" {
		t.Errorf("Expected \"HTTPDir/\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "old HTTPSrcDir" {
		t.Errorf("Expected \"old HTTPSrcDir\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "old ScriptsDir" {
		t.Errorf("Expected \"old ScriptsDir\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "old ScriptsSrcDir" {
		t.Errorf("Expected \"old ScriptsSrcDir\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{HTTPSrcDir: "HTTPSrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "old HTTPDir" {
		t.Errorf("Expected \"old HTTPDir\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "HTTPSrcDir/" {
		t.Errorf("Expected \"HTTPSrcDir/\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "old ScriptsDir" {
		t.Errorf("Expected \"old ScriptsDir\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "old ScriptsSrcDir" {
		t.Errorf("Expected \"old ScriptsSrcDir\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{OutDir: "OutDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "old HTTPDir" {
		t.Errorf("Expected \"old HTTPDir\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "old HTTPSrcDir" {
		t.Errorf("Expected \"old HTTPSrcDir\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "OutDir/" {
		t.Errorf("Expected \"OutDir/\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "old ScriptsDir" {
		t.Errorf("Expected \"old ScriptsDi\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "old ScriptsSrcDir" {
		t.Errorf("Expected \"old ScriptsSrcDir\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
	}

	oldIODirInf = IODirInf{CommandsSrcDir: "old CommandsSrcDir", HTTPDir: "old HTTPDir", HTTPSrcDir: "old HTTPSrcDir", OutDir: "old OutDir", ScriptsDir: "old ScriptsDir", ScriptsSrcDir: "old ScriptsSrcDir", SrcDir: "old SrcDir"}
	newIODirInf = IODirInf{ScriptsDir: "ScriptsDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.CommandsSrcDir != "old CommandsSrcDir" {
		t.Errorf("Expected \"old CommandsSrcDir\", got %q", oldIODirInf.CommandsSrcDir)
	}
	if oldIODirInf.HTTPDir != "old HTTPDir" {
		t.Errorf("Expected \"old HTTPDir\", got %q", oldIODirInf.HTTPDir)
	}
	if oldIODirInf.HTTPSrcDir != "old HTTPSrcDir" {
		t.Errorf("Expected \"old HTTPSrcDir\", got %q", oldIODirInf.HTTPSrcDir)
	}
	if oldIODirInf.OutDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutDir)
	}
	if oldIODirInf.ScriptsDir != "ScriptsDir/" {
		t.Errorf("Expected \"ScriptsDir/\", got %q", oldIODirInf.ScriptsDir)
	}
	if oldIODirInf.ScriptsSrcDir != "old ScriptsSrcDir" {
		t.Errorf("Expected \"old ScriptsSrcDir\", got %q", oldIODirInf.ScriptsSrcDir)
	}
	if oldIODirInf.SrcDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SrcDir)
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

func TestRawTemplateISOInfo(t *testing.T) {
	err := testDistroDefaultUbuntu.ISOInfo(VirtualBoxISO, []string{"iso_checksum_type = sha256", "http_directory=http"})
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if testDistroDefaultUbuntu.BaseURL != "http://releases.ubuntu.org/" {
			t.Errorf("Expected \"http://releases.ubuntu.org\", got %q", testDistroDefaultUbuntu.BaseURL)
		}
		if testDistroDefaultUbuntu.releaseISO.(*ubuntu).ChecksumType != "sha256" {
			t.Errorf("Expected \"sha256\", got %q", testDistroDefaultUbuntu.releaseISO.(*ubuntu).ChecksumType)
		}
		if testDistroDefaultUbuntu.releaseISO.(*ubuntu).Name != "ubuntu-12.04-server-amd64.iso" {
			t.Errorf("Expected \"ubuntu-12.04-server-amd64.iso\", got %q", testDistroDefaultUbuntu.releaseISO.(*ubuntu).Name)
		}
	}

	err = testDistroDefaultCentOS.ISOInfo(VirtualBoxOVF, []string{"iso_checksum_type = sha256", "http_directory=http"})
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if testDistroDefaultCentOS.BaseURL != "" {
			t.Errorf("Expected \"\", got %q", testDistroDefaultCentOS.BaseURL)
		}
		if testDistroDefaultCentOS.releaseISO.(*centOS).ChecksumType != "sha256" {
			t.Errorf("Expected \"sha256\", got %q", testDistroDefaultCentOS.releaseISO.(*centOS).ChecksumType)
		}
		// TODO, the actual release number may change, split on . and compare parts, stripping the port up to - in the second element
		if testDistroDefaultCentOS.releaseISO.(*centOS).Name != "CentOS-6.6-x86_64-minimal.iso" {
			t.Errorf("Expected \"CentOS-6.6-x86_64-minimal.iso\", got %q", testDistroDefaultCentOS.releaseISO.(*centOS).Name)
		}
	}
}

func TestRawTemplateMergeSrcDir(t *testing.T) {
	tests := []struct {
		CustomSrcDir         bool
		SrcDir               string
		ExpectedCustomSrcDir bool
		ExpectedSrcDir       string
	}{
		{false, "src/", false, "src"},
		{true, "src/custom/", true, "src/custom"},
		{true, "src/:distro/", true, "src/ubuntu"},
		{false, "src/:distro/", true, "src/ubuntu"},
		{false, "src/files/", false, "src/files"},
	}
	rawTpl := newRawTemplate()
	rawTpl.delim = ":"
	rawTpl.Distro = "ubuntu"
	rawTpl.setBaseVarVals()
	for i, test := range tests {
		rawTpl.CustomSrcDir = test.CustomSrcDir
		rawTpl.SrcDir = test.SrcDir
		rawTpl.mergeSrcDir()
		if rawTpl.CustomSrcDir != test.ExpectedCustomSrcDir {
			t.Errorf("MergeSrcDir test %d: expected CustomSrcDir to be %t; got %t", i, test.ExpectedCustomSrcDir, rawTpl.CustomSrcDir)
		}
		if rawTpl.SrcDir != test.ExpectedSrcDir {
			t.Errorf("MergeSrcDir test %d: expected SrcDir to be %s; got %s", i, test.ExpectedSrcDir, rawTpl.SrcDir)
		}
	}
}

func TestRawTemplateMergeOutDir(t *testing.T) {
	tests := []struct {
		CustomOutDir         bool
		OutDir               string
		ExpectedCustomOutDir bool
		ExpectedOutDir       string
	}{
		{false, "out", false, "out"},
		{true, "out/custom/", true, "out/custom"},
		{true, "out/:distro/", true, "out/ubuntu"},
		{false, "out/:distro/", true, "out/ubuntu"},
		{false, "out/files/", false, "out/files"},
	}
	rawTpl := newRawTemplate()
	rawTpl.delim = ":"
	rawTpl.Distro = "ubuntu"
	rawTpl.setBaseVarVals()
	for i, test := range tests {
		rawTpl.CustomOutDir = test.CustomOutDir
		rawTpl.OutDir = test.OutDir
		rawTpl.mergeOutDir()
		if rawTpl.CustomOutDir != test.ExpectedCustomOutDir {
			t.Errorf("MergeOutDir test %d: expected CustomOutDir to be %t; got %t", i, test.ExpectedCustomOutDir, rawTpl.CustomOutDir)
		}
		if rawTpl.OutDir != test.ExpectedOutDir {
			t.Errorf("MergeOutDirtest %d: expected OutDir to be %s; got %s", i, test.ExpectedOutDir, rawTpl.OutDir)
		}
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
				t.Errorf("%d: expected :distro to be %q, got %q", test.Distro, tmp)
			}
		}
		tmp, ok = r.varVals[":release"]
		if !ok {
			t.Errorf("%d: expected :release to be in map, it wasn't", i)
		} else {
			if tmp != test.Release {
				t.Errorf("%d: expected :release to be %q, got %q", test.Release, tmp)
			}
		}
		tmp, ok = r.varVals[":arch"]
		if !ok {
			t.Errorf("%d: expected :arch to be in map, it wasn't", i)
		} else {
			if tmp != test.Arch {
				t.Errorf("%d: expected :arch to be %q, got %q", test.Arch, tmp)
			}
		}
		tmp, ok = r.varVals[":image"]
		if !ok {
			t.Errorf("%d: expected :image to be in map, it wasn't", i)
		} else {
			if tmp != test.Image {
				t.Errorf("%d: expected :image to be %q, got %q", test.Image, tmp)
			}
		}
		tmp, ok = r.varVals[":date"]
		if !ok {
			t.Errorf("%d: expected :date to be in map, it wasn't", i)
		} else {
			if tmp != splitDate[0] {
				t.Errorf("%d: expected :date to be %q, got %q", splitDate[0], tmp)
			}
		}
		tmp, ok = r.varVals[":build_name"]
		if !ok {
			t.Errorf("%d: expected :build_name to be in map, it wasn't", i)
		} else {
			if tmp != test.BuildName {
				t.Errorf("%d: expected :build_name to be %q, got %q", test.BuildName, tmp)
			}
		}
	}
}

/*
func TestRawTemplateMergeString(t *testing.t) {
	tests := []struct{
		value string
		default string
		expected string
	}{
		{"", "", ""},
		{"", "http", "http"},
		{"dir", "http", "dir"},
		{"dir/", "http", "dir"},
		{"dir", "", "dir"},
	}
	r := newRawTemplate()
	for i, test := ragne tests {
		v := r.mergeString(test.value, test.default)
		if v != test.expected {
			t.Errorf("mergeString %d: expected %q, got %q", test.expected, v)
		}
	}
}
*/
