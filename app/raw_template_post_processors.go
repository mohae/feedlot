package app

import (
	"fmt"
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

func postProcessorErr(p PostProcessor, err error) error {
	return fmt.Errorf("%s post-processor error: %s", p.String(), err)
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
	// Convert the existing postProcessors to Componenter.
	var oldC = make(map[string]Componenter, len(r.PostProcessors))
	oldC = DeepCopyMapStringPostProcessor(r.PostProcessors)
	// Convert the new postProcessors to Componenter
	var newC = make(map[string]Componenter, len(newP))
	newC = DeepCopyMapStringPostProcessor(newP)
	// Get the all keys from both maps
	var keys []string
	keys = mergeKeysFromComponentMaps(oldC, newC)
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
func (r *rawTemplate) createPostProcessors() (pp []interface{}, err error) {
	if r.PostProcessorIDs == nil || len(r.PostProcessorIDs) <= 0 {
		return nil, nil
	}
	var tmpS map[string]interface{}
	var ndx int
	pp = make([]interface{}, len(r.PostProcessorIDs))
	// Generate the postProcessor for each postProcessor type.
	for _, ID := range r.PostProcessorIDs {
		tmpPP, ok := r.PostProcessors[ID]
		if !ok {
			return nil, fmt.Errorf("post-processor configuration for %s not found", ID)
		}
		typ := PostProcessorFromString(tmpPP.Type)
		switch typ {
		case Atlas:
			tmpS, err = r.createAtlas(ID)
			if err != nil {
				return nil, postProcessorErr(Atlas, err)
			}
		case Compress:
			tmpS, err = r.createCompress(ID)
			if err != nil {
				return nil, postProcessorErr(Compress, err)
			}
		case DockerImport:
			tmpS, err = r.createDockerImport(ID)
			if err != nil {
				return nil, postProcessorErr(DockerImport, err)
			}
		case DockerPush:
			tmpS, err = r.createDockerPush(ID)
			if err != nil {
				return nil, postProcessorErr(DockerPush, err)
			}
		case DockerSave:
			tmpS, err = r.createDockerSave(ID)
			if err != nil {
				return nil, postProcessorErr(DockerSave, err)
			}
		case DockerTag:
			tmpS, err = r.createDockerTag(ID)
			if err != nil {
				return nil, postProcessorErr(DockerTag, err)
			}
		case Vagrant:
			tmpS, err = r.createVagrant(ID)
			if err != nil {
				return nil, postProcessorErr(Vagrant, err)
			}
		case VagrantCloud:
			// Create the settings
			tmpS, err = r.createVagrantCloud(ID)
			if err != nil {
				return nil, postProcessorErr(VagrantCloud, err)
			}
		case VSphere:
			tmpS, err = r.createVSphere(ID)
			if err != nil {
				return nil, postProcessorErr(VSphere, err)
			}
		default:
			return nil, postProcessorErr(UnsupportedPostProcessor, fmt.Errorf("%q is not supported", tmpPP.Type))
		}
		pp[ndx] = tmpS
		ndx++
	}
	return pp, nil
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
func (r *rawTemplate) createAtlas(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
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
	for _, s := range r.PostProcessors[ID].Settings {
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
		return nil, requiredSettingErr(Atlas.String(), "artifact")
	}
	if !hasArtifactType {
		return nil, requiredSettingErr(Atlas.String(), "artifact_type")
	}
	if !hasToken {
		return nil, requiredSettingErr(Atlas.String(), "token")
	}
	for name, val := range r.PostProcessors[ID].Arrays {
		if name == "metadata" {
			array := deepcopy.Iface(val)
			if array != nil {
				settings[name] = array
			}
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
//   none
// Optional configuration options:
//   compression_level    int
//   keep_input_artifact  bool
//   output               string
func (r *rawTemplate) createCompress(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = Compress.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.PostProcessors[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "compression_level":
			i, err := strconv.Atoi(v)
			if err != nil {
				return errPostProcessor(Compress, err)
			}
			settings[k] = i
		case "keep_input_artifact":
			// Invalid values are treated as false so the error is
			// ignored.
			settings[k], _ = strconv.ParseBool(v)
		case "output":
			settings[k] = v
		}
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
//   tag         string
// Optional configuration options:
//   none
func (r *rawTemplate) createDockerImport(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = DockerImport.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasRepository, hasTag bool
	for _, s := range r.PostProcessors[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "repository":
			settings[k] = v
			hasRepository = true
		case "tag":
			settings[k] = v
			hasTag = true
		}
	}
	if !hasRepository {
		return nil, requiredSettingErr(DockerImport.String(), "repository")
	}
	if !hasTag {
		return nil, requiredSettingerr(DockerImport.String(), "tag")
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
func (r *rawTemplate) createDockerPush(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = DockerPush.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	for _, s := range r.PostProcessors[ID].Settings {
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
func (r *rawTemplate) createDockerSave(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
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
	for _, s := range r.PostProcessors[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "path":
			settings[k] = v
			hasPath = true
		}
	}
	if !hasPath {
		return nil, requiredSettingErr(DockerSave.String(), "path")
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
//   force       bool
//   tag         string
func (r *rawTemplate) createDockerTag(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
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
	for _, s := range r.PostProcessors[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "repository":
			settings[k] = v
			hasRepository = true
		case "force":
			// Invalid values are treated as false so the error is
			// ignored.
			settings[k], _ = strconv.ParseBool(v)
		case "tag":
			settings[k] = v
		}
	}
	if !hasRepository {
		return nil, requiredSettingErr(DockerTag.String(), "repository")
	}
	return settings, nil
}

// createVagrant() creates a map of settings for Packer's Vagrant
// post-processor.  Any values that aren't supported by the Vagrant
// post-processor are ignored. For more information refer to
// https://packer.io/docs/post-processors/vagrant.html.
//
// Required configuration options:
//   none
// Optional configuration options:
//   compression_level     int
//   include               array of strings
//   keep_input_artifact   bool
//   output                string
//   vagrantfile_template  string
// Provider-Specific Overrides:
//   override	           array of strings
func (r *rawTemplate) createVagrant(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = Vagrant.String()
	// For each value, extract its key value pair and then process. Only process the supported keys.
	// Key validation isn't done here, leaving that for Packer.
	var k, v string
	for _, s := range r.PostProcessors[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "output":
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
		case "vagrantfile_template":
			// locate the file
			src, err := r.findComponentSource(Vagrant.String(), v, false)
			if err != nil {
				return nil, settingErr(k, err)
			}
			jww.ERROR.Printf("vagrantfile_template: %v", src)
			// if the source couldn't be found and an error wasn't generated, replace
			// s with the original value; this occurs when it is an example.
			// Nothing should be copied in this instancel it should not be added
			// to the copy info
			if src != "" {
				r.files[r.buildOutPath(Vagrant.String(), v)] = src
			}
			settings[k] = r.buildTemplateResourcePath(Vagrant.String(), v)
		}
	}
	// Process the Arrays.
	for name, val := range r.PostProcessors[ID].Arrays {
		switch name {
		case "except":
		case "only":
		case "override":
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
func (r *rawTemplate) createVagrantCloud(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = VagrantCloud.String()
	var hasAccessToken, hasBoxTag, hasVersion bool
	for _, s := range r.PostProcessors[ID].Settings {
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
//   datastore*      string
//   host            string
//   password        string
//   resource_pool*  string
//   username        string
//   vm_name         string
// Optional configuration options:
//   datastore       string
//   disk_mode       string
//   insecure        boolean
//   vm_folder       string
//   vm_network      string
//
// Notes:
//   * datastore is not required if resource_pool is specified
func (r *rawTemplate) createVSphere(ID string) (settings map[string]interface{}, err error) {
	_, ok := r.PostProcessors[ID]
	if !ok {
		return nil, configNotFoundErr()
	}
	settings = make(map[string]interface{})
	settings["type"] = VSphere.String()
	// For each value, extract its key value pair and then process. Only
	// process the supported keys. Key validation isn't done here, leaving
	// that for Packer.
	var k, v string
	var hasCluster, hasDatacenter, hasDatastore, hasHost, hasPassword, hasResourcePool, hasUsername, hasVMName bool
	for _, s := range r.PostProcessors[ID].Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "cluster":
			settings[k] = v
			hasCluster = true
		case "datacenter":
			settings[k] = v
			hasDatacenter = true
		case "datastore":
			settings[k] = v
			hasDatastore = true
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
		case "disk_mode", "vm_folder", "vm_network":
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
	if !hasDatastore {
		if !hasResourcePool {
			return nil, fmt.Errorf("%s; if the datastore is not set a resource_pool must be specified.", requiredSettingErr("datastore/resource_pool"))
		}
	}
	if !hasHost {
		return nil, requiredSettingErr("host")
	}
	if !hasPassword {
		return nil, requiredSettingErr("password")
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
func DeepCopyMapStringPostProcessor(p map[string]postProcessor) map[string]Componenter {
	c := map[string]Componenter{}
	for k, v := range p {
		tmpP := postProcessor{}
		tmpP = v.DeepCopy()
		c[k] = tmpP
	}
	return c
}
