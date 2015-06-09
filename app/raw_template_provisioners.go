package app

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
)

// Provisioner constants
const (
	UnsupportedProvisioner Provisioner = iota
	Ansible
	ChefClient
	ChefSolo
	FileUploads
	PuppetMasterless
	PuppetServer
	Salt
	ShellScript
)

// Provisioner is a packer supported provisioner
type Provisioner int

var provisioners = [...]string{
	"unsupported",
	"ansible-local",     //ansible is the name of the Ansible Provisioner
	"chef-client",       //chef-client is the name of the ChefClient Provisioner
	"chef-solo",         //chef-solo is the name of the ChefSolo Provisioner
	"file",              //file is the name of the FileUploads Provisioner
	"puppet-masterless", //puppet-masterless is the name of the PuppetMasterless Provisioner
	"puppet-server",     // puppet-server is the name of the PuppetServer Provisioner
	"salt-masterless",   //salt is the name of the Salt Provisioner
	"shell",             // shell is the name for the Shell provisioner
}

func (p Provisioner) String() string { return provisioners[p] }

// ProvisionerFromString returns the Provisioner constant for the passed string
// or unsupported. All incoming strings are normalized to lowercase
func ProvisionerFromString(s string) Provisioner {
	s = strings.ToLower(s)
	switch s {
	case "ansible-local":
		return Ansible
	case "chef-client":
		return ChefClient
	case "chef-solo":
		return ChefSolo
	case "file":
		return FileUploads
	case "puppet-masterless":
		return PuppetMasterless
	case "puppet-server":
		return PuppetServer
	case "salt-masterless":
		return Salt
	case "shell":
		return ShellScript
	}
	return UnsupportedProvisioner
}

// createProvisioner creates the provisioners for a build.
func (r *rawTemplate) createProvisioners() (p []interface{}, err error) {
	if r.ProvisionerTypes == nil || len(r.ProvisionerTypes) <= 0 {
		err = fmt.Errorf("unable to create provisioners: none specified")
		return nil, err
	}
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.ProvisionerTypes))
	// Generate the postProcessor for each postProcessor type.
	for _, pType := range r.ProvisionerTypes {
		typ := ProvisionerFromString(pType)
		switch typ {
		case Ansible:
			tmpS, err = r.createAnsible()
			if err != nil {
				return nil, provisionerErr(Ansible, err)
			}
		case FileUploads:
			tmpS, err = r.createFileUploads()
			if err != nil {
				return nil, provisionerErr(FileUploads, err)
			}
		case Salt:
			tmpS, err = r.createSalt()
			if err != nil {
				return nil, provisionerErr(Salt, err)
			}
		case ShellScript:
			tmpS, err = r.createShellScript()
			if err != nil {
				return nil, provisionerErr(ShellScript, err)
			}
		case ChefClient:
			tmpS, err = r.createChefClient()
			if err != nil {
				return nil, provisionerErr(ChefClient, err)
			}
		case ChefSolo:
			tmpS, err = r.createChefSolo()
			if err != nil {
				return nil, provisionerErr(ChefSolo, err)
			}
			/*
				case PuppetClient:
					// not implemented
				case PuppetServer:
					// not implemented
			*/
		default:
			return nil, provisionerErr(UnsupportedProvisioner, fmt.Errorf("%s is not supported", pType))
		}
		p[ndx] = tmpS
		ndx++
	}
	return p, nil
}

// createAnsible() creates a map of settings for Packer's ansible-local
// provisioner.  Any values that aren't supported by the file provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/ansible-local.html
//
// Required configuration options:
//   playbook_file      string
// Optional configuration options:
//   command            string
//   extra_arguments    array of strings
//   inventory_file     string
//   group_vars         string
//   host_vars          string
//   playbook_dir	      string
//   playbook_paths     array of strings
//   role_paths         array of strings
//   staging_directory  string
func (r *rawTemplate) createAnsible() (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[Ansible.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = Ansible.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasPlaybook bool
	for _, s := range r.Provisioners[Ansible.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "playbook_file":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(Ansible.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.files[filepath.Join(r.OutDir, Ansible.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(Ansible.String(), v)
			hasPlaybook = true
		case "inventory_file":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(Ansible.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.files[r.buildOutPath(Ansible.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(Ansible.String(), v)
		case "playbook_dir", "host_vars", "group_vars":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(Ansible.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.dirs[r.buildOutPath(Ansible.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(Ansible.String(), v)
		case "command", "staging_directory":
			settings[k] = v
		}
	}
	if !hasPlaybook {
		return nil, requiredSettingErr("playbook_file")
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[Ansible.String()].Arrays {
		// playbook_paths, role_paths
		if name == "playbook_paths" || name == "role_paths" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			for i, v := range array {
				v = r.replaceVariables(v)
				s, err := r.findComponentSource(Ansible.String(), v)
				if err != nil {
					return nil, settingErr(k, err)
				}
				r.files[r.buildOutPath(Ansible.String(), v)] = s
				array[i] = r.buildTemplateResourcePath(Ansible.String(), v)
			}
			settings[name] = array
			continue
		}
		// extra_arguments
		if name == "extra_arguments" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
			continue
		}
	}
	return settings, nil
}

// createChefClient() creates a map of settings for Packer's chef-client
// provisioner.  Any values that aren't supported by the chef-client
// provisioner are ignored. For more information, refer to:
// https://www.packer.io/docs/provisioners/chef-client.html
//
// Required configuration options:
//   none
// Optional configuraiton options:
//   chef_environment        string
//   config_template         string
//   execute_command         string
//   install_command         string
///  node_name               string
//   prevent_sudo            bool
//   run_list                array of strings
//   server_url              string
//   skip_clean_client       bool
//   skip_clean_node         bool
//   skip_install            bool
//   staging_directory       string
//   validation_client_name  string
//   validation_key_path     string
// Unsopported configuration options:
//   json                    object
func (r *rawTemplate) createChefClient() (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ChefClient.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = ChefClient.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	for _, s := range r.Provisioners[ChefClient.String()].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "chef_environment", "node_name", "server_url", "staging_directory", "validation_client_name":
			settings[k] = v
		case "prevent_sudo", "skip_clean_client", "skip_clean_node", "skip_install":
			settings[k], _ = strconv.ParseBool(v)
		case "config_template", "validation_key_path":
			// find the actual location of the source file and add it to the files map for copying
			src, err := r.findComponentSource(ChefClient.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.files[r.buildOutPath(ChefClient.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(ChefClient.String(), v)
		case "execute_command", "install_command":
			// if the value ends with .command, find the referenced command file and use its
			// contents as the command, otherwise just use the value
			if strings.HasSuffix(v, ".command") {
				commands, err := r.commandsFromFile(ChefClient.String(), v)
				if err != nil {
					return nil, commandFileErr(k, v, err)
				}
				if len(commands) == 0 {
					return nil, noCommandsFoundErr(k, v)
				}
				settings[k] = commands[0]
				continue
			}
			settings[k] = v
		}
	}
	for name, val := range r.Provisioners[ChefClient.String()].Arrays {
		if name == "run_list" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			settings[name] = array
			continue
		}
	}
	return settings, nil
}

// createChefSolo() creates a map of settings for Packer's chef-solo
// provisioner.  Any values that aren't supported by the chef-solo provisioner
// are ignored.  For more information, refer to
// https://www.packer.io/docs/provisioners/chef-solo.html
//
// Required configuration options:
//   none
// Optional configuraiton options:
//   config_template                 string
//   cookbook_paths                  array of strings
//   data_bags_path                  string
//   encrypted_data_bag_secret_path  string
//   environments_path               string
//   execute_command                 string
//   install_command                 string
//   prevent_sudo                    bool
//   remote_cookbook_paths           array of strings
//   roles_path                      string
//   run_list                        array of strings
//   skip_install                    bool
//   staging_directory               string
// Unsopported configuration options:
//   json                            object
func (r *rawTemplate) createChefSolo() (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ChefSolo.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = ChefSolo.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	for _, s := range r.Provisioners[ChefSolo.String()].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "staging_directory":
			settings[k] = v
		case "prevent_sudo", "skip_install":
			settings[k], _ = strconv.ParseBool(v)
		case "config_template", "encrypted_data_bag_secret_path":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(ChefSolo.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.files[r.buildOutPath(ChefSolo.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(ChefSolo.String(), v)
		case "data_bags_path", "environments_path", "roles_path":
			src, err := r.findComponentSource(ChefSolo.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.dirs[r.buildOutPath(ChefSolo.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(ChefSolo.String(), v)
		case "execute_command", "install_command":
			// if the value ends with .command, find the referenced command file and use its
			// contents as the command, otherwise just use the value
			if strings.HasSuffix(v, ".command") {
				commands, err := r.commandsFromFile(ChefSolo.String(), v)
				if err != nil {
					return nil, commandFileErr(k, v, err)
				}
				if len(commands) == 0 {
					return nil, noCommandsFoundErr(k, v)
				}
				settings[k] = commands[0]
				continue
			}
			settings[k] = v
		}
	}
	for name, val := range r.Provisioners[ChefSolo.String()].Arrays {
		if name == "cookbook_paths" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			for i, v := range array {
				v = r.replaceVariables(v)
				// find the actual location and add it to the files map for copying
				src, err := r.findComponentSource(ChefSolo.String(), v)
				if err != nil {
					return nil, settingErr(name, err)
				}
				array[i] = r.buildTemplateResourcePath(ChefSolo.String(), v)
				r.dirs[r.buildOutPath(ChefSolo.String(), v)] = src
			}
			settings[name] = array
			continue
		}
		if name == "run_list" || name == "remote_cookbook_paths" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			settings[name] = array
			continue
		}
	}
	return settings, nil
}

// createFileUploads() creates a map of settings for Packer's file uploads
// provisioner. Any values that aren't supported by the file provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/file.html
//
// Required configuration options:
//   destination  string
//   source       string
func (r *rawTemplate) createFileUploads() (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[FileUploads.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = FileUploads.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var k, v string
	var hasSource, hasDestination bool
	for _, s := range r.Provisioners[FileUploads.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "source":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(FileUploads.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			// add to files
			r.files[r.buildOutPath(FileUploads.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(FileUploads.String(), v)
			hasSource = true
		case "destination":
			settings[k] = v
			hasDestination = true
		}
	}
	if !hasSource {
		return nil, requiredSettingErr("source")
	}
	if !hasDestination {
		return nil, requiredSettingErr("destination")
	}
	return settings, nil
}

// createSalt() creates a map of settings for Packer's salt provisioner. Any
// values that aren't supported by the salt provisioner are ignored. For more
// information, refer to
// https://packer.io/docs/provisioners/salt-masterless.html
//
// Required configuration options:
//   local_state_tree     string
// Optional configuration options
//   bootstrap_args       string
//   local_pillar_roots   string
//   local_state_tree     string
//   minion_config        string
//   skip_bootstrap       boolean
//   temp_config_dir      string
func (r *rawTemplate) createSalt() (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[Salt.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = Salt.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var k, v string
	var hasLocalStateTree bool
	for _, s := range r.Provisioners[Salt.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "local_state_tree":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(Salt.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.dirs[r.buildOutPath(Salt.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(Salt.String(), v)
			hasLocalStateTree = true
		case "local_pillar_roots":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(Salt.String(), v)
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.dirs[r.buildOutPath(Salt.String(), v)] = src
			settings[k] = r.buildTemplateResourcePath(Salt.String(), v)
		case "minion_config":
			// find the actual location and add it to the files map for copying
			src, err := r.findComponentSource(Salt.String(), filepath.Join(v, "minion"))
			if err != nil {
				return nil, settingErr(k, err)
			}
			r.files[r.buildOutPath(Salt.String(), filepath.Join(v, "minion"))] = src
			settings[k] = r.buildTemplateResourcePath(Salt.String(), v)
		case "bootstrap_args", "temp_config_dir":
			settings[k] = v
		case "skip_bootstrap":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	if !hasLocalStateTree {
		return nil, requiredSettingErr("local_state_tree")
	}
	// salt does not have any arrays to support
	return settings, nil
}

// createShellScript() creates a map of settings for Packer's shell script
// provisioner. Any values that aren't supported by the shell provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/shell.html
//
// Of the "inline", "script", and "scripts" options, only "scripts" is
// currently supported.
//
// Required configuration options:
//   scripts              array of strings
// Optional confinguration parameters:
//   binary               boolean
//   environment_vars     array of strings
//   execute_command      string
//   inline_shebang       string
//   remote_path          string
//   start_retry_timeout  string
func (r *rawTemplate) createShellScript() (settings map[string]interface{}, err error) {
	fmt.Println(r.IncludeComponentString)
	_, ok := r.Provisioners[ShellScript.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = ShellScript.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var k, v string
	for _, s := range r.Provisioners[ShellScript.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "execute_command":
			// If the execute_command references a file, parse that for the command
			// Otherwise assume that it contains the command
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile(ShellScript.String(), v)
				if err != nil {
					return nil, commandFileErr(k, v, err)
				}
				if len(commands) == 0 {
					return nil, noCommandsFoundErr(k, v)
				}
				settings[k] = commands[0] // for execute_command, only the first element is used
				continue
			}
			settings[k] = v
		case "inline_shebang", "remote_path", "start_retry_timeout":
			settings[k] = v
		case "binary":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	// Process the Arrays.
	var scripts []string
	for name, val := range r.Provisioners[ShellScript.String()].Arrays {
		// if this is a scripts array, special processing needs to be done.
		if name == "scripts" {
			scripts = deepcopy.InterfaceToSliceOfStrings(val)
			for i, v := range scripts {
				v = r.replaceVariables(v)
				// find the source
				s, err := r.findComponentSource(ShellScript.String(), v)
				if err != nil {
					return nil, settingErr(k, err)
				}
				r.files[r.buildOutPath(ShellScript.String(), v)] = s
				scripts[i] = r.buildTemplateResourcePath(ShellScript.String(), v)
			}
			settings[name] = scripts
			continue
		}
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	if len(scripts) == 0 {
		return nil, requiredSettingErr("scripts")
	}
	return settings, nil
}

// Merges the new config with the old. The updates occur as follows:
//  * The existing configuration is used when no `new` provisioners are
//    specified.
//  * When 1 or more `new` provisioner are specified, they will replace all
//    existing provisioners.  In this situation, if a provisioner exists in the
//   `old` map but it does not exist in the `new` map, that provisioner will be
//   orphaned.
//  * If there isn't a new config, return the existing as there are no
//    overrides.
func (r *rawTemplate) updateProvisioners(newP map[string]provisioner) error {
	// If there is nothing new, old equals merged.
	if len(newP) <= 0 || newP == nil {
		return nil
	}
	// Convert the existing provisioners to interface.
	var ifaceOld = make(map[string]interface{}, len(r.Provisioners))
	ifaceOld = DeepCopyMapStringProvisioner(r.Provisioners)
	// Convert the new provisioners to interface.
	var ifaceNew = make(map[string]interface{}, len(newP))
	ifaceNew = DeepCopyMapStringProvisioner(newP)
	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	if r.Provisioners == nil {
		r.Provisioners = map[string]provisioner{}
	}
	// Copy: if the key exists in the new provisioners only.
	// Ignore: if the key does not exist in the new provisioners.
	// Merge: if the key exists in both the new and old provisioners.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		p, ok := r.Provisioners[v]
		if !ok {
			pp, _ := newP[v]
			r.Provisioners[v] = pp.DeepCopy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		pp, ok := newP[v]
		if !ok {
			continue
		}
		err := p.mergeSettings(pp.Settings)
		if err != nil {
			return mergeSettingsErr(err)
		}
		p.mergeArrays(pp.Arrays)
		r.Provisioners[v] = p
	}
	return nil
}

// Go through all of the Settings and convert them to a map. Each setting is
// parsed into its constituent parts. The value then goes through variable
// replacement to ensure that the settings are properly resolved.
func (p *provisioner) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	var k string
	var v interface{}
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type
	for _, s := range p.Settings {
		k, v = parseVar(s)
		switch k {
		case "binary":
			v, _ = strconv.ParseBool(v.(string))
		default:
			v = r.replaceVariables(v.(string))
		}
		m[k] = v
	}
	return m
}

// DeepCopyMapStringProvisioner makes a deep copy of each builder passed and
// returns the copie map[string]provisioner as a map[string]interface{}
func DeepCopyMapStringProvisioner(p map[string]provisioner) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range p {
		tmpP := provisioner{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}
