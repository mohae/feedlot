// raw_template_builders.go contains all of the builder related functionality
// for rawTemplates. Any new builders should be added here.
package ranchr

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	json "github.com/mohae/customjson"
	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// r.createBuilders takes a raw builder and create the appropriate Packer
// Builders along with a slice of variables for that section builder type.
// Some Settings are in-lined instead of adding them to the variable section.
//
// At this point, all of the settings
//
// * update BuilderCommon with the ne, as this may be used by any of the Packer
// builders.
// * For each Builder in the template, create it's Packer Template version
//
func (r *rawTemplate) createBuilders() (bldrs []interface{}, vars map[string]interface{}, err error) {
	if r.BuilderTypes == nil || len(r.BuilderTypes) <= 0 {
		err = fmt.Errorf("rawTemplate.createBuilders: no builder types were configured, unable to create builders")
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	bldrs = make([]interface{}, len(r.BuilderTypes))

	// Set the BuilderCommon settings. Only the builder.Settings field is used
	// for BuilderCommon as everything else is usually builder specific, even
	// if they have common names, e.g. difference between specifying memory
	// between VMWare and VirtualBox.
	//	r.updateBuilderCommon

	// Generate the builders for each builder type.
	for _, bType := range r.BuilderTypes {
		jww.TRACE.Println(bType)

		// TODO calculate the length of the two longest Settings and VMSettings sections and make it
		// that length. That will prevent a panic should there be more than 50 options. Besides its
		// stupid, on so many levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})

		switch bType {
		case BuilderVMWareISO:
			tmpS, tmpVar, err = r.createBuilderVMWareISO()
		case BuilderVMWareVMX:
			//			tmpS, tmpVar, err = r.createBuilderVMWareOVF()

		case BuilderVirtualBoxISO:
			tmpS, tmpVar, err = r.createBuilderVirtualBoxISO()

		case BuilderVirtualBoxOVF:
			//			tmpS, tmpVar, err = r.createVirtualBoxOVF()

		default:
			err = errors.New("The requested builder, '" + bType + "', is not supported by Rancher")
			jww.ERROR.Println(err.Error())
			return nil, nil, err
		}

		bldrs[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}

	return bldrs, vars, nil
}

// Go through all of the Settings and convert them to a map. Each setting
// is parsed into its constituent parts. The value then goes through
// variable replacement to ensure that the settings are properly resolved.
func (b *builder) settingsToMap(r *rawTemplate) map[string]interface{} {
	var k, v string
	m := make(map[string]interface{})

	for _, s := range b.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		m[k] = v
	}

	return m
}

// r.createBuilderVirtualboxISO generates the settings for a virtualbox-iso builder.
func (r *rawTemplate) createBuilderVirtualBoxISO() (settings map[string]interface{}, vars []string, err error) {
	settings = make(map[string]interface{})

	// Each create function is responsible for setting its own type.
	settings["type"] = BuilderVirtualBoxISO

	// Merge the settings between common and this builders.
	mergedSlice := mergeSettingsSlices(r.Builders[BuilderCommon].Settings, r.Builders[BuilderVirtualBoxISO].Settings)

	var k, v string

	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range mergedSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			settings[k] = commands

		case "boot_wait", "export_opts", "floppy_files", "format", "guest_additions_mode",
			"guest_additions_path", "guest_additions_sha256", "guest_additions_url",
			"hard_drive_interface", "http_directory", "ssh_key_path", "ssh_password",
			"ssh_username", "ssh_wait_timeout", "vboxmanage", "vboxmanage_post",
			"virtualbox_version_file", "vm_name":
			settings[k] = v

		case "guest_os_type":
			if v == "" {
				settings[k] = v
			} else {
				settings[k] = r.osType
			}

		case "headless":
			if strings.ToLower(v) == "true" {
				settings[k] = true
			} else {
				settings[k] = false
			}

		case "iso_checksum_type":
			// First set the ISO info for the desired release, if it's not already set
			if r.osType == "" {
				err = r.ISOInfo(BuilderVirtualBoxISO, mergedSlice)
				if err != nil {
					jww.ERROR.Println(err.Error())
					return nil, nil, err
				}
			}

			switch r.Type {

			case "ubuntu":
				settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
				settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType

			case "centos":
				settings["iso_url"] = r.releaseISO.(*centOS).isoURL
				settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*centOS).ChecksumType

			case "default":
				err = errors.New("rawTemplate.createBuilderVirtualBoxISO: " + k + " is not a supported builder type")
				jww.ERROR.Println(err.Error())
				return nil, nil, err

			}

		// For the fields of int value, only set if it converts to a valid int.
		// Otherwise, throw an error
		case "disk_size", "ssh_host_port_min", "ssh_host_port_max", "ssh_port":
			// only add if its an int
			if _, err := strconv.Atoi(v); err != nil {
				return nil, nil, errors.New("rawTemplate.createBuilderVirtualBoxISO: An error occurred while trying to set " + k + "'s value, '" + v + "': " + err.Error())
			}
			settings[k] = v

		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			// Assume it's the first element.
			settings[k] = commands[0]

		}
	}

	// Generate Packer Variables
	// Generate builder specific section
	l, err := getSliceLenFromIface(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings])
	if err != nil {
		return nil, nil, err
	}

	if l > 0 {
		tmpVB := make([][]string, l)
	
		tmp := reflect.ValueOf(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings])
		jww.TRACE.Printf("%v\n", tmp)
	
		var vm_settings interface{}

		switch tmp.Type() {
		case TypeOfSliceInterfaces:
			vm_settings = deepcopy.Iface(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings]).([]interface{})
		case TypeOfSliceStrings:
			vm_settings = deepcopy.Iface(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings]).([]string)
		}		
	
		vms := deepcopy.InterfaceToSliceStrings(vm_settings)

		for i, v := range vms {
			vo := reflect.ValueOf(v)
			jww.TRACE.Printf("TTYT%v\t%v\n", vo, vo.Kind(), vo.Type())
			k, val := parseVar(vo.Interface().(string))
			val = r.replaceVariables(val)
			tmpVB[i] = make([]string, 4)
			tmpVB[i][0] = "modifyvm"
			tmpVB[i][1] = "{{.Name}}"
			tmpVB[i][2] = "--" + k
			tmpVB[i][3] = val
		}

		settings["vboxmanage"] = tmpVB
	}

	return settings, nil, nil
}

/*
vmx
*/

/*
// r.createBuilderVMWareISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVMWareISO() (settings map[string]interface{}, vars []string, err error) {
	// Generate the common Settings and their vars
	if tmpS, tmpVar, err = r.commonVMSettings(bType, r.Builders[BuilderCommon].Settings, r.Builders[bType].Settings); err != nil {
		jww.ERROR.Println(err.Error())
		return nil, nil, err
	}

	tmpS["type"] = bType

	// Generate builder specific section
	tmpvm := make(map[string]string, len(r.Builders[bType].Arrays[VMSettings]))

	for i, v = range r.Builders[bType].Arrays[VMSettings] {
		k, val = parseVar(v)
		val = r.replaceVariables(val)
		tmpvm[k] = val
		tmpS["vmx_data"] = tmpvm
	}
}
*/
// r.createBuilderVMWareISO generates the settings for a vmware-iso builder.
func (r *rawTemplate) createBuilderVMWareISO() (settings map[string]interface{}, vars []string, err error) {
	settings = make(map[string]interface{})

	// Each create function is responsible for setting its own type.
	settings["type"] = BuilderVMWareISO

	// Merge the settings between common and this builders.
	mergedSlice := mergeSettingsSlices(r.Builders[BuilderCommon].Settings, r.Builders[BuilderVMWareISO].Settings)

	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range mergedSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			settings[k] = commands

		case "boot_wait", "disk_size_id", "floppy_files", "fusion_app_path", "http_directory",
			"iso_urls", "output_directory", "remote_datastore", "remote_host", "remote_password",
			"remote_type", "remote_username", "shutdown_timeout", "ssh_host", "ssh_key_path",
			"ssh_password", "ssh_username", "ssh_wait_timeout", "tools_upload_flavor",
			"tools_upload_path", "vm_name", "vmdk_name", "vmx_data", "vmx_data_post",
			"vmx_template_path":
			settings[k] = v

		case "guest_os_type":
			if v == "" {
				settings[k] = v
			} else {
				settings[k] = r.osType
			}

		case "headless", "skip_compaction", "ssh_skip_request_pty":
			if strings.ToLower(v) == "true" {
				settings[k] = true
			} else {
				settings[k] = false
			}

		case "iso_checksum_type":
			// First set the ISO info for the desired release, if it's not already set
			if r.osType == "" {
				err = r.ISOInfo(BuilderVMWareISO, mergedSlice)
				if err != nil {
					jww.ERROR.Println(err.Error())
					return nil, nil, err
				}
			}

			switch r.Type {

			case "ubuntu":
				settings["iso_url"] = r.releaseISO.(*ubuntu).isoURL
				settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType

			case "centos":
				settings["iso_url"] = r.releaseISO.(*centOS).isoURL
				settings["iso_checksum"] = r.releaseISO.(*centOS).Checksum
				settings["iso_checksum_type"] = r.releaseISO.(*centOS).ChecksumType

			case "default":
				err = errors.New("rawTemplate.createBuilderVirtualBoxISO: " + k + " is not a supported builder type")
				jww.ERROR.Println(err.Error())
				return nil, nil, err

			}

		// For the fields of int value, only set if it converts to a valid int.
		// Otherwise, throw an error
		case "disk_size", "http_port_min", "http_port_max", "ssh_host_port_min", "ssh_host_port_max",
			"ssh_port", "vnc_port_min", "vnc_port_max":
			// only add if its an int
			if _, err := strconv.Atoi(v); err != nil {
				return nil, nil, errors.New("rawTemplate.createBuilderVirtualBoxISO: An error occurred while trying to set " + k + "'s value, '" + v + "': " + err.Error())
			}
			settings[k] = v

		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			var commands []string

			if commands, err = commandsFromFile(v); err != nil {
				jww.ERROR.Println(err.Error())
				return nil, nil, err
			}

			// Assume it's the first element.
			settings[k] = commands[0]

		}
	}

	// Generate builder specific section
	tmpVB := map[string]string{}
	vm_settings := deepcopy.InterfaceToSliceStrings(r.Builders[BuilderVirtualBoxISO].Arrays[VMSettings])
	for _, v := range vm_settings {
		k, val := parseVar(v)
		val = r.replaceVariables(val)
		tmpVB[k] = val
	}

	settings["vmx_data"] = tmpVB
	return settings, nil, nil
}

// rawTemplate.updateBuilders updates the rawTemplate's builders with the
// passed new builder.
//
// Builder Update rules:
// 	* If r's old builder does not have a matching builder in the new
// 	  builder map, new, nothing is done.
//	* If the builder exists in both r and new, the new builder updates r's
//	  builder.
//	* If the new builder does not have a matching builder in r, the new
//	  builder is added to r's builder map.
//
// Settings update rules:
//
//	* If the setting exists in r's builder but not in new, nothing is done.
//	  This means that deletion of settings via not having them exist in the
//	  new builder is not supported. This is to simplify overriding
//	  templates in the configuration files.
//	* If the setting exists in both r's builder and new, r's builder is
//	  updated with new's value.
//	* If the setting exists in new, but not r's builder, new's setting is
//	  added to r's builder.
//	* To unset a setting, specify the key, without a value:
//	      `"key="`
//	  In most situations, Rancher will interprete an key without a value as
//	  a deletion of that key. There are exceptions:
//
//	  	* `guest_os_type`: This is generally set at Packer Template
//		  generation time by Rancher.
func (r *rawTemplate) updateBuilders(new map[string]*builder) {
	fmt.Printf("Entering rawTemplate.updateBuilders with: %v\n", json.MarshalToString(new))
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return
	}

	// Convert the existing Builders to interfaces.
	var ifaceOld map[string]interface{} = make(map[string]interface{}, len(r.Builders))
	ifaceOld = DeepCopyMapStringPBuilder(r.Builders)
	//	for i, o := range r.Builders {
	//		ifaceOld[i] = o
	//	}
	// Convert the new Builders to interfaces.
	var ifaceNew map[string]interface{} = make(map[string]interface{}, len(new))
	ifaceNew = DeepCopyMapStringPBuilder(new)

	// Make the slice as long as the slices in both builders, odds are its
	// shorter, but this is the worst case.
	var keys []string

	// Convert the keys to a map
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	var vm_settings []string

	// If there's a builder with the key BuilderCommon, merge them. This is
	// a special case for builders only.
	if _, ok := new[BuilderCommon]; ok {
		r.updateBuilderCommon(new[BuilderCommon])
	}

	b := &builder{}

	// Copy: if the key exists in the new builder only.
	// Ignore: if the key does not exist in the new builder.
	// Merge: if the key exists in both the new and old builder.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		if _, ok := r.Builders[v]; !ok {
			r.Builders[v] = new[v].DeepCopy()
			continue
		}

		// If the element for this key doesn't exist, skip it.
		if _, ok := new[v]; !ok {
			continue
		}

		b = r.Builders[v].DeepCopy()
		vm_settings = deepcopy.InterfaceToSliceStrings(new[v].Arrays[VMSettings])

		// If there is anything to merge, do so
		if vm_settings != nil {
			b.Arrays[VMSettings] = vm_settings
			r.Builders[v] = b
		}

	}

	return
}

// r.updateBuilderCommonSettings updates rawTemplate's BuilderCommon settings
// Update rules:
//	* When both the existing BuilderCommon, r, and the new one, b, have the
//	  same setting, b's value replaces r's; the new setting value replaces
//        the existing.
//	* When the setting in b is new, it is added to r: new settings are
//	  inserted into r's BuilderCommon setting list.
//	* When r has a setting that does not exist in b, nothing is done. This
//	  method does not delete any settings that already exist in R.
func (r *rawTemplate) updateBuilderCommon(new *builder) {
	if r.Builders == nil {
		r.Builders = map[string]*builder{}
	}

	// If the existing builder doesn't have a BuilderCommon section, just add it
	if _, ok := r.Builders[BuilderCommon]; !ok {
		r.Builders[BuilderCommon] = &builder{templateSection: templateSection{Settings: new.Settings, Arrays: new.Arrays}}
		return
	}

	// Otherwise merge the two
	r.Builders[BuilderCommon].mergeSettings(new.Settings)

	return
}

// DeepCopyMapStringPBuilder makes a deep copy of each builder passed and
// returns the copy map[string]*builder as a map[string]interface{}
// notes:
//	P means pointer
func DeepCopyMapStringPBuilder(b map[string]*builder) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range b {
		tmpB := &builder{}
		tmpB = v.DeepCopy()
		c[k] = tmpB
	}
	return c
}
