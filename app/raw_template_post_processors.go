package app

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mohae/utilitybelt/deepcopy"
)

// PostProcessor constants
const (
	UnsupportedPostProcessor PostProcessor = iota
	Atlas
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
	"unsupported",
	"atlas",
	"compress",
	"docker-import",
	"docker-push",
	"docker-save",
	"docker-tag",
	"vagrant",
	"vagrant-cloud",
	"vsphere",
}

func (p PostProcessor) String() string { return postProcessors[p] }

// PostProcessorFromString returns the PostProcessor constant for the passed
// string, or unsupported. All incoming strings are normalized to lowercase.
func PostProcessorFromString(s string) PostProcessor {
	s = strings.ToLower(s)
	switch s {
	case "atlas":
		return Atlas
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
//   * The existing configuration is used when no `new` postProcessors are
//     specified.
//   * When 1 or more `new` postProcessors are specified, they will replace all
//     existing postProcessors.  In this situation, if a postProcessor exists
//     in the `old` map but it does not exist in the `new` map, that
//     postProcessor will be orphaned.
//   * If there isn't a new config, the existing one is used
func (r *rawTemplate) updatePostProcessors(newP map[string]postProcessor) error {
	// If there is nothing new, old equals merged.
	if len(newP) == 0 || newP == nil {
		return nil
	}
	// Convert the existing postProcessors to interface.
	var ifaceOld = make(map[string]interface{}, len(r.PostProcessors))
	ifaceOld = DeepCopyMapStringPostProcessor(r.PostProcessors)
	// Convert the new postProcessors to interfaces
	var ifaceNew = make(map[string]interface{}, len(newP))
	ifaceNew = DeepCopyMapStringPostProcessor(newP)
	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)
	if r.PostProcessors == nil {
		r.PostProcessors = map[string]postProcessor{}
	}
	// Copy: if the key exists in the new postProcessors only.
	// Ignore: if the key does not exist in the new postProcessors.
	// Merge: if the key exists in both the new and old postProcessors.
	for _, v := range keys {
		// If it doesn't exist in the old builder, add it.
		p, ok := r.PostProcessors[v]
		if !ok {
			pp, _ := newP[v]
			r.PostProcessors[v] = pp.DeepCopy()
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
		r.PostProcessors[v] = p
	}
	return nil
}

// r.createPostProcessors creates the PostProcessors for a build.
func (r *rawTemplate) createPostProcessors() (p []interface{}, err error) {
	if r.PostProcessorTypes == nil || len(r.PostProcessorTypes) <= 0 {
		return nil, nil
	}
	var ndx int
	p = make([]interface{}, len(r.PostProcessorTypes))
	// Generate the postProcessor for each postProcessor type.
	for _, pType := range r.PostProcessorTypes {
		tmpS := make(map[string]interface{})
		typ := PostProcessorFromString(pType)
		switch typ {
		case Atlas:
			tmpS, err = r.createAtlas()
			if err != nil {
				return nil, postProcessorErr(Atlas, err)
			}
		case Compress:
			tmpS, err = r.createCompress()
			if err != nil {
				return nil, postProcessorErr(Compress, err)
			}
		case DockerImport:
			tmpS, err = r.createDockerImport()
			if err != nil {
				return nil, postProcessorErr(DockerImport, err)
			}
		case DockerPush:
			tmpS, err = r.createDockerPush()
			if err != nil {
				return nil, postProcessorErr(DockerPush, err)
			}
		case DockerSave:
			tmpS, err = r.createDockerSave()
			if err != nil {
				return nil, postProcessorErr(DockerSave, err)
			}
		case DockerTag:
			tmpS, err = r.createDockerTag()
			if err != nil {
				return nil, postProcessorErr(DockerTag, err)
			}
		case Vagrant:
			tmpS, err = r.createVagrant()
			if err != nil {
				return nil, postProcessorErr(Vagrant, err)
			}
		case VagrantCloud:
			// Create the settings
			tmpS, err = r.createVagrantCloud()
			if err != nil {
				return nil, postProcessorErr(VagrantCloud, err)
			}
		case VSphere:
			tmpS, err = r.createVSphere()
			if err != nil {
				return nil, postProcessorErr(VSphere, err)
			}
		default:
			return nil, postProcessorErr(UnsupportedPostProcessor, fmt.Errorf("%q is not supported", pType))
		}
		p[ndx] = tmpS
		ndx++
	}
	return p, nil
}

// createAtlas() creates a map of settings for Packer's atlas post-processor.
// Any values that aren't supported by the atlas post-processor are ignored.
// For more information refer to
// https://packer.io/docs/post-processors/compress.html
//
// Required configuration options:
//   artifact      string
//   artifact_type string
//   token         string
// Optional configuration options:
//   atlas_url     string
//   metadata      object of key/value strings
func (r *rawTemplate) createAtlas() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[Atlas.String()]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = Atlas.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasArtifact, hasArtifactType, hasToken bool
	for _, s := range r.PostProcessors[Atlas.String()].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "artifact":
			settings[k] = v
			hasArtifact = true
		case "artifact_type":
			settings[k] = v
			hasArtifactType = true
		case "token":
			settings[k] = v
			hasToken = true
		case "atlas_url":
			settings[k] = v
		}
	}
	if !hasArtifact {
		return nil, requiredSettingErr("artifact")
	}
	if !hasArtifactType {
		return nil, requiredSettingErr("artifact_type")
	}
	if !hasToken {
		return nil, requiredSettingErr("token")
	}
	for name, val := range r.PostProcessors[Atlas.String()].Arrays {
		if name == "metadata" {
			settings[name] = val
		}
	}
	return settings, nil
}

// createCompress() creates a map of settings for Packer's compress
// post-processor.  Any values that aren't supported by the compress
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/compress.html
//
// Required configuration options:
//   output  string
// Optional configuration options:
//   none
func (r *rawTemplate) createCompress() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[Compress.String()]
	if !ok {
		return nil, configNotFoundErr()
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
		return nil, requiredSettingErr("output")
	}
	return settings, nil
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
func (r *rawTemplate) createDockerImport() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[DockerImport.String()]
	if !ok {
		return nil, configNotFoundErr()
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
		return nil, requiredSettingErr("repository")
	}
	return settings, nil
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
func (r *rawTemplate) createDockerPush() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[DockerPush.String()]
	if !ok {
		return nil, configNotFoundErr()
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
	return settings, nil
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
func (r *rawTemplate) createDockerSave() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[DockerSave.String()]
	if !ok {
		return nil, configNotFoundErr()
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
		return nil, requiredSettingErr("path")
	}
	return settings, nil
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
func (r *rawTemplate) createDockerTag() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[DockerTag.String()]
	if !ok {
		return nil, configNotFoundErr()
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
		return nil, requiredSettingErr("repository")
	}
	return settings, nil
}

// createVagrant() creates a map of settings for Packer's Vagrant
// post-processor.  Any values that aren't supported by the Vagrant
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/vagrant.html.
//
// Configuration options:
//   compression_level     integer
//   include               array of strings
//   keep_input_artifact   boolean
//   output                string
//   vagrantfile_template  string
// Provider-Specific Overrides:
//   override	              array of strings
func (r *rawTemplate) createVagrant() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[Vagrant.String()]
	if !ok {
		return nil, configNotFoundErr()
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
				return nil, settingErr(k, err)
			}
			settings[k] = i
		}
	}
	// Process the Arrays.
	for name, val := range r.PostProcessors[Vagrant.String()].Arrays {
		switch name {
		case "include":
			array := deepcopy.InterfaceToSliceOfStrings(val)
			for i, v := range array {
				v = r.replaceVariables(v)
				src, err := r.findComponentSource(Vagrant.String(), v)
				if err != nil {
					return nil, settingErr(v, err)
				}
				array[i] = v
				r.files[filepath.Join(r.OutDir, Vagrant.String(), v)] = src
			}
			settings[name] = array
		case "override", "except", "only":
			array := deepcopy.Iface(val)
			if array != nil {
				settings[name] = array
			}
		}
	}
	return settings, nil
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
func (r *rawTemplate) createVagrantCloud() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[VagrantCloud.String()]
	if !ok {
		return nil, configNotFoundErr()
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
		return nil, requiredSettingErr("access_token")
	}
	if !hasBoxTag {
		return nil, requiredSettingErr("box_tag")
	}
	if !hasVersion {
		return nil, requiredSettingErr("version")
	}
	return settings, nil
}

// createVSphere() creates a map of settings for Packer's vSphere
// post-processor.  Any values that aren't supported by the vSphere
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/vsphere.html.
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
func (r *rawTemplate) createVSphere() (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[VSphere.String()]
	if !ok {
		return nil, configNotFoundErr()
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
		return nil, requiredSettingErr("cluster")
	}
	if !hasDatacenter {
		return nil, requiredSettingErr("datacenter")
	}
	if !hasHost {
		return nil, requiredSettingErr("host")
	}
	if !hasPassword {
		return nil, requiredSettingErr("password")
	}
	if !hasResourcePool {
		return nil, requiredSettingErr("resource_pool")
	}
	if !hasUsername {
		return nil, requiredSettingErr("username")
	}
	if !hasVMName {
		return nil, requiredSettingErr("vm_name")
	}
	return settings, nil
}

// Go through all of the Settings and convert them to a map. Each setting is
// parsed into its constituent parts. The value then goes through variable
// replacement to ensure that the settings are properly resolved.
func (p *postProcessor) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	var k string
	var v interface{}
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type
	for _, s := range p.Settings {
		k, v = parseVar(s)
		switch k {
		case "keep_input_artifact":
			v, _ = strconv.ParseBool(v.(string))
		default:
			v = r.replaceVariables(v.(string))
		}
		m[k] = v
	}
	return m
}

// DeepCopyMapStringPostProcessor makes a deep copy of each builder passed and
// returns the copie map[string]postProcessor as a map[string]interface{}
// Note: This currently only supports string slices.
func DeepCopyMapStringPostProcessor(p map[string]postProcessor) map[string]interface{} {
	c := map[string]interface{}{}
	for k, v := range p {
		tmpP := postProcessor{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}
