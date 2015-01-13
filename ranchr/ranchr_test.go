package ranchr

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mohae/deepcopy"
)

func init() {
	setCommonTestData()
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
	// Preserve current state.
	tmpConfig := os.Getenv(EnvRancherFile)
	tmpBuildsFile := os.Getenv(EnvBuildsFile)
	tmpBuildListsFile := os.Getenv(EnvBuildListsFile)
	tmpDefaultsFile := os.Getenv(EnvDefaultsFile)
	tmpLogToFile := os.Getenv(EnvLogToFile)
	tmpLogFilename := os.Getenv(EnvLogFilename)
	tmpLogLevelFile := os.Getenv(EnvLogLevelFile)
	tmpLogLevelStdout := os.Getenv(EnvLogLevelStdout)
	tmpParamDelimStart := os.Getenv(EnvParamDelimStart)
	tmpSupportedFile := os.Getenv(EnvSupportedFile)

	os.Setenv(EnvRancherFile, "")
	os.Setenv(EnvBuildsFile, "")
	os.Setenv(EnvBuildListsFile, "")
	os.Setenv(EnvDefaultsFile, "")
	os.Setenv(EnvLogToFile, "")
	os.Setenv(EnvLogFilename, "")
	os.Setenv(EnvLogLevelFile, "")
	os.Setenv(EnvLogLevelStdout, "")
	os.Setenv(EnvParamDelimStart, "")
	os.Setenv(EnvSupportedFile, "")

	err := SetEnv()
	if err == nil {
		t.Error("Expected an error, was nil")
	} else {
		if err.Error() != "open rancher.cfg: no such file or directory" {
			t.Errorf("Expected \"open rancher.cfg: no such file or directory\", %q", err.Error())
		}
	}

	os.Setenv(EnvRancherFile, testRancherCfg)
	err = SetEnv()
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if os.Getenv(EnvBuildsFile) != testBuildsFile {
			t.Errorf("Expected %q, got %q", testBuildsFile, os.Getenv(EnvBuildsFile))
		}
		if os.Getenv(EnvBuildListsFile) != testBuildListsFile {
			t.Errorf("Expected %q, got %q", testBuildListsFile, os.Getenv(EnvBuildListsFile))
		}
		if os.Getenv(EnvDefaultsFile) != testDefaultsFile {
			t.Errorf("Expected %q, got %q", testDefaultsFile, os.Getenv(EnvDefaultsFile))
		}
		if os.Getenv(EnvLogToFile) != "false" {
			t.Errorf("Expected \"false\", got %q", os.Getenv(EnvLogToFile))
		}
		if os.Getenv(EnvLogToFile) != "false" {
			t.Errorf("Expected \"false\", got %q", os.Getenv(EnvLogToFile))
		}
		if os.Getenv(EnvLogLevelFile) != "INFO" {
			t.Errorf("Expected \"INFO\", got %q", os.Getenv(EnvLogLevelFile))
		}
		if os.Getenv(EnvLogLevelStdout) != "TRACE" {
			t.Errorf("Expected \"TRACE\", got %q", os.Getenv(EnvLogLevelStdout))
		}
		if os.Getenv(EnvParamDelimStart) != ":" {
			t.Errorf("Expected \":\", got %q", os.Getenv(EnvParamDelimStart))
		}
		if os.Getenv(EnvSupportedFile) != testSupportedFile {
			t.Errorf("Expected %q, got %q", testSupportedFile, os.Getenv(EnvSupportedFile))
		}
	}

	os.Setenv(EnvRancherFile, testRancherCfg)
	err = SetEnv()
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	} else {
		if os.Getenv(EnvDefaultsFile) != testDefaultsFile {
			t.Errorf("Expected %q, got %q", testDefaultsFile, os.Getenv(EnvDefaultsFile))
		}
	}

	// Restore the state
	os.Setenv(EnvRancherFile, tmpConfig)
	os.Setenv(EnvBuildsFile, tmpBuildsFile)
	os.Setenv(EnvBuildListsFile, tmpBuildListsFile)
	os.Setenv(EnvDefaultsFile, tmpDefaultsFile)
	os.Setenv(EnvLogToFile, tmpLogToFile)
	os.Setenv(EnvLogFilename, tmpLogFilename)
	os.Setenv(EnvLogLevelFile, tmpLogLevelFile)
	os.Setenv(EnvLogLevelStdout, tmpLogLevelStdout)
	os.Setenv(EnvParamDelimStart, tmpParamDelimStart)
	os.Setenv(EnvSupportedFile, tmpSupportedFile)
}

// TODO add check of results other than error state and fix
func TestBuildDistro(t *testing.T) {
	tmpEnvRancherFile := os.Getenv(EnvRancherFile)
	tmpEnvBuildsFile := os.Getenv(EnvBuildsFile)
	tmpEnvDefaultsFile := os.Getenv(EnvDefaultsFile)
	tmpEnvParamDelimStart := os.Getenv(EnvParamDelimStart)
	tmpEnvSupportedFile := os.Getenv(EnvSupportedFile)
	os.Setenv(EnvRancherFile, testRancherCfg)
	os.Setenv(EnvBuildsFile, testBuildsFile)
	os.Setenv(EnvDefaultsFile, testDefaultsFile)
	os.Setenv(EnvParamDelimStart, ":")
	os.Setenv(EnvSupportedFile, testSupportedFile)

	err := DistroDefaults.Set()
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}
	aFilter := ArgsFilter{Arch: "amd64", Distro: "ubuntu", Image: "server", Release: "14.04"}
	err = BuildDistro(aFilter)
	if err != nil {
		t.Errorf("Expected error to be nil, got %q", err.Error())
	}

	os.Setenv(EnvRancherFile, tmpEnvRancherFile)
	os.Setenv(EnvBuildsFile, tmpEnvBuildsFile)
	os.Setenv(EnvDefaultsFile, tmpEnvDefaultsFile)
	os.Setenv(EnvParamDelimStart, tmpEnvParamDelimStart)
	os.Setenv(EnvSupportedFile, tmpEnvSupportedFile)
}

func TestbuildPackerTemplateFromDistros(t *testing.T) {
	a := ArgsFilter{}
	tmp := os.Getenv(EnvRancherFile)
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

	_ = os.Setenv(EnvRancherFile, testRancherCfg)
	a = ArgsFilter{Distro: "ubuntu", Arch: "amd64", Image: "desktop", Release: "14.04"}
	err = buildPackerTemplateFromDistro(a)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	}

	os.Setenv(EnvRancherFile, tmp)
}

/*
TODO: rewrite this test to work with refactored code
func TestBuildBuilds(t *testing.T) {
	tmpEnvRancherFile := os.Getenv(EnvRancherFile)
	tmpEnvBuildsFile := os.Getenv(EnvBuildsFile)
	tmpEnvDefaultsFile := os.Getenv(EnvDefaultsFile)
	tmpEnvParamDelimStart := os.Getenv(EnvParamDelimStart)
	tmpEnvSupportedFile := os.Getenv(EnvSupportedFile)
	os.Setenv(EnvRancherFile, testRancherCfg)
	os.Setenv(EnvBuildsFile, testBuildsFile)
	os.Setenv(EnvDefaultsFile, testDefaultsFile)
	os.Setenv(EnvParamDelimStart, ":")
	os.Setenv(EnvSupportedFile, testSupportedFile)
	_ = loadSupported()

	bldName := ""
	resultString, err := BuildBuilds(bldName)
	if err == nil {
		t.Error("Expected an error, none received")
	} else {
		if err.Error() != "Nothing to build. No build name was passed" {
			t.Errorf("Expected \"Nothing to build. No build name was passed\", got %q", err.Error())
		}
	}
	if resultString != "" {
		t.Errorf("Expected an empty string, got %q", resultString)
	}

	bldName = "test1"
	resultString, err = BuildBuilds(bldName)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	}
	if resultString != "Create Packer templates from named builds: 1 Builds were successfully processed and 0 Builds resulted in an error." {
		t.Errorf("Expected \"Create Packer templates from named builds: 1 Builds were successfully processed and 0 Builds resulted in an error.\", got %q", resultString)
	}

	bldName1 := "test1"
	bldName2 := "test2"
	resultString, err = BuildBuilds(bldName1, bldName2)
	resultString, err = BuildBuilds(bldName)
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	}
	if resultString != "Create Packer templates from named builds: 1 Builds were successfully processed and 1 Builds resulted in an error." {
		t.Errorf("Expected \"Create Packer templates from named builds: 1 Builds were successfully processed and 1 Builds resulted in an error.\", got %q", resultString)
	}

	os.Setenv(EnvRancherFile, tmpEnvRancherFile)
	os.Setenv(EnvBuildsFile, tmpEnvBuildsFile)
	os.Setenv(EnvDefaultsFile, tmpEnvDefaultsFile)
	os.Setenv(EnvParamDelimStart, tmpEnvParamDelimStart)
	os.Setenv(EnvSupportedFile, tmpEnvSupportedFile)
}
*/

func TestbuildPackerTemplateFromNamedBuild(t *testing.T) {
	tmp := os.Getenv(EnvRancherFile)
	tmpBuildsFile := os.Getenv(EnvBuildsFile)

	os.Setenv(EnvRancherFile, testRancherCfg)
	os.Setenv(EnvBuildsFile, "look/for/it/here/")
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

	go buildPackerTemplateFromNamedBuild("test1", doneCh)
	err = <-doneCh
	if err != nil {
		t.Errorf("Expected no error, received %q", err.Error())
	}

	//	doneCh := make(chan error, 1)
	go buildPackerTemplateFromNamedBuild("test11", doneCh)
	err = <-doneCh
	if err != nil {
		t.Errorf("Expected no error, received %q", err.Error())
	}

	//	doneCh := make(chan error, 1)
	go buildPackerTemplateFromNamedBuild("test2", doneCh)
	err = <-doneCh
	if err != nil {
		t.Errorf("Expected no error, received %q", err.Error())
	}

	close(doneCh)
	os.Setenv(EnvRancherFile, tmp)
	os.Setenv(EnvBuildsFile, tmpBuildsFile)
}

func TestCommandsFromFile(t *testing.T) {
	var commands []string
	var err error

	commands, err = commandsFromFile("")
	if err.Error() != "the passed Command filename was empty" {
		t.Errorf("Expected \"the passed Command filename was empty\", got %q", err.Error())
	}

	commands, err = commandsFromFile(testDir + "src/ubuntu/commands/execute_test.command")
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err.Error())
	}
	if commands == nil {
		t.Error("expected commands to no be nil, got nil")
	} else {
		if len(commands) != 1 {
			t.Errorf("expected commands to have 1 member, had %d", len(commands))
		}
		if !stringSliceContains(commands, "\"echo 'vagrant'|sudo -S sh '{{.Path}}'\"") {
			t.Error("expected commands to have member \"\"echo 'vagrant'|sudo -S sh '{{.Path}}'\"\", not a member")
		}
	}
	commands, err = commandsFromFile(testDir + "src/ubuntu/commands/boot_test.command")
	if err != nil {
		t.Errorf("expected error to be nil, got %q", err.Error())
	}
	if commands == nil {
		t.Error("expected commands to no be nil, got nil")
	} else {
		if len(commands) != 22 {
			t.Errorf("expected commands to have 22 member, had %d", len(commands))
		} else {
			// check the slice in order
			if commands[0] != "<esc><wait>" {
				t.Errorf("expected \"<esc><wait>\", got %q", commands[0])
			}
			if commands[1] != "<esc><wait>" {
				t.Errorf("expected \"<esc><wait>\", got %q", commands[1])
			}
			if commands[2] != "<enter><wait>" {
				t.Errorf("expected \"<enter><wait>\", got %q", commands[2])
			}
			if commands[3] != "/install/vmlinuz<wait>" {
				t.Errorf("expected \"/install/vmlinuz<wait>\", got %q", commands[3])
			}
			if commands[4] != " auto<wait>" {
				t.Errorf("expected \" auto<wait>\", got %q", commands[4])
			}
			if commands[5] != " console-setup/ask_detect=false<wait>" {
				t.Errorf("expected \" console-setup/ask_detect=false<wait>\", got %q", commands[5])
			}
			if commands[6] != " console-setup/layoutcode=us<wait>" {
				t.Errorf("expected \" console-setup/layoutcode=us<wait>\", got %q", commands[6])
			}
			if commands[7] != " console-setup/modelcode=pc105<wait>" {
				t.Errorf("expected \" console-setup/modelcode=pc105<wait>\", got %q", commands[7])
			}
			if commands[8] != " debconf/frontend=noninteractive<wait>" {
				t.Errorf("expected \" debconf/frontend=noninteractive<wait>\", got %q", commands[8])
			}
			if commands[9] != " debian-installer=en_US<wait>" {
				t.Errorf("expected \" debian-installer=en_US<wait>\", got %q", commands[9])
			}
			if commands[10] != " fb=false<wait>" {
				t.Errorf("expected \" fb=false<wait>\", got %q", commands[10])
			}
			if commands[11] != " initrd=/install/initrd.gz<wait>" {
				t.Errorf("expected \" initrd=/install/initrd.gz<wait>\", got %q", commands[11])
			}
			if commands[12] != " kbd-chooser/method=us<wait>" {
				t.Errorf("expected \" kbd-chooser/method=us<wait>\", got %q", commands[12])
			}
			if commands[13] != " keyboard-configuration/layout=USA<wait>" {
				t.Errorf("expected \" keyboard-configuration/layout=USA<wait>\", got %q", commands[13])
			}
			if commands[14] != " keyboard-configuration/variant=USA<wait>" {
				t.Errorf("expected \" keyboard-configuration/variant=USA<wait>\", got %q", commands[14])
			}
			if commands[15] != " locale=en_US<wait>" {
				t.Errorf("expected \" locale=en_US<wait>\", got %q", commands[15])
			}
			if commands[16] != " netcfg/get_hostname=ubuntu-1204<wait>" {
				t.Errorf("expected \" netcfg/get_hostname=ubuntu-1204<wait>\", got %q", commands[16])
			}
			if commands[17] != " netcfg/get_domain=vagrantup.com<wait>" {
				t.Errorf("expected \" netcfg/get_domain=vagrantup.com<wait>\", got %q", commands[17])
			}
			if commands[18] != " noapic<wait>" {
				t.Errorf("expected \" noapic<wait>\", got %q", commands[18])
			}
			if commands[19] != " preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg<wait>" {
				t.Errorf("expected \" preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg<wait>\", got %q", commands[19])
			}
			if commands[20] != " -- <wait>" {
				t.Errorf("expected \" -- <wait>\", got %q", commands[20])
			}
			if commands[21] != "<enter><wait>" {
				t.Errorf("expected \"<enter><wait>\", got %q", commands[21])
			}
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
	wB, err := copyFile("", "", testDir+"test")
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

	wB, err = copyFile("", testDir+"conf", "")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "no destination directory received" {
			t.Errorf("Expected \"copyFile: no destination directory passed\", got %q", err.Error())
		}
	}

	wB, err = copyFile("", testDir+"conf", testDir+"test")
	if err == nil {
		t.Error("Expected an error, no received")
	} else {
		if err.Error() != "no filename received" {
			t.Errorf("Expected \"copyFile: no filename passed\", got %q", err.Error())
		}
	}
	if wB != 0 {
		t.Errorf("Expected 0 bytes written, %d were written", wB)
	}

	wB, err = copyFile("builds_test.toml", testDir+"conf", testDir+"test")
	if err != nil {
		t.Errorf("Expected no error, got %q", err.Error())
	}
	if wB != 2571 {
		t.Errorf("Expected 2571 bytes written, %d were written", wB)
	}
}

func TestCopyDirContent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test")
	os.MkdirAll(tmpDir+"test", os.FileMode(0766))
	err = copyDirContent("../test_files/conf", tmpDir)
	if err != nil {
		t.Errorf("expected no error, got %q", err.Error())
	}

	err = copyDirContent("../test_files/buildbuild", tmpDir)
	if err == nil {
		t.Error("Expected an error, none received")
	} else {
		if err.Error() != "nothing copied: the source, ../test_files/buildbuild, does not exist" {
			t.Errorf("Expected \"nothing copied: the source, ../test_files/buildbuild, does not exist\", got %q", err.Error())
		}
	}
}

func TestDeleteDirContent(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "testdel")
	testFile1, err := os.Create(filepath.Join(tmpDir, "test1.txt"))
	if err != nil {
		t.Errorf("no error expected, got %q", err.Error())
	} else {
		testFile1.Close()
	}

	testFile2, err := os.Create(testDir + "test2.txt")
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

	err = deleteDirContent(tmpDir)
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
