package app

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	json "github.com/mohae/customjson"
	"github.com/mohae/utilitybelt/deepcopy"
)

// Builder constants
const (
	UnsupportedBuilder Builder = iota
	Common
	Custom
	AmazonChroot
	AmazonEBS
	AmazonInstance
	DigitalOcean
	Docker
	GoogleCompute
	Null
	OpenStack
	Parallels
	QEMU
	VirtualBoxISO
	VirtualBoxOVF
	VMWareISO
	VMWareVMX
)

// Builder is a Packer supported builder.
type Builder int

var builders = [...]string{
	"unsupported builder",
	"common",
	"custom",
	"amazon-chroot",
	"amazon-ebs",
	"amazon-instance",
	"digitalocean",
	"docker",
	"googlecompute",
	"null",
	"openstack",
	"parallels",
	"qemu",
	"virtualbox-iso",
	"virtualbox-ovf",
	"vmware-iso",
	"vmware-vmx",
}

func (b Builder) String() string { return builders[b] }

// BuilderFromString returns the builder constant for the passed string or
// unsupported. All incoming strings are normalized to lowercase.
func BuilderFromString(s string) Builder {
	s = strings.ToLower(s)
	switch s {
	case "common":
		return Common
	case "custom":
		return Custom
	case "amazon-chroot":
		return AmazonChroot
	case "amazon-ebs":
		return AmazonEBS
	case "amazon-instance":
		return AmazonInstance
	case "digitalocean":
		return DigitalOcean
	case "docker":
		return Docker
	case "googlecompute":
		return GoogleCompute
	case "null":
		return Null
	case "openstack":
		return OpenStack
	case "parallels":
		return Parallels
	case "qemu":
		return QEMU
	case "virtualbox-iso":
		return VirtualBoxISO
	case "virtualbox-ovf":
		return VirtualBoxOVF
	case "vmware-iso":
		return VMWareISO
	case "vmware-vmx":
		return VMWareVMX
	}
	return UnsupportedBuilder
}

// r.createBuilders takes a raw builder and create the appropriate Packer
// Builder
func (r *rawTemplate) createBuilders() (bldrs []interface{}, err error) {
	if r.BuilderIDs == nil || len(r.BuilderIDs) <= 0 {
		return nil, fmt.Errorf("unable to create builders: none specified")
	}
	var tmpS map[string]interface{}
	var ndx int
	bldrs = make([]interface{}, len(r.BuilderIDs))
	// Set the CommonBuilder settings. Only the builder.Settings field is used
	// for CommonBuilder as everything else is usually builder specific, even
	// if they have common names, e.g. difference between specifying memory
	// between VMWare and VirtualBox.
	//	r.updateCommonBuilder
	//
	// Generate the builders for each builder type.
	for _, ID := range r.BuilderIDs {
		bldr, ok := r.Builders[ID]
		if !ok {
			return nil, fmt.Errorf("builder configuration for %s not found", ID)
		}
		typ := BuilderFromString(bldr.Type)
		switch typ {
		case AmazonChroot:
			tmpS, err = r.createAmazonChroot(ID)
			if err != nil {
				return nil, &Error{AmazonChroot.String(), err}
			}
		case AmazonEBS:
			tmpS, err = r.createAmazonEBS(ID)
			if err != nil {
				return nil, &Error{AmazonEBS.String(), err}
			}
		case AmazonInstance:
			tmpS, err = r.createAmazonInstance(ID)
			if err != nil {
				return nil, &Error{AmazonInstance.String(), err}
			}
		case DigitalOcean:
			tmpS, err = r.createDigitalOcean(ID)
			if err != nil {
				return nil, &Error{DigitalOcean.String(), err}
			}
		case Docker:
			tmpS, err = r.createDocker(ID)
			if err != nil {
				return nil, &Error{Docker.String(), err}
			}
		case GoogleCompute:
			tmpS, err = r.createGoogleCompute(ID)
			if err != nil {
				return nil, &Error{GoogleCompute.String(), err}
			}
		case Null:
			tmpS, err = r.createNull(ID)
			if err != nil {
				return nil, &Error{Null.String(), err}
			}
		case OpenStack:
			tmpS, err = r.createOpenStack(ID)
			if err != nil {
				return nil, &Error{Null.String(), err}
			}
		//	case ParallelsISO, ParallelsPVM:
		case QEMU:
			tmpS, err = r.createQEMU(ID)
			if err != nil {
				return nil, &Error{QEMU.String(), err}
			}
		case VirtualBoxISO:
			tmpS, err = r.createVirtualBoxISO(ID)
			if err != nil {
				return nil, &Error{VirtualBoxISO.String(), err}
			}
		case VirtualBoxOVF:
			tmpS, err = r.createVirtualBoxOVF(ID)
			if err != nil {
				return nil, &Error{VirtualBoxOVF.String(), err}
			}
		case VMWareISO:
			tmpS, err = r.createVMWareISO(ID)
			if err != nil {
				return nil, &Error{VMWareISO.String(), err}
			}
		case VMWareVMX:
			tmpS, err = r.createVMWareVMX(ID)
			if err != nil {
				return nil, &Error{VMWareVMX.String(), err}
			}
		default:
			return nil, &Error{UnsupportedBuilder.String(), fmt.Errorf("%q is not supported", typ.String())}
		}
		bldrs[ndx] = tmpS
		ndx++
	}
	return bldrs, nil
}

// Go through all of the Settings and convert them to a map.  Each setting is
// parsed into its constituent parts.  The value then goes through variable
// replacement to ensure that the settings are properly resolved.
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

// createAmazonChroot creates a map of settings for Packer's amazon-chroot
// builder.  Any values that aren't supported by the amazon-ebs builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/amazon-instance.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   access_key               string
//   ami_name                 string
//   secret_key               string
//   source_ami               string
// Optional configuration options:
//   ami_description          string
//   ami_groups               array of strings
//   ami_product_codes        array of strings
//   ami_regions              array of strings
//   ami_users                array of strings
//   ami_virtualization_type  string
//   chroot_mounts            array of array of strings
//   command_wrapper          string
//   copy_files               array of strings
//   device_path              string
//   enhanced_networking      bool
//   force_deregister         bool
//   mount_options            array of strings
//   mount_path               string
//   root_volume_size         int
//   tags                     object of key/value strings
func (r *rawTemplate) createAmazonChroot(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[AmazonChroot.String()]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = AmazonChroot.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}

	} else {
		workSlice = r.Builders[ID].Settings
	}
	var k, v string
	var hasAccessKey, hasAmiName, hasSecretKey, hasSourceAmi bool
	// check for communicator first
	_, err = r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "access_key":
			settings[k] = v
			hasAccessKey = true
		case "ami_name":
			settings[k] = v
			hasAmiName = true
		case "ami_description":
			settings[k] = v
		case "ami_virtualization_type":
			settings[k] = v
		case "command_wrapper":
			settings[k] = v
		case "device_path":
			settings[k] = v
		case "enhanced_networking":
			settings[k], _ = strconv.ParseBool(v)
		case "force_deregister":
			settings[k], _ = strconv.ParseBool(v)
		case "mount_path":
			settings[k] = v
		case "root_volume_size":
			settings[k], err = strconv.Atoi(v)
			return nil, &SettingError{ID, k, v, err}
		case "secret_key":
			settings[k] = v
			hasSecretKey = true
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		}
	}
	if !hasAccessKey {
		return nil, &RequiredSettingError{ID, "access_key"}
	}
	if !hasAmiName {
		return nil, &RequiredSettingError{ID, "ami_name"}
	}
	if !hasSecretKey {
		return nil, &RequiredSettingError{ID, "secret_key"}
	}
	if !hasSourceAmi {
		return nil, &RequiredSettingError{ID, "source_ami"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		// if it's not a supported array group, log a warning and move on
		switch name {
		case "ami_groups":
		case "ami_product_codes":
		case "ami_regions":
		case "ami_users":
		case "chroot_mounts":
		case "copy_files":
		case "mount_options":
		case "tags":
		default:
			// not supported; skip
			continue
		}
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, nil
}

// createAmazonEBS creates a map of settings for Packer's amazon-ebs builder.
// Any values that aren't supported by the amazon-ebs builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/amazon-ebs.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   access_key                   string
//   ami_name                     string
//   instance_type                string
//   region                       string
//   secret_key                   string
//   source_ami                   string
//   ssh_username                 string
// Optional configuration options:
//   ami_block_device_mappings     array of block device mappings
//     delete_on_termination       bool
//     device_name                 string
//     encrypted                   bool
//     iops                        int
//     no_device                   bool
//     snapshot_id                 string
//     virtual_name                string
//     volume_type                 string
//     volume_size                 int
//   ami_description               string
//   ami_groups                    array of strings
//   ami_product_codes             array of strings
//   ami_regions                   array of strings
//   ami_users                     array of strings
//   associate_public_ip_address   bool
//   availability_zone             string
//   ebs_optimized                 bool
//   enhanced_networking           bool
//   force_deregister              bool
//   iam_instance_profile          string
//   launch_block_device_mappings  array of block device mappings
//   run_tags                      object of key/value strings
//   security_group_id             string
//   security_group_ids            array of strings
//   spot_price                    string
//   spot_price_auto_product       string
//   ssh_keypair_name              string
//   ssh_private_ip                bool
//   ssh_private_key_file          string
//   subnet_id                     string
//   tags                          object of key/value strings
//   temporary_key_pair_name       string
//   token                         string
//   user_data                     string
//   user_data_file                string
//   volume_run_tags               object of key/value strings
//   vpc_id                        string
//   windows_password_timeout      string
func (r *rawTemplate) createAmazonEBS(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = AmazonEBS.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var k, v string
	var hasAccessKey, hasAmiName, hasInstanceType, hasRegion, hasSecretKey bool
	var hasSourceAmi, hasUsername, hasCommunicator bool
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "access_key":
			settings[k] = v
			hasAccessKey = true
		case "ami_description":
			settings[k] = v
		case "ami_name":
			settings[k] = v
			hasAmiName = true
		case "associate_public_ip_address":
			settings[k], _ = strconv.ParseBool(v)
		case "availability_zone":
			settings[k] = v
		case "enhanced_networking":
			settings[k], _ = strconv.ParseBool(v)
		case "force_deregister":
			settings[k], _ = strconv.ParseBool(v)
		case "iam_instance_profile":
			settings[k] = v
		case "instance_type":
			settings[k] = v
			hasInstanceType = true
		case "region":
			settings[k] = v
			hasRegion = true
		case "secret_key":
			settings[k] = v
			hasSecretKey = true
		case "security_group_id":
			settings[k] = v
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		case "spot_price":
			settings[k] = v
		case "spot_price_auto_product":
			settings[k] = v
		case "ssh_keypair_name":
			// Only process if there's no communicator or if the communicator is SSH.
			if hasCommunicator && prefix != "ssh" {
				continue
			}
			settings[k] = v
		case "ssh_private_key_file":
			// Only process if there's no communicator or if the communicator is SSH.
			if hasCommunicator && prefix != "ssh" {
				continue
			}
			settings[k] = v
		case "ssh_username":
			// Only set if there wasn't a communicator to process.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		case "subnet_id":
			settings[k] = v
		case "temporary_key_pair_name":
			settings[k] = v
		case "token":
			settings[k] = v
		case "user_data":
			settings[k] = v
		case "user_data_file":
			src, err := r.findComponentSource(AmazonEBS.String(), v, false)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(AmazonEBS.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(AmazonEBS.String(), v)
		case "vpc_id":
			settings[k] = v
		case "windows_password_timeout":
			// Don't set if there's a non WinRM communicator.
			if hasCommunicator && prefix != "winrm" {
				continue
			}
			settings[k] = v
		}
	}
	if !hasAccessKey {
		return nil, &RequiredSettingError{ID, "access_key"}
	}
	if !hasAmiName {
		return nil, &RequiredSettingError{ID, "ami_name"}
	}
	if !hasInstanceType {
		return nil, &RequiredSettingError{ID, "instance_type"}
	}
	if !hasRegion {
		return nil, &RequiredSettingError{ID, "region"}
	}
	if !hasSecretKey {
		return nil, &RequiredSettingError{ID, "secret_key"}
	}
	if !hasSourceAmi {
		return nil, &RequiredSettingError{ID, "source_ami"}
	}
	if !hasUsername {
		// If there isn't a prefix, use ssh as that's the setting
		// that's required according to the docs.
		if prefix == "" {
			prefix = "ssh"
		}
		return nil, &RequiredSettingError{ID, prefix + "_username"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		// only process supported array stuff
		switch name {
		case "ami_block_device_mappings":
			// do ami_block_device_mappings processing
			settings[name], err = r.processAMIBlockDeviceMappings(val)
			if err != nil {
				return nil, &SettingError{ID, "ami_block_device_mappings", "", err}
			}
			continue
		case "ami_groups":
		case "ami_product_codes":
		case "ami_regions":
		case "ami_users":
		case "launch_block_device_mappings":
		case "run_tags":
		case "security_group_ids":
		case "tags":
		default:
			continue
		}
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, nil
}

// createAmazonInstance creates a map of settings for Packer's amazon-instance
// builder.  Any values that aren't supported by the amazon-ebs builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/amazon-ebs.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   access_key                    string
//   account_id                    string
//   ami_name                      string
//   instance_type                 string
//   region                        string
//   s3_bucket                     string
//   secret_key                    string
//   source_ami                    string
//   ssh_username                  string
//   x509_cert_path                string
//   x509_key_path                 string
// Optional configuration options:
//   ami_block_device_mappings     array of block device mappings
//     delete_on_termination       bool
//     device_name                 string
//     encrypted                   bool
//     iops                        int
//     no_device                   bool
//     snapshot_id                 string
//     virtual_name                string
//     volume_size                 int
//     volume_type                 string
//   ami_description               string
//   ami_groups                    array of strings
//   ami_product_codes             array of strings
//   ami_regions                   array of strings
//   ami_users                     array of strings
//   ami_virtualization_type       string
//   associate_public_ip_address   bool
//   availability_zone             string
//   bundle_destination            string
//   bundle_prefix                 string
//   bundle_upload_command         string
//   bundle_vol_command            string
//   ebs_optimized                 bool
//   enhanced_networking           bool
//   force_deregister              bool
//   iam_instance_profile          string
//   launch_block_device_mappings  array of block device mappings
//   run_tags                      object of key/value strings
//   security_group_id             string
//   security_group_ids            array of strings
//   spot_price                    string
//   spot_price_auto_product       string
//   ssh_keypair_name              string
//   ssh_private_ip                bool
//   ssh_private_key_file          string
//   subnet_id                     string
//   tags                          object of key/value strings
//   temporary_key_pair_name       string
//   user_data                     string
//   user_data_file                string
//   vpc_id                        string
//   x509_upload_path              string
//   windows_password_timeout      string
func (r *rawTemplate) createAmazonInstance(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = AmazonInstance.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var k, v string
	var hasAccessKey, hasAccountID, hasAmiName, hasInstanceType, hasRegion, hasS3Bucket bool
	var hasSecretKey, hasSourceAmi, hasUsername, hasX509CertPath, hasX509KeyPath, hasCommunicator bool
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "access_key":
			settings[k] = v
			hasAccessKey = true
		case "account_id":
			settings[k] = v
			hasAccountID = true
		case "ami_description":
			settings[k] = v
		case "ami_name":
			settings[k] = v
			hasAmiName = true
		case "ami_virtualization_type":
			settings[k] = v
		case "associate_public_ip_address":
			settings[k], _ = strconv.ParseBool(v)
		case "availability_zone":
			settings[k] = v
		case "bundle_destination":
			settings[k] = v
		case "bundle_prefix":
			settings[k] = v
		case "bundle_upload_command":
			if !strings.HasSuffix(v, ".command") {
				// The value is the command.
				settings[k] = v
				continue
			}
			// The value is a command file, load the contents of the
			// file.
			cmds, err := r.commandsFromFile(AmazonInstance.String(), v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			//
			cmd := commandFromSlice(cmds)
			if cmd == "" {
				return nil, &SettingError{ID, k, v, ErrNoCommands}
			}
			settings[k] = cmd
		case "bundle_vol_command":
			if !strings.HasSuffix(v, ".command") {
				// The value is the command.
				settings[k] = v
				continue
			}
			// The value is a command file, load the contents of the
			// file.
			cmds, err := r.commandsFromFile(AmazonInstance.String(), v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			//
			cmd := commandFromSlice(cmds)
			if cmd == "" {
				return nil, &SettingError{ID, k, v, ErrNoCommands}
			}
			settings[k] = cmd
		case "ebs_optimized":
			settings[k], _ = strconv.ParseBool(v)
		case "enhanced_networking":
			settings[k], _ = strconv.ParseBool(v)
		case "force_deregister":
			settings[k], _ = strconv.ParseBool(v)
		case "iam_instance_profile":
			settings[k] = v
		case "instance_type":
			settings[k] = v
			hasInstanceType = true
		case "region":
			settings[k] = v
			hasRegion = true
		case "s3_bucket":
			settings[k] = v
			hasS3Bucket = true
		case "secret_key":
			settings[k] = v
			hasSecretKey = true
		case "security_group_id":
			settings[k] = v
		case "spot_price":
			settings[k] = v
		case "spot_price_auto_product":
			settings[k] = v
		case "ssh_keypair_name":
			// Don't process if there's a communicator and it wasn't SSH.
			if hasCommunicator && prefix != "ssh" {
				continue
			}
			settings[k] = v
		case "ssh_private_ip":
			// Don't process if there's a communicator and it wasn't SSH.
			if hasCommunicator && prefix != "ssh" {
				continue
			}
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_private_key_file":
			// Don't process if there was a communicator.
			if hasCommunicator {
				continue
			}
			settings[k] = v
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		case "ssh_username":
			// Don't process if there was a communicator.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		case "subnet_id":
			settings[k] = v
		case "temporary_key_pair_name":
			settings[k] = v
		case "user_data":
			settings[k] = v
		case "user_data_file":
			src, err := r.findComponentSource(AmazonInstance.String(), v, false)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// src with the original value; this occurs when it is an example.
			// Nothing should be copied in this instance and it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(AmazonEBS.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(AmazonInstance.String(), v)

		case "vpc_id":
			settings[k] = v
		case "windows_password_timeout":
			// Don't process if there was a communicator and it wasn't WinRM.
			if hasCommunicator && prefix != "winrm" {
				continue
			}
			settings[k] = v
		case "x509_cert_path":
			settings[k] = v
			hasX509CertPath = true
		case "x509_key_path":
			settings[k] = v
			hasX509KeyPath = true
		case "x509_upload_path":
			settings[k] = v
		}
	}
	if !hasAccessKey {
		return nil, &RequiredSettingError{ID, "access_key"}
	}
	if !hasAccountID {
		return nil, &RequiredSettingError{ID, "account_id"}
	}
	if !hasAmiName {
		return nil, &RequiredSettingError{ID, "ami_name"}
	}
	if !hasInstanceType {
		return nil, &RequiredSettingError{ID, "instance_type"}
	}
	if !hasRegion {
		return nil, &RequiredSettingError{ID, "region"}
	}
	if !hasS3Bucket {
		return nil, &RequiredSettingError{ID, "s3_bucket"}
	}
	if !hasSecretKey {
		return nil, &RequiredSettingError{ID, "secret_key"}
	}
	if !hasSourceAmi {
		return nil, &RequiredSettingError{ID, "source_ami"}
	}
	if !hasUsername {
		// if prefix was empty, no communicator was used which means
		// ssh_username is expected.
		if prefix == "" {
			prefix = "ssh"
		}
		return nil, &RequiredSettingError{ID, prefix + "_username"}
	}
	if !hasX509CertPath {
		return nil, &RequiredSettingError{ID, "x509_cert_path"}
	}
	if !hasX509KeyPath {
		return nil, &RequiredSettingError{ID, "x509_key_path"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		// if it's not a supported array group skip
		switch name {
		case "ami_block_device_mappings":
			// do ami_block_device_mappings processing
			settings[name], err = r.processAMIBlockDeviceMappings(val)
			if err != nil {
				return nil, &SettingError{ID, "ami_block_device_mappings", "", err}
			}
			continue
		case "ami_groups":
		case "ami_product_codes":
		case "ami_regions":
		case "ami_users":
		case "launch_block_device_mappings":
		case "run_tags":
		case "security_group_ids":
		case "tags":
		default:
			continue
		}
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, nil
}

// processAMIBlockDeviceMappings handles the ami_block_device_mappings
// array for Amazon builders.  The mappings must be in the form of either
// []map[string]interface{} or [][]string.  An error will occur is the
// data is anything else.
//
// For []map[string]interface{}, the data is returned without additional
// processing.  Processing of the []map to only use valid keys may be added
// at some point in the future.
//
// For [][]string, processing will be done to convert the strings into
// key value pairs and place them in a map[string]interface{}.  Values that
// are not supported settings for ami_block_device_mappings are ignored.
// The returned interface{} only includes the supported settings.  When
// settings that are ints have invalid values specified, an error will be
// returned.
func (r *rawTemplate) processAMIBlockDeviceMappings(v interface{}) (interface{}, error) {
	if reflect.TypeOf(v) == reflect.TypeOf([]map[string]interface{}{}) {
		return v, nil
	}

	// Process the [][]string into a []map[string]interface{}
	slices, ok := v.([][]string)
	if !ok {
		return nil, errors.New("not in a supported format")
	}
	ret := make([]map[string]interface{}, len(slices))
	for i, settings := range slices {
		vals := map[string]interface{}{}
		for _, setting := range settings {
			k, v := parseVar(setting)
			switch k {
			case "delete_on_termination":
				vals[k], _ = strconv.ParseBool(v)
			case "device_name":
				vals[k] = v
			case "encrypted":
				vals[k], _ = strconv.ParseBool(v)
			case "iops":
				i, err := strconv.Atoi(v)
				if err != nil {
					return nil, fmt.Errorf("iops: %s", err)
				}
				vals[k] = i
			case "no_device":
				vals[k], _ = strconv.ParseBool(v)
			case "snapshot_id":
				vals[k] = v
			case "virtual_name":
				vals[k] = v
			case "volume_size":
				i, err := strconv.Atoi(v)
				if err != nil {
					return nil, fmt.Errorf("iops: %s", err)
				}
				vals[k] = i
			case "volume_type":
				vals[k] = v
			}
		}
		ret[i] = vals
	}
	return ret, nil
}

// createDigitalOcean creates a map of settings for Packer's digitalocean
// builder.  Any values that aren't supported by the digitalocean builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/digitalocean.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   api_token           string
//   image               string
//   region              string
//   size                string
// Optional configuration options:
//   droplet_name        string
//   private_networking  bool
//   snapshot_name       string
//   state_timeout       string
//   user_date           string
func (r *rawTemplate) createDigitalOcean(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = DigitalOcean.String()
	// If a common builder was defined, merge the settings between common and this builders.
	_, ok = r.Builders[Common.String()]
	var workSlice []string
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	_, err = r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasAPIToken, hasImage, hasRegion, hasSize bool
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "api_token":
			settings[k] = v
			hasAPIToken = true
		case "droplet_name":
			settings[k] = v
		case "image":
			settings[k] = v
			hasImage = true
		case "private_networking":
			settings[k], _ = strconv.ParseBool(v)
		case "region":
			settings[k] = v
			hasRegion = true
		case "size":
			settings[k] = v
			hasSize = true
		case "snapshot_name":
			settings[k] = v
		case "state_timeout":
			settings[k] = v
		case "user_data":
			settings[k] = v
		}
	}
	if !hasAPIToken {
		return nil, &RequiredSettingError{ID, "api_token"}
	}
	if !hasImage {
		return nil, &RequiredSettingError{ID, "image"}
	}
	if !hasRegion {
		return nil, &RequiredSettingError{ID, "region"}
	}
	if !hasSize {
		return nil, &RequiredSettingError{ID, "size"}
	}
	return settings, nil
}

// createDocker creates a map of settings for Packer's docker builder. Any
// values that aren't supported by the digitalocean builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped. For more information, refer to
// https://packer.io/docs/builders/docker.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   commit          bool
//   discard         bool
//   export_path     string
//   image           string
// Optional configuration options:
//   login           bool
//   login_email     string
//   login_username  string
//   login_password  string
//   login_server    string
//   pull            bool
//   run_command     array of strings
//   volumes         map of strings to strings
func (r *rawTemplate) createDocker(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = Docker.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	// Process the communicator settings first, if there are any.
	_, err = r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasCommit, hasDiscard, hasExportPath, hasImage, hasRunCommandArray bool
	var runCommandFile string
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "commit":
			settings[k], _ = strconv.ParseBool(v)
			hasCommit = true
		case "discard":
			settings[k], _ = strconv.ParseBool(v)
			hasDiscard = true
		case "export_path":
			settings[k] = v
			hasExportPath = true
		case "image":
			settings[k] = v
			hasImage = true
		case "login":
			settings[k], _ = strconv.ParseBool(v)
		case "login_email":
			settings[k] = v
		case "login_password":
			settings[k] = v
		case "login_username":
			settings[k] = v
		case "login_server":
			settings[k] = v
		case "pull":
			settings[k], _ = strconv.ParseBool(v)
		case "run_command":
			// if it's here, cache the value, delay processing until arrays section
			if v != "" {
				runCommandFile = v
			}
		}
	}
	if !hasCommit {
		return nil, &RequiredSettingError{ID, "commit"}
	}
	if !hasDiscard {
		return nil, &RequiredSettingError{ID, "discard"}
	}
	if !hasExportPath {
		return nil, &RequiredSettingError{ID, "export_path"}
	}
	if !hasImage {
		return nil, &RequiredSettingError{ID, "image"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "run_command":
			array := deepcopy.Iface(val)
			if array != nil {
				settings[name] = array
			}
			hasRunCommandArray = true
			continue
		case "volumes":
		default:
			continue
		}
		settings[name] = deepcopy.Iface(val)
	}
	// if there wasn't an array of run commands, check to see if they should be loaded
	// from a file
	if !hasRunCommandArray {
		if runCommandFile != "" {
			commands, err := r.commandsFromFile(Docker.String(), runCommandFile)
			if err != nil {
				return nil, &SettingError{ID, "run_command", runCommandFile, err}
			}
			if len(commands) == 0 {
				return nil, &SettingError{ID, "run_command", runCommandFile, ErrNoCommands}
			}
			settings["run_command"] = commands
		}
	}
	return settings, nil
}

// createGoogleCompute creates a map of settings for Packer's googlecompute
// builder.  Any values that aren't supported by the googlecompute builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/googlecompute.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   project_id         string
//   source_image       string
//   zone               string
// Optional configuration options:
//   account_file       string
//   address            string
//   disk_size          int
//   image_name         string
//   image_description  string
//   instance_name      string
//   machine_type       string
//   metadata           object of key/value strings
//   network            string
//   preemtipble        bool
//   state_timeout      string
//   tags               array of strings
//   use_internal_ip    bool
func (r *rawTemplate) createGoogleCompute(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = GoogleCompute.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var hasProjectID, hasSourceImage, hasZone bool
	// process communicator stuff first
	_, err = r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "account_file":
			// Account file contains account credentials: the value
			// is taken as is.
			settings[k] = v
		case "address":
			settings[k] = v
		case "disk_size":
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "image_name":
			settings[k] = v
		case "image_description":
			settings[k] = v
		case "instance_name":
			settings[k] = v
		case "machine_type":
			settings[k] = v
		case "network":
			settings[k] = v
		case "preemtible":
			settings[k], _ = strconv.ParseBool(v)
		case "project_id":
			settings[k] = v
			hasProjectID = true
		case "source_image":
			settings[k] = v
			hasSourceImage = true
		case "state_timeout":
			settings[k] = v
		case "use_internal_ip":
			settings[k], _ = strconv.ParseBool(v)
		case "zone":
			settings[k] = v
			hasZone = true
		}
	}
	if !hasProjectID {
		return nil, &RequiredSettingError{ID, "project_id"}
	}
	if !hasSourceImage {
		return nil, &RequiredSettingError{ID, "source_image"}
	}
	if !hasZone {
		return nil, &RequiredSettingError{ID, "zone"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "metadata":
		case "tags":
		default:
			continue
		}
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, nil
}

// createNull creates a map of settings for Packer's null builder. Any  values
// that aren't supported by the null builder are ignored.  Any required
// settings that don't exist result in an error and processing of the builder
// is stopped.  For more information, refer to
// https://packer.io/docs/builders/null.html
//
// Configuration options:
//   Only settings provided by communicators are supported.  See communicator
//   documentation.
//
//   communicator == none is considered invalid.
func (r *rawTemplate) createNull(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = Null.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[Null.String()].Settings
	}
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	if prefix == "" {
		// communicator == none; there must be a communicator
		return nil, fmt.Errorf("%s: %s builder requires a communicator other than \"none\"", ID, Null.String())
	}
	return settings, nil
}

// createOpenStack creates a map of settings for Packer's OpenStack builder.
// Any values that aren't supported by the QEMU builder are ignored.  Any
// required settings that doesn't exist result in an error and processing
// of the builder is stopped. For more information, refer to
// https://packer.io/docs/builders/openstack.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   flavor               string
//   image_name           string
//   source_image         string
//   username             string
//   password             string
// Optional configuration options:
//   api_key              string
//   availability_zone    string
//   config_drive         bool
//   floating_ip          string
//   floating_ip_pool     string
//   insecure             bool
//   metadata             bool
//   networks             array of strings
//   rackconnect_wait     bool
//   region               string
//   security_groups      array of strings
//   ssh_interface        string
//   tenant_id            string
//   tenant_name          string
//   use_floating_ip      bool
func (r *rawTemplate) createOpenStack(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = map[string]interface{}{}
	// Each create function is responsible for setting its own type.
	settings["type"] = OpenStack.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var hasFlavor, hasImageName, hasSourceImage, hasUsername, hasPassword, hasCommunicator bool
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		_, ok = settings[prefix+"_password"]
		if ok {
			hasPassword = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "api_key":
			settings[k] = v
		case "availability_zone":
			settings[k] = v
		case "config_drive":
			settings[k], _ = strconv.ParseBool(v)
		case "flavor":
			settings[k] = v
			hasFlavor = true
		case "floating_ip":
			settings[k] = v
		case "floating_ip_pool":
			settings[k] = v
		case "image_name":
			settings[k] = v
			hasImageName = true
		case "insecure":
			settings[k], _ = strconv.ParseBool(v)
		case "metadata":
			settings[k], _ = strconv.ParseBool(v)
		case "password":
			// skip if communicator was processed
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasPassword = true
		case "rackconnect_wait":
			settings[k] = v
			settings[k], _ = strconv.ParseBool(v)
		case "region":
			settings[k] = v
		case "ssh_interface":
			// If there's a communicator and it's not SSH skip.
			if hasCommunicator && prefix != "ssh" {
				continue
			}
			settings[k] = v
		case "source_image":
			settings[k] = v
			hasSourceImage = true
		case "tenant_id":
			settings[k] = v
		case "tenant_name":
			settings[k] = v
		case "use_floating_ip":
			settings[k], _ = strconv.ParseBool(v)
		case "username":
			// skip if communicator was processed.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		}
	}
	// flavor is required
	if !hasFlavor {
		return nil, &RequiredSettingError{ID, "flavor"}
	}
	// image_name is required
	if !hasImageName {
		return nil, &RequiredSettingError{ID, "image_name"}
	}
	// source_image is required
	if !hasSourceImage {
		return nil, &RequiredSettingError{ID, "source_image"}
	}
	// Password is required
	if !hasPassword {
		if prefix == "" {
			return nil, &RequiredSettingError{ID, "password"}
		}
		return nil, &RequiredSettingError{ID, prefix + "_password"}
	}
	// Username is required
	if !hasUsername {
		if prefix == "" {
			return nil, &RequiredSettingError{ID, "username"}
		}
		return nil, &RequiredSettingError{ID, prefix + "_username"}
	}

	// Process arrays, iso_urls is only valid if iso_url is not set
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "metadata":
		case "networks":
		case "security_groups":
		default:
			continue
		}
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, nil
}

// createQEMU creates a map of settings for Packer's QEMU builder.  Any
// values that aren't supported by the QEMU builder are ignored.  Any
// required settings that doesn't exist result in an error and processing
// of the builder is stopped. For more information, refer to
// https://packer.io/docs/builders/qemu.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   iso_checksum             string
//   iso_checksum_type        string
//   iso_url                  string
//   ssh_username             string
// Optional configuration options:
//   accelerator           string
//   boot_command          array of strings
//   boot_wait             string
//   disk_cache            string
//   disk_compression      bool
//   disk_discard          string
//   disk_image            bool
//   disk_interface        string
//   disk_size             int
//   floppy_files          array_of_strings
//   format                string
//   headless              bool
//   http_directory        string
//   http_port_max         int
//   http_port_min         int
//   iso_target_path       string
//   iso_urls              array of strings
//   net_device            string
//   output_directory      string
//   qemuargs              array of array of strings
//   qemu_binary           string
//   skip_compaction       bool
func (r *rawTemplate) createQEMU(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = map[string]interface{}{}
	// Each create function is responsible for setting its own type.
	settings["type"] = QEMU.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var bootCmdProcessed, hasChecksum, hasChecksumType, hasISOURL, hasUsername, hasCommunicator bool
	var bootCommandFile string
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "accelerator":
			settings[k] = v
		case "boot_command":
			bootCommandFile = v
		case "boot_wait":
			settings[k] = v
		case "disk_cache":
			settings[k] = v
		case "disk_compression":
			settings[k], _ = strconv.ParseBool(v)
		case "disk_discard":
			settings[k] = v
		case "disk_image":
			settings[k], _ = strconv.ParseBool(v)
		case "disk_interface":
			settings[k] = v
		case "disk_size":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "format":
			settings[k] = v
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "http_directory":
			settings[k] = v
		case "http_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "http_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "iso_checksum":
			settings[k] = v
			hasChecksum = true
		case "iso_checksum_type":
			settings[k] = v
			hasChecksumType = true
		case "iso_target_path":
			// TODO should this have path location?
			settings[k] = v
		case "iso_url":
			settings[k] = v
			hasISOURL = true
		case "net_device":
			settings[k] = v
		case "output_directory":
			settings[k] = v
		case "qemu_binary":
			settings[k] = v
		case "skip_compaction":
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_username":
			// Skip if communicator exists; this was already processed during communicator processing.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		}
	}
	// Username is required
	if !hasUsername {
		return nil, &RequiredSettingError{ID, prefix + "_username"}
	}
	// make sure http_directory is set and add to dir list
	// TODO reconcile with above
	err = r.setHTTP(QEMU.String(), settings)
	if err != nil {
		return nil, err
	}
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
		case "floppy_files":
		case "iso_urls":
			// iso_url takes precedence
			if hasISOURL {
				continue
			}
			// This is processed here because we need to know if it exists or not
			array := deepcopy.Iface(val)
			if array != nil {
				settings[name] = array
				hasISOURL = true
			}
			continue
		case "qemuargs":
		default:
			continue
		}
		// copy the array and set it
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	if !hasISOURL {
		return nil, &RequiredSettingError{ID, "iso_url"}
	}
	// If the iso info wasn't set from the Settings, get it from the distro's release
	if !hasISOURL {
		//handle iso lookup vs set in file
		switch r.Distro {
		case CentOS.String():
			settings["iso_url"] = r.releaseISO.(*centos).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*centos).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*centos).ChecksumType
		case Debian.String():
			settings["iso_url"] = r.releaseISO.(*debian).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*debian).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*debian).ChecksumType

		case Ubuntu.String():
			settings["iso_url"] = r.releaseISO.(*ubuntu).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType
		default:
			err = fmt.Errorf("%q is not a supported Distro", r.Distro)
			return nil, err
		}
		return settings, nil
	}
	if !hasChecksum {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_checksum"}
	}
	if !hasChecksumType {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_checksum_type"}
	}
	return settings, nil
}

// createVirtualBoxISO creates a map of settings for Packer's virtualbox-iso
// builder.  Any values that aren't supported by the virtualbox-iso builder are
// ignored.  Any required settings that doesn't exist result in an error and
// processing of the builder is stopped. For more information, refer to
// https://packer.io/docs/builders/virtualbox-iso.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   iso_checksum             string
//   iso_checksum_type        string
//   iso_url                  string
//   ssh_password             string
//   ssh_username             string
// Optional configuration options:
//   boot_command             array of strings
//   boot_wait                string
//   disk_size                int
//   export_opts              array of strings
//   floppy_files             array of strings
//   format                   string; "ovf" or "ova"
//   guest_additions_mode     string
//   guest_additions_path     string
//   guest_additions_sha256   string
//   guest_additions_url      string
//   guest_os_type            string; if empty, generated by rancher
//   hard_drive_interface     string
//   headless                 bool
//   http_directory           string
//   http_port_min            int
//   http_port_max            int
//   iso_interface            string
//   iso_target_path          string
//   iso_urls                 array_of_strings
//   output_directory         string
//   shutdown_command         string
//   shutdown_timeout         string
//   ssh_host_port_min        int
//   ssh_host_port_max        int
//   ssh_skip_nat_mapping     bool
//   vboxmanage               array of array of strings
//   vboxmanage_post          array of array of strings
//   virtualbox_version_file  string
//   vm_name                  string
func (r *rawTemplate) createVirtualBoxISO(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = map[string]interface{}{}
	// Each create function is responsible for setting its own type.
	settings["type"] = VirtualBoxISO.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var bootCmdProcessed, hasChecksum, hasChecksumType, hasISOURL, hasUsername, hasPassword, hasCommunicator bool
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		_, ok = settings[prefix+"_password"]
		if ok {
			hasPassword = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "boot_wait":
			settings[k] = v
		case "disk_size":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "format":
			settings[k] = v
		case "guest_additions_mode":
			settings[k] = v
		case "guest_additions_path":
			settings[k] = v
		case "guest_additions_sha256":
			settings[k] = v
		case "guest_additions_url":
			settings[k] = v
		case "guest_os_type":
			settings[k] = v
		case "hard_drive_interface":
			settings[k] = v
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "http_directory":
			settings[k] = v
		case "http_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "http_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "iso_checksum":
			settings[k] = v
			hasChecksum = true
		case "iso_checksum_type":
			settings[k] = v
			hasChecksumType = true
		case "iso_interface":
			settings[k] = v
		case "iso_target_path":
			// TODO should this have path location?
			settings[k] = v
		case "iso_url":
			settings[k] = v
			hasISOURL = true
		case "output_directory":
			settings[k] = v
		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v // the value is the command
			}
		case "shutdown_timeout":
			settings[k] = v
		case "ssh_host_port_min":
			// Skip if prefix == winrm as SSH settings don't apply to WinRM
			if prefix == "winrm" {
				continue
			}
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "ssh_host_port_max":
			// Skip if prefix == winrm as SSH settings don't apply to WinRM
			if prefix == "winrm" {
				continue
			}
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "ssh_password":
			// Skip if communicator exists; this was already processed during communicator processing.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasPassword = true
		case "ssh_username":
			// Skip if communicator exists; this was already processed during communicator processing.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		case "virtualbox_version_file":
			// TODO: should this have path resolution?
			settings[k] = v
		case "vm_name":
			settings[k] = v
		}
	}
	// Username is required
	if !hasUsername {
		return nil, &RequiredSettingError{ID, prefix + "_username"}
	}
	// Password is required
	if !hasPassword {
		return nil, &RequiredSettingError{ID, prefix + "_password"}
	}
	// make sure http_directory is set and add to dir list
	// TODO reconcile with above
	err = r.setHTTP(VirtualBoxISO.String(), settings)
	if err != nil {
		return nil, err
	}
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "export_opts":
			settings[name] = val
		case "floppy_files":
			settings[name] = val
		case "iso_urls":
			// iso_url takes precedence
			if hasISOURL {
				continue
			}
			settings[name] = val
			hasISOURL = true
		case "vboxmanage":
			settings[name] = r.createVBoxManage(val)
		case "vboxmanage_post":
			settings[name] = r.createVBoxManage(val)
		}
	}
	if !hasISOURL {
		return nil, &RequiredSettingError{ID, "iso_url"}
	}
	if r.osType == "" { // if the os type hasn't been set, the ISO info hasn't been retrieved
		err = r.ISOInfo(VirtualBoxISO, workSlice)
		if err != nil {
			return nil, err
		}
	}
	// TODO: modify to select the proper virtualbox value based on distro and arch
	/*
		// set the guest_os_type
		if tmpGuestOSType == "" {
			tmpGuestOSType = r.osType
		}
		settings["guest_os_type"] = tmpGuestOSType
	*/
	// If the iso info wasn't set from the Settings, get it from the distro's release
	if !hasISOURL {
		//handle iso lookup vs set in file
		switch r.Distro {
		case CentOS.String():
			settings["iso_url"] = r.releaseISO.(*centos).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*centos).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*centos).ChecksumType
		case Debian.String():
			settings["iso_url"] = r.releaseISO.(*debian).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*debian).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*debian).ChecksumType

		case Ubuntu.String():
			settings["iso_url"] = r.releaseISO.(*ubuntu).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType
		default:
			err = fmt.Errorf("%q is not a supported Distro", r.Distro)
			return nil, err
		}
		return settings, nil
	}
	if !hasChecksum {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_checksum"}
	}
	if !hasChecksumType {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_checksum_type"}
	}
	return settings, nil
}

// createVirtualBoxOVF creates a map of settings for Packer's virtualbox-ovf
// builder.  Any values that aren't supported by the virtualbox-ovf builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped. For more information, refer to
// https://packer.io/docs/builders/virtualbox-ovf.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   source_path              string
//   ssh_username             string
// Optional configuration options:
//   boot_command             array of strings
//   boot_wait                string
//   export_opts              array of strings
//   floppy_files             array of strings
//   format                   string
//   guest_additions_mode     string
//   guest_additions_path     string
//   guest_additions_sha256   string
//   guest_additions_url      string
//   headless                 bool
//   http_directory           string
//   http_port_min            int
//   http_port_max            int
//   import_flags             array of strings
//   import_opts              string
//   output_directory         string
//   shutdown_command         string
//   shutdown_timeout         string
//   ssh_host_port_min        int
//   ssh_host_port_max        int
//   ssh_skip_nat_mapping     bool
//   vboxmanage               array of strings
//   vboxmanage_post          array of strings
//   virtualbox_version_file  string
//   vm_name                  string
func (r *rawTemplate) createVirtualBoxOVF(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = map[string]interface{}{}
	// Each create function is responsible for setting its own type.
	settings["type"] = VirtualBoxOVF.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasSourcePath, hasUsername, bootCmdProcessed, hasCommunicator, hasWinRMCommunicator bool
	var userNameVal string
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix == "" {
		// for communicator == none or no communicator setting assume ssh_username
		// since the docs have that as required.
		// TODO: revist after communicator doc clarification
		userNameVal = "ssh_username"
	} else {
		userNameVal = prefix + "_username"
		_, ok = settings[userNameVal]
		if ok {
			hasUsername = true
		}
		hasCommunicator = true
		if prefix == "winrm" {
			hasWinRMCommunicator = true
		}
	}
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "boot_wait":
			settings[k] = v
		case "format":
			settings[k] = v
		case "guest_additions_mode":
			settings[k] = v
		case "guest_additions_path":
			settings[k] = v
		case "guest_additions_sha256":
			settings[k] = v
		case "guest_additions_url":
			settings[k] = v
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "http_directory":
			settings[k] = v
		case "http_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = &SettingError{ID, k, v, err}
				return nil, err
			}
			settings[k] = i
		case "http_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = &SettingError{ID, k, v, err}
				return nil, err
			}
			settings[k] = i
		case "import_opts":
			settings[k] = v
		case "output_directory":
			settings[k] = v
		case "shutdown_command":
			if strings.HasSuffix(v, ".command") {
				//If it ends in .command, replace it with the command from the filepath
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v
			}
		case "shutdown_timeout":
			settings[k] = v
		case "source_path":
			src, err := r.findComponentSource(VirtualBoxOVF.String(), v, true)
			if err != nil {
				return nil, err
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(VirtualBoxOVF.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(VirtualBoxOVF.String(), v)
			hasSourcePath = true
		case "ssh_host_port_min":
			if hasWinRMCommunicator {
				continue
			}
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = &SettingError{ID, k, v, err}
				return nil, err
			}
			settings[k] = i
		case "ssh_host_port_max":
			if hasWinRMCommunicator {
				continue
			}
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = &SettingError{ID, k, v, err}
				return nil, err
			}
			settings[k] = i
		case "ssh_skip_nat_mapping":
			// SSH settings don't apply to winrm
			if hasWinRMCommunicator {
				continue
			}
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_username":
			// skip if communicator exists (prefix will be empty)
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		case "virtualbox_version_file":
			settings[k] = v
		case "vm_name":
			settings[k] = v
		}
	}
	// Check to see if the required info was processed.
	if !hasUsername {
		return nil, &RequiredSettingError{ID, userNameVal}
	}
	if !hasSourcePath {
		return nil, &RequiredSettingError{ID, "source_path"}
	}

	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VirtualBoxOVF.String(), settings)
	if err != nil {
		return nil, err
	}

	// Generate Packer Variables
	// Generate builder specific section
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "export_opts":
			settings[name] = val
		case "floppy_files":
			settings[name] = val
		case "import_flags":
			settings[name] = val
		case "vboxmanage":
			settings[name] = r.createVBoxManage(val)
		case "vboxmanage_post":
			settings[name] = r.createVBoxManage(val)
		}
	}
	return settings, nil
}

// createVBoxManage creates the vboxmanage and vboxmanage_post arrays from the
// received interface.
func (r *rawTemplate) createVBoxManage(v interface{}) [][]string {
	vms := deepcopy.InterfaceToSliceOfStrings(v)
	tmp := make([][]string, len(vms))
	for i, v := range vms {
		k, vv := parseVar(v)
		// ensure that the key starts with --. A naive concatonation is done.
		if !strings.HasPrefix(k, "--") {
			k = "--" + k
		}
		vv = r.replaceVariables(vv)
		tmp[i] = make([]string, 4)
		tmp[i][0] = "modifyvm"
		tmp[i][1] = "{{.Name}}"
		tmp[i][2] = k
		tmp[i][3] = vv
	}
	return tmp
}

// createVMWareISO creates a map of settings for Packer's vmware-iso builder.
// Any values that aren't supported by the vmware-iso builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped. For more information, refer to
// https://packer.io/docs/builders/vmware-iso.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   iso_checksum            string
//   iso_checksum_type       string
//   iso_url                 string
//   ssh_username            string
// Optional configuration options
//   boot_command            array of strings
//   boot_wait               string
//   disk_additional_size    array of ints
//   disk_size               int
//   disk_type_id            string
//   floppy_files            array of strings
//   fusion_app_path         string
//   guest_os_type           string; if not set, will be generated
//   headless                bool
//   http_directory          string
//   http_port_min           int
//   http_port_max           int
//   iso_target_path         string
//   iso_urls                array of strings
//   output_directory        string
//   remote_cache_datastore  string
//   remote_cache_directory  string
//   remote_datastore        string
//   remote_host             string
//   remote_password         string
//   remote_private_key_file string
//   remote_type             string
//   remote_username         string
//   shutdown_command        string
//   shutdown_timeout        string
//   skip_compaction         bool
//   tools_upload_flavor     string
//   tools_upload_path       string
//   version                 string
//   vm_name                 string
//   vmdk_name               string
//   vmx_data                object of key/value strings
//   vmx_data_post           object of key/value strings
//   vmx_template_path       string
//   vnc_port_min            int
//   vnc_port_max            int
func (r *rawTemplate) createVMWareISO(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = VMWareISO.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var bootCmdProcessed, hasChecksum, hasChecksumType, hasISOURL, hasUsername, hasCommunicator bool
	var guestOSType string
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	for _, s := range workSlice {
		// to this builder.
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "boot_wait":
			settings[k] = v
		case "disk_size":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "disk_type_id":
			settings[k] = v
		case "fusion_app_path":
			settings[k] = v
		case "guest_os_type":
			guestOSType = v
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "http_directory":
			settings[k] = v
		case "http_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "http_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "iso_checksum":
			settings[k] = v
			hasChecksum = true
		case "iso_checksum_type":
			settings[k] = v
			hasChecksumType = true
		case "iso_target_path":
			settings[k] = v
		case "iso_url":
			settings[k] = v
			hasISOURL = true
		case "output_directory":
			settings[k] = v
		case "remote_cache_datastore":
			settings[k] = v
		case "remote_cache_directory":
			settings[k] = v
		case "remote_datastore":
			settings[k] = v
		case "remote_host":
			settings[k] = v
		case "remote_password":
			settings[k] = v
		case "remote_private_key_file":
			settings[k] = v
		case "remote_type":
			settings[k] = v
		case "remote_username":
			settings[k] = v
		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				// Assume it's the first element.
				settings[k] = commands[0]
				continue
			}
			settings[k] = v // the value is the command
		case "shutdown_timeout":
			settings[k] = v
		case "skip_compaction":
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_username":
			// Skip if communicator exists; this was already processed during communicator processing.
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		case "tools_upload_flavor":
			settings[k] = v
		case "tools_upload_path":
			settings[k] = v
		case "version":
			settings[k] = v
		case "vm_name":
			settings[k] = v
		case "vmdk_name":
			settings[k] = v
		case "vmx_template_path":
			settings[k] = v
		case "vnc_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "vnc_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		}
	}
	// Only check to see if the required ssh_username field was set. The required iso info is checked after Array processing
	if !hasUsername {
		return nil, &RequiredSettingError{ID, prefix + "_username"}
	}
	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VMWareISO.String(), settings)
	if err != nil {
		return nil, err
	}
	// Process arrays, iso_urls is only valid if iso_url is not set
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "disk_additional_size":
			var tmp []int
			// TODO it is assumed that it is a slice of strings.  Is this a good assumption?
			vals, ok := val.([]string)
			if !ok {
				return nil, &SettingError{ID, name, json.MarshalToString(val), fmt.Errorf("expected a string array")}
			}
			for _, v := range vals {
				i, err := strconv.Atoi(v)
				if err != nil {
					return nil, &SettingError{ID, name, json.MarshalToString(val), err}
				}
				tmp = append(tmp, i)
			}
			settings[name] = tmp
		case "floppy_files":
			settings[name] = val
		case "iso_urls":
			// these are only added if iso_url isn't set
			if hasISOURL {
				continue
			}
			settings[name] = val
			hasISOURL = true
		case "vmx_data":
			settings[name] = r.createVMXData(val)
		case "vmx_data_post":
			settings[name] = r.createVMXData(val)
		}
	}
	// TODO how is this affected by checksum being set in the template?
	if r.osType == "" { // if the os type hasn't been set, the ISO info hasn't been retrieved
		err = r.ISOInfo(VirtualBoxISO, workSlice)
		if err != nil {
			return nil, err
		}
	}
	// set the guest_os_type
	if guestOSType == "" {
		guestOSType = r.osType
	}
	settings["guest_os_type"] = guestOSType
	// If the iso info wasn't set from the Settings, get it from the distro's release
	if !hasISOURL {
		//handle iso lookup vs set in file
		switch r.Distro {
		case CentOS.String():
			settings["iso_url"] = r.releaseISO.(*centos).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*centos).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*centos).ChecksumType
		case Debian.String():
			settings["iso_url"] = r.releaseISO.(*debian).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*debian).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*debian).ChecksumType
		case Ubuntu.String():
			settings["iso_url"] = r.releaseISO.(*ubuntu).imageURL()
			settings["iso_checksum"] = r.releaseISO.(*ubuntu).Checksum
			settings["iso_checksum_type"] = r.releaseISO.(*ubuntu).ChecksumType
		default:
			err = fmt.Errorf("%q is not a supported Distro", r.Distro)
			return nil, err
		}
		return settings, nil
	}
	if !hasChecksum {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_checksum"}
	}
	if !hasChecksumType {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_checksum_type"}
	}
	return settings, nil
}

// createVMWareVMX creates a map of settings for Packer's vmware-vmx builder.
// Any values that aren't supported by the vmware-vmx builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/vmware-vmx.html
//
// In addition to the following options, Packer communicators are supported.
// Check the communicator docs for valid options.
//
// Required configuration options:
//   source_name              string
//   ssh_username             string
// Optional configuration options
//   boot_command             array of strings*
//   boot_wait                string
//   floppy_files             array of strings
//   fusion_app_path          string
//   headless                 bool
//   http_directory           string
//   http_port_min            int
//   http_port_max            int
//   output_directory         string
//   shutdown_command         string
//   shutdown_timeout         string
//   skip_compaction          bool
//   vm_name                  string
//   vmx_data                 object of key/value strings
//   vmx_data_post            object of key/value strings
//   vnc_port_min             int
//   vnc_port_max             int
func (r *rawTemplate) createVMWareVMX(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Builders[ID]
	if !ok {
		return nil, NewErrConfigNotFound(ID)
	}
	settings = make(map[string]interface{})
	// Each create function is responsible for setting its own type.
	settings["type"] = VMWareVMX.String()
	// Merge the settings between common and this builders.
	var workSlice []string
	_, ok = r.Builders[Common.String()]
	if ok {
		workSlice, err = mergeSettingsSlices(r.Builders[Common.String()].Settings, r.Builders[ID].Settings)
		if err != nil {
			return nil, err
		}
	} else {
		workSlice = r.Builders[ID].Settings
	}
	var hasSourcePath, hasUsername, bootCmdProcessed, hasCommunicator bool
	// check for communicator first
	prefix, err := r.processCommunicator(ID, workSlice, settings)
	if err != nil {
		return nil, err
	}
	// see if the required settings include username/password
	if prefix != "" {
		_, ok = settings[prefix+"_username"]
		if ok {
			hasUsername = true
		}
		hasCommunicator = true
	}
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "boot_command":
			// if the boot_command exists in the Settings section, it should
			// reference a file. This boot_command takes precedence over any
			// boot_command in the array defined in the Arrays section.
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				settings[k] = commands
				bootCmdProcessed = true
			}
		case "boot_wait":
			settings[k] = v
		case "fusion_app_path":
			settings[k] = v
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "http_directory":
			settings[k] = v
		case "http_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "http_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "output_directory":
			settings[k] = v
		case "shutdown_timeout":
			settings[k] = v
		case "shutdown_command":
			//If it ends in .command, replace it with the command from the filepath
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile("", v)
				if err != nil {
					return nil, &SettingError{ID, k, v, err}
				}
				if len(commands) == 0 {
					return nil, &SettingError{ID, k, v, ErrNoCommands}
				}
				// Assume it's the first element.
				settings[k] = commands[0]
			} else {
				settings[k] = v // the value is the command
			}
		case "skip_compaction":
			settings[k], _ = strconv.ParseBool(v)
		case "source_path":
			src, err := r.findComponentSource(VMWareVMX.String(), v, true)
			if err != nil {
				return nil, err
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(VMWareVMX.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(VMWareVMX.String(), v)
			hasSourcePath = true
		case "ssh_username":
			if hasCommunicator {
				continue
			}
			settings[k] = v
			hasUsername = true
		case "vm_name":
			settings[k] = v
		case "vnc_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "vnc_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		}
	}
	// Check if required fields were processed
	if !hasUsername {
		return nil, &RequiredSettingError{ID, "ssh_username"}
	}
	if !hasSourcePath {
		return nil, &RequiredSettingError{ID, "source_path"}
	}
	// make sure http_directory is set and add to dir list
	err = r.setHTTP(VMWareVMX.String(), settings)
	if err != nil {
		return nil, err
	}
	// Process arrays, iso_urls is only valid if iso_url is not set
	for name, val := range r.Builders[ID].Arrays {
		switch name {
		case "boot_command":
			if bootCmdProcessed {
				continue // if the boot command was already set, don't use this array
			}
			settings[name] = val
		case "floppy_files":
			settings[name] = val
		case "vmx_data":
			settings[name] = r.createVMXData(val)
		case "vmx_data_post":
			settings[name] = r.createVMXData(val)
		}
	}
	return settings, nil
}

func (r *rawTemplate) createVMXData(v interface{}) map[string]string {
	vms := deepcopy.InterfaceToSliceOfStrings(v)
	tmp := make(map[string]string, len(vms))
	for _, v := range vms {
		k, val := parseVar(v)
		val = r.replaceVariables(val)
		tmp[k] = val
	}
	return tmp
}

// updateBuilders updates the rawTemplate's builders with the passed new
// builder.
// Builder Update rules:
//   * If r's old builder does not have a matching builder in the new builder
//     map, new, nothing is done.
//   * If the builder exists in both r and new, the new builder updates r's
//     builder.
//   * If the new builder does not have a matching builder in r, the new
//     builder is added to r's builder map.
//
// Settings update rules:
//   * If the setting exists in r's builder but not in new, nothing is done.
//     This means that deletion of settings via not having them exist in the
//     new builder is not supported. This is to simplify overriding templates
//     in the configuration files.
//   * If the setting exists in both r's builder and new, r's builder is
//     updated with new's value.
//   * If the setting exists in new, but not r's builder, new's setting is
//     added to r's builder.
//   * To unset a setting, specify the key, without a value: `"key="`.  In most
//     situations, Rancher will interpret an key without a value as a deletion
//     of that key. There is an exception:
//   * `guest_os_type`: This is generally set at Packer Template generation
//     time by Rancher.
func (r *rawTemplate) updateBuilders(newB map[string]builder) error {
	// If there is nothing new, old equals merged.
	if len(newB) == 0 || newB == nil {
		return nil
	}
	// Convert the existing Builders to Componenter.
	var oldC = make(map[string]Componenter, len(r.Builders))
	oldC = DeepCopyMapStringBuilder(r.Builders)
	// Convert the new Builders to Componenter.
	var newC = make(map[string]Componenter, len(newB))
	newC = DeepCopyMapStringBuilder(newB)
	// Make the slice as long as the slices in both builders, odds are its shorter, but this is the worst case.
	var keys []string
	// Convert the keys to a map
	keys = mergeKeysFromComponentMaps(oldC, newC)

	// If there's a builder with the key CommonBuilder, merge them. This is a special case for builders only.
	_, ok := newB[Common.String()]
	if ok {
		r.updateCommon(newB[Common.String()])
	}
	// Copy: if the key exists in the new builder only.
	// Ignore: if the key does not exist in the new builder.
	// Merge: if the key exists in both the new and old builder.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		b, ok := r.Builders[v]
		if !ok {
			bb, _ := newB[v]
			r.Builders[v] = bb.DeepCopy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		bb, ok := newB[v]
		if !ok {
			continue
		}
		err := b.mergeSettings(bb.Settings)
		if err != nil {
			return fmt.Errorf("merge of settings failed: %s", err)
		}
		b.mergeArrays(bb.Arrays)
		r.Builders[v] = b
	}
	return nil
}

// updateCommon updates rawTemplate's common builder settings.
// Update rules:
//   * When both the existing common builder, r, and the new one, b, have the
//     same setting, b's value replaces r's; the new value replaces the
//     existing value.
//   * When the setting in b is new, it is added to r: new settings are
//     inserted into r's CommonBuilder setting list.
//   * When r has a setting that does not exist in b, nothing is done.  This
//     method does not delete any settings that already exist in r.
func (r *rawTemplate) updateCommon(newB builder) error {
	if r.Builders == nil {
		r.Builders = map[string]builder{}
	}
	// If the existing builder doesn't have a CommonBuilder section, just add it
	b, ok := r.Builders[Common.String()]
	if !ok {
		r.Builders[Common.String()] = builder{templateSection: templateSection{Type: newB.Type, Settings: newB.Settings, Arrays: newB.Arrays}}
		return nil
	}
	// Otherwise merge the two
	err := b.mergeSettings(b.Settings)
	if err != nil {
		return err
	}
	r.Builders[Common.String()] = b
	return nil
}

// setHTTP ensures that http setting is set and adds it to the dirs info so that its
// contents can be copied. If it is not set, http is assumed.
//
// The http_directory doesn't include component
func (r *rawTemplate) setHTTP(component string, m map[string]interface{}) error {
	v, ok := m["http_directory"]
	if !ok {
		v = "http"
	}
	src, err := r.findComponentSource(component, v.(string), true)
	if err != nil {
		return fmt.Errorf("setHTTP error: %s", err)
	}
	// if the source couldn't be found and an error wasn't generated, replace
	// s with the original value; this occurs when it is an example.
	// Nothing should be copied in this instancel it should not be added
	// to the copy info
	if src != "" {
		r.dirs[r.buildOutPath("", v.(string))] = src
	}
	m["http_directory"] = r.buildTemplateResourcePath("", v.(string))
	return nil
}

// DeepCopyMapStringBuilder makes a deep copy of each builder passed and
// returns the copy map[string]builder as a map[string]Componenter{}
func DeepCopyMapStringBuilder(b map[string]builder) map[string]Componenter {
	c := map[string]Componenter{}
	for k, v := range b {
		tmpB := builder{}
		tmpB = v.DeepCopy()
		c[k] = tmpB
	}
	return c
}

// commandFromSlice takes a []string and returns it as a string.  If there is
// only 1 element, that is returned as the command without any additional
// processing.  Otherwise each element of the slice is processed.
//
// Processing multi-line commands are done by trimming space characters
// (space, tabs, newlines) and joining them to form a single command string.
// The `\` character is used when a single line is split across multiple
// lines.  As such, a line without one signals the end of the command being
// processed and if there are any additional lines in the string slice being
// processed, they will be ignored.
//
// Once a line without a `\` is encountered that line is added to the
// command string and the resulting command string is returned.
func commandFromSlice(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	if len(lines) == 1 {
		return lines[0]
	}
	var cmd string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasSuffix(line, `\`) {
			cmd += line
			return cmd
		}
		cmd += strings.TrimSuffix(line, `\`)
	}
	return cmd
}
