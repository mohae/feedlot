package app

import (
	"fmt"
	"testing"

	"github.com/mohae/contour"
)

var testDefaults = &defaults{
	IODirInf: IODirInf{
		IncludeComponentString: "true",
		OutputDir:              "out/:build_name",
		SourceDir:              "src",
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
		},
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
						"output_directory = :out_dir",
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
						"guest_additions_path = VBoxGuestAdditions_{{ .Version }}.iso",
						"virtualbox_version_file = .vbox_version",
					},
					Arrays: map[string]interface{}{
						"vboxmanage": []string{
							"cpus=1",
							"memory=1024",
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
						"compression_level = 9",
						"keep_input_artifact = false",
						"output = :build_name.box",
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

var testSupported = map[string]distro{
	"centos": distro{
		BuildInf: BuildInf{
			BaseURL: "",
		},
		IODirInf: IODirInf{},
		PackerInf: PackerInf{
			MinPackerVersion: "",
			Description:      "Default template config and Rancher options for CentOS",
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
	},
	"debian": distro{
		BuildInf: BuildInf{
			BaseURL: "http://cdimage.debian.org/debian-cd/",
		},
		IODirInf: IODirInf{},
		PackerInf: PackerInf{
			MinPackerVersion: "",
			Description:      "Default template config and Rancher options for Debian",
		},
		Arch: []string{
			"i386",
			"amd64",
		},
		Image: []string{
			"lxde-CD-1",
			"netinst",
			"xfce-CD-1",
		},
		Release: []string{
			"8",
		},
		DefImage: []string{
			"release = 8",
			"image = netinst",
			"arch = amd64",
		},
	},
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
		Builders: map[string]builder{
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
		PostProcessors: map[string]postProcessor{
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
		Provisioners: map[string]provisioner{
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

var testBuild = map[string]rawTemplate{
	"1204-amd64": rawTemplate{
		Distro: "ubuntu",
		PackerInf: PackerInf{
			Description: "ubuntu LTS 1204 amd64 server build, minimal install",
		},
		Arch:    "amd64",
		Image:   "server",
		Release: "12.04",
		build: build{
			BuilderTypes: []string{
				"virtualbox-iso",
			},
			Builders: map[string]builder{
				"common": {
					templateSection{
						ID: "",
						Settings: []string{
							"ssh_wait_timeout = 300m",
						},
					},
				},
				"virtualbox-iso": {
					templateSection{
						ID: "",
						Arrays: map[string]interface{}{
							"vboxmanage": []string{
								"memory=4096",
							},
						},
					},
				},
			},
		},
	},
	"centos6": rawTemplate{
		Distro: "centos",
		PackerInf: PackerInf{
			Description: "Centos 6 w virtualbox-iso only",
		},
		build: build{
			BuilderTypes: []string{
				"virtualbox-iso",
			},
		},
	},
	"jessie": rawTemplate{
		Distro: "debian",
		PackerInf: PackerInf{
			Description: "debian jessie",
		},
		Arch: "amd64",
		build: build{
			BuilderTypes: []string{
				"virtualbox-iso",
			},
			Builders: map[string]builder{
				"virtualbox-iso": {
					templateSection{
						ID: "",
						Arrays: map[string]interface{}{
							"vboxmanage": []string{
								"--memory=4096",
							},
						},
					},
				},
			},
			PostProcessorTypes: []string{
				"vagrant",
			},
			ProvisionerTypes: []string{
				"basic-shell",
			},
			Provisioners: map[string]provisioner{
				"shell": {
					templateSection{
						ID: "basic-shell",
						Arrays: map[string]interface{}{
							"scripts": []string{
								"setup.sh",
								"sudoers.sh",
								"vagrant.sh",
								"customize.sh",
								"cleanup.sh",
							},
						},
					},
				},
			},
		},
	},
}

var testBuildList = map[string]list{
	"ubuntu-all": list{Builds: []string{"1204-amd64-server", "1310-amd64-desktop"}},
}

func TestBuildCopy(t *testing.T) {
	tstTpl := testBuild["jessie"]
	newBuild := tstTpl.build.copy()
	if fmt.Sprintf("%p", newBuild.BuilderTypes) == fmt.Sprintf("%p", tstTpl.build.BuilderTypes) {
		t.Errorf("The pointer for BuilderTypes is the same for both newBuild and testBuild: %x, expected them to be different.", fmt.Sprintf("%p", tstTpl.build.BuilderTypes))
		goto builderTypesEnd
	}
	if len(newBuild.BuilderTypes) != len(tstTpl.BuilderTypes) {
		t.Errorf("Expected newBuild.BuilderTypoes to have a length of %d; got %d", len(tstTpl.BuilderTypes), len(newBuild.BuilderTypes))
		goto builderTypesEnd
	}
	for i, v := range tstTpl.BuilderTypes {
		if v != newBuild.BuilderTypes[i] {
			t.Errorf("Expected builder type at index %d to be %q; got %q", i, v, newBuild.BuilderTypes[i])
		}
	}
builderTypesEnd:
	if fmt.Sprintf("%p", newBuild.Builders) == fmt.Sprintf("%p", tstTpl.build.Builders) {
		t.Errorf("The pointer for BuilderTypes is the same for both newBuild and testBuild: %x, expected them to be different.", fmt.Sprintf("%p", tstTpl.build.Builders))
		goto buildersEnd
	}
	if len(newBuild.Builders) != len(tstTpl.Builders) {
		t.Errorf("Expected newBuild.BuilderTypoes to have a length of %d; got %d", len(tstTpl.Builders), len(newBuild.Builders))
		goto buildersEnd
	}
	for k, _ := range tstTpl.Builders {
		_, ok := newBuild.Builders[k]
		if !ok {
			t.Errorf("Expected %s to be a builder in the copy, but it wasn't", k)
		}
	}
buildersEnd:
	if len(newBuild.PostProcessorTypes) != len(tstTpl.PostProcessorTypes) {
		t.Errorf("Expected newBuild.PostProcessorTypes to have a length of %d; got %d", len(tstTpl.PostProcessorTypes), len(newBuild.PostProcessorTypes))
		goto postProcessorTypesEnd
	}
	for i, v := range tstTpl.PostProcessorTypes {
		if v != newBuild.PostProcessorTypes[i] {
			t.Errorf("Expected PostProcessor type at index %d to be %q; got %q", i, v, newBuild.PostProcessorTypes[i])
		}
	}
postProcessorTypesEnd:
	if fmt.Sprintf("%p", newBuild.PostProcessors) == fmt.Sprintf("%p", tstTpl.build.PostProcessors) {
		t.Errorf("The pointer for PostProcessors is the same for both newBuild and testBuild: %x, expected them to be different.", fmt.Sprintf("%p", tstTpl.build.PostProcessors))
		goto postProcessorsEnd
	}
	if len(newBuild.PostProcessors) != len(tstTpl.PostProcessors) {
		t.Errorf("Expected newBuild.PostProcessors to have a length of %d; got %d", len(tstTpl.PostProcessors), len(newBuild.PostProcessors))
		goto postProcessorsEnd
	}
	for k, _ := range tstTpl.PostProcessors {
		_, ok := newBuild.PostProcessors[k]
		if !ok {
			t.Errorf("Expected %s to be a PostProcessors in the copy, but it wasn't", k)
		}
	}
postProcessorsEnd:
	if len(newBuild.ProvisionerTypes) != len(tstTpl.ProvisionerTypes) {
		t.Errorf("Expected newBuild.ProvisionerTypes to have a length of %d; got %d", len(tstTpl.ProvisionerTypes), len(newBuild.PostProcessorTypes))
		goto provisionerTypesEnd
	}
	for i, v := range tstTpl.ProvisionerTypes {
		if v != newBuild.ProvisionerTypes[i] {
			t.Errorf("Expected provisioner type at index %d to be %q; got %q", i, v, newBuild.ProvisionerTypes[i])
		}
	}
provisionerTypesEnd:
	if fmt.Sprintf("%p", newBuild.Provisioners) == fmt.Sprintf("%p", tstTpl.build.Provisioners) {
		t.Errorf("The pointer for Provisioners is the same for both newBuild and testBuild: %x, expected them to be different.", fmt.Sprintf("%p", tstTpl.build.Provisioners))
		goto provisionersEnd
	}
	if len(newBuild.Provisioners) != len(tstTpl.Provisioners) {
		t.Errorf("Expected newBuild.Provisioners types to have a length of %d; got %d", len(tstTpl.Provisioners), len(newBuild.Provisioners))
		goto provisionersEnd
	}
	for k, _ := range tstTpl.Provisioners {
		_, ok := newBuild.Provisioners[k]
		if !ok {
			t.Errorf("Expected %s to be a Provisioners in the copy, but it wasn't", k)
		}
	}
provisionersEnd:
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
	tests := []struct {
		format      string
		expectedErr string
	}{
		{"", "unsupported format"},
		{"yaml", "unsupported format"},
		{"toml", ""},
		{"json", ""},
	}
	contour.UpdateString(ConfDir, "conf")
	for i, test := range tests {
		contour.UpdateString(Format, test.format)
		d := defaults{}
		err := d.Load("../test_files")
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expepcted an error: %q, got none", i, test.expectedErr)
			continue
		}
		if MarshalJSONToString.Get(d) != MarshalJSONToString.Get(testDefaults) {
			t.Errorf("%d: expected %q, got %q", i, MarshalJSONToString.Get(testDefaults), MarshalJSONToString.Get(d))
		}
	}
}

func TestSupported(t *testing.T) {
	tests := []struct {
		format      string
		p           string
		expectedErr string
	}{
		{"", "", "unsupported format"},
		{"yaml", "", "unsupported format"},
		{"toml", "../test_files", ""},
		{"json", "../test_files", ""},
	}
	for i, test := range tests {
		contour.UpdateString(Format, test.format)
		s := supported{}
		err := s.Load(test.p)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expected an error: %q, got none", i, test.expectedErr)
			continue
		}
		if MarshalJSONToString.Get(s.Distro) != MarshalJSONToString.Get(testSupported) {
			t.Errorf("%d: expected %q, got %q", i, MarshalJSONToString.Get(testSupported), MarshalJSONToString.Get(s.Distro))
		}
	}
}

func TestBuildStuff(t *testing.T) {
	tests := []struct {
		filename    string
		format      string
		expectedErr string
	}{
		{"", "", "\"build\" not set, unable to retrieve the build file"},
		{"", "yaml", "\"build\" not set, unable to retrieve the build file"},
		{"", "toml", "\"build\" not set, unable to retrieve the build file"},
		{"", "json", "\"build\" not set, unable to retrieve the build file"},
		{"../test_files/conf/build2.yaml", "yaml", "unsupported format"},
		{"../test_files/conf/build2.toml", "toml", ""},
		{"../test_files/conf/build2.json", "json", ""},
	}
	contour.UpdateString(ConfDir, "../test_files/conf")
	for i, test := range tests {
		contour.UpdateString(Format, test.format)
		b := builds{}
		err := b.Load(test.filename)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expepcted an error: %q, got none", i, test.expectedErr)
			continue
		}
		if MarshalJSONToString.Get(b.Build) != MarshalJSONToString.Get(testBuild) {
			t.Errorf("%d: expected %q, got %q", i, MarshalJSONToString.Get(testBuild), MarshalJSONToString.Get(b.Build))
		}
	}
}

func TestBuildListStuff(t *testing.T) {
	tests := []struct {
		format      string
		expectedErr string
	}{
		{"", "unsupported format"},
		{"yaml", "unsupported format"},
		{"toml", ""},
		{"json", ""},
	}
	contour.UpdateString(ConfDir, "conf")
	for i, test := range tests {
		contour.UpdateString(Format, test.format)
		b := &buildLists{List: map[string]list{}}
		err := b.Load("../test_files")
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf("%d: expected %q, got %q", i, test.expectedErr, err)
			}
			continue
		}
		if test.expectedErr != "" {
			t.Errorf("%d: expected an error: %q, got none", i, test.expectedErr)
			continue
		}
		if MarshalJSONToString.Get(b.List) != MarshalJSONToString.Get(testBuildList) {
			t.Errorf("%d: expected %q, got %q", i, MarshalJSONToString.Get(testBuildList), MarshalJSONToString.Get(b.List))
		}
	}
}

func TestIODirInfUpdate(t *testing.T) {
	oldIODirInf := IODirInf{OutputDir: "old OutDir", SourceDir: "old SrcDir"}
	newIODirInf := IODirInf{}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.OutputDir != "old OutDir" {
		t.Errorf("Expected \"old OutDir\", got %q", oldIODirInf.OutputDir)
	}
	if oldIODirInf.SourceDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SourceDir)
	}

	oldIODirInf = IODirInf{OutputDir: "old OutDir", SourceDir: "old SrcDir"}
	newIODirInf = IODirInf{OutputDir: "new OutDir", SourceDir: "new SrcDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.OutputDir != "new OutDir/" {
		t.Errorf("Expected \"new OutDir/\", got %q", oldIODirInf.OutputDir)
	}
	if oldIODirInf.SourceDir != "new SrcDir/" {
		t.Errorf("Expected \"new SrcDir/\", got %q", oldIODirInf.SourceDir)
	}

	oldIODirInf = IODirInf{OutputDir: "old OutDir", SourceDir: "old SrcDir"}
	newIODirInf = IODirInf{OutputDir: "OutDir"}
	oldIODirInf.update(newIODirInf)
	if oldIODirInf.OutputDir != "OutDir/" {
		t.Errorf("Expected \"OutDir/\", got %q", oldIODirInf.OutputDir)
	}
	if oldIODirInf.SourceDir != "old SrcDir" {
		t.Errorf("Expected \"old SrcDir\", got %q", oldIODirInf.SourceDir)
	}
}

func TestGetAltJSONName(t *testing.T) {
	tests := []struct{
		fname string
		expected string
		isJSON bool
	}{
		{"test.json", "test.cjsn", true},
		{"test.cjsn", "test.json", true},
		{"test.toml", "", false},
	}

	for i, test := range tests {
		name, ok := getAltJSONName(test.fname)
		if ok != test.isJSON {
			t.Errorf("%d: expected %t, got %t", i, test.isJSON, ok)
		}
		if name != test.expected {
			t.Errorf("%d: expected %q, got %q", test.expected, name)
		}
	}
}
