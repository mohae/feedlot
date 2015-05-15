package ranchr

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	json "github.com/mohae/customjson"
)

var MarshalJSONToString = json.NewMarshalString()
var today = time.Now().Local().Format("2006-01-02")

// Simple funcs to help handle testing returned stuff
func stringSliceContains(sl []string, val string) bool {
	for _, v := range sl {
		if v == val {
			return true
		}
	}
	return false
}

var testDistroDefaultUbuntu = &rawTemplate{
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
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]*builder{
			"common": {
				templateSection{
					Settings: []string{
						"boot_wait = 5s",
						"disk_size = 20000",
						"guest_os_type = ",
						"headless = true",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = shutdown",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
					Arrays: map[string]interface{}{},
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
					Settings: []string{},
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
						"include": []string{
							"include1",
							"include2",
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
		},
		ProvisionerTypes: []string{"shell-scripts"},
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
}

var testDistroDefaultCentOS = &rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description:      "Test template config and Rancher options for CentOS",
	},
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
		BaseURL:   "",
	},
	date:    today,
	delim:   ":",
	Distro:  "centos",
	Arch:    "x86_64",
	Image:   "minimal",
	Release: "6",
	varVals: map[string]string{},
	vars:    map[string]string{},
	build: build{
		BuilderTypes: []string{"virtualbox-iso", "vmware-iso"},
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
					Arrays: map[string]interface{}{},
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
					Settings: []string{},
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
}

var testDistroDefaults distroDefaults

func init() {
	testDistroDefaults = distroDefaults{Templates: map[Distro]*rawTemplate{}, IsSet: true}
	testDistroDefaults.Templates[Ubuntu] = testDistroDefaultUbuntu
	testDistroDefaults.Templates[CentOS] = testDistroDefaultCentOS
}

func TestDistroDefaultsGetTemplate(t *testing.T) {
	var err error
	var emptyRawTemplate *rawTemplate
	r := &rawTemplate{}
	r, err = testDistroDefaults.GetTemplate("invalid")
	if err == nil {
		t.Error("expected \"unsupported distro: invalid\", got nil")
	} else {
		if err.Error() != "unsupported distro: invalid" {
			t.Errorf("unsupported distro: invalid, got %q", err.Error())
		}
		if MarshalJSONToString.Get(r) != MarshalJSONToString.Get(emptyRawTemplate) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(emptyRawTemplate), MarshalJSONToString.Get(r))
		}
	}

	r, err = testDistroDefaults.GetTemplate("ubuntu")
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(r) != MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu]) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu]), MarshalJSONToString.Get(r))
		}
	}
}

func TestSetEnv(t *testing.T) {
	err := SetEnv()
	if err == nil {
		t.Error("Expected an error, was nil")
	} else {
		if err.Error() != "open rancher.toml: no such file or directory" {
			t.Errorf("Expected \"open rancher.toml: no such file or directory\", %q", err.Error())
		}
	}
}

func TestbuildPackerTemplateFromDistros(t *testing.T) {
	a := ArgsFilter{}
	err := buildPackerTemplateFromDistro(a)
	if err == nil {
		t.Error("Expected an error, none occurred")
	} else {
		if err.Error() != "Cannot build requested packer template, the supported data structure was empty." {
			t.Errorf("Expected \"Cannot build requested packer template, the supported data structure was empty.\", got %q", err.Error())
		}
	}

	err = buildPackerTemplateFromDistro(a)
	if err == nil {
		t.Error("Expected an error, none occurred")
	} else {
		if err.Error() != "Cannot build a packer template because no target distro information was passed." {
			t.Errorf("Expected \"Cannot build a packer template because no target distro information was passed.\", got %q", err.Error())
		}
	}

	a.Distro = "ubuntu"
	err = buildPackerTemplateFromDistro(a)
	if err == nil {
		t.Error("Expected an error, none occurred")
	} else {
		if err.Error() != "Cannot build a packer template from passed distro: ubuntu is not supported. Please pass a supported distribution." {
			t.Errorf("Expected \"Cannot build a packer template from passed distro: ubuntu is not supported. Please pass a supported distribution.\". got %q", err.Error())
		}
	}

	a.Distro = "slackware"
	err = buildPackerTemplateFromDistro(a)
	if err.Error() != "Cannot build a packer template from passed distro: slackware is not supported. Please pass a supported distribution." {
		t.Errorf("Expected \"Cannot build a packer template from passed distro: slackware is not supported. Please pass a supported distribution.\", got %q", err.Error())
	}
}

func TestbuildPackerTemplateFromNamedBuild(t *testing.T) {
	doneCh := make(chan error)
	go buildPackerTemplateFromNamedBuild("", doneCh)
	err := <-doneCh
	if err == nil {
		t.Error("Expected an error, received none")
	} else {
		if err.Error() != "open look/for/it/here/: no such file or directory" {
			t.Errorf("Expected \"open look/for/it/here/: no such file or directory\", got %q", err.Error())
		}
	}

	os.Setenv(EnvBuildsFile, "../test_files/conf/builds_test.toml")
	go buildPackerTemplateFromNamedBuild("", doneCh)
	err = <-doneCh
	if err == nil {
		t.Error("Expected an error, received none")
	} else {
		if err.Error() != "buildPackerTemplateFromNamedBuild error: no build names were passed. Nothing was built." {
			t.Errorf("Expected \"buildPackerTemplateFromNamedBuild error: no build names were passed. Nothing was built.\", got %q", err.Error())
		}
	}

	close(doneCh)
}

func TestMergeSlices(t *testing.T) {
	// The private implementation only merges two slices at a time.
	var s1, s2, res []string
	res = mergeSlices(s1, s2)
	if res != nil {
		t.Errorf("expected nil, got %+v", res)
	}

	s1 = []string{"element1", "element2", "element3"}
	res = mergeSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 3 {
			t.Errorf("Expected slice to have 3 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "element1") {
			t.Error("Expected slice to contain \"element1\", it didn't")
		}
		if !stringSliceContains(res, "element2") {
			t.Error("Expected slice to contain \"element2\", it didn't")
		}
		if !stringSliceContains(res, "element3") {
			t.Error("Expected slice to contain \"element3\", it didn't")
		}
	}

	s2 = []string{"element1", "element2", "element3"}
	res = mergeSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 3 {
			t.Errorf("Expected slice to have 3 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "element1") {
			t.Error("Expected slice to contain \"element1\", it didn't")
		}
		if !stringSliceContains(res, "element2") {
			t.Error("Expected slice to contain \"element2\", it didn't")
		}
		if !stringSliceContains(res, "element3") {
			t.Error("Expected slice to contain \"element3\", it didn't")
		}
	}

	s1 = []string{"element1", "element2", "element3"}
	s2 = []string{"element3", "element4"}
	res = mergeSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 4 {
			t.Errorf("Expected slice to have 4 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "element1") {
			t.Error("Expected slice to contain \"element1\", it didn't")
		}
		if !stringSliceContains(res, "element2") {
			t.Error("Expected slice to contain \"element2\", it didn't")
		}
		if !stringSliceContains(res, "element3") {
			t.Error("Expected slice to contain \"element3\", it didn't")
		}
		if !stringSliceContains(res, "element4") {
			t.Error("Expected slice to contain \"element4\", it didn't")
		}
	}

	// The public implementation uses a variadic argument to enable merging of n elements.
	var s3 []string
	s1 = []string{}
	s2 = []string{}
	res = MergeSlices(s1, s2)
	if res != nil {
		t.Errorf("expected nil, got %+v", res)
	}

	s1 = []string{"element1", "element2", "element3"}
	res = MergeSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 3 {
			t.Errorf("Expected slice to have 3 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "element1") {
			t.Error("Expected slice to contain \"element1\", it didn't")
		}
		if !stringSliceContains(res, "element2") {
			t.Error("Expected slice to contain \"element2\", it didn't")
		}
		if !stringSliceContains(res, "element3") {
			t.Error("Expected slice to contain \"element3\", it didn't")
		}
	}

	s2 = []string{"element1", "element2", "element3"}
	res = MergeSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 3 {
			t.Errorf("Expected slice to have 3 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "element1") {
			t.Error("Expected slice to contain \"element1\", it didn't")
		}
		if !stringSliceContains(res, "element2") {
			t.Error("Expected slice to contain \"element2\", it didn't")
		}
		if !stringSliceContains(res, "element3") {
			t.Error("Expected slice to contain \"element3\", it didn't")
		}
	}

	s1 = []string{"element1", "element2", "element3"}
	s2 = []string{"element3", "element4"}
	res = MergeSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 4 {
			t.Errorf("Expected slice to have 4 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "element1") {
			t.Error("Expected slice to contain \"element1\", it didn't")
		}
		if !stringSliceContains(res, "element2") {
			t.Error("Expected slice to contain \"element2\", it didn't")
		}
		if !stringSliceContains(res, "element3") {
			t.Error("Expected slice to contain \"element3\", it didn't")
		}
		if !stringSliceContains(res, "element4") {
			t.Error("Expected slice to contain \"element4\", it didn't")
		}
	}

	s1 = []string{"apples = 1", "bananas=bunches", "oranges=10lbs"}
	s2 = []string{"apples= a bushel", "oranges=10lbs", "strawberries=some"}
	s3 = []string{"strawberries=2lbs", "strawberries=some", "mangoes=2", "starfruit=4"}
	res = MergeSlices(s1, s2, s3)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 8 {
			t.Errorf("Expected slice to have 8 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "apples = 1") {
			t.Error("Expected slice to contain \"apples = 1\", it didn't")
		}
		if !stringSliceContains(res, "bananas=bunches") {
			t.Error("Expected slice to contain \"bananas=bunches\", it didn't")
		}
		if !stringSliceContains(res, "oranges=10lbs") {
			t.Error("Expected slice to contain \"oranges=10lbs\", it didn't")
		}
		if !stringSliceContains(res, "apples= a bushel") {
			t.Error("Expected slice to contain \"apples= a bushel\", it didn't")
		}
		if !stringSliceContains(res, "strawberries=some") {
			t.Error("Expected slice to contain \"strawberries=some\", it didn't")
		}
		if !stringSliceContains(res, "strawberries=2lbs") {
			t.Error("Expected slice to contain \"strawberries=2lbs\", it didn't")
		}
		if !stringSliceContains(res, "mangoes=2") {
			t.Error("Expected slice to contain \"mangoes=2\", it didn't")
		}
		if !stringSliceContains(res, "starfruit=4") {
			t.Error("Expected slice to contain \"starfruit=4\", it didn't")
		}
	}
}

func TestMergeSettingsSlices(t *testing.T) {
	var s1, s2, res []string
	res = mergeSettingsSlices(s1, s2)
	if res != nil {
		t.Errorf("expected nil, got %+v", res)
	}

	s1 = []string{"key1=value1", "key2=value2", "key3=value3"}
	res = mergeSettingsSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 3 {
			t.Errorf("Expected slice to have 3 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "key1=value1") {
			t.Error("Expected slice to contain \"key1=value1\", it didn't")
		}
		if !stringSliceContains(res, "key2=value2") {
			t.Error("Expected slice to contain \"key2=value2\", it didn't")
		}
		if !stringSliceContains(res, "key3=value3") {
			t.Error("Expected slice to contain \"key3=value3\", it didn't")
		}
	}

	s2 = []string{"key1=value1", "key2=value2", "key3=value3"}
	res = mergeSettingsSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 3 {
			t.Errorf("Expected slice to have 3 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "key1=value1") {
			t.Error("Expected slice to contain \"key1=value1\", it didn't")
		}
		if !stringSliceContains(res, "key2=value2") {
			t.Error("Expected slice to contain \"key2=value2\", it didn't")
		}
		if !stringSliceContains(res, "key3=value3") {
			t.Error("Expected slice to contain \"key3=value3\", it didn't")
		}
	}

	s1 = []string{"key1=value1", "key2=value2", "key3=value3"}
	s2 = []string{"key2=value22", "key4=value4"}
	res = mergeSettingsSlices(s1, s2)
	if res == nil {
		t.Error("Expected a non-nil slice, got nil")
	} else {
		if len(res) != 4 {
			t.Errorf("Expected slice to have 4 elements, had %d", len(res))
		}
		if !stringSliceContains(res, "key1=value1") {
			t.Error("Expected slice to contain \"key1=value1\", it didn't")
		}
		if !stringSliceContains(res, "key2=value22") {
			t.Error("Expected slice to contain \"key2=value22\", it didn't")
		}
		if !stringSliceContains(res, "key3=value3") {
			t.Error("Expected slice to contain \"key3=value3\", it didn't")
		}
		if !stringSliceContains(res, "key4=value4") {
			t.Error("Expected slice to contain \"key4=value4\", it didn't")
		}
	}

}

func TestVarMapFromSlice(t *testing.T) {
	res := varMapFromSlice([]string{})
	if len(res) != 0 {
		t.Errorf("expected res to have no members, had %d", len(res))
	}

	res = varMapFromSlice(nil)
	if res != nil {
		t.Errorf("Expected res to be nil, got %v", res)
	}

	sl := []string{"key1=value1", "key2=value2"}
	res = varMapFromSlice(sl)
	if res == nil {
		t.Error("Did not expect res to be nil, but it was")
	} else {
		if len(res) != 2 {
			t.Errorf("Expected length == 2, was %d", len(res))
		}
		v, ok := res["key1"]
		if !ok {
			t.Error("Expected \"key1\" to exist, but it didn't")
		} else {
			if v != "value1" {
				t.Errorf("expected value of \"key1\" to be \"value1\", was %q", v)
			}
		}
		v, ok = res["key2"]
		if !ok {
			t.Error("Expected \"key2\" to exist, but it didn't")
		} else {
			if v != "value2" {
				t.Errorf("expected value of \"key2\" to be \"value2\", was %q", v)
			}
		}
	}
}

func TestParseVar(t *testing.T) {
	k, v := parseVar("")
	if k != "" {
		t.Errorf("Expected key to be empty, was %q", k)
	}
	if v != "" {
		t.Errorf("Expected value to be empty, was %q", v)
	}

	k, v = parseVar("key1=value1")
	if k != "key1" {
		t.Errorf("Expected  \"key1\", got %q", k)
	}
	if v != "value1" {
		t.Errorf("Expected \"value1\", got %q", v)
	}
}

func TestIndexOfKeyInVarSlice(t *testing.T) {
	sl := []string{"key1=value1", "key2=value2", "key3=value3", "key4=value4", "key5=value5"}
	i := indexOfKeyInVarSlice("key3", sl)
	if i != 2 {
		t.Errorf("Expected index of 2, got %d", i)
	}

	i = indexOfKeyInVarSlice("key6", sl)
	if i != -1 {
		t.Errorf("Expected index of -1, got %d", i)
	}
}

func TestGetPackerVariableFromString(t *testing.T) {
	res := getPackerVariableFromString("")
	if res != "" {
		t.Errorf("Expected empty, got %q", res)
	}

	res = getPackerVariableFromString("variableName")
	if res != "{{user `variableName` }}" {
		t.Errorf("Expected \"{{user `variableName` }}\", got %q", res)
	}
}

func TestGetDefaultISOInfo(t *testing.T) {
	d := []string{"arch=amd64", "image=desktop", "release=14.04", "notakey=notavalue"}
	arch, image, release := getDefaultISOInfo(d)
	if arch != "amd64" {
		t.Errorf("Expected \"amd64\", got %q", arch)
	}
	if image != "desktop" {
		t.Errorf("Expected \"desktop\", got %q", image)
	}
	if release != "14.04" {
		t.Errorf("Expected \"14.04\", got %q", release)
	}
}

func TestAppendSlash(t *testing.T) {
	s := appendSlash("")
	if s != "" {
		t.Errorf("Expected an empty string ,got %q", s)
	}

	s = appendSlash("test")
	if s != "test/" {
		t.Errorf("Expected \"test/\", got %q", s)
	}

	s = appendSlash("test/")
	if s != "test/" {
		t.Errorf("Expected \"test/\", got %q", s)
	}
}

func TestTrimSuffix(t *testing.T) {
	s := trimSuffix("", "")
	if s != "" {
		t.Errorf("Expected an empty string, got %q", s)
	}

	s = trimSuffix("aStringWithaSuffix", "")
	if s != "aStringWithaSuffix" {
		t.Errorf("Expected \"aStringWithaSuffix\", got %q", s)
	}

	s = trimSuffix("aStringWithaSuffix", "aszc")
	if s != "aStringWithaSuffix" {
		t.Errorf("Expected \"aStringWithaSuffix\", got %q", s)
	}

	s = trimSuffix("aStringWithaSuffix", "Suffix")
	if s != "aStringWitha" {
		t.Errorf("Expected \"aStringWitha\", got %q", s)
	}
}

func TestCopyFile(t *testing.T) {
	wB, err := copyFile("", "", "test")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "no source directory received" {
			t.Errorf("Expected \"copyFile: no source directory passed\", got %q", err.Error())
		}
	}
	if wB != 0 {
		t.Errorf("Expected 0 bytes written, %d were written", wB)
	}

	wB, err = copyFile("", "conf", "")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "no destination directory received" {
			t.Errorf("Expected \"copyFile: no destination directory passed\", got %q", err.Error())
		}
	}

	wB, err = copyFile("", "conf", "test")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "no filename received" {
			t.Errorf("Expected \"copyFile: no filename passed\", got %q", err.Error())
		}
	}
}

func TestCopyDirContent(t *testing.T) {
	origDir, err := ioutil.TempDir("", "orig")
	os.MkdirAll(origDir, os.FileMode(0766))

	copyDir, err := ioutil.TempDir("", "test")
	os.MkdirAll(copyDir, os.FileMode(0766))
	ioutil.WriteFile(filepath.Join(origDir, "test1"), []byte("this is a test file"), 0777)
	ioutil.WriteFile(filepath.Join(origDir, "test2"), []byte("this is another test file"), 0777)

	notADir := filepath.Join(origDir, "zzz")
	err = copyDirContent(notADir, copyDir)
	if err == nil {
		t.Error("Expected an error, none received")
	} else {
		if err.Error() != fmt.Sprintf("nothing copied: the source, %s, does not exist", notADir) {
			t.Errorf("Expected \"nothing copied: the source, %q, does not exist\", got %q", notADir, err.Error())
		}
	}

	err = copyDirContent(origDir, copyDir)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}

	os.RemoveAll(copyDir)
	os.RemoveAll(origDir)
}

func TestDeleteDirContent(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "testdel")
	testFile1, err := os.Create(filepath.Join(tmpDir, "test1.txt"))
	if err != nil {
		t.Errorf("no error expected, got %q", err.Error())
	} else {
		testFile1.Close()
	}

	testFile2, err := os.Create(filepath.Join(tmpDir + "test2.txt"))
	if err != nil {
		t.Errorf("no error expected, got %q", err.Error())
	} else {
		testFile2.Close()
	}

	err = deleteDirContent(filepath.Join(tmpDir, "testtest"))
	if err == nil {
		t.Error("Expected an error, none occurred")
	} else {
		if err.Error() != "stat "+tmpDir+"/testtest: no such file or directory" {
			t.Errorf("expected \"stat "+tmpDir+"/testtest: no such file or directory\", got %q", err.Error())
		}
	}

	err = os.RemoveAll(tmpDir)
	if err != nil {
		t.Errorf("Expected no error: got %q", err.Error())
	}
}

func TestSubString(t *testing.T) {
	testString := "This is a test"
	res := Substring(testString, -1, 0)
	if res != "" {
		t.Errorf("Expected empty string, \"\", got %q", res)
	}

	res = Substring(testString, 4, 0)
	if res != "" {
		t.Errorf("Expected empty string, \"\", got %q", res)
	}

	res = Substring(testString, 4, -3)
	if res != "" {
		t.Errorf("Expected empty string, \"\", got %q", res)
	}

	res = Substring(testString, 4, 12)
	if res == "" {
		t.Error("Expected a substring, got an empty string \"\"")
	} else {
		if res != " is a test" {
			t.Errorf("Expected \" is a test\". got %q", res)
		}
	}

	res = Substring(testString, 4, 4)
	if res == "" {
		t.Error("Expected a substring, got an empty string \"\"")
	} else {
		if res != " is " {
			t.Errorf("Expected \" is \". got %q", res)
		}
	}
}

func TestMergedKeysFromMaps(t *testing.T) {
	map1 := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": []string{
			"element1",
			"element2",
		},
	}

	keys := mergedKeysFromMaps(map1)
	if len(keys) != 3 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	} else {
		b := stringSliceContains(keys, "key1")
		if !b {
			t.Error("expected \"key1\" to be in merged keys, not found")
		}
		b = stringSliceContains(keys, "key2")
		if !b {
			t.Error("expected \"key2\" to be in merged keys, not found")
		}
		b = stringSliceContains(keys, "key3")
		if !b {
			t.Error("expected \"key3\" to be in merged keys, not found")
		}
	}
	map2 := map[string]interface{}{
		"key1": "another value",
		"key4": "value4",
	}

	keys = mergedKeysFromMaps(map1, map2)
	if len(keys) != 4 {
		t.Errorf("expected 3 keys, got %d", len(keys))
	} else {
		b := stringSliceContains(keys, "key1")
		if !b {
			t.Error("expected \"key1\" to be in merged keys, not found")
		}
		b = stringSliceContains(keys, "key2")
		if !b {
			t.Error("expected \"key2\" to be in merged keys, not found")
		}
		b = stringSliceContains(keys, "key3")
		if !b {
			t.Error("expected \"key3\" to be in merged keys, not found")
		}
		b = stringSliceContains(keys, "key4")
		if !b {
			t.Error("expected \"key4\" to be in merged keys, not found")
		}
		b = stringSliceContains(keys, "")
		if b {
			t.Error("expected merged keys to not have \"\", it was in merged keys slice")
		}

	}
}
