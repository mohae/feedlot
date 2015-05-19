package ranchr

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// Provisioner constants
const (
	UnsupportedProvisioner Provisioner = iota
	AnsibleLocal
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
	"unsupported provisioner",
	"ansible-local",     //ansible is the name of the Ansible Provisioner
	"chef-client",       //chef-client is the name of the ChefClient Provisioner
	"chef-solo",         //chef-solo is the name of the ChefSolo Provisioner
	"file-uploads",      //file-uploads is the name of the FileUploads Provisioner
	"puppet-masterless", //puppet-masterless is the name of the PuppetMasterless Provisioner
	"puppet-server",     // puppet-server is the name of the PuppetServer Provisioner
	"salt-masterless",   //salt is the name of the Salt Provisioner
	"shell",             // shell is the name for the Shell provisioner
}

func (p Provisioner) String() string { return provisioners[p] }

// ProvisionerFromString returns the Provisioner constant for the passed string or
// unsupported. All incoming strings are normalized to lowercase
func ProvisionerFromString(s string) Provisioner {
	s = strings.ToLower(s)
	switch s {
	case "ansible-local":
		return AnsibleLocal
	case "chef-client":
		return ChefClient
	case "chef-solo":
		return ChefSolo
	case "file-uploads":
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
		case AnsibleLocal:
			tmpS, tmpVar, err = r.createAnsibleLocal()
			if err != nil {
				return nil, nil, err
			}
		case FileUploads:
			tmpS, tmpVar, err = r.createFileUploads()
			if err != nil {
				return nil, nil, err
			}
		case Salt:
			tmpS, tmpVar, err = r.createSalt()
			if err != nil {
				return nil, nil, err
			}
		case ShellScript:
			tmpS, tmpVar, err = r.createShellScript()
			if err != nil {
				return nil, nil, err
			}
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

// createAnsibleLocal() creates a map of settings for Packer's ansible-local
// provisioner. Any values that aren't supported by the file provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/ansible-local.html
//
// Required configuration options:
//   playbook_file		// string
// Optional configuration options:
//   command			    // string
//   extra_arguments	    // array of strings
//   inventory_file		// string
//   group_vars			// string
//   host_vars			// string
//   playbook_dir	         // string
//   playbook_paths		// array of strings
//   role_paths			// array of strings
//   staging_directory	// string
func (r *rawTemplate) createAnsibleLocal() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[AnsibleLocal.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", AnsibleLocal.String())
		jww.ERROR.Print(err)
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = AnsibleLocal.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasPlaybook bool
	for _, s := range r.Provisioners[AnsibleLocal.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "playbook_file":
			settings[k] = v
			hasPlaybook = true
		case "command", "inventory_file", "group_vars", "host_vars", "playbook_dir", "staging_directory":
			settings[k] = v
		default:
			jww.WARN.Printf("unsupported ansible-masterless key was encountered: %q" + k)
		}
	}
	if !hasPlaybook {
		err := fmt.Errorf("\"playbook_file\" setting is required for ansible-local, not found")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	// Process the Arrays.
	for name, val := range r.Provisioners[AnsibleLocal.String()].Arrays {
		array := deepcopy.InterfaceToSliceOfStrings(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, vars, err
}

// createFileUploads() creates a map of settings for Packer's file uploads
// provisioner. Any values that aren't supported by the file provisioner are
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/file.html
//
//	Required configuration options:
//   destination	// string
//   source		// string
func (r *rawTemplate) createFileUploads() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[FileUploads.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", FileUploads.String())
		jww.ERROR.Print(err)
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = FileUploads.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasSource, hasDestination bool
	for _, s := range r.Provisioners[FileUploads.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "source":
			settings[k] = v
			hasSource = true
		case "destination":
			settings[k] = v
			hasDestination = true
		default:
			jww.WARN.Printf("unsupported %s key was encountered: %q", FileUploads.String(), k)
		}
	}
	if !hasSource {
		err := fmt.Errorf("\"source\" setting is required for file-uploads, not found")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	if !hasDestination {
		err := fmt.Errorf("\"destination\" setting is required for file-uploads, not found")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	return settings, vars, nil
}

// createSalt() creates a map of settings for Packer's salt provisioner. Any values
// that aren't supported by the salt provisioner are logged as a WARN and are then
// ignored. For more information, refer to
// https://packer.io/docs/provisioners/salt-masterless.html
//
// Required configuration options:
//   local_state_tree		// string
// Optional configuration options
//   bootstrap_args		// string
//   local_pillar_roots	// string
//   local_state_tree     // string
//   minion_config		// string
//   skip_bootstrap		// boolean
//   temp_config_dir		// string
func (r *rawTemplate) createSalt() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[Salt.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", Salt.String())
		jww.ERROR.Print(err)
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = Salt.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasLocalStateTree bool
	for _, s := range r.Provisioners[Salt.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "local_state_tree":
			// prepend the path with salt-masterless if there isn't a parent dir
			v = setParentDir(Salt.String(), v)
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v)
			if err != nil {
				jww.ERROR.Println(err)
				return nil, nil, err
			}
			r.dirs[filepath.Join(r.OutDir, v)] = src
			settings[k] = v
			hasLocalStateTree = true
		case "local_pillar_roots":
			// prepend the path with salt-masterless if there isn't a parent dir
			v = setParentDir(Salt.String(), v)
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(v)
			if err != nil {
				jww.ERROR.Println(err)
				return nil, nil, err
			}
			r.dirs[filepath.Join(r.OutDir, v)] = src
			settings[k] = v
		case "minion_config":
			// prepend the path with salt-masterless if there isn't a parent dir
			v = setParentDir(Salt.String(), v)
			// find the actual location and add it to the files map for copying
			src, err := r.findSource(filepath.Join(v, "minion"))
			if err != nil {
				jww.ERROR.Println(err)
				return nil, nil, err
			}
			r.files[filepath.Join(r.OutDir, v, "minion")] = src
			settings[k] = v
		case "bootstrap_args", "temp_config_dir":
			settings[k] = v
		case "skip_bootstrap":
			settings[k], _ = strconv.ParseBool(v)
		default:
			jww.WARN.Println("unsupported " + Salt.String() + " key was encountered: " + k)
		}
	}
	if !hasLocalStateTree {
		err := fmt.Errorf("\"local_state_tree\" setting is required for salt, not found")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	// salt does not have any arrays to support
	return settings, vars, nil
}

// createShellScriptl() creates a map of settings for Packer's shell script
// provisioner. Any values that aren't supported by the shell provisioner generate a
// warning and are otherwise ignored. For more information, refer to
// https://packer.io/docs/provisioners/shell.html
//
// Of the "inline", "script", and "scripts" options, only "scripts" is
// currently supported.
//
// Required configuration options:
//   scripts				// array of strings
// Optional confinguration parameters:
//   binary				// boolean
//   environment_vars		// array of strings
//   execute_command		// string
//   inline_shebang		// string
//   remote_path			// string
//   start_retry_timeout	// string
func (r *rawTemplate) createShellScript() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.Provisioners[ShellScript.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", ShellScript.String())
		jww.ERROR.Print(err)
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = ShellScript.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
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
				commands, err = r.commandsFromFile(v)
				if err != nil {
					jww.ERROR.Println(err)
					return nil, nil, err
				}
				settings[k] = commands[0] // for execute_command, only the first element is used
				continue
			}
			settings[k] = v
		case "inline_shebang", "remote_path", "start_retry_timeout":
			settings[k] = v
		case "binary":
			settings[k], _ = strconv.ParseBool(v)
		default:
			jww.WARN.Println("unsupported " + ShellScript.String() + " key was encountered: " + k)
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
				// prepend the path with salt-masterless if there isn't a parent dir
				scripts[i] = setParentDir(ShellScript.String(), v)
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
		err := fmt.Errorf("\"scripts\" setting is required for shell, not found")
		jww.ERROR.Println(err)
		return nil, nil, err
	}
	// go through the scripts, find their source, and add to the files map. error if
	// the script source cannot be deteremined.
	for _, script := range scripts {
		s, err := r.findSource(script)
		if err != nil {
			jww.ERROR.Printf("error while adding file to file map: %s", err)
			return nil, nil, err
		}
		r.files[filepath.Join(r.OutDir, script)] = s
	}
	for k, v := range r.files {
		fmt.Printf("%s: %s\n", k, v)
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
