package ranchr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// Merges the new config with the old. The updates occur as follows:
//
//	* The existing configuration is used when no `new` provisioners are
//	  specified.
//	* When 1 or more `new` provisioner are specified, they will replace all
//	  existing provisioner. In this situation, if a provisioner exists in
//	  the `old` map but it does not exist in the `new` map, that
//	  provisioner will be orphaned.
// If there isn't a new config, return the existing as there are no
// overrides
func (r *rawTemplate) updateProvisioners(new map[string]*provisioner) {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return
	}
	// Convert the existing provisioners to interface.
	var ifaceOld = make(map[string]interface{}, len(r.Provisioners))
	ifaceOld = DeepCopyMapStringPProvisioner(r.Provisioners)
	// Convert the new provisioners to interface.
	var ifaceNew = make(map[string]interface{}, len(new))
	ifaceNew = DeepCopyMapStringPProvisioner(new)
	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	p := &provisioner{}
	if r.Provisioners == nil {
		r.Provisioners = map[string]*provisioner{}
	}
	// Copy: if the key exists in the new provisioners only.
	// Ignore: if the key does not exist in the new provisioners.
	// Merge: if the key exists in both the new and old provisioners.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		_, ok := r.Provisioners[v]
		if !ok {
			r.Provisioners[v] = new[v].DeepCopy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		_, ok = new[v]
		if !ok {
			continue
		}
		p = r.Provisioners[v].DeepCopy()
		if p == nil {
			p = &provisioner{templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
		}
		p.mergeSettings(new[v].Settings)
		p.mergeArrays(new[v].Arrays)
		r.Provisioners[v] = p
	}
}

// Go through all of the Settings and convert them to a map. Each setting
// is parsed into its constituent parts. The value then goes through
// variable replacement to ensure that the settings are properly resolved.
func (p *provisioner) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	var k string
	var v interface{}
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type
	for _, s := range p.Settings {
		k, v = parseVar(s)
		switch k {
		// If its not set to 'true' then false. This is a case insensitive
		// comparison.
		case "binary":
			if strings.ToLower(fmt.Sprint(v)) == "true" {
				v = true
			} else {
				v = false
			}
		default:
			v = r.replaceVariables(fmt.Sprint(v))
		}
		m[k] = v
	}
	return m
}

// r.createProvisioner creates the provisioners for a build.
func (r *rawTemplate) createProvisioners() (p []interface{}, vars map[string]interface{}, err error) {
	if r.ProvisionerTypes == nil || len(r.ProvisionerTypes) <= 0 {
		err = fmt.Errorf("unable to create provisioners: none specified")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.ProvisionerTypes))
	// Generate the postProcessor for each postProcessor type.
	for _, pType := range r.ProvisionerTypes {
		// TODO calculate the length of the two longest Settings sections
		// and make it that length. That will prevent a panic unless
		// there are more than 50 options. Besides its stupid, on so many
		// levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})
		typ := ProvisionerFromString(pType)
		switch typ {
		case Ansible:
			tmpS, tmpVar, err = r.createAnsible()
		case FileUploads:
			tmpS, tmpVar, err = r.createFileUploads()
		case Salt:
			tmpS, tmpVar, err = r.createSalt()
		case ShellScripts:
			tmpS, tmpVar, err = r.createShellScripts()
			/*
				case ChefClient:
					// not implemented
				case ChefSolo:
					// not implemented
				case PuppetClient:
					// not implemented
				case PuppetServer:
					// not implemented
			*/
		default:
			err = fmt.Errorf("%s provisioner is not supported", pType)
			jww.ERROR.Println(err)
			return nil, nil, err
		}
		p[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}
	return p, vars, nil
}

// createAnsible() creates a map of settings for Packer's ansible-local
// provisioner. Any values that aren't supported by the file provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/ansible-local.html
//
// Required configuration options:
//   playbook_file		// string
// Optional configuration options:
//   command			// string
//   extra_arguments	// array of strings
//   inventory_file		// string
//   group_vars			// string
//   host_vars			// string
//   playbook_dir		// string
//   playbook_paths		// array of strings
//   role_paths			// array of strings
//   staging_directory	// string
func (r *rawTemplate) createAnsible() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[Ansible.String()]
	if !ok {
		err = fmt.Errorf("no configuration for %q found", Ansible.String())
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
		switch k {
		case "playbook_file":
			settings[k] = v
			hasPlaybook = true
		case "command", "inventory_file", "group_vars", "host_vars", "playbook_dir", "staging_directory":
			settings[k] = v
		default:
			jww.WARN.Println("unsupported " + Ansible.String() + " key was encountered: " + k)
		}
	}
	if !hasPlaybook {
		err := fmt.Errorf("\"playbook_file\" setting is required for ansible-local, not found")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[Ansible.String()].Arrays {
		array := deepcopy.InterfaceToSliceStrings(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, vars, err
}

// createSalt() creates a map of settings for Packer's salt-masterless
// provisioner. Any values that aren't supported by the salt-masterless
// provisioner are ignored. For more information, refer to
// https://packer.io/docs/provisioners/salt-masterless.html
//
// Required configuration options:
//   local_state_tree		// string
// Optional configuration options
//   bootstrap_args			// string
//   local_pillar_roots		// string
//   minion_config			// string
//   skip_bootstrap			// boolean
//   temp_config_dir			// string
func (r *rawTemplate) createSalt() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[Salt.String()]
	if !ok {
		err = fmt.Errorf("no configuration for %q found", Salt.String())
	}
	settings = make(map[string]interface{})
	settings["type"] = Salt.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.Provisioners[Salt.String()].Settings {
		k, v = parseVar(s)
		switch k {
		case "bootstrap_args", "local_pillar_roots", "local_state_tree", "minion_config", "temp_config_dir":
			settings[k] = v
		case "skip_bootstrap":
			settings[k], _ = strconv.ParseBool(v)
		default:
			jww.WARN.Println("unsupported " + Salt.String() + " key was encountered: " + k)
		}
	}
	// salt-masterless does not have any arrays to support
	return settings, vars, nil
}

// createShellScripts() creates a map of settings for Packer's shell
// provisioner. Any values that aren't supported by the shell provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/shell.html
//
// Of the "inline", "script", and "scripts" options, only "scripts" is
// currently supported.
//
// Required configuration options:
//   scripts				// array of strings
// Optional confinguration parameters:
//   binary					// boolean
//   environment_vars		// array of strings
//   execute_command		// string
//   inline_shebang			// string
//   remote_path			// string
//   start_retry_timeout	// string
func (r *rawTemplate) createShellScripts() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[ShellScripts.String()]
	if !ok {
		err = fmt.Errorf("no configuration for %q found", ShellScripts.String())
	}
	settings = make(map[string]interface{})
	settings["type"] = ShellScripts.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.Provisioners[ShellScripts.String()].Settings {
		k, v = parseVar(s)
		switch k {
		case "execute_command", "inline_shebang", "remote_path", "start_retry_timeout":
			settings[s] = v
		case "binary":
			settings[k], _ = strconv.ParseBool(v)
		default:
			jww.WARN.Println("unsupported " + ShellScripts.String() + " key was encountered: " + k)
		}
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[ShellScripts.String()].Arrays {
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, vars, nil
}

// createFileUploads() creates a map of settings for Packer's file uploads
// provisioner. Any values that aren't supported by the file provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/file.html
//
//	Required configuration options:
//   destination	// string
//   source			// string
func (r *rawTemplate) createFileUploads() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[FileUploads.String()]
	if !ok {
		err = fmt.Errorf("no configuration for %q found", FileUploads.String())
	}
	settings = make(map[string]interface{})
	settings["type"] = FileUploads.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.Provisioners[FileUploads.String()].Settings {
		k, v = parseVar(s)
		switch k {
		case "source", "destination":
			settings[k] = v
		default:
			jww.WARN.Printf("unsupported %s key was encountered: %q", FileUploads.String(), k)
		}
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[FileUploads.String()].Arrays {
		array := deepcopy.InterfaceToSliceStrings(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, vars, nil
}

// DeepCopyMapStringPProvisioner makes a deep copy of each builder passed and
// returns the copie map[string]*provisioner as a map[string]interface{}
// notes: This currently only supports string slices.
func DeepCopyMapStringPProvisioner(p map[string]*provisioner) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range p {
		tmpP := &provisioner{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}
