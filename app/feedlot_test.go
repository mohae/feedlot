package app

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	json "github.com/mohae/unsafejson"
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

// simple func to create and populate a test directory.
// it is the callers responisibility to clean it up when done.
func createTmpTestDirFiles(s string) (dir string, files []string, err error) {
	dir, err = ioutil.TempDir("", s)
	if err != nil {
		return dir, files, err
	}
	// we make a subdirectory because archive. All tests expect dir/test as a result.
	tmpDir := filepath.Join(dir, "test")
	err = os.MkdirAll(tmpDir, 0777)
	if err != nil {
		return dir, files, err
	}
	data := []byte("abcdefghijklmnopqrstuvwxyz")
	for i := 0; i < 3; i++ {
		err := ioutil.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("test-%d", i)), data, 0777)
		if err != nil {
			return dir, files, err
		}
		files = append(files, filepath.Join(tmpDir, fmt.Sprintf("test-%d", i)))
	}
	return dir, files, nil
}

var testDistroDefaultUbuntu = rawTemplate{
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
		BuilderIDs: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
			"common": {
				templateSection{
					//ID: "common",
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
					//ID: "virtualbox-iso",
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
		PostProcessorIDs: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
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
		ProvisionerIDs: []string{"shell"},
		Provisioners: map[string]provisioner{
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
}

var testDistroDefaultCentOS = rawTemplate{
	PackerInf: PackerInf{
		MinPackerVersion: "0.4.0",
		Description:      "Test template config and feedlot options for CentOS",
	},
	IODirInf: IODirInf{
		TemplateOutputDir: "../test_files/out/:distro/:build_name",
		PackerOutputDir:   "boxes/:distro/:build_name",
		SourceDir:         "../test_files/src/:distro",
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
	dirs:    map[string]string{},
	files:   map[string]string{},
	build: build{
		BuilderIDs: []string{"virtualbox-iso", "vmware-iso"},
		Builders: map[string]builder{
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
		PostProcessorIDs: []string{
			"vagrant",
			"vagrant-cloud",
		},
		PostProcessors: map[string]postProcessor{
			"vagrant": {
				templateSection{
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
		ProvisionerIDs: []string{
			"shell",
		},
		Provisioners: map[string]provisioner{
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
}

var testDistroDefaults distroDefaults

func init() {
	var b bool
	testSupportedCentOS.IODirInf.SourceDirIsRelative = &b
	testSupportedCentOS.IODirInf.TemplateOutputDirIsRelative = &b
	testDistroDefaults = distroDefaults{Templates: map[Distro]rawTemplate{}, IsSet: true}
	testDistroDefaults.Templates[Ubuntu] = testDistroDefaultUbuntu
	testDistroDefaults.Templates[CentOS] = testDistroDefaultCentOS
}

func TestDistroFromString(t *testing.T) {
	tests := []struct {
		value    string
		expected Distro
	}{
		{"centos", CentOS},
		{"CentOS", CentOS},
		{"CENTOS", CentOS},
		{"debian", Debian},
		{"Debian", Debian},
		{"DEBIAN", Debian},
		{"ubuntu", Ubuntu},
		{"Ubuntu", Ubuntu},
		{"UBUNTU", Ubuntu},
		{"slackware", UnsupportedDistro},
		{"", UnsupportedDistro},
	}
	for i, test := range tests {
		d := ParseDistro(test.value)
		if d != test.expected {
			t.Errorf("%d: expected %s got %s", i, test.expected, d)
		}
	}

}

func TestDistroDefaultsGetTemplate(t *testing.T) {
	r, err := testDistroDefaults.GetTemplate("invalid")
	if err == nil {
		t.Error("expected \"unsupported distro: invalid\", got nil")
	} else {
		if err.Error() != "unsupported distro: invalid" {
			t.Errorf("unsupported distro: invalid, got %q", err)
		}
		if r != nil {
			t.Errorf("Expected nil, got %q", MarshalJSONToString.Get(r))
		}
	}

	r, err = testDistroDefaults.GetTemplate("ubuntu")
	if err != nil {
		t.Errorf("Expected no error, got %q", err)
	} else {
		if MarshalJSONToString.Get(r) != MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu]) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDistroDefaults.Templates[Ubuntu]), MarshalJSONToString.Get(r))
		}
	}
}

func TestGetSliceLenFromIface(t *testing.T) {
	ssl := []string{"a", "b", "c"}
	isl := []int{1, 2, 3}
	bsl := []byte("abc")
	s := "hello"
	i, err := getSliceLenFromIface(ssl)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if i != 3 {
			t.Errorf("Expected the len to be 3, got %d", i)
		}
	}
	i, err = getSliceLenFromIface(isl)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if i != 3 {
			t.Errorf("Expected the len to be 3, got %d", i)
		}
	}
	i, err = getSliceLenFromIface(bsl)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	} else {
		if i != 3 {
			t.Errorf("Expected the len to be 3, got %d", i)
		}
	}
	_, err = getSliceLenFromIface(s)
	if err == nil {
		t.Error("Expected an error, got nil")
	} else {
		if err.Error() != "getSliceLenFromIface expected a slice, got \"string\"" {
			t.Errorf("Expected errror to be \"getSliceLenFromIface expected a slice, got \"string\"\", got %q", err)
		}
	}

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
	var err error
	res, err = mergeSettingsSlices(s1, s2)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
	if res != nil {
		t.Errorf("expected nil, got %+v", res)
	}

	s1 = []string{"key1=value1", "key2=value2", "key3=value3"}
	res, err = mergeSettingsSlices(s1, s2)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
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
	res, err = mergeSettingsSlices(s1, s2)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
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
	res, err = mergeSettingsSlices(s1, s2)
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
	}
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

func TestCopyFile(t *testing.T) {
	_, err := copyFile("", "test")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "copyfile error: source name was empty" {
			t.Errorf("Expected \"copyfile error: source name was empty\", got %q", err)
		}
	}

	_, err = copyFile("conf", "")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "copyfile error: destination name was empty" {
			t.Errorf("Expected \"copyfile error: destination name was empty\", got %q", err)
		}
	}

	_, err = copyFile("conf", "test")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "copyfile error: destination name, \"test\", did not include a directory" {
			t.Errorf("Expected \"copyfile error: destination name, \"test\", did not include a directory\", got %q", err)
		}
	}
	dir, files, err := createTmpTestDirFiles("feedlot-copyfile-")
	if err != nil {
		t.Errorf("unexpected error while setting up a copy test: %s", err)
	}
	fname := filepath.Base(files[1])
	toDir, err := ioutil.TempDir("", "copyfile")
	_, err = copyFile(files[1], path.Join(toDir, fname))
	if err != nil {
		t.Errorf("Expected no error, got %q", err)
	}
	_, err = os.Stat(path.Join(toDir, fname))
	if err != nil {
		t.Errorf("expected no error, got %q", err)
	}
	os.RemoveAll(dir)
	os.RemoveAll(toDir)
}

func TestCopyDirContent(t *testing.T) {
	dir, files, err := createTmpTestDirFiles("feedlot-copydircontent-")
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err)
		return
	}
	toDir, err := ioutil.TempDir("", "copyto")
	if err != nil {
		t.Errorf("cannot create destination directory for copy: %q", err)
	}
	notADir := filepath.Join(dir, "zzz")
	err = copyDir(notADir, toDir)
	if err == nil {
		t.Error("Expected an error, none received")
	} else {
		if err.Error() != fmt.Sprintf("copyDir error: %s does not exist", notADir) {
			t.Errorf("Expected \"copyDir error: %s, does not exist\", got %q", notADir, err)
		}
	}
	err = copyDir(dir, toDir)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err)
	}
	// make sure that all the files have been copied
	for _, file := range files {
		fname := filepath.Base(file)
		_, err := os.Stat(filepath.Join(toDir, "test", fname))
		if err != nil {
			t.Errorf("expected no error, got %q", err)
		}
	}
	os.RemoveAll(dir)
	os.RemoveAll(toDir)
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

func TestMergedKeysFromComponentMaps(t *testing.T) {
	map1 := map[string]Componenter{
		"key1": provisioner{},
		"key2": provisioner{},
		"key3": builder{},
	}

	keys := mergeKeysFromComponentMaps(map1)
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
	map2 := map[string]Componenter{
		"key1": provisioner{},
		"key4": builder{},
	}

	keys = mergeKeysFromComponentMaps(map1, map2)
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

func TestSetParentDir(t *testing.T) {
	tests := []struct {
		d        string
		p        string
		expected string
	}{
		{"", "", ""},
		{"", "path", "path"},
		{"dir", "path", "dir/path"},
		{"dir", "some/path", "some/path"},
	}
	for i, test := range tests {
		r := setParentDir(test.d, test.p)
		if r != test.expected {
			t.Errorf("setParentDir %d: expected %q, got %q", i, test.expected, r)
		}
	}
}

func TestGetUniqueFilename(t *testing.T) {
	tests := []struct {
		filename    string
		layout      string
		expected    string
		expectedErr string
	}{
		{"", "", "", ""},
		{"../test_files/notthere.txt", "", "../test_files/notthere.txt", ""},
		{"../test_files/notthere.txt", "2006", "../test_files/notthere.txt", ""},
		{"../test_files/not.there.txt", "", "../test_files/not.there.txt", ""},
		{"../test_files/not.there.txt", "2006", "../test_files/not.there.txt", ""},
		{"../test_files/test.txt", "", "../test_files/test-3.txt", ""},
		{"../test_files/test.txt", "2006", "../test_files/test.2015-2.txt", ""},
		{"../test_files/test.file.txt", "", "../test_files/test.file-2.txt", ""},
		{"../test_files/test.file.txt", "2006", "../test_files/test.file.2015-1.txt", ""},
	}
	for i, test := range tests {
		f, err := getUniqueFilename(test.filename, test.layout)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("TestGetUniqueFilename %d:  Expected error to be %q. got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("TestGetUniqueFilename %d: Expected no error, got %q", i, err)
			continue
		}
		if test.expected != f {
			// test that paths are the same
			dirW, tmpW := filepath.Split(test.expected)
			dirG, tmpG := filepath.Split(f)
			if dirW != dirG {
				t.Errorf("TestGetUniqueFilename %d: Expected %q, got %q", i, test.expected, f)
				continue
			}
			// test that the ext is t he same
			if filepath.Ext(tmpW) != filepath.Ext(tmpG) {
				t.Errorf("TestGetUniqueFilename %d: Expected %q, got %q", i, test.expected, f)
				continue
			}
			// test that the first part of the file is the same
			// portion between first element and ext is ignored; makes it
			// stable over time.
			partsW := strings.Split(tmpW, ".")
			partsG := strings.Split(tmpG, ".")
			if partsW[0] != partsG[0] {
				t.Errorf("TestGetUniqueFilename %d: Expected %q, got %q", i, test.expected, f)
			}
		}
	}
}

func TestIndexDir(t *testing.T) {
	dir, files, err := createTmpTestDirFiles("feedlot-indexDir")
	if err != nil {
		t.Errorf("An error occurred during test file creation, aborting IndexDir tests: %q", err)
		return
	}
	_, _, err = indexDir("")
	if err == nil {
		t.Errorf("Expected an error, got nil")
	} else {
		if err.Error() != "received an empty parameter, expected a value" {
			t.Errorf("Expected \"received an empty parameter, expected a value\", got %q", err)
		}
	}
	_, _, err = indexDir(filepath.Join(dir, "notthere"))
	if err == nil {
		t.Errorf("Expected an error, got nil")
	} else {
		if !os.IsNotExist(err) {
			t.Errorf("Expected os.IsNotExist(), got %q", err)
		}
	}
	_, _, err = indexDir(files[1])
	if err == nil {
		t.Errorf("Expected an error, got none")
	} else {
		if err.Error() != fmt.Sprintf("cannot index %s: not a directory", files[1]) {
			t.Errorf("Expected error to be \"cannot index %s: not a directory\": got %q", files[1], err)
		}
	}
	dirs, fnames, err := indexDir(dir)
	if err != nil {
		t.Errorf("Expected error to be nil; got %q", err)
		goto next
	}
	if len(dirs) != 1 {
		t.Errorf("Expected 1 directory in the results, got %v", dirs)
	}
	if dirs[0] != "test" {
		t.Errorf("Expected the directory \"test\" to be indexed; got %q", dirs[1])
	}
	if len(fnames) > 0 {
		t.Errorf("Expected 0 files to be indexed; got %d", len(fnames))
	}
next:
	dirs, fnames, err = indexDir(filepath.Join(dir, "test"))
	if err != nil {
		t.Errorf("Expected error to be nil; got %q", err)
		goto done
	}
	if len(dirs) > 0 {
		t.Errorf("Expected 0 files to be idnexed; got %d", len(dirs))
	}
	if len(files) != len(fnames) {
		t.Errorf("Expected %d files to be indexed; got %d", len(files), len(fnames))
	}
	for _, file := range files {
		if !stringSliceContains(fnames, filepath.Base(file)) {
			t.Errorf("Expected %q to be in the indexed files list; it wasn't", file)
		}
	}
done:
	_ = os.RemoveAll(dir)
}
