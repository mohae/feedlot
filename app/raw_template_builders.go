package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
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
	Openstack
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
		return Openstack
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
		//	case Openstack:
		//	case ParallelsISO, ParallelsPVM:
		//	case QEMU:
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
		case "secret_key":
			settings[k] = v
			hasSecretKey = true
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		case "ami_description", "ami_virtualization_type", "command_wrapper",
			"device_path", "mount_path":
			settings[k] = v
		case "enhanced_networking", "force_deregister":
			settings[k], _ = strconv.ParseBool(v)
		case "root_volume_size":
			settings[k], err = strconv.Atoi(v)
			return nil, &SettingError{ID, k, v, err}
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
//   ami_description               string
//   ami_groups                    array of strings
//   ami_product_codes             array of strings
//   ami_regions                   array of strings
//   ami_users                     array of strings
//   associate_public_ip_address   bool
//   availability_zone             string
//   enhanced_networking           bool
//   delete_on_termination         bool
//   device_name                   string
//   encrypted                     bool
//   force_deregister              bool
//   iam_instance_profile          string
//   iops                          bool
//   launch_block_device_mappings  array of block device mappings
//   no_device                     bool
//   run_tags                      object of key/value strings
//   security_group_id             string
//   security_group_ids            array of strings
//   snapshot_id                   string
//   spot_price                    string
//   spot_price_auto_product       string
//   ssh_keypair_name              string
//   ssh_port                      int
//   ssh_private_ip                bool
//   ssh_private_key_file          string
//   subnet_id                     string
//   tags                          object of key/value strings
//   temporary_key_pair_name       string
//   token                         string
//   user_data                     string
//   user_data_file                string
//   virtual_name                  string
//   volume_size                   int
//   volume_type                   int
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
	var hasAccessKey, hasAmiName, hasInstanceType, hasRegion, hasSecretKey, hasSourceAmi, hasSSHUsername bool
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
		case "instance_type":
			settings[k] = v
			hasInstanceType = true
		case "region":
			settings[k] = v
			hasRegion = true
		case "secret_key":
			settings[k] = v
			hasSecretKey = true
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "ami_description", "availability_zone", "device_name",
			"iam_instance_profile", "security_group_id", "snapshot_id",
			"spot_price", "spot_price_auto_product", "ssh_keypair_name",
			"ssh_private_key_file", "subnet_id", "temporary_key_pair_name",
			"token", "user_data", "virtual_name",
			"vpc_id", "windows_password_timeout":
			settings[k] = v
		case "ssh_port", "volume_size", "volume_type":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "user_data_file":
			src, err := r.findComponentSource(AmazonEBS.String(), v, false)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			jww.ERROR.Printf("EBS user_data_file: %v", src)
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(AmazonEBS.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(AmazonEBS.String(), v)
		case "associate_public_ip_address", "enhanced_networking", "delete_termination",
			"encrypted", "force_deregister", "iops",
			"no_device", "ssh_private_ip":
			settings[k], _ = strconv.ParseBool(v)
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
	if !hasSSHUsername {
		return nil, &RequiredSettingError{ID, "ssh_username"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		// only process supported array stuff
		switch name {
		case "ami_block_device_mappings":
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
//   delete_on_termination         bool
//   device_name                   string
//   encrypted                     bool
//   enhanced_networking           bool
//   force_deregister              bool
//   iam_instance_profile          string
//   iops                          int
//   launch_block_device_mappings  array of block device mappings
//   no_device                     bool
//   run_tags                      object of key/value strings
//   security_group_id             string
//   security_group_ids            array of strings
//   snapshot_id                   string
//   spot_price                    string
//   spot_price_auto_product       string
//   ssh_keypair_name              string
//   shh_port                      int
//   ssh_private_key_file          string
//   ssh_private_ip                bool
//   subnet_id                     string
//   tags                          object of key/value strings
//   temporary_key_pair_name       string
//   user_data                     string
//   user_data_file                string
//   virtual_name                  string
//   volume_size                   int
//   volume_type                   string
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
	var hasSecretKey, hasSourceAmi, hasSSHUsername, hasX509CertPath, hasX509KeyPath bool
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
		case "ami_name":
			settings[k] = v
			hasAmiName = true
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
		case "source_ami":
			settings[k] = v
			hasSourceAmi = true
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "x509_cert_path":
			settings[k] = v
			hasX509CertPath = true
		case "x509_key_path":
			settings[k] = v
			hasX509KeyPath = true
		case "ami_description", "ami_virtualization_type", "availability_zone",
			"bundle_destination", "bundle_prefix", "device_name",
			"iam_instance_profile", "security_group_id", "snapshot_id",
			"spot_price", "spot_price_auto_product", "ssh_keypair_name",
			"ssh_private_key_file", "subnet_id", "temporary_key_pair_name",
			"user_data", "virtual_name", "vpc_id",
			"x509_upload_path", "windows_password_timeout":
			settings[k] = v
		case "iops", "ssh_port", "volume_size":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		case "user_data_file":
			src, err := r.findComponentSource(AmazonInstance.String(), v, false)
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
			settings[k] = r.buildTemplateResourcePath(AmazonInstance.String(), v)
		case "associate_public_ip_address", "delete_on_termination", "encrypted",
			"enhanced_networking", "force_deregister", "no_device",
			"ssh_private_ip":
			settings[k], _ = strconv.ParseBool(v)
		case "bundle_upload_command", "bundle_vol_command":
			cmds, err := r.commandsFromFile(AmazonInstance.String(), v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			if len(cmds) == 0 {
				return nil, &SettingError{ID, k, v, ErrNoCommands}
			}
			// the setting is a string so don't use the full slice
			settings[k] = cmds[0]
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
	if !hasSSHUsername {
		return nil, &RequiredSettingError{ID, "ssh_username"}
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

// createDigitalOcean creates a map of settings for Packer's digitalocean
// builder.  Any values that aren't supported by the digitalocean builder are
// ignored.  Any required settings that don't exist result in an error and
// processing of the builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/digitalocean.html
//
// NOTE: The deprecated image_id, region_id, and size_id options are not
//       supported.
//
// Required V1 api configuration options:
//   api_key             string
//   client_id           string
// Required V2 api configuration options:
//   api_token           string
// Optional configuration options:
//   api_url             string
//   droplet_name        string
//   image               string
//   image_id            int
//   private_networking  bool
//   region              string
//   region_id           int
//   size                string
//   size_id             int
//   snapshot_name       string
//   ssh_port            int
//   ssh_timeout         string
//   ssh_username        string
//   state_timeout       string
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
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	// TODO look at snapshot name handling--it should be unique, e.g. timestamp
	var hasAPIToken, hasAPIKey, hasClientID bool
	for _, s := range workSlice {
		// var tmp interface{}
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "api_key":
			settings[k] = v
			hasAPIKey = true
		case "api_token":
			settings[k] = v
			hasAPIToken = true
		case "client_id":
			settings[k] = v
			hasClientID = true
		case "api_url", "droplet_name", "image", "region", "size", "snapshot_name", "ssh_timeout", "ssh_username", "state_timeout":
			settings[k] = v
		case "private_networking":
			settings[k], _ = strconv.ParseBool(v)
		case "ssh_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		}
	}
	if hasAPIToken {
		return settings, nil
	}
	if hasAPIKey && hasClientID {
		return settings, nil
	}
	return nil, &RequiredSettingError{ID, "(either api_token or (api_key && client_id))"}
}

// createDocker creates a map of settings for Packer's docker builder. Any
// values that aren't supported by the digitalocean builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped. For more information, refer to
// https://packer.io/docs/builders/docker.html
//
// Required configuration options:
//   commit         bool
//   export_path    string
//   image          string
// Optional configuration options:
//   login          bool
//   login_email    string
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
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasCommit, hasExportPath, hasImage, hasRunCommandArray bool
	var runCommandFile string
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "export_path":
			settings[k] = v
			hasExportPath = true
		case "image":
			settings[k] = v
			hasImage = true
		case "login_email", "login_username", "login_password", "login_server":
			settings[k] = v
		case "commit":
			settings[k], _ = strconv.ParseBool(v)
			hasCommit = true
		case "login", "pull":
			settings[k], _ = strconv.ParseBool(v)
		case "run_command":
			// if it's here, cache the value, delay processing until arrays section
			runCommandFile = v
		}
	}
	if !hasCommit {
		return nil, &RequiredSettingError{ID, "commit"}
	}
	if !hasExportPath {
		return nil, &RequiredSettingError{ID, "export_path"}
	}
	if !hasImage {
		return nil, &RequiredSettingError{ID, "image"}
	}
	// Process the Arrays.
	for name, val := range r.Builders[ID].Arrays {
		if name == "run_command" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
			hasRunCommandArray = true
			continue
		}
		if name == "volumes" {
			settings[name] = val
		}
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
// Required configuration options:
//   project_id         string
//   source_image       string
//   zone               string
// Optional configuration options:
//   account_file       string
//   disk_size          int
//   image_name         string
//   image_description  string
//   instance_name      string
//   machine_type       string
//   metadata           object of key/value strings
//   network            string
//   ssh_port           int
//   ssh_timeout        string
//   ssh_username       string
//   state_timeout      string
//   tags               array of strings
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
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var hasProjectID, hasSourceImage, hasZone bool
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "project_id":
			settings[k] = v
			hasProjectID = true
		case "source_image":
			settings[k] = v
			hasSourceImage = true
		case "zone":
			settings[k] = v
			hasZone = true
		case "image_name", "image_description", "instance_name",
			"machine_type", "network", "ssh_timeout", "ssh_username", "state_timeout":
			settings[k] = v
		case "account_file":
			src, err := r.findComponentSource(GoogleCompute.String(), v, false)
			if err != nil {
				return nil, err
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(GoogleCompute.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(GoogleCompute.String(), v)
		case "disk_size", "ssh_port":
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
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
		if name == "metadata" {
			settings[name] = val
			continue
		}
		if name == "tags" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
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
// Required configuration options:
//   host string
//   ssh_password         string
//   ssh_privateKey_file  string
//   ssh_username         string
// Optional configuration options:
//   port                 int
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
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	for _, s := range workSlice {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "host", "ssh_password", "ssh_private_key_file", "ssh_username":
			settings[k] = v
		case "port":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		}
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
// Required configuration options:
//   iso_checksum            string
//   iso_checksum_type       string
//   iso_url                 string
//   ssh_username            string
// Optional configuration options
//   boot_command            array of strings
//   boot_wait               string
//   disk_size               int
//   disk_type_id            string
//   floppy_files            array of strings
//   fusion_app_path         string
//   guest_os_type           string; if not set, will be generated
//   headless                bool
//   http_directory          string
//   http_port_min           int
//   http_port_max           int
//   iso_urls                array of strings
//   output_directory        string
//   remote_cache_datastore  string
//   remote_cache_directory  string
//   remote_datastore        string
//   remote_host             string
//   remote_password         string
//   remote_type             string
//   remote_username         string
//   shutdown_command        string
//   shutdown_timeout        string
//   skip_compaction         bool
//   ssh_host                string
//   ssh_key_path            string
//   ssh_password            string
//   ssh_port                int
//   ssh_skip_request_pty    bool
//   ssh_timeout        string
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
	// Go through each element in the slice, only take the ones that matter
	// to this builder.
	var bootCmdProcessed, hasSSHUsername bool
	var tmpISOChecksum, tmpISOChecksumType, tmpISOUrl, tmpGuestOSType string
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
		case "boot_wait", "disk_type_id", "fusion_app_path", "http_directory",
			"output_directory", "remote_cache_datastore", "remote_cache_directory",
			"remote_datastore", "remote_host", "remote_password", "remote_type",
			"remote_username", "shutdown_timeout", "ssh_host", "ssh_key_path",
			"ssh_password", "ssh_timeout", "tools_upload_flavor", "tools_upload_path",
			"vm_name", "vmdk_name", "vmx_template_path":
			settings[k] = v
		case "guest_os_type":
			tmpGuestOSType = v
		case "ssh_username":
			settings[k] = v
			hasSSHUsername = true
		case "headless":
			settings[k], _ = strconv.ParseBool(v)
		case "iso_checksum_type":
			settings[k] = v
			tmpISOChecksumType = v
		case "iso_checksum":
			settings[k] = v
			tmpISOChecksum = v
		case "iso_url":
			settings[k] = v
			tmpISOUrl = v
		case "disk_size", "http_port_min", "http_port_max", "ssh_port", "vnc_port_min",
			"vnc_port_max":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		}
	}
	// Only check to see if the required ssh_username field was set. The required iso info is checked after Array processing
	if !hasSSHUsername {
		return nil, &RequiredSettingError{ID, "ssh_username"}
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
		case "floppy_files":
			settings[name] = val
		case "iso_urls":
			// these are only added if iso_url isn't set
			if tmpISOUrl == "" {
				if tmpISOChecksum == "" {
					return nil, &RequiredSettingError{ID: ID, Key: "iso_url, iso_checksum"}
				}
				if tmpISOChecksumType == "" {
					return nil, &RequiredSettingError{ID: ID, Key: "iso_url, iso_checksum_type"}
				}
				settings[name] = val
			}
		case "vmx_data", "vmx_data_post":
			settings[name] = r.createVMXData(val)
		}
	}
	if r.osType == "" { // if the os type hasn't been set, the ISO info hasn't been retrieved
		err = r.ISOInfo(VirtualBoxISO, workSlice)
		if err != nil {
			return nil, err
		}
	}
	// set the guest_os_type
	if tmpGuestOSType == "" {
		tmpGuestOSType = r.osType
	}
	settings["guest_os_type"] = tmpGuestOSType
	// If the iso info wasn't set from the Settings, get it from the distro's release
	if tmpISOUrl == "" {
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
	if tmpISOChecksum == "" {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_url, iso_checksum"}
	}
	if tmpISOChecksumType == "" {
		return nil, &RequiredSettingError{ID: ID, Key: "iso_url, iso_checksum_type"}
	}
	return settings, nil
}

// createVMWareVMX creates a map of settings for Packer's vmware-vmx builder.
// Any values that aren't supported by the vmware-vmx builder are ignored.  Any
// required settings that don't exist result in an error and processing of the
// builder is stopped.  For more information, refer to
// https://packer.io/docs/builders/vmware-vmx.html
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
//   ssh_key_path             string
//   ssh_password             string
//   ssh_port                 int
//   ssh_skip_request_pty     bool
//   ssh_timeout         string
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
	var hasSourcePath, hasSSHUsername, bootCmdProcessed bool
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
			settings[k] = v
			hasSSHUsername = true
		case "boot_wait", "fusion_app_path", "http_directory", "output_directory", "shutdown_timeout",
			"ssh_key_path", "ssh_password", "ssh_timeout", "vm_name":
			settings[k] = v
		case "headless", "skip_compaction", "ssh_skip_request_pty":
			settings[k], _ = strconv.ParseBool(v)
		// For the fields of int value, only set if it converts to a valid int.
		// Otherwise, throw an error
		case "http_port_max", "http_port_min", "ssh_port", "vnc_port_max", "vnc_port_min":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				return nil, &SettingError{ID, k, v, err}
			}
			settings[k] = i
		}
	}
	// Check if required fields were processed
	if !hasSSHUsername {
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
		case "vmx_data", "vmx_data_post":
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
