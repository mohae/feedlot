package ranchr

import (
	_ "errors"
	_ "fmt"
	"reflect"
	"testing"
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
	{"Merge Slices-alternating alphabet", []string{"a=1", "c=2", "e=3", "g=4", "i=5"}, []string{"b=6", "d=7", "f=8", "h=9", "j=10"}, []string{"a=1", "c=2", "e=3", "g=4", "i=5", "b=6", "d=7", "f=8", "h=9", "j=10"}},
	{"Merge Slices-duplicate values", []string{"a=1", "b=2", "c=3", "d=4", "e=5", "f=6"}, []string{"c=33", "f=66", "g=7", "h=8", "i=9", "j=10"}, []string{"a=1", "b=2", "c=33", "d=4", "e=5", "f=66", "g=7", "h=8", "i=9", "j=10"}},
	{"Merge Slices-duplicates, unordered", []string{"d=1", "c=2", "x=3", "p=4", "e=5", "f=6"}, []string{"c=22", "f=66", "a=7", "x=33", "i=8", "j=9"}, []string{"d=1", "c=22", "x=33", "p=4", "e=5", "f=66", "a=7", "i=8", "j=9"}},
}

// test variable slice to map function
type varMapFromSliceTest struct {
	name     string
	sl       []string
	expected map[string]interface{}
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
		},
	},
	{
		"Create []variable From slice T2",
		[]string{"memory=2048", "ssh_port=222", "ssh_username=vagrant"},
		map[string]interface{}{
			"memory": "2048", "ssh_port": "222", "ssh_username": "vagrant",
		},
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
	/*	{
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
	*/{
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
	},
}

type getMergedValueStringTest struct {
	name     string
	old      string
	new      string
	expected string
}

var TestsGetMergedValueStringCases = []getMergedValueStringTest{
	{"test Merge Value Strings, empty new value", "old", "", "old"},
	{"test Merge Value Strings", "old", "new", "new"},
}

func TestRanchr(t *testing.T) {
	// test parsing of a string into its key:value components
	// test converging the default variables with distro variables
	for _, test := range TestsParseVarCases {
		k, v := parseVar(test.variable)
		if k != test.key || v != test.value {
			t.Error("Expected:", test.key, "Got:", k, "Expected:", test.value, "Got:", v)
		} else {
			t.Logf(test.name, "OK")
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
			} else {
				t.Logf(test.name, "OK")
			}
		}
	}

	// test merging of settings slices
	for _, test := range TestsMergeSettingsSlicesCases {
		results := mergeSettingsSlices(test.s1, test.s2)
		if results == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: Nil")
		} else {
			if !reflect.DeepEqual(test.expected, results) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", results)
			} else {
				t.Logf(test.name, "OK")
			}
		}
	}

	// test creation of variable slice
	for _, test := range TestsVarMapFromSliceCases {
		vars := varMapFromSlice(test.sl)
		if vars == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: nil")
		} else {
			if !reflect.DeepEqual(test.expected, vars) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", vars)

			} else {
				t.Logf(test.name, "OK")
			}
		}
	}

	// test retrieval of key from a variable slice (keys are embedded in the string on variable slices)
	for _, test := range TestsKeyIndexInVarSliceCases {
		i := keyIndexInVarSlice(test.key, test.sl)
		if i != test.expected {
			t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
		} else {
			t.Logf(test.name, "OK")
		}
	}

	// test merging of value strings
	for _, test := range TestsGetMergedValueStringCases {
		i := getMergedValueString(test.old, test.new)
		if i != test.expected {
			t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
		} else {
			t.Logf(test.name, "OK")
		}
	}

	// test getting variable names
	for _, test := range TestsGetVariableNameCases {
		if i, err := getVariableName(test.variable); err != nil {
			if err.Error() != test.expected {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", i)
			} else {
				t.Logf(test.name, "OK")
			}
		} else {
			t.Logf(test.name, "OK")
		}
	}

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

	mergedPP := map[string]postProcessors{}
	// test merging of postProcessors
	for _, test := range TestGetMergedPostProcessorsCases {
		mergedPP = getMergedPostProcessors(test.old, test.new)
		if mergedPP == nil {
			t.Errorf(test.name, "Expected:", test.expected, "Got: nil")
		} else {
			if !reflect.DeepEqual(test.expected, mergedPP) {
				t.Errorf(test.name, "Expected:", test.expected, "Got:", mergedPP)
			} else {
				t.Logf(test.name, "OK")
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
			} else {
				t.Logf(test.name, "OK")
			}
		}
	}
}
