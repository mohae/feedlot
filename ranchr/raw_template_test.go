package ranchr

import (
	"testing"
	_ "time"
)

var updatedBuilders = map[string]*builder{
	"common": {
		templateSection{
			Settings: []string{
				"ssh_wait_timeout = 300m",
			},
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
	"shell": {
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

func init() {
	setCommonTestData()
}

/*
func TestNewRawTemplate(t *testing.T) {
	Convey("Testing NewRawTemplate", t, func() {
		Convey("Given a request for a newRawTemplate()", func() {
			rawTpl := newRawTemplate()
			Convey("The raw template should equal--we don't test the date because it is always changeing", func() {
				if rawTpl, ShouldResemble, testRawTemplate)
			})
		})
	})
}

func TestCreatePackerTemplate(t *testing.T) {
	Convey("Given a template", t, func() {
		r := &rawTemplate{}
		r = testDistroDefaults.Templates[Ubuntu]
		Convey("Calling rawTemplate.CreatePackerTemplate() should result in", func() {
			var pTpl packerTemplate
			var err error
			pTpl, err = r.createPackerTemplate()
			if err, ShouldBeNil)
			if pTpl, ShouldNotResemble, r)
		})
	})
}
*/

/*
func TestCreateBuilders(t *testing.T) {
	Convey("Given a template", t, func() {
		r := &rawTemplate{}
		r = testDistroDefaults.Templates[Ubuntu]
		var bldrs []interface{}
		var vars map[string]interface{}
		var err error
		Convey("Given a call to createBuilders", func() {
			Convey("Given a valid Builder Type", func() {
				//first merge the variables so that create builders will work
				r.mergeVariables()
				bldrs, vars, err = r.createBuilders()
				if err, ShouldBeNil)
				if vars, ShouldBeNil)
				if bldrs, ShouldNotResemble, r)
			})
			Convey("Given an unsupported Builder Types", func() {
				r.BuilderTypes[0] = "unsupported"
				bldrs, vars, err = r.createBuilders()
				if err.Error() != "The requested builder, 'unsupported', is not supported by Rancher")
				if vars, ShouldBeNil)
				if bldrs, ShouldBeNil)
			})
			Convey("Given no Builder Types", func() {
				r.BuilderTypes = nil
				bldrs, vars, err = r.createBuilders()
				if err.Error() != "rawTemplate.createBuilders: no builder types were configured, unable to create builders")
				if vars, ShouldBeNil)
				if bldrs, ShouldBeNil)
			})

		})
	})
}
*/

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

func TestRawTemplateVariableSection(t *testing.T) {
	r := newRawTemplate()
	res, err := r.variableSection()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if len(res) != 0 {
			t.Errorf("Expected an empty map, got %+v", res)
		}
	}
}

/*
func TestRawTemplateSetDefaults(t *testing.T) {
	Convey("Given a raw template", t, func() {
		r := newRawTemplate()
		Convey("setting the distro defaults", func() {
			r.setDefaults(testSupported.Distro["centos"])
			Convey("should result in setting that match the defaults", func() {
				if r.IODirInf, ShouldResemble, testSupported.Distro["centos"].IODirInf)
				if r.PackerInf, ShouldResemble, testSupported.Distro["centos"].PackerInf)
				if r.BuildInf, ShouldResemble, testSupported.Distro["centos"].BuildInf)
				if r.BuilderTypes, ShouldResemble, testSupported.Distro["centos"].BuilderTypes)
				if r.PostProcessorTypes, ShouldResemble, testSupported.Distro["centos"].PostProcessorTypes)
				if r.ProvisionerTypes, ShouldResemble, testSupported.Distro["centos"].ProvisionerTypes)
				if r.Builders, ShouldResemble, testSupported.Distro["centos"].Builders)
				if r.PostProcessors, ShouldResemble, testSupported.Distro["centos"].PostProcessors)
				if r.Provisioners, ShouldResemble, testSupported.Distro["centos"].Provisioners)
			})
		})
	})
}

func TestRawTemplateUpdateBuildSettings(t *testing.T) {
	Convey("Given a raw template with defaults set", t, func() {
		r := newRawTemplate()
		r.setDefaults(testSupported.Distro["centos"])
		Convey("updating the build settings", func() {
			r.updateBuildSettings(testBuilds.Build["test1"])
			Convey("should result in updated build settings", func() {
				if r.IODirInf, ShouldResemble, testSupported.Distro["centos"].IODirInf)
				if r.PackerInf, ShouldResemble, testBuilds.Build["test1"].PackerInf)
				if r.BuildInf, ShouldResemble, testSupported.Distro["centos"].BuildInf)
				if r.BuilderTypes, ShouldResemble, testBuilds.Build["test1"].BuilderTypes)
				if r.PostProcessorTypes, ShouldResemble, testBuilds.Build["test1"].PostProcessorTypes)
				if r.ProvisionerTypes, ShouldResemble, testBuilds.Build["test1"].ProvisionerTypes)
				if MarshalJSONToString.Get(r.Builders), ShouldResemble, MarshalJSONToString.Get(updatedBuilders))
				if MarshalJSONToString.Get(r.PostProcessors), ShouldResemble, MarshalJSONToString.Get(comparePostProcessors))
				if MarshalJSONToString.Get(r.Provisioners), ShouldResemble, MarshalJSONToString.Get(compareProvisioners))
			})
		})
	})
}
*/

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

	err = testDistroDefaultCentOS.ISOInfo(VirtualBoxISO, []string{"iso_checksum_type = sha256", "http_directory=http"})
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
