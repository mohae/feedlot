package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// ProvisionerErr is an error processing provisioner, its Err field may contain
// additional type information.
type ProvisionerErr struct {
	id string
	Provisioner
	Err error
}

func (e ProvisionerErr) Error() string {
	var s string
	if e.Provisioner != UnsupportedProvisioner {
		s = e.Provisioner.String()
	}
	if e.id != "" {
		if s == "" {
			s = e.id
		} else {
			s += ": " + e.id
		}
	}
	if s == "" {
		return e.Err.Error()
	}
	return s + ": " + e.Err.Error()
}

// ErrProvisionerNotFound occurs when a provisioner with a matching ID is not
// found in the definition.
var ErrProvisionerNotFound = errors.New("provisioner not found")

// Provisioner constants
const (
	UnsupportedProvisioner Provisioner = iota
	Ansible
	AnsibleLocal
	ChefClient
	ChefSolo
	File
	PuppetMasterless
	PuppetServer
	Salt
	Shell
	ShellLocal
)

// Provisioner is a packer supported provisioner
type Provisioner int

func (p Provisioner) String() string { return provisioners[p] }

var provisioners = [...]string{
	"unsupported provisioner",
	"ansible",
	"ansible-local",
	"chef-client",
	"chef-solo",
	"file",
	"puppet-masterless",
	"puppet-server",
	"salt-masterless",
	"shell",
	"shell-local",
}

// ParseProvisioner returns the Provisioner constant for s. If no match is
// found, UnsupportedProvisioner is returned. All incoming strings are
// normalized to lowercase
func ParseProvisioner(s string) Provisioner {
	s = strings.ToLower(s)
	switch s {
	case "ansible":
		return Ansible
	case "ansible-local":
		return AnsibleLocal
	case "chef-client":
		return ChefClient
	case "chef-solo":
		return ChefSolo
	case "file":
		return File
	case "puppet-masterless":
		return PuppetMasterless
	case "puppet-server":
		return PuppetServer
	case "salt-masterless":
		return Salt
	case "shell":
		return Shell
	case "shell-local":
		return ShellLocal
	}
	return UnsupportedProvisioner
}

// createProvisioner creates the provisioners for a build.
func (r *rawTemplate) createProvisioners() (p []interface{}, err error) {
	if r.ProvisionerIDs == nil || len(r.ProvisionerIDs) <= 0 {
		return nil, nil
	}
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.ProvisionerIDs))
	// Generate the provisioners for each provisioners ID.
	for _, ID := range r.ProvisionerIDs {
		tmpP, ok := r.Provisioners[ID]
		if !ok {
			return nil, ProvisionerErr{id: ID, Err: ErrProvisionerNotFound}
		}
		jww.DEBUG.Printf("processing provisioner id: %s\n", ID)
		typ := ParseProvisioner(tmpP.Type)
		switch typ {
		case Ansible:
			tmpS, err = r.createAnsible(ID)
			if err != nil {
				return nil, err
			}
		case AnsibleLocal:
			tmpS, err = r.createAnsibleLocal(ID)
			if err != nil {
				return nil, err
			}
		case ChefClient:
			tmpS, err = r.createChefClient(ID)
			if err != nil {
				return nil, err
			}
		case ChefSolo:
			tmpS, err = r.createChefSolo(ID)
			if err != nil {
				return nil, err
			}
		case File:
			tmpS, err = r.createFile(ID)
			if err != nil {
				return nil, err
			}
		case PuppetMasterless:
			tmpS, err = r.createPuppetMasterless(ID)
			if err != nil {
				return nil, err
			}
		case PuppetServer:
			tmpS, err = r.createPuppetServer(ID)
			if err != nil {
				return nil, err
			}
		case Salt:
			tmpS, err = r.createSalt(ID)
			if err != nil {
				return nil, err
			}
		case Shell:
			tmpS, err = r.createShell(ID)
			if err != nil {
				return nil, err
			}
		case ShellLocal:
			tmpS, err = r.createShellLocal(ID)
			if err != nil {
				return nil, err
			}
		default:
			return nil, InvalidComponentErr{cTyp: "provisioner", s: tmpP.Type}
		}
		p[ndx] = tmpS
		ndx++
		jww.DEBUG.Printf("processed provisioner id: %s\n", ID)
	}
	return p, nil
}

// createAnsible creates a map of settings for Packer's ansible provisioner.
//  Any values that aren't supported by the file provisioner are ignored. For
// more information, refer to
// https://packer.io/docs/provisioners/ansible-local.html
//
// Required configuration options:
//   playbook_file            string
// Optional configuration options:
//   ansible_env_vars         array of strings
//   empty_groups             array of strings
//   extra_arguments          array of strings
//   groups                   array of strings
//   host_alias               string
//   local_port               string
//   sftp_command             string
//   ssh_authorized_key_file  string
//   ssh_host_key_file        string
//   user                     string
func (r *rawTemplate) createAnsible(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: Ansible, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = Ansible.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasPlaybook bool
	for _, s := range r.Provisioners[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "playbook_file":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, Ansible.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: Ansible, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[filepath.Join(r.TemplateOutputDir, Ansible.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(Ansible.String(), v, false)
			hasPlaybook = true
		case "sftp_command":
			if !stringIsCommandFilename(v) {
				// The value is the command.
				settings[k] = v
				continue
			}
			// The value is a command file, load the contents of the file.
			cmds, err := r.commandsFromFile(v, Ansible.String())
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: Ansible, Err: SettingErr{k, v, err}}
			}
			// Make the cmds slice a single string, if it was split into multiple lines.
			cmd := commandFromSlice(cmds)
			if cmd == "" {
				return nil, ProvisionerErr{id: ID, Provisioner: Ansible, Err: SettingErr{k, v, ErrNoCommands}}
			}
			settings[k] = cmd
		case "host_alias", "local_port", "ssh_authorized_key_file", "ssh_host_key_file", "user":
			settings[k] = v
		}
	}
	if !hasPlaybook {
		return nil, ProvisionerErr{id: ID, Provisioner: Ansible, Err: RequiredSettingErr{"playbook_file"}}
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "ansible_env_vars" || name == "empty_groups" || name == "extra_arguments" || name == "groups" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
			continue
		}
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createAnsibleLocal creates a map of settings for Packer's ansible
// provisioner.  Any values that aren't supported by the file provisioner
// are ignored. For more information, refer to
// https://packer.io/docs/provisioners/ansible-local.html
//
// Required configuration options:
//   playbook_file      string
// Optional configuration options:
//   command            string
//   extra_arguments    array of strings
//   group_vars         string
//   host_vars          string
//   inventory_groups   string
//   inventory_file     string
//   playbook_dir	    string
//   playbook_paths     array of strings
//   role_paths         array of strings
//   staging_directory  string
func (r *rawTemplate) createAnsibleLocal(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: AnsibleLocal, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = AnsibleLocal.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasPlaybook bool
	for _, s := range r.Provisioners[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "playbook_file":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, AnsibleLocal.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: AnsibleLocal, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[filepath.Join(r.TemplateOutputDir, AnsibleLocal.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(AnsibleLocal.String(), v, false)
			hasPlaybook = true
		case "inventory_file":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, AnsibleLocal.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: AnsibleLocal, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(AnsibleLocal.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(AnsibleLocal.String(), v, false)
		case "playbook_dir", "host_vars", "group_vars":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, AnsibleLocal.String(), true)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: AnsibleLocal, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.dirs[r.buildOutPath(AnsibleLocal.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(AnsibleLocal.String(), v, false)
		case "command", "staging_directory", "inventory_groups":
			settings[k] = v
		}
	}
	if !hasPlaybook {
		return nil, ProvisionerErr{id: ID, Provisioner: AnsibleLocal, Err: RequiredSettingErr{"playbook_file"}}
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[ID].Arrays {
		// playbook_paths, role_paths
		if name == "playbook_paths" || name == "role_paths" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			for i, v := range array {
				v = r.replaceVariables(v)
				src, err := r.findSource(v, AnsibleLocal.String(), true)
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: AnsibleLocal, Err: SettingErr{k, v, err}}
				}
				// if the source couldn't be found and an error wasn't generated, replace
				// s with the original value; this occurs when it is an example.
				// Nothing should be copied in this instancel it should not be added
				// to the copy info
				if src != "" {
					r.files[r.buildOutPath(AnsibleLocal.String(), v)] = src
				}
				array[i] = r.buildTemplateResourcePath(AnsibleLocal.String(), v, false)
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
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createChefClient creates a map of settings for Packer's chef-client
// provisioner.  Any values that aren't supported by the chef-client
// provisioner are ignored. For more information, refer to:
// https://www.packer.io/docs/provisioners/chef-client.html
//
// Required configuration options:
//   none
// Optional configuraiton options:
//   chef_environment               string
//   client_key                     string
//   config_template                string
//   encrypted_data_bag_secret_path string
//   execute_command                string
//   guest_os_type                  string
//   install_command                string
//   node_name                      string
//   prevent_sudo                   bool
//   run_list                       array of strings
//   server_url                     string
//   skip_clean_client              bool
//   skip_clean_node                bool
//   skip_install                   bool
//   ssl_verify_mode                string
//   staging_directory              string
//   validation_client_name         string
//   validation_key_path            string
// Unsopported configuration options:
//   json                           object
func (r *rawTemplate) createChefClient(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ChefClient.String()]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: ChefClient, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = ChefClient.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	for _, s := range r.Provisioners[ID].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "chef_environment", "client_key", "encrypted_data_bag_secret_path", "guest_os_type",
			"node_name", "server_url", "ssl_verify_mode", "staging_directory",
			"validation_client_name", "validation_key_path":
			settings[k] = v
		case "prevent_sudo", "skip_clean_client", "skip_clean_node", "skip_install":
			settings[k], _ = strconv.ParseBool(v)
		case "config_template":
			// find the actual location of the source file and add it to the files map for copying
			src, err := r.findSource(v, ChefClient.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: ChefClient, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(ChefClient.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(ChefClient.String(), v, false)
		case "execute_command", "install_command":
			// if the value ends with .command, find the referenced command file and use its
			// contents as the command, otherwise just use the value
			if strings.HasSuffix(v, ".command") {
				commands, err := r.commandsFromFile(v, ChefClient.String())
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: ChefClient, Err: SettingErr{k, v, err}}
				}
				if len(commands) == 0 {
					return nil, ProvisionerErr{id: ID, Provisioner: ChefClient, Err: SettingErr{k, v, ErrNoCommands}}
				}
				settings[k] = commands[0]
				continue
			}
			settings[k] = v
		}
	}
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "run_list" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			settings[name] = array
			continue
		}
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createChefSolo creates a map of settings for Packer's chef-solo provisioner.
// Any values that aren't supported by the chef-solo provisioner are ignored.
//  For more information, refer to
// https://www.packer.io/docs/provisioners/chef-solo.html
//
// Required configuration options:
//   none
// Optional configuraiton options:
//   chef_environment                string
//   config_template                 string
//   cookbook_paths                  array of strings
//   data_bags_path                  string
//   encrypted_data_bag_secret_path  string
//   environments_path               string
//   execute_command                 string
//   guest_os_type                   string
//   install_command                 string
//   prevent_sudo                    bool
//   remote_cookbook_paths           array of strings
//   roles_path                      string
//   run_list                        array of strings
//   skip_install                    bool
//   staging_directory               string
// Unsopported configuration options:
//   json                            object
func (r *rawTemplate) createChefSolo(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: ChefSolo, Err: ErrProvisionerNotFound}
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
		case "chef_environment", "encrypted_data_bag_secret_path", "guest_os_type",
			"staging_directory":
			settings[k] = v
		case "prevent_sudo", "skip_install":
			settings[k], _ = strconv.ParseBool(v)
		case "config_template":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, ChefSolo.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: ChefSolo, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(ChefSolo.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(ChefSolo.String(), v, false)
		case "data_bags_path", "environments_path", "roles_path":
			src, err := r.findSource(v, ChefSolo.String(), true)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: ChefSolo, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.dirs[r.buildOutPath(ChefSolo.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(ChefSolo.String(), v, false)
		case "execute_command", "install_command":
			// if the value ends with .command, find the referenced command file and use its
			// contents as the command, otherwise just use the value
			if strings.HasSuffix(v, ".command") {
				commands, err := r.commandsFromFile(v, ChefSolo.String())
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: ChefSolo, Err: SettingErr{k, v, err}}
				}
				if len(commands) == 0 {
					return nil, ProvisionerErr{id: ID, Provisioner: ChefSolo, Err: SettingErr{k, v, ErrNoCommands}}
				}
				settings[k] = commands[0]
				continue
			}
			settings[k] = v
		}
	}
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "cookbook_paths" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			for i, v := range array {
				v = r.replaceVariables(v)
				// find the actual location and add it to the files map for copying
				src, err := r.findSource(v, ChefSolo.String(), true)
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: ChefSolo, Err: SettingErr{name, v, err}}
				}
				// if the source couldn't be found and an error wasn't generated, replace
				// s with the original value; this occurs when it is an example.
				// Nothing should be copied in this instancel it should not be added
				// to the copy info
				if src != "" {
					r.dirs[r.buildOutPath(ChefSolo.String(), v)] = src
				}
				array[i] = r.buildTemplateResourcePath(ChefSolo.String(), v, false)
			}
			settings[name] = array
			continue
		}
		if name == "run_list" || name == "remote_cookbook_paths" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			settings[name] = array
			continue
		}
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createFile creates a map of settings for Packer's file provisioner. Any
// values that aren't supported by the file provisioner are ignored. For
// more information, refer to https://packer.io/docs/provisioners/file.html
//
// Required configuration options:
//   destination  string
//   source       string
// Optional configuraiton options:
//   direction    string
func (r *rawTemplate) createFile(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: File, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = File.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var k, v string
	var hasSource, hasDestination bool
	for _, s := range r.Provisioners[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "source":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, File.String(), true)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: File, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instance it should not be added
			// to the copy info
			var isDir bool // used to track if the path is a dir
			if src != "" {
				// see if this is a dir
				inf, err := os.Stat(src)
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: File, Err: SettingErr{k, v, err}}
				}
				if inf.IsDir() {
					isDir = true
					r.dirs[r.buildOutPath(File.String(), v)] = src
				} else {
					r.files[r.buildOutPath(File.String(), v)] = src
				}
			}
			settings[k] = r.buildTemplateResourcePath(File.String(), v, isDir)
			hasSource = true
		case "destination":
			settings[k] = v
			hasDestination = true
		}
	}
	if !hasSource {
		return nil, ProvisionerErr{id: ID, Provisioner: File, Err: RequiredSettingErr{"source"}}
	}
	if !hasDestination {
		return nil, ProvisionerErr{id: ID, Provisioner: File, Err: RequiredSettingErr{"destination"}}
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createPuppetMasterless creates a map of settings for Packer's puppet-client
// provisioner.  Any values that aren't supported by the puppet-client
// provisioner are ignored. For more information, refer to:
// https://www.packer.io/docs/provisioners/chef-client.html
//
// Required configuration options:
//   manifest_file
// Optional configuraiton options:
//   execute_command    string
//   extra_arguments    array of strings
//   facter             object, string key and values
//   hiera_config_path  string
//   ignore_exit_codes  bool
//   manifest_dir       string
//   module_paths       array of strings
//   prevent_sudo       bool
//   staging_directroy  string
//   working_directory  string
func (r *rawTemplate) createPuppetMasterless(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = PuppetMasterless.String()
	var hasManifestFile bool
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	for _, s := range r.Provisioners[ID].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "manifest_file":
			src, err := r.findSource(v, PuppetMasterless.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(PuppetMasterless.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(PuppetMasterless.String(), v, false)
			hasManifestFile = true
		case "staging_directory", "working_directory":
			settings[k] = v
		case "ignore_exit_codes", "prevent_sudo":
			settings[k], _ = strconv.ParseBool(v)
		case "hiera_config_path":
			// find the actual location of the source file and add it to the files map for copying
			src, err := r.findSource(v, PuppetMasterless.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(PuppetMasterless.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(PuppetMasterless.String(), v, false)
		case "manifest_dir":
			// find the actual location of the directory and add it to the dir map for copying contents
			src, err := r.findSource(v, PuppetMasterless.String(), true)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.dirs[r.buildOutPath(PuppetMasterless.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(PuppetMasterless.String(), v, false)
		case "execute_command":
			// if the value ends with .command, find the referenced command file and use its
			// contents as the command, otherwise just use the value
			if strings.HasSuffix(v, ".command") {
				commands, err := r.commandsFromFile(v, PuppetMasterless.String())
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: SettingErr{k, v, err}}
				}
				if len(commands) == 0 {
					return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: SettingErr{k, v, ErrNoCommands}}
				}
				settings[k] = commands[0]
				continue
			}
			settings[k] = v
		}
	}
	if !hasManifestFile {
		return nil, ProvisionerErr{id: ID, Provisioner: PuppetMasterless, Err: RequiredSettingErr{"manifest_file"}}
	}
	for name, val := range r.Provisioners[ID].Arrays {
		switch name {
		case "extra_arguments", "module_paths", "only", "except":
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		case "facter":
			settings[name] = deepcopy.Iface(val)
		}
	}
	return settings, nil
}

// createPuppetServer creates a map of settings for Packer's puppet-client
// provisioner.  Any values that aren't supported by the puppet-client
// provisioner are ignored. For more information, refer to:
// https://www.packer.io/docs/provisioners/chef-client.html
//
// Required configuration options:
//  none
// Optional configuraiton options:
//   client_cert_path         string
//   client_private_key_path  string
//   facter                   object, string key and values
//   ignore_exit_codes        bool
//   options                  string
//   prevent_sudo             bool
//   puppet_node              string
//   puppet_server            string
//   staging_directroy        string
func (r *rawTemplate) createPuppetServer(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: PuppetServer, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = PuppetServer.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	for _, s := range r.Provisioners[PuppetServer.String()].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "client_cert_path", "client_private_key_path", "options", "puppet_node", "puppet_server", "staging_directory":
			settings[k] = v
		case "ignore_exit_codes", "prevent_sudo":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "facter" {
			settings[name] = deepcopy.Iface(val)
			continue
		}
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createSalt creates a map of settings for Packer's salt provisioner. Any
// values that aren't supported by the salt provisioner are ignored. For more
// information, refer to
// https://packer.io/docs/provisioners/salt-masterless.html
//
// Required configuration options:
//   local_state_tree     string
// Optional configuration options
//   bootstrap_args       string
//   disable_sudo         bool
//   local_pillar_roots   string
//   local_state_tree     string
//   log_level            string
//   minion_config        string
//   no_exit_on_failure   bool
//   remote_pillar_roots  string
//   remote_state_tree    string
//   skip_bootstrap       bool
//   temp_config_dir      string
func (r *rawTemplate) createSalt(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = Salt.String()
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var (
		k, v, remotePillarRoots, remoteStateTree string
		hasLocalStateTree, hasMinion             bool
	)
	for _, s := range r.Provisioners[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "local_state_tree":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, Salt.String(), true)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info)
			if src != "" {
				r.dirs[r.buildOutPath(Salt.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(Salt.String(), v, false)
			hasLocalStateTree = true
		case "local_pillar_roots":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v, Salt.String(), true)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.dirs[r.buildOutPath(Salt.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(Salt.String(), v, false)
		case "minion_config":
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(filepath.Join(v, "minion"), Salt.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(Salt.String(), filepath.Join(v, "minion"))] = src
			}
			settings[k] = r.buildTemplateResourcePath(Salt.String(), v, false)
			hasMinion = true
		case "bootstrap_args", "log_level", "temp_config_dir":
			settings[k] = v
		case "remote_pillar_roots":
			settings[k] = v
			remotePillarRoots = v
		case "remote_state_tree":
			settings[k] = v
			remoteStateTree = v
		case "disable_sudo", "no_exit_on_failure", "skip_bootstrap":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	if !hasLocalStateTree {
		return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: RequiredSettingErr{"local_state_tree"}}
	}
	// If minion is set, remote_pilar_roots and remote_state_tree cannot be used.
	if hasMinion {
		if remotePillarRoots != "" {
			return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: SettingErr{Key: "remote_pillar_roots", Value: remotePillarRoots, err: errors.New("cannot be used with the 'minion_config' setting")}}
		}
		if remoteStateTree != "" {
			return nil, ProvisionerErr{id: ID, Provisioner: Salt, Err: SettingErr{Key: "remote_state_tree", Value: remoteStateTree, err: errors.New("cannot be used with the 'minion_config' setting")}}
		}
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createShell creates a map of settings for Packer's shell provisioner.  Any
// values that aren't supported by the shell provisioner are ignored. For
// more information, refer to https://packer.io/docs/provisioners/shell.html
//
// Of the "inline", "script", and "scripts" options, only one can be used:
// they are mutually exclusive.  If multiple are specified, the one with the
// highest precedence will be used.  They are listed in order of precedence,
// from high to low:
//
// Required configuration options:
//   inline               array of strings
//   script               string
//   scripts              array of strings
// Optional confinguration parameters:
//   binary               bool
//   environment_vars     array of strings
//   execute_command      string
//   inline_shebang       string
//   remote_file          string
//   remote_folder        string
//   remote_path          string
//   skip_clean           bool
//   start_retry_timeout  string
func (r *rawTemplate) createShell(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: Shell, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = Shell.String()

	var script string
	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var k, v string
	for _, s := range r.Provisioners[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "execute_command":
			// If the execute_command references a file, parse that for the command
			// Otherwise assume that it contains the command
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile(v, Shell.String())
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: Shell, Err: SettingErr{k, v, err}}
				}
				if len(commands) == 0 {
					return nil, ProvisionerErr{id: ID, Provisioner: Shell, Err: SettingErr{k, v, ErrNoCommands}}
				}
				settings[k] = commands[0] // for execute_command, only the first element is used
				continue
			}
			settings[k] = v
		case "inline_shebang", "remote_file", "remote_folder", "remote_path", "start_retry_timeout":
			settings[k] = v
		case "script":
			// defer resolution of
			script = v
		case "binary", "skip_clean":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	// Check for inline and/or scripts in the Arrays.  This is so that the
	// setting with the highest precedence is used in situations where
	// inline, script, and/or scripts are not exclusive (at least two of them
	// exist in the settings).
	var (
		vals interface{}
		key  string
	)
	vals, ok = r.Provisioners[ID].Arrays["inline"]
	if ok {
		key = "inline"
	} else {
		vals, ok = r.Provisioners[ID].Arrays["scripts"]
		if ok {
			key = "scripts"
		}
	}
	if key == "inline" {
		settings[key] = deepcopy.InterfaceToSliceOfStrings(vals)
		goto arrays
	}

	if script != "" {
		script = r.replaceVariables(script)
		// find the source
		src, err := r.findSource(script, Shell.String(), false)
		if err != nil {
			return nil, ProvisionerErr{id: ID, Provisioner: Shell, Err: SettingErr{"script", script, err}}
		}
		// if the source couldn't be found and an error wasn't generated, replace
		// s with the original value; this occurs when it is an example.
		// Nothing should be copied in this instancel it should not be added
		// to the copy info
		if src != "" {
			r.files[r.buildOutPath(Shell.String(), script)] = src
		}
		settings["script"] = r.buildTemplateResourcePath(Shell.String(), script, false)
		goto arrays
	}
	if key == "scripts" {
		scripts := deepcopy.InterfaceToSliceOfStrings(vals)
		for i, v := range scripts {
			v = r.replaceVariables(v)
			// find the source
			src, err := r.findSource(v, Shell.String(), false)
			if err != nil {
				return nil, ProvisionerErr{id: ID, Provisioner: Shell, Err: SettingErr{k, v, err}}
			}
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(Shell.String(), v)] = src
			}
			scripts[i] = r.buildTemplateResourcePath(Shell.String(), v, false)
		}
		settings[key] = scripts
		goto arrays
	}
	// This means the a required setting was not found.
	return nil, ProvisionerErr{id: ID, Provisioner: Shell, Err: RequiredSettingErr{"inline, script, scripts"}}

arrays:
	// Process the Arrays.
	for name, val := range r.Provisioners[ID].Arrays {
		if name == "environment_vars" || name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
}

// createShellLocal creates a map of settings for Packer's shell local
// provisioner.  Any values that aren't supported by the shell local
// provisioner are ignored. For/ more information, refer to
// https://packer.io/docs/provisioners/shell.html
//
// Of the "inline", "script", and "scripts" options, only one can be used:
// they are mutually exclusive.  If multiple are specified, the one with the
// highest precedence will be used.  They are listed in order of precedence,
// from high to low:
//
// Required configuration options:
//   command              string
// Optional confinguration parameters:
//   execute_command      string
func (r *rawTemplate) createShellLocal(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.Provisioners[ID]
	if !ok {
		return nil, ProvisionerErr{id: ID, Provisioner: ShellLocal, Err: ErrProvisionerNotFound}
	}
	settings = make(map[string]interface{})
	settings["type"] = ShellLocal.String()

	// For each value, extract its key value pair and then process. Only process the supported
	// keys. Key validation isn't done here, leaving that for Packer.
	var hasCommand bool
	for _, s := range r.Provisioners[ID].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "command", "execute_command":
			if k == "command" && v != "" {
				hasCommand = true
			}
			// If the execute_command references a file, parse that for the command
			// Otherwise assume that it contains the command
			if strings.HasSuffix(v, ".command") {
				var commands []string
				commands, err = r.commandsFromFile(v, Shell.String())
				if err != nil {
					return nil, ProvisionerErr{id: ID, Provisioner: ShellLocal, Err: SettingErr{k, v, err}}
				}
				if len(commands) == 0 {
					return nil, ProvisionerErr{id: ID, Provisioner: ShellLocal, Err: SettingErr{k, v, ErrNoCommands}}
				}
				settings[k] = commands[0] // for execute_command, only the first element is used
				continue
			}
			settings[k] = v
		}
	}
	if !hasCommand {
		return nil, ProvisionerErr{id: ID, Provisioner: ShellLocal, Err: RequiredSettingErr{"command"}}
	}

	for name, val := range r.Provisioners[ID].Arrays {
		if name == "environment_vars" || name == "only" || name == "except" {
			array := deepcopy.InterfaceToSliceOfStrings(val)
			if array != nil {
				settings[name] = array
			}
		}
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
	// Convert the existing provisioners to Componenter.
	oldC := DeepCopyMapStringProvisioner(r.Provisioners)
	// Convert the new provisioners to Componenter.
	newC := DeepCopyMapStringProvisioner(newP)
	// Get the all keys from both maps
	keys := mergeKeysFromComponentMaps(oldC, newC)
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
			pp, ok := newP[v]
			if !ok { // if the key exists in neither then something is wrong
				return fmt.Errorf("provisioner merge failed: %s key not found in either template", v)
			}
			r.Provisioners[v] = pp.Copy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		pp, ok := newP[v]
		if !ok {
			continue
		}
		err := p.mergeSettings(pp.Settings)
		if err != nil {
			return err
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
// returns the copied map[string]provisioner as a map[string]interface{}
func DeepCopyMapStringProvisioner(p map[string]provisioner) map[string]Componenter {
	c := map[string]Componenter{}
	for k, v := range p {
		c[k] = v.Copy()
	}
	return c
}
