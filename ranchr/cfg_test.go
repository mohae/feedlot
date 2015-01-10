package ranchr

import (
	"os"
	"reflect"
	"testing"
)

func init() {
	setCommonTestData()
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
					"scripts/base.sh",
					"scripts/vagrant.sh",
					"scripts/virtualbox.sh",
					"scripts/cleanup.sh",
				},
			},
		},
	}

	nw := map[string]interface{}{
		"type": "shell",
		"override": map[string]interface{}{
			"vmware-iso": map[string]interface{}{
				"scripts": []string{
					"scripts/base.sh",
					"scripts/vagrant.sh",
					"scripts/virtualbox.sh",
					"scripts/cleanup.sh",
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
					"scripts/base.sh",
					"scripts/vagrant.sh",
					"scripts/virtualbox.sh",
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
		if err.Error() != "an error occurred while loading the default settings; check the log for more information" {
			t.Errorf("Expected \"an error occurred while loading the default settings; check the log for more information\", got %q", err.Error())
		} else {
			if d.MinPackerVersion != "" {
				t.Errorf("Expected \"\", got %q", d.MinPackerVersion)
			}
		}
	}

	os.Setenv(EnvDefaultsFile, "../test_files/conf/defaults_test.toml")
	os.Setenv(EnvRancherFile, "../test_files/rancher.cfg")
	d = defaults{}
	err = d.LoadOnce()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	} else {
		if MarshalJSONToString.Get(d) != MarshalJSONToString.Get(testDefaults) {
			t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDefaults), MarshalJSONToString.Get(d))
		}
		/*
			if MarshalJSONToString.Get(d.IODirInf) != MarshalJSONToString.Get(testDefaults.IODirInf) {
				t.Errorf("Expected %q, got %q", MarshalJSONToString.Get(testDefaults.IODirInf), MarshalJSONToString.Get(d.IODirInf))
			}
			/*
				So(d.PackerInf, ShouldResemble, testDefaults.PackerInf)
				So(d.BuildInf, ShouldResemble, testDefaults.BuildInf)
				So(d.build.BuilderTypes, ShouldResemble, testDefaults.build.BuilderTypes)
			/
			So(MarshalJSONToString.Get(d.build.Builders[BuilderVirtualBoxISO]) != MarshalJSONToString.Get(testDefaults.build.Builders[BuilderVirtualBoxISO]))
			So(d.build.PostProcessorTypes != testDefaults.build.PostProcessorTypes)
			So(MarshalJSONToString.Get(d.build.PostProcessors[PostProcessorVagrant]) != MarshalJSONToString.Get(testDefaults.build.PostProcessors[PostProcessorVagrant]))
			So(MarshalJSONToString.Get(d.build.PostProcessors[PostProcessorVagrantCloud])  MarshalJSONToString.Get(testDefaults.build.PostProcessors[PostProcessorVagrantCloud]))
			So(d.build.ProvisionerTypes, ShouldResemble, testDefaults.build.ProvisionerTypes)
			So(MarshalJSONToString.Get(d.build.Provisioners[ProvisionerShell]), ShouldEqual, MarshalJSONToString.Get(testDefaults.build.Provisioners[ProvisionerShell]))
		*/
	}
	_ = os.Setenv(EnvDefaultsFile, tmpEnvDefaultsFile)
}

func TestSupported(t *testing.T) {
	tmpEnv := os.Getenv(EnvSupportedFile)
	s := supported{}
	os.Setenv(EnvSupportedFile, "")
	err := s.LoadOnce()
	if err == nil {
		t.Errorf("expected error, none occurred")
	} else {
		if err.Error() != "an error occurred while loading the Supported information, please check the log" {
			t.Errorf("expected \"an error occurred while loading the Supported information, please check the log\" got %q", err.Error())
		}
		if s.loaded {
			t.Errorf("expected Supported info not to be loaded, but it was")
		}
	}

	s = supported{}
	os.Setenv(EnvSupportedFile, "../test_files/conf/supported_test.toml")
	err = s.LoadOnce()
	if err != nil {
		t.Errorf("unexpected error %q", err.Error())
	} else {
		if !s.loaded {
			t.Errorf("expected the Supported info to be loaded, but it wasn't")
		} else {
			if MarshalJSONToString.GetIndented(s.Distro[Ubuntu.String()]) != MarshalJSONToString.GetIndented(testSupported.Distro[Ubuntu.String()]) {
				t.Errorf("expected %q, got %q", MarshalJSONToString.GetIndented(testSupported.Distro[Ubuntu.String()]), MarshalJSONToString.GetIndented(s.Distro[Ubuntu.String()]))
			}
			if MarshalJSONToString.GetIndented(s.Distro[CentOS.String()]) != MarshalJSONToString.GetIndented(testSupported.Distro[CentOS.String()]) {
				t.Errorf("expected %q, got %q", MarshalJSONToString.GetIndented(testSupported.Distro[CentOS.String()]), MarshalJSONToString.GetIndented(s.Distro[CentOS.String()]))
			}
		}
	}
	// Set this because, for some reason it isn't set in testing >.>
	//	testSupported.Distro["ubuntu"].BaseURL = "http://releases.ubuntu.com/"
	_ = os.Setenv(EnvSupportedFile, tmpEnv)
}

func TestBuildsStuff(t *testing.T) {
	b := builds{}
	tmpEnv := os.Getenv(EnvBuildsFile)
	/*
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
	*/
	os.Setenv(EnvBuildsFile, "../test_files/conf/builds_test.toml")
	b.LoadOnce()
	//	t.Errorf("%+v", b)
	if b.loaded == false {
		t.Errorf("expected Build info to be loaded, but it wasn't")
	} else {
		if MarshalJSONToString.GetIndented(testBuilds.Build["test1"]) != MarshalJSONToString.GetIndented(b.Build["test1"]) {
			t.Errorf("expected %q, got %q", MarshalJSONToString.GetIndented(testBuilds.Build["test1"]), MarshalJSONToString.GetIndented(b.Build["test1"]))
		}
		if MarshalJSONToString.GetIndented(testBuilds.Build["test2"]) != MarshalJSONToString.GetIndented(b.Build["test2"]) {
			t.Errorf("expected %q, got %q", MarshalJSONToString.GetIndented(testBuilds.Build["test2"]), MarshalJSONToString.GetIndented(b.Build["test2"]))
		}
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

	os.Setenv(EnvBuildListsFile, "../test_files/conf/build_lists_test.toml")
	err = b.Load()
	if err != nil {
		t.Errorf("Did not expect an error, got %q", err.Error())
	} else {
		//check if testlist-1 exists
		lst, ok := b.List["testlist-1"]
		if !ok {
			t.Error("Expected \"testlist-1\" to exist in Build list map, not found")
		} else {
			if len(lst.Builds) != 2 {
				t.Errorf("Expected \"testlist-2\" to contain 4 elements, had %d", len(lst.Builds))
			}
			if !stringSliceContains(lst.Builds, "test1") {
				t.Error("Expected \"test1\" to be in \"testlist-1\" slice, not found")
			}
			if !stringSliceContains(lst.Builds, "test2") {
				t.Error("Expected \"test2\" to be in \"testlist-1\" slice, not found")
			}
		}
		lst, ok = b.List["testlist-2"]
		if !ok {
			t.Error("Expected \"testlist-2\" to exist in Build list map, not found")
		} else {
			if len(lst.Builds) != 4 {
				t.Errorf("Expected \"testlist-2\" to contain 4 elements, had %d", len(lst.Builds))
			}
			if !stringSliceContains(lst.Builds, "test1") {
				t.Error("Expected \"test1\" to be in \"testlist-2\" slice, not found")
			}
			if !stringSliceContains(lst.Builds, "test2") {
				t.Error("Expected \"test2\" to be in \"testlist-2\" slice, not found")
			}
			if !stringSliceContains(lst.Builds, "test3") {
				t.Error("Expected \"test3\" to be in \"testlist-2\" slice, not found")
			}
			if !stringSliceContains(lst.Builds, "test4") {
				t.Error("Expected \"test4\" to be in \"testlist-2\" slice, not found")
			}
		}
	}

	os.Setenv(EnvBuildListsFile, tmpEnv)
}
