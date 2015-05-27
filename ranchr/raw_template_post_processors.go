package ranchr

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
)

// PostProcessor constants
const (
	UnsupportedPostProcessor PostProcessor = iota
	Compress
	DockerImport
	DockerPush
	DockerSave
	DockerTag
	Vagrant
	VagrantCloud
	VSphere
)

// PostProcessor is a Packer supported post-processor.
type PostProcessor int

var postProcessors = [...]string{
	"unsupported post-processor",
	"compress",      // Compress is the name of the compress PostProcessor
	"docker-import", // DockerImport is the name of the DockerImport PostProcessor
	"docker-push",   // DockerPush is the name of the DockerPush PostProcessor
	"docker-save",   // DockerSave is the name of the DockerSave PostProcessor
	"docker-tag",    // DockerTag is the name of the DockerTag PostProcessor
	"vagrant",       // Vagrant is the name of the Vagrant PostProcessor
	"vagrant-cloud", // Vagrant is the name of the Vagrant CloudPostProcessor
	"vsphere",       // VSphere is the name of the VSphere PostProcessor
}

func (p PostProcessor) String() string { return postProcessors[p] }

// PostProcessorFromString returns the PostProcessor constant for the passed
// string, or unsupported. All incoming strings are normalized to lowercase.
func PostProcessorFromString(s string) PostProcessor {
	s = strings.ToLower(s)
	switch s {
	case "compress":
		return Compress
	case "docker-import":
		return DockerImport
	case "docker-push":
		return DockerPush
	case "docker-save":
		return DockerSave
	case "docker-tag":
		return DockerTag
	case "vagrant":
		return Vagrant
	case "vagrant-cloud":
		return VagrantCloud
	case "vsphere":
		return VSphere
	}
	return UnsupportedPostProcessor
}

// Merges the new config with the old. The updates occur as follows:
//
//	* The existing configuration is used when no `new` postProcessors are
//	  specified.
//	* When 1 or more `new` postProcessors are specified, they will replace
//        all existing postProcessors. In this situation, if a postProcessor
//	  exists in the `old` map but it does not exist in the `new` map, that
//        postProcessor will be orphaned.
// If there isn't a new config, return the existing as there are no
// overrides
func (r *rawTemplate) updatePostProcessors(new map[string]*postProcessor) {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return
	}
	// Convert the existing postProcessors to interface.
	var ifaceOld = make(map[string]interface{}, len(r.PostProcessors))
	ifaceOld = DeepCopyMapStringPPostProcessor(r.PostProcessors)
	// Convert the new postProcessors to interfaces
	var ifaceNew = make(map[string]interface{}, len(new))
	ifaceNew = DeepCopyMapStringPPostProcessor(new)
	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	p := &postProcessor{}
	if r.PostProcessors == nil {
		r.PostProcessors = map[string]*postProcessor{}
	}
	// Copy: if the key exists in the new postProcessors only.
	// Ignore: if the key does not exist in the new postProcessors.
	// Merge: if the key exists in both the new and old postProcessors.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		_, ok := r.PostProcessors[v]
		if !ok {
			r.PostProcessors[v] = new[v].DeepCopy()
			continue
		}
		// If the element for this key doesn't exist, skip it.
		_, ok = new[v]
		if !ok {
			continue
		}
		p = r.PostProcessors[v].DeepCopy()
		if p == nil {
			p = &postProcessor{templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
		}
		p.mergeSettings(new[v].Settings)
		p.mergeArrays(new[v].Arrays)
		r.PostProcessors[v] = p
	}
}

// Go through all of the Settings and convert them to a map. Each setting
// is parsed into its constituent parts. The value then goes through
// variable replacement to ensure that the settings are properly resolved.
func (p *postProcessor) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	var k string
	var v interface{}
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type
	for _, s := range p.Settings {
		k, v = parseVar(s)
		switch k {
		// If its not set to 'true' then false. This is a case insensitive
		// comparison.
		// TODO why am I using fmt.Sprint(v) here?
		case "keep_input_artifact":
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

// r.createPostProcessors creates the PostProcessors for a build.
func (r *rawTemplate) createPostProcessors() (p []interface{}, vars map[string]interface{}, err error) {
	if r.PostProcessorTypes == nil || len(r.PostProcessorTypes) <= 0 {
		err = fmt.Errorf("unable to create post-processors: none specified")
		return nil, nil, err
	}
	var vrbls, tmpVar []string
	var tmpS map[string]interface{}
	var ndx int
	p = make([]interface{}, len(r.PostProcessorTypes))
	// Generate the postProcessor for each postProcessor type.
	for _, pType := range r.PostProcessorTypes {
		// TODO calculate the length of the two longest Settings sections
		// and make it that length. That will prevent a panic unless
		// there are more than 50 options. Besides its stupid, on so many
		// levels, to hard code this...which makes me...d'oh!
		tmpVar = make([]string, 50)
		tmpS = make(map[string]interface{})
		typ := PostProcessorFromString(pType)
		switch typ {
		case Compress:
			tmpS, tmpVar, err = r.createCompress()
			if err != nil {
				return nil, nil, err
			}
		case DockerImport:
			tmpS, tmpVar, err = r.createDockerImport()
			if err != nil {
				return nil, nil, err
			}
		case DockerPush:
			tmpS, tmpVar, err = r.createDockerPush()
			if err != nil {
				return nil, nil, err
			}
		case DockerSave:
			tmpS, tmpVar, err = r.createDockerSave()
			if err != nil {
				return nil, nil, err
			}
		case DockerTag:
			tmpS, tmpVar, err = r.createDockerTag()
			if err != nil {
				return nil, nil, err
			}
		case Vagrant:
			tmpS, tmpVar, err = r.createVagrant()
			if err != nil {
				return nil, nil, err
			}
		case VagrantCloud:
			// Create the settings
			tmpS, tmpVar, err = r.createVagrantCloud()
			if err != nil {
				return nil, nil, err
			}
		case VSphere:
			tmpS, tmpVar, err = r.createVSphere()
			if err != nil {
				return nil, nil, err
			}
		default:
			err = fmt.Errorf("%s is not supported", pType)
			return nil, nil, err
		}
		p[ndx] = tmpS
		ndx++
		vrbls = append(vrbls, tmpVar...)
	}
	return p, vars, nil
}

// createCompress() creates a map of settings for Packer's compress post-processor.
//  Any values that aren't supported by the compress post-processor are ignored. For
// more information refer to https://packer.io/docs/post-processors/compress.html
//
// Required configuration options:
//   output  string
// Optional configuration options:
//   none
func (r *rawTemplate) createCompress() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[Compress.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", Compress.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = Compress.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasOutput bool
	for _, s := range r.PostProcessors[Compress.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "output":
			settings[k] = v
			hasOutput = true
		}
	}
	if !hasOutput {
		err := fmt.Errorf("\"output\" setting is required for compress, not found")
		return nil, nil, err
	}
	return settings, vars, nil
}

// createDockerImport() creates a map of settings for Packer's docker-import
// post-processor.  Any values that aren't supported by the docker-import
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/docker-import.html.
//
// Required configuration options:
//   repository  string
// Optional configuration options:
//   tag         string
func (r *rawTemplate) createDockerImport() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[DockerImport.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", DockerImport.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = DockerImport.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasRepository bool
	for _, s := range r.PostProcessors[DockerImport.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "repository":
			settings[k] = v
			hasRepository = true
		case "tag":
			settings[k] = v
		}
	}
	if !hasRepository {
		err := fmt.Errorf("\"repository\" setting is required for docker-import, not found")
		return nil, nil, err
	}
	return settings, vars, nil
}

// createDockerPush() creates a map of settings for Packer's docker-push
// post-processor.  Any values that aren't supported by the docker-push
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/docker-push.html.
//
// Required configuration options:
//   none
// Optional configuration options:
//   login           boolean
//   login_email     string
//   login_username  string
//   login_password  string
//   login_server    string
func (r *rawTemplate) createDockerPush() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[DockerPush.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", DockerPush.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = DockerPush.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.PostProcessors[DockerPush.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "login_email", "login_username", "login_password", "login_server":
			settings[k] = v
		case "login":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	return settings, vars, nil
}

// createDockerSave() creates a map of settings for Packer's docker-save
// post-processor.  Any values that aren't supported by the docker-save
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/docker-save.html.
//
// Required configuration options:
//   path  // string
// Optional configuration options:
//   none
func (r *rawTemplate) createDockerSave() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[DockerSave.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", DockerSave.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = DockerSave.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasPath bool
	for _, s := range r.PostProcessors[DockerSave.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "path":
			settings[k] = v
			hasPath = true
		}
	}
	if !hasPath {
		err := fmt.Errorf("\"path\" setting is required for docker-save, not found")
		return nil, nil, err
	}
	return settings, vars, nil
}

// createDockerTag() creates a map of settings for Packer's docker-tag
// post-processor.  Any values that aren't supported by the docker-tag
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/docker-tag.html.
//
// Required configuration options:
//   repository  string
// Optional configuration options:
//   tag         string
func (r *rawTemplate) createDockerTag() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[DockerTag.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", DockerTag.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = DockerTag.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasRepository bool
	for _, s := range r.PostProcessors[DockerTag.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "repository":
			settings[k] = v
			hasRepository = true
		case "tag":
			settings[k] = v
		}
	}
	if !hasRepository {
		err := fmt.Errorf("\"repository\" setting is required for docker-tag, not found")
		return nil, nil, err
	}
	return settings, vars, nil
}

// createVagrant() creates a map of settings for Packer's Vagrant post-processor.
//  Any values that aren't supported by the Vagrant post-processor are ignored. For
//  more information refer to https://packer.io/docs/post-processors/vagrant.html.
//
// Configuration options:
//   compression_level     integer
//   include               array of strings
//   keep_input_artifact   boolean
//   output                string
//   vagrantfile_template  string
// Provider-Specific Overrides:
//   override	              array of strings
func (r *rawTemplate) createVagrant() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[Vagrant.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", Vagrant.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = Vagrant.String()
	// For each value, extract its key value pair and then process. Only process the supported keys.
	// Key validation isn't done here, leaving that for Packer.
	var k, v string
	for _, s := range r.PostProcessors[Vagrant.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "output", "vagrantfile_template":
			settings[k] = v
		case "keep_input_artifact":
			settings[k], _ = strconv.ParseBool(v)
		case "compression_level":
			// only add if its an int
			i, err := strconv.Atoi(v)
			if err != nil {
				err = fmt.Errorf("Vagrant builder error while trying to set %q to %q: %s", k, v, err)
				return nil, nil, err
			}
			settings[k] = i
		}
	}
	// Process the Arrays.
	for name, val := range r.PostProcessors[Vagrant.String()].Arrays {
		array := deepcopy.Iface(val)
		if array != nil {
			settings[name] = array
		}
	}
	return settings, vars, nil
}

// createVagrantCloud() creates a map of settings for Packer's Vagrant-Cloud
// post-processor.  Any values that aren't supported by the Vagrant-Cloud
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/vagrant-cloud.html
//
// Required configuration options:
//   access_token         string
//   box_tag              string
//   version              string
// Optional configuration options
//   no_release           string
//   vagrant_cloud_url    string
//   version_description  string
//   box_download_url     string
func (r *rawTemplate) createVagrantCloud() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[VagrantCloud.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", VagrantCloud.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = VagrantCloud.String()
	var hasAccessToken, hasBoxTag, hasVersion bool
	for _, s := range r.PostProcessors[VagrantCloud.String()].Settings {
		k, v := parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "access_token":
			settings[k] = v
			hasAccessToken = true
		case "box_tag":
			settings[k] = v
			hasBoxTag = true
		case "version":
			settings[k] = v
			hasVersion = true
		case "box_download_url", "no_release", "vagrant_cloud_url", "version_description":
			settings[k] = v
		}
	}
	if !hasAccessToken {
		err := fmt.Errorf("\"access_token\" setting is required for vagrant-cloud, not found")
		return nil, nil, err
	}
	if !hasBoxTag {
		err := fmt.Errorf("\"box_tag\" setting is required for vagrant-cloud, not found")
		return nil, nil, err
	}
	if !hasVersion {
		err := fmt.Errorf("\"version\" setting is required for vagrant-cloud, not found")
		return nil, nil, err
	}
	return settings, nil, nil
}

// createvSphere() creates a map of settings for Packer's vSphere post-processor.
// Any values that aren't supported by the vSphere post-processor are ignored. For
// more information refer to https://packer.io/docs/post-processors/vsphere.html.
//
// Required configuration options:
//   cluster         string
//   datacenter      string
//   host            string
//   password        string
//   resource_pool   string
//   username        string
//   vm_name         string
// Optional configuration options:
//   datastore       string
//   disk_mode       string
//   insecure        boolean
//   vm_folder       string
//   vm_network      string
func (r *rawTemplate) createVSphere() (settings map[string]interface{}, vars []string, err error) {
	_, ok := r.PostProcessors[VSphere.String()]
	if !ok {
		err = fmt.Errorf("no configuration found for %q", VSphere.String())
		return nil, nil, err
	}
	settings = make(map[string]interface{})
	settings["type"] = VSphere.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasCluster, hasDatacenter, hasHost, hasPassword, hasResourcePool, hasUsername, hasVMName bool
	for _, s := range r.PostProcessors[VSphere.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "cluster":
			settings[k] = v
			hasCluster = true
		case "datacenter":
			settings[k] = v
			hasDatacenter = true
		case "host":
			settings[k] = v
			hasHost = true
		case "password":
			settings[k] = v
			hasPassword = true
		case "resource_pool":
			settings[k] = v
			hasResourcePool = true
		case "username":
			settings[k] = v
			hasUsername = true
		case "vm_name":
			settings[k] = v
			hasVMName = true
		case "datastore", "disk_mode", "vm_folder", "vm_network":
			settings[k] = v
		case "insecure":
			settings[k], _ = strconv.ParseBool(v)
		}
	}
	if !hasCluster {
		err := fmt.Errorf("\"cluster\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	if !hasDatacenter {
		err := fmt.Errorf("\"datacenter\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	if !hasHost {
		err := fmt.Errorf("\"host\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	if !hasPassword {
		err := fmt.Errorf("\"password\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	if !hasResourcePool {
		err := fmt.Errorf("\"resource_pool\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	if !hasUsername {
		err := fmt.Errorf("\"username\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	if !hasVMName {
		err := fmt.Errorf("\"vm_name\" setting is required for vSphere, not found")
		return nil, nil, err
	}
	return settings, vars, nil
}

// DeepCopyMapStringPPostProcessor makes a deep copy of each builder passed and
// returns the copie map[string]*builder as a map[string]interface{}
// notes: This currently only supports string slices.
func DeepCopyMapStringPPostProcessor(p map[string]*postProcessor) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range p {
		tmpP := &postProcessor{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}
