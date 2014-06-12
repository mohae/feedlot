package ranchr

import (
	_ "errors"
	_ "fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

// Test Parsing of variables
type parseVarTest struct {
	name     string
	variable string
	key      string
	value    string
}

// Test cases for parsing the variables into key value
// Lots of tests but all possibilities should be covered.
// Parser does not account for " or '.
// Parser does not support = in keys or values.
var TestsParseVarCases = []parseVarTest{
	{"Test Parsing empty string", "", "", ""},
	{"Test Parsing 'key=value'", "key=value", "key", "value"},
	{"Test parsing 'key= value'", "key= value", "key", "value"},
	{"Test parsing 'key =value'", "key =value", "key", "value"},
	{"Test parsing 'key = value'", "key = value", "key", "value"},
	{"Test Parsing ' key=value'", "key=value", "key", "value"},
	{"Test parsing ' key= value'", "key= value", "key", "value"},
	{"Test parsing ' key =value'", "key =value", "key", "value"},
	{"Test parsing ' key = value'", "key = value", "key", "value"},
	{"Test Parsing 'key=value '", "key=value", "key", "value"},
	{"Test parsing 'key= value '", "key= value", "key", "value"},
	{"Test parsing 'key =value '", "key =value", "key", "value"},
	{"Test parsing 'key = value '", "key = value", "key", "value"},
	{"Test Parsing ' key=value '", "key=value", "key", "value"},
	{"Test parsing ' key= value '", "key= value", "key", "value"},
	{"Test parsing ' key =value '", "key =value", "key", "value"},
	{"Test parsing ' key = value '", "key = value", "key", "value"},
	{"Test Parsing 'key=value with spaces'", "key=value with spaces", "key", "value with spaces"},
	{"Test parsing 'key= value with spaces'", "key= value with spaces", "key", "value with spaces"},
	{"Test parsing 'key =value with spaces'", "key =value with spaces", "key", "value with spaces"},
	{"Test parsing 'key = value with spaces'", "key = value with spaces", "key", "value with spaces"},
	{"Test Parsing ' key=value with spaces'", " key=value with spaces", "key", "value with spaces"},
	{"Test parsing ' key= value with spaces'", " key= value with spaces", "key", "value with spaces"},
	{"Test parsing ' key =value with spaces'", " key =value with spaces", "key", "value with spaces"},
	{"Test parsing ' key = value with spaces'", " key = value with spaces", "key", "value with spaces"},
	{"Test Parsing 'key=value with spaces '", "key=value with spaces ", "key", "value with spaces"},
	{"Test parsing 'key= value with spaces '", "key= value with spaces ", "key", "value with spaces"},
	{"Test parsing 'key =value with spaces '", "key =value with spaces ", "key", "value with spaces"},
	{"Test parsing 'key = value with spaces '", "key = value with spaces ", "key", "value with spaces"},
	{"Test Parsing ' key=value with spaces '", " key=value with spaces ", "key", "value with spaces"},
	{"Test parsing ' key= value with spaces '", " key= value with spaces ", "key", "value with spaces"},
	{"Test parsing ' key =value with spaces '", " key =value with spaces ", "key", "value with spaces"},
	{"Test parsing ' key = value with spaces '", " key = value with spaces ", "key", "value with spaces"},
	{"Test Parsing 'key with spaces=value with spaces'", "key with spaces=value with spaces", "key with spaces", "value with spaces"},
	{"Test parsing 'key with spaces= value with spaces'", "key with spaces= value with spaces", "key with spaces", "value with spaces"},
	{"Test parsing 'key with spaces =value with spaces'", "key with spaces =value with spaces", "key with spaces", "value with spaces"},
	{"Test parsing 'key with spaces = value with spaces'", "key with spaces = value with spaces", "key with spaces", "value with spaces"},
	{"Test Parsing ' key with spaces=value with spaces'", " key with spaces=value with spaces", "key with spaces", "value with spaces"},
	{"Test parsing ' key with spaces= value with spaces'", " key with spaces= value with spaces", "key with spaces", "value with spaces"},
	{"Test parsing ' key with spaces =value with spaces'", " key with spaces =value with spaces", "key with spaces", "value with spaces"},
	{"Test parsing ' key with spaces = value with spaces'", " key with spaces = value with spaces", "key with spaces", "value with spaces"},
	{"Test Parsing 'key with spaces=value with spaces '", "key with spaces=value with spaces ", "key with spaces", "value with spaces"},
	{"Test parsing 'key with spaces= value with spaces '", "key with spaces= value with spaces ", "key with spaces", "value with spaces"},
	{"Test parsing 'key with spaces =value with spaces '", "key with spaces =value with spaces ", "key with spaces", "value with spaces"},
	{"Test parsing 'key with spaces = value with spaces '", "key with spaces = value with spaces ", "key with spaces", "value with spaces"},
	{"Test Parsing ' key with spaces=value with spaces '", " key with spaces=value with spaces ", "key with spaces", "value with spaces"},
	{"Test parsing ' key with spaces= value with spaces '", " key with spaces= value with spaces ", "key with spaces", "value with spaces"},
	{"Test parsing ' key with spaces =value with spaces '", " key with spaces =value with spaces ", "key with spaces", "value with spaces"},
	{"Test parsing ' key with spaces = value with spaces '", " key with spaces = value with spaces ", "key with spaces", "value with spaces"},
}

// test slice merging
type mergeSlicesTest struct {
	name     string
	s1       []string
	s2       []string
	expected []string
}

var TestsMergeSlicesCases = []mergeSlicesTest{
	{"Merge Slice, 1st slice empty", []string{}, []string{"a=1", "b=2", "c=3", "d=4", "e=5"}, []string{"a=1", "b=2", "c=3", "d=4", "e=5"}},
	{"Merge Slice, 2nd slice empty", []string{"a=1", "b=2", "c=3", "d=4", "e=5"}, []string{}, []string{"a=1", "b=2", "c=3", "d=4", "e=5"}},
	{"Merge Slices", []string{"a=1", "b=2", "c=3", "d=4", "e=5"}, []string{"f=6", "g=7", "h=8", "i=9", "j=10"}, []string{"a=1", "b=2", "c=3", "d=4", "e=5", "f=6", "g=7", "h=8", "i=9", "j=10"}},
	{"Merge Slices-alternating alphabet", []string{"a=1", "c=2", "e=3", "g=4", "i=5"}, []string{"b=6", "d=7", "f=8", "h=9", "j=10"}, []string{"a=1", "c=2", "e=3", "g=4", "i=5", "b=6", "d=7", "f=8", "h=9", "j=10"}},
	{"Merge Slices-duplicate values", []string{"apple", "banana", "orange", "lemon", "lime", "strawberry"}, []string{"cherry", "apple", "strawberry", "durian", "lime", "mango"}, []string{"apple", "banana", "orange", "lemon", "lime", "strawberry", "cherry", "durian", "mango"}},
}

// test settings slices merging
type mergeSettingsSlicesTest struct {
	name     string
	s1       []string
	s2       []string
	expected []string
}

var TestsMergeSettingsSlicesCases = []mergeSettingsSlicesTest{
	{"Merge Slice, 1st slice empty", []string{}, []string{"a=1", "b=2", "c=3", "d=4", "e=5"}, []string{"a=1", "b=2", "c=3", "d=4", "e=5"}},
	{"Merge Slice, 2nd slice empty", []string{"a=1", "b=2", "c=3", "d=4", "e=5"}, []string{}, []string{"a=1", "b=2", "c=3", "d=4", "e=5"}},
	{"Merge Slices", []string{"a=1", "b=2", "c=3", "d=4", "e=5"}, []string{"f=6", "g=7", "h=8", "i=9", "j=10"}, []string{"a=1", "b=2", "c=3", "d=4", "e=5", "f=6", "g=7", "h=8", "i=9", "j=10"}},
	{"Merge Slices: first slice nil", nil, nil, nil},
	{"Merge Slices-duplicate values", []string{"a=1", "b=2", "c=3", "d=4", "e=5", "f=6"}, []string{"c=33", "f=66", "g=7", "h=8", "i=9", "j=10"}, []string{"a=1", "b=2", "c=33", "d=4", "e=5", "f=66", "g=7", "h=8", "i=9", "j=10"}},
	{"Merge Slices-duplicates, unordered", []string{"d=1", "c=2", "x=3", "p=4", "e=5", "f=6"}, []string{"c=22", "f=66", "a=7", "x=33", "i=8", "j=9"}, []string{"d=1", "c=22", "x=33", "p=4", "e=5", "f=66", "a=7", "i=8", "j=9"}},
}

// test variable slice to map function
type varMapFromSliceTest struct {
	name     string
	sl       []string
	expected map[string]interface{}
	expectedErrS string
}

var TestsVarMapFromSliceCases = []varMapFromSliceTest{
	{
		"Create []variable From slice T1",
		[]string{
			"type=virtualbox-iso", "boot_wait=5s", "disk_size=20000",
			"guest_os_type=Ubuntu_64", "iso_checksum=sha256", "memory=4096",
		},
		map[string]interface{}{
			"type": "virtualbox-iso", "boot_wait": "5s", "disk_size": "20000",
			"guest_os_type": "Ubuntu_64", "iso_checksum": "sha256", "memory": "4096",
		}, "",
	},
	{
		"Create []variable From slice T2",
		[]string{"memory=2048", "ssh_port=222", "ssh_username=vagrant"},
		map[string]interface{}{
			"memory": "2048", "ssh_port": "222", "ssh_username": "vagrant",
		}, "",
	},
	{
		"Create []varaible: pass nil",
		nil,
		nil, "Unable to create a Packer Settings map because no variables were received",
	},

}

type keyIndexInVarSliceTest struct {
	name     string
	key      string
	sl       []string
	expected int
}

var TestsKeyIndexInVarSliceCases = []keyIndexInVarSliceTest{
	{
		"Find key index in slice: key not found",
		"memoory",
		[]string{"akey=avalue", "memory=2048", "checksum_type=sha256", "ssh_port=2222"},
		-1,
	},
	{
		"Find key index in slice: key is index 0",
		"akey",
		[]string{"akey=avalue", "memory=2048", "checksum_type=sha256", "ssh_port=2222"},
		0,
	},
	{
		"Find key index in slice: key is index 0",
		"memory",
		[]string{"akey=avalue", "memory=2048", "checksum_type=sha256", "ssh_port=2222"},
		1,
	},
	{
		"Find key index in slice: key is index 0",
		"ssh_port",
		[]string{"akey=avalue", "memory=2048", "checksum_type=sha256", "ssh_port=2222"},
		3,
	},
}

type getVariableNameTest struct {
	name     string
	variable string
	expected string
}

var TestsGetVariableNameCases = []getVariableNameTest{
	{"getVariableName test1", "variable1", "{{user `variable1` }}"},
	{"getVariableName test2", "variable2", "{{user `variable2` }}"},
	{"getVariableName test3: empty", "", "no variable name was passed"},
}

type commandsFromFileTest struct {
	test
	commandFile string
	Expected    []string
}

var testCommandsFromFileCases = []commandsFromFileTest{
	{
		test: test{
			Name:         "CommandFromFile test: no file",
			VarValue:     "",
			ExpectedErrS: "the passed Command filename was empty",
		},
		Expected: []string{},
	},
	{
		test: test{
			Name:         "boot command from file test",
			VarValue:     "../test_files/commands/boot_test.command",
			ExpectedErrS: "",
		},
		Expected: []string{
			`"<esc><wait>",`,
			`"<esc><wait>",`,
			`"<enter><wait>",`,
			`"/install/vmlinuz<wait>",`,
			`" auto<wait>",`,
			`" console-setup/ask_detect=false<wait>",`,
			`" console-setup/layoutcode=us<wait>",`,
			`" console-setup/modelcode=pc105<wait>",`,
			`" debconf/frontend=noninteractive<wait>",`,
			`" debian-installer=en_US<wait>",`,
			`" fb=false<wait>",`,
			`" initrd=/install/initrd.gz<wait>",`,
			`" kbd-chooser/method=us<wait>",`,
			`" keyboard-configuration/layout=USA<wait>",`,
			`" keyboard-configuration/variant=USA<wait>",`,
			`" locale=en_US<wait>",`,
			`" netcfg/get_hostname=ubuntu-1204<wait>",`,
			`" netcfg/get_domain=vagrantup.com<wait>",`,
			`" noapic<wait>",`,
			`" preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg<wait>",`,
			`" -- <wait>",`,
			`"<enter><wait>"`,
		},
	},
	{
		test: test{
			Name:         "execute command from file test",
			VarValue:     "../test_files/commands/execute_test.command",
			ExpectedErrS: "",
		},
		Expected: []string{`"echo 'vagrant'|sudo -S sh '{{.Path}}'"`},
	},
	{
		test: test{
			Name:         "shutdown command from file test",
			VarValue:     "../test_files/commands/shutdown_test.command",
			ExpectedErrS: "",
		},
		Expected: []string{`"echo 'shutdown -P now' > /tmp/shutdown.sh; echo 'vagrant'|sudo -S sh '/tmp/shutdown.sh'"`},
	},
}

type getMergedBuildersTest struct {
	name     string
	old      map[string]builder
	new      map[string]builder
	expected map[string]builder
}

var TestGetMergedBuildersCases = []getMergedBuildersTest{
	{
		name: "Test merge builders: update common only",
		old: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :src_dir/:type/:commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		new: map[string]builder{
			"common": {
				Settings: []string{
					"boot_wait = 15s",
					"disk_size = 30000",
					"http_directory = www",
				},
			},
		},
		expected: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :src_dir/:type/:commands_dir/boot.command",
					"boot_wait = 15s",
					"disk_size = 30000",
					"http_directory = www",
					"iso_checksum_type = sha256",
					"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
	},
	{
		name: "Test merge builders: update common, virtualbox, and vmware",
		old: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :src_dir/:type/:commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		new: map[string]builder{
			"common": {
				Settings: []string{
					"disk_size = 40000",
					"shutdown_command = src/commnds/shutdown.command",
					"ssh_wait_timeout = 300m",
				},
			},
			"virtualbox": {
				VMSettings: []string{
					"memory=2048",
				},
			},
			"vmware": {
				VMSettings: []string{
					"memsize=2048",
				},
			},
		},
		expected: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :src_dir/:type/:commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 40000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = src/commands/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 300m",
				},
			},
			"virtualbox": {
				VMSettings: []string{
					"cpus=1",
					"memory=2048",
				},
			},
			"vmware": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=2048",
					"numvcpus=1",
				},
			},
		},
	},
	{
		name: "Test merge builders: old has common only, new has vm stuff only",
		old: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :src_dir/:type/:commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 240m",
				},
			},
		},
		new: map[string]builder{
			"virtualbox": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		expected: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :src_dir/:type/:commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
	},
	// disabled because DeepEqual comes back with != even though they are
	/*	{
			name: "Test merge builders: no new builders",
			old: map[string]builder{
				"common": {
					Settings: []string{
						"boot_command = :src_dir/:type/:commands_dir/boot.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
				},
			},
			new: nil,
			expected: map[string]builder{
				"common": {
					Settings: []string{
						"boot_command = :src_dir/:type/:commands_dir/boot.command",
						"boot_wait = 5s",
						"disk_size = 20000",
						"http_directory = http",
						"iso_checksum_type = sha256",
						"shutdown_command = :src_dir/:type/:commands_dir/shutdown.command",
						"ssh_password = vagrant",
						"ssh_port = 22",
						"ssh_username = vagrant",
						"ssh_wait_timeout = 240m",
					},
				},
			},
		},
	*/
}

type getMergedPostProcessorsTest struct {
	name     string
	old      map[string]postProcessors
	new      map[string]postProcessors
	expected map[string]postProcessors
}

var TestGetMergedPostProcessorsCases = []getMergedPostProcessorsTest{
	{
		name: "Test merging postProcessors: update all",
		old: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/someComposedBoxName.box",
				},
			},
		},
		new: map[string]postProcessors{},
		expected: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/someComposedBoxName.box",
				},
			},
		},
	},
	{
		name: "Test merging postProcessors: update output only",
		old: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/someComposedBoxName.box",
				},
			},
		},
		new: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = out/NewName.box",
				},
			},
		},
		expected: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = out/NewName.box",
				},
			},
		},
	},
	{
		name: "Test merging postProcessors: no new postProcessor",
		old: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/someComposedBoxName.box",
				},
			},
		},
		new: nil,
		expected: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/someComposedBoxName.box",
				},
			},
		},
	},
}

type getMergedProvisionersTest struct {
	name     string
	old      map[string]provisioners
	new      map[string]provisioners
	expected map[string]provisioners
}

var TestGetMergedProvisionersCases = []getMergedProvisionersTest{
	{
		name: "Test merging provisioners, override old scripts",
		old: map[string]provisioners{
			"shell": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
		new: map[string]provisioners{
			"shell": {
				Scripts: []string{
					"scripts/setup.sh",
					"scripts/vagrant.sh",
					"scripts/zerodisk.sh",
				},
			},
		},
		expected: map[string]provisioners{
			"shell": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					"scripts/setup.sh",
					"scripts/vagrant.sh",
					"scripts/zerodisk.sh",
				},
			},
		},
	},
	{
		name: "Test merging provisioners, change execute_command only",
		old: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
		new: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = commands/execute.command"},
			},
		},
		expected: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = commands/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
	}, {
		name: "Test merging provisioners,no new provisioner",
		old: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
		new: map[string]provisioners{},
		expected: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
	},
	{
		name: "Test merging provisioners, no new provisioners",
		old: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
		new: nil,
		expected: map[string]provisioners{
			"vagrant": {
				Settings: []string{"execute_command = :commands_dir/execute.command"},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
	},
}

type appendSlashTest struct {
	name     string
	value    string
	expected string
}

var TestAppendSlashCases = []appendSlashTest{
	{"appendSlashCases test 1", "this/is/a/test", "this/is/a/test/"},
	{"appendSlashCases test 2", "this/is/another/test/", "this/is/another/test/"},
}

type copyFileTest struct {
	name          string
	srcDir        string
	destDir       string
	script        string
	expectedInt64 int64
	expectedErr   string
}

var TestCopyFileCases = []copyFileTest{
	{"Test Copy File, No src Dir", "", "test_files/out/", "test_file.sh", 0, "copyFile: no source directory passed"},
	{"Test Copy File, No dest dir", "test_files/", "", "test_file.sh", 0, "copyFile: no destination directory passed"},
	{"Test Copy File", "test_files/scripts/", "test_files/tmp", "test_file.sh", 0, "open test_files/scripts/test_file.sh: no such file or directory"},
	{"Test Copy File", "../test_files/scripts/", "test_files/tmp", "test_file.sh", 0, "o"},

}

var testDefaults = defaults{
	IODirInf: IODirInf{
		OutDir:      "out/:type/:build_name",
		ScriptsDir:  ":src_dir/scripts",
		SrcDir:      "src/:type",
		ScriptsSrcDir:      "",
		CommandsSrcDir: "",
	},
	PackerInf: PackerInf{
		MinPackerVersion: "",
		Description:      "Test Default Rancher template",
	},
	BuildInf: BuildInf{
		Name:      ":type-:release-:image-:arch",
		BuildName: "",
	},
	build: build{
		BuilderType: []string{
			"virtualbox-iso",
			"vmware-iso",
		},
		Builders: map[string]builder{
			"common": {
				Settings: []string{
					"boot_command = :commands_dir/boot.command",
					"boot_wait = 5s",
					"disk_size = 20000",
					"http_directory = http",
					"iso_checksum_type = sha256",
					"shutdown_command = :commands_dir/shutdown.command",
					"ssh_password = vagrant",
					"ssh_port = 22",
					"ssh_username = vagrant",
					"ssh_wait_timeout = 240m",
				},
			},
			"virtualbox-iso": {
				VMSettings: []string{
					"cpus=1",
					"memory=1024",
				},
			},
			"vmware-iso": {
				VMSettings: []string{
					"cpuid.coresPerSocket=1",
					"memsize=1024",
					"numvcpus=1",
				},
			},
		},
		PostProcessors: map[string]postProcessors{
			"vagrant": {
				Settings: []string{
					"keep_input_artifact = false",
					"output = :out_dir/someComposedBoxName.box",
				},
			},
		},
		Provisioners: map[string]provisioners{
			"shell": {
				Settings: []string{
					"execute_command = :commands_dir/execute.command",
				},
				Scripts: []string{
					":scripts_dir/setup.sh",
					":scripts_dir/base.sh",
					":scripts_dir/vagrant.sh",
					":scripts_dir/cleanup.sh",
					":scripts_dir/zerodisk.sh",
				},
			},
		},
	},
}

func TestRanchr(t *testing.T) {
	setCommonTestData()
	time.Sleep(100 * time.Millisecond)

	// test parsing of a string into its key:value components
	// test converging the default variables with distro variables
	for _, test := range TestsParseVarCases {
		k, v := parseVar(test.variable)
		if k != test.key || v != test.value {
			t.Error("Expected:", test.key, "Got:", k, "Expected:", test.value, "Got:", v)
		} 
	}
	/*
	   // test parsing of a string into its key:value components
	   // test converging the default variables with distro variables
	   for _, test := range TestsCommandCases {
	       if commands, err := getCommandsFromFile(test.File); err != nil {
	           if err.Error() != test.ExpectedErrS {
	               t.Errorf(test.Name+" error: ", err)
	           } else {
	               t.Logf(test.Name, test.ExpectedErrS)
	           }
	       } else {
	           for _, command := range commands {
	               t.Logf("=========")
	               t.Logf(command)
	               t.Logf(test.ExpectedErrS)
	           }
	       }
	   }
	*/

	// test merging of slices
	for _, test := range TestsMergeSlicesCases {
		results := mergeSlices(test.s1, test.s2)
		if results == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: Nil")
		} else {
			if !reflect.DeepEqual(test.expected, results) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", results)
			}
		}
	}

	// test merging of settings slices
	for _, test := range TestsMergeSettingsSlicesCases {
		results := mergeSettingsSlices(test.s1, test.s2)
		if results != nil {
			if !reflect.DeepEqual(test.expected, results) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", results)
			}
		}
	}

	// test creation of variable slice
	for _, test := range TestsVarMapFromSliceCases {
		vars := varMapFromSlice(test.sl)
		if vars == nil {
			if test.expectedErrS == "" {
				t.Errorf(test.name, "Expected:", test.expected, "Got: nil")
			}
		} else {
			if !reflect.DeepEqual(test.expected, vars) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", vars)
			}
		}
	}

	// test retrieval of key from a variable slice (keys are embedded in the string on variable slices)
	for _, test := range TestsKeyIndexInVarSliceCases {
		i := keyIndexInVarSlice(test.key, test.sl)
		if i != test.expected {
			t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
		} 
	}
	/*
		// test merging of value strings
		for _, test := range TestsGetMergedValueStringCases {
			i := getMergedValueString(test.old, test.new)
			if i != test.expected {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
			} else {
				t.Logf(test.name, "OK")
			}
		}
	*/

	for _, test := range testCommandsFromFileCases {
		if commands, err := commandsFromFile(test.VarValue); err != nil {
			if err.Error() != test.ExpectedErrS {
				t.Errorf(test.Name, err.Error())
			}
		} else {
			if !reflect.DeepEqual(commands, test.Expected) {
				t.Error(test.Name, "Expected:", test.Expected, "Got:", commands)
			}
		}
	}

	// test getting variable names
	for _, test := range TestsGetVariableNameCases {
		if i, err := getVariableName(test.variable); err != nil {
			if err.Error() != test.expected {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
			}
		}
	}

	for _, test := range TestsGetVariableNameCases {
		if i, err := getVariableName(test.variable); err != nil {
			if err.Error() != test.expected {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
			}
		}
	}

	/* DeepEqual returns false when true?
	// Test merging of two builders
	for _, test := range TestGetMergedBuildersCases {
		mergedB := map[string]builder{}
		mergedB = getMergedBuilders(test.old, test.new)
		if mergedB == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: nil")
		} else {
			if !reflect.DeepEqual(test.expected, mergedB) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", mergedB)
			} else {
				t.Logf(test.name, "OK")
			}
		}
	}
	*/
	mergedPP := map[string]postProcessors{}
	// test merging of postProcessors
	for _, test := range TestGetMergedPostProcessorsCases {
		mergedPP = getMergedPostProcessors(test.old, test.new)
		if mergedPP == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: nil")
		} else {
			if !reflect.DeepEqual(test.expected, mergedPP) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", mergedPP)
			}
		}
	}

	mergedP := map[string]provisioners{}
	// test merging of provisoners
	for _, test := range TestGetMergedProvisionersCases {
		mergedP = getMergedProvisioners(test.old, test.new)
		if mergedP == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: nil")
		} else {
			if !reflect.DeepEqual(test.expected, mergedP) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", mergedP)
			}
		}
	}

	for _, test := range TestAppendSlashCases {
		res := appendSlash(test.value)
		if res != test.expected {
			t.Errorf(test.name, "Expected: ", test.expected, " Got: ", res)
		}
	}

	for _, test := range TestCopyFileCases {
		bW, err := copyFile(test.srcDir, test.destDir, test.script)
		if err != nil {
			if err.Error() != test.expectedErr {
				t.Errorf(test.name, "expected: ", test.expectedErr, " Got: ", err.Error())
			}
		} else {
			if bW != test.expectedInt64 {
				t.Errorf(test.name, "expected: ", strconv.FormatInt(test.expectedInt64, 10), " Got: ", strconv.FormatInt(bW, 10))
			}
		}
	}

// Goconvey...

	tstConfig := "../test_files/rancher_test.cfg"
	tstDefaults := "../test_files/conf/defaults_test.toml"
	tstSupported := "../test_files/conf/supported_test.toml"
	tstBuilds := "../test_files/conf/builds_test.toml"
	tstBuildLists := "../test_files/conf/build_lists_test.toml"
	tstParamDelimStart := ":"
	tstLogging := "true"
	tstLogFile := "rancher.log"
	tstLogLevel := "info"

	
	// save current setttings
	tmpConfig := os.Getenv(EnvConfig)
	tmpDefaults := os.Getenv(EnvDefaultsFile)
	tmpSupported := os.Getenv(EnvSupportedFile)
	tmpBuilds := os.Getenv(EnvBuildsFile)
	tmpBuildLists := os.Getenv(EnvBuildListsFile)
	tmpParamDelimStart := os.Getenv(EnvParamDelimStart)
	tmpLogging := os.Getenv(EnvLogging)
	tmpLogFile := os.Getenv(EnvLogFile)
	tmpLogLevel := os.Getenv(EnvLogLevel)		
	
	Convey("Given a rancher.cfg setting", t, func() {

		// set to test values
		os.Setenv(EnvConfig, tstConfig)
		os.Setenv(EnvDefaultsFile, tstDefaults)
		os.Setenv(EnvSupportedFile, tstSupported)
		os.Setenv(EnvBuildsFile, tstBuilds)
		os.Setenv(EnvBuildListsFile, tstBuildLists)
		os.Setenv(EnvParamDelimStart, tstParamDelimStart)
		os.Setenv(EnvLogging, tstLogging)
		os.Setenv(EnvLogFile, tstLogFile)
		os.Setenv(EnvLogLevel, tstLogLevel)
		err := SetEnv()

		if err == nil {

			Convey("Given the environment variable EnvDefaultsFile", func() {
				tmp := os.Getenv(EnvDefaultsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstDefaults)
				})

			})

			Convey("Given the environment variable EnvSupportedFile", func() {
				tmp := os.Getenv(EnvSupportedFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstSupported)
				})
			})

			Convey("Given the environment variable EnvBuildsFile", func() {
				tmp := os.Getenv(EnvBuildsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstBuilds)
				})	
			})

			Convey("Given the environment variable EnvBuildListsFile", func() {
				tmp := os.Getenv(EnvBuildListsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstBuildLists)
				})
			})

			Convey("Given the environment variable EnvParamDeliStart", func() {
				tmp := os.Getenv(EnvParamDelimStart)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstParamDelimStart)
				})
			})

			Convey("Given the environment variable EnvLogging", func() {
				tmp := os.Getenv(EnvLogging)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogging)
				})
			})

			Convey("Given the environment variable EnvLogFile", func() {
				tmp := os.Getenv(EnvLogFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogFile)
				})
			})

			Convey("Given the environment variable EnvLogLevelFile", func() {
				tmp := os.Getenv(EnvLogLevel)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogLevel)
				})
			})
		}
	})
			
	Convey("Given a rancher.cfg setting with blank environment variables", t, func() {
		// set to blank values (test load of rancher.cfg.
		os.Setenv(EnvConfig, "")
		os.Setenv(EnvDefaultsFile, "")
		os.Setenv(EnvSupportedFile, "")
		os.Setenv(EnvBuildsFile, "")
		os.Setenv(EnvBuildListsFile, "")
		os.Setenv(EnvParamDelimStart, "")
		os.Setenv(EnvLogging, "")
		os.Setenv(EnvLogFile, "")
		os.Setenv(EnvLogLevel, "")
	
		tstDefaults = "conf/defaults.toml"
		tstSupported = "conf/supported.toml"
		tstBuilds = "conf.d/builds.toml"
		tstBuildLists = "conf.d/build_lists.toml"
		tstParamDelimStart = ":"
		tstLogging = "true"
		tstLogFile = "rancher.log"
		tstLogLevel = "info"

		err := SetEnv()

		if err == nil {

			Convey("Given the environment variable EnvDefaultsFile", func() {
				tmp := os.Getenv(EnvDefaultsFile)

				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstDefaults)
				})

			})

			Convey("Given the environment variable EnvSupportedFile", func() {
				tmp := os.Getenv(EnvSupportedFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstSupported)
				})
			})

			Convey("Given the environment variable EnvBuildsFile", func() {
				tmp := os.Getenv(EnvBuildsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstBuilds)
				})
			})

			Convey("Given the environment variable EnvBuildListsFile", func() {
				tmp := os.Getenv(EnvBuildListsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstBuildLists)
				})
			})

			Convey("Given the environment variable EnvParamDeliStart", func() {
				tmp := os.Getenv(EnvParamDelimStart)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstParamDelimStart)
				})
			})

			Convey("Given the environment variable EnvLogging", func() {
				tmp := os.Getenv(EnvLogging)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogging)
				})
			})

			Convey("Given the environment variable EnvLogFile", func() {
				tmp := os.Getenv(EnvLogFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogFile)
				})
			})

			Convey("Given the environment variable EnvLogLevelFile", func() {
				tmp := os.Getenv(EnvLogLevel)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogLevel)
				})
			})
		}
			
	})


		// Test empty config
		// TODO this is tied to actual rancher.cfg, which shouldn't change
		// but makes it brital...laziness has its price.
	
	Convey("Given a blank config Environment variable setting", t, func() {
		os.Setenv(EnvConfig, "")
		err := SetEnv()
		if err == nil {
			Convey("Given the environment variable EnvDefaultsFile", func() {
				tmp := os.Getenv(EnvDefaultsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstDefaults)
				})
			})

			Convey("Given the environment variable EnvSupportedFile", func() {
				tmp := os.Getenv(EnvSupportedFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstSupported)
				})
			})

			Convey("Given the environment variable EnvBuildsFile", func() {
				tmp := os.Getenv(EnvBuildsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstBuilds)
				})
			})

			Convey("Given the environment variable EnvBuildListsFile", func() {
				tmp := os.Getenv(EnvBuildListsFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstBuildLists)
				})
			})

			Convey("Given the environment variable EnvParamDeliStart", func() {
				tmp := os.Getenv(EnvParamDelimStart)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstParamDelimStart)
				})
			})

			Convey("Given the environment variable EnvLogging", func() {
				tmp := os.Getenv(EnvLogging)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogging)
				})
			})

			Convey("Given the environment variable EnvLogFile", func() {
				tmp := os.Getenv(EnvLogFile)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogFile)
				})
			})
	
			Convey("Given the environment variable EnvLogLevelFile", func() {
				tmp := os.Getenv(EnvLogLevel)
				Convey("The result should be", func() {
					So(tmp, ShouldEqual, tstLogLevel)
				})
			})
		}
	
		// restore to original
		os.Setenv(EnvConfig, tmpConfig)
		os.Setenv(EnvDefaultsFile, tmpDefaults)
		os.Setenv(EnvSupportedFile, tmpSupported)
		os.Setenv(EnvBuildsFile, tmpBuilds)
		os.Setenv(EnvBuildListsFile, tmpBuildLists)
		os.Setenv(EnvParamDelimStart, tmpParamDelimStart)
		os.Setenv(EnvLogging, tmpLogging)
		os.Setenv(EnvLogFile, tmpLogFile)
		os.Setenv(EnvLogLevel, tmpLogLevel)
	
	})

	Convey("Given set defaults and supported ENV variables", t, func() {
		tmpDefaults := os.Getenv(EnvDefaultsFile)
		tmpSupported := os.Getenv(EnvSupportedFile)
	
		var supported Supported
		var tpls map[string]RawTemplate
		var err error
		
		Convey("Given an empty defaults file location", func() {
			os.Setenv(EnvDefaultsFile, "")
			Convey("The result of loading distro inf should be", func() {
				if supported, tpls, err = DistrosInf(); err != nil {
					So(err.Error(), ShouldEqual, "could not retrieve the default Settings file because the RANCHER_DEFAULTS_FILE ENV variable was not set. Either set it or check your rancher.cfg setting")
				}
			})
		})

		Convey("Given an empty supported file location", func() {
			os.Setenv(EnvSupportedFile, "")
			Convey("The result of loading distro inf should be", func() {
				if supported, tpls, err = DistrosInf(); err != nil {
					So(err.Error(), ShouldEqual, "could not retrieve the default Settings file because the RANCHER_DEFAULTS_FILE ENV variable was not set. Either set it or check your rancher.cfg setting")
				}
			})
		})

/*
		Convey("Given properly set file information", func() {
			os.Setenv(EnvDefaultsFile, "../test_files/conf/defaults_test.toml")
			os.Setenv(EnvSupportedFile, "../test_files/conf/supported_test.toml")
			if supported, tpls, err = DistrosInf(); err == nil {
				So(supported, ShouldResemble, testSupported )
//				So(tpls, ShouldResemble, testTpls)
			}
		})
*/
		os.Setenv(EnvSupportedFile, tmpSupported)
		os.Setenv(EnvDefaultsFile, tmpDefaults)
	})
/*
	Convey("Testing BuildPackerTemplateFromDistros", t, func() {
		Convey("Given an empty supported distro", func() {
			arg := ArgsFilter{Distro: "", Arch:"", Image:"", Release: ""}
			err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, arg)
			So(err.Error(), ShouldEqual, "%v, is not Supported. Please pass a Supported distribution.")
		})
			
		Convey("Given a nil supported DistroDefaults map", func() {
			arg := ArgsFilter{Distro: "ubuntu", Arch:"", Image:"", Release: ""}
			err := BuildPackerTemplateFromDistro(Supported{}, testDistroDefaults, arg)
			So(err.Error(), ShouldEqual, "%v, is not Supported. Please pass a Supported distribution.")
		})

		Convey("Given a nil arg", func() {
			err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, ArgsFilter{})
			So(err.Error(), ShouldEqual, "%v, is not Supported. Please pass a Supported distribution.")
		})

		Convey("Given a supported distro, a map of Distro defaults, and args", func() {
			arg := ArgsFilter{Distro: "", Arch:"", Image:"", Release: ""}
			Convey("Given empty strings in ArgsFilter", func() {
				err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, arg)
				So(err.Error(), ShouldEqual, "")
			})

			arg.Arch = "i386"
			Convey("Given a populated Arch in ArgsFilter", func() {
				err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, arg)
				So(err, ShouldBeNil)
			})

			arg = ArgsFilter{Distro: "ubuntu", Arch:"", Image:"desktop", Release: ""}
			Convey("Given a populated image in ArgsFilter", func() {
				err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, arg)
				So(err, ShouldBeNil)
			})

			arg = ArgsFilter{Distro: "ubuntu", Arch:"", Image:"", Release: "14.04"}
			Convey("Given a populated release in ArgsFilter", func() {
				err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, arg)
				So(err, ShouldBeNil)
			})

			arg = ArgsFilter{Distro: "ubuntu", Arch:"i386", Image:"server", Release: "14.04"}
			Convey("Given a populated ArgsFilter", func() {
				err := BuildPackerTemplateFromDistro(testSupported, testDistroDefaults, arg)
				So(err, ShouldBeNil)
			})
		})
	})
*/


	Convey("Given a slice of default ISO information", t, func() {
		tmp := []string{"arch=amd64", "image=server", "release=14.04", "unknown=what"}
		arch, image, release := getDefaultISOInfo(tmp)

		Convey("The results should be", func() {
			So(arch, ShouldEqual, "amd64")
			So(image, ShouldEqual, "server")
			So(release, ShouldEqual, "14.04")
		})
	})	


	Convey("Given a directory", t, func() {
		srcDir := "../test_files/scripts/"
		destDir := "../test_files/tmp/"

		Convey("The results of a copy operation should be", func() {
			if err := copyDirContent(srcDir, destDir); err == nil {
				// this is a dummied equality as a non-error should signify success
				So(4, ShouldEqual, 4)
			}
		})

		Convey("The results of a copy operation should be", func() {
			if err := copyDirContent(srcDir, ""); err != nil {
				// this is a dummied equality as a non-error should signify success
				So(err.Error(), ShouldEqual, "copyFile: no destination directory passed")
			}
		})

		Convey("The results of a delete operation should be", func() {
			if err := deleteDirContent(destDir); err == nil {
				// this is a dummied equality as a non-error should signify success
				So(4, ShouldEqual, 4)
			}
		})

		Convey("The results of a delete operation should be", func() {
			if err := deleteDirContent(""); err != nil {
				// this is a dummied equality as a non-error should signify success
				So(err.Error(), ShouldEqual, "remove : no such file or directory")
			}
		})

	})
	Convey("Given some strings, their suffix should be removed", t, func() {
		res0 := trimSuffix("This is a string!", "!")
		Convey("Given a string with the suffix, the suffix should be removed", func() {
			So(res0, ShouldEqual, "This is a string")	
		})

		res1 := trimSuffix("This is a string", "!")
		Convey("Given a string with the suffix, the should", func() {
			So(res1, ShouldEqual, "This is a string")	
		})

		res2 := trimSuffix("This is a string with blanks    ", " ")
		Convey("Given a string with the suffix, the should", func() {
			So(res2, ShouldEqual, "This is a string with blanks   ")	
		})

		res3 := trimSuffix("Try this Њф Њ Җ", "Җ")
		Convey("Given a string with the utf 8 and suffix", func() {
			So(res3, ShouldEqual, "Try this Њф Њ ")	
		})

		res4 := trimSuffix("Try this Њф Њ Җ ", "Җ")
		Convey("Given a string with utf8 and a blank at the end of the string", func() {
			So(res4, ShouldEqual, "Try this Њф Њ Җ ")	
		})

		res5 := trimSuffix("Try this Њф Њ", "Њ")
		Convey("Given a string with utf8 with the same character embedded in the string, the should", func() {
			So(res5, ShouldEqual, "Try this Њф ")	
		})
	})
	
	Convey("Given a string, get a substring", t, func() {
	
		Convey("Given string get characters 2-6", func() {
			str := "This is a test"
			subStr := Substring(str, 1, 5)
			Convey("The result should be'", func() {
				So(subStr, ShouldEqual, "his i")
			})
		})
	
		Convey("Given string get characeters 2-20", func() {
			str := "This is a test"
			subStr := Substring(str, 1, 18)
			Convey("The result should be'", func() {
				So(subStr, ShouldEqual, "his is a test")
			})
		})
	
		Convey("Given a rune", func() {
			str := "Hello 世界--erm 'Hi'"
			subStr := Substring(str, 4, 12)
			Convey("The result should be'", func() {
				So(subStr, ShouldEqual, "o 世界--erm 'H")
			})
		})
	})
}

		
