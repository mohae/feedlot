// Contains structs and methods for Rancher's configuration files.
//
// With the exception of rancher.cfg, the configuration files use TOML.
package ranchr

import (
	"errors"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// build holds most of the data needed to generate a Packer template. A build
// generically models Packer build templates.
type build struct {
	// Targeted builders: the values are consistent with Packer's, e.g.
	// `virtualbox.iso` is used for VirtualBox.
	BuilderTypes []string `toml:"builder_types"`

	// A map of builder configuration. There should always be a `common`
	// builder, which has settings common to both VMWare and VirtualBox.
	Builders map[string]*builder `toml:"builders"`

	// Targeted post-processors: the values are consistent with Packer's, e.g.
	// `vagrant` is used for Vagrant.
	PostProcessorTypes []string `toml:"post_processor_types"`

	// A map of post-processor configurations.
	PostProcessors map[string]*postProcessor `toml:"post_processors"`

	// Targeted provisioners: the values are consistent with Packer's, e.g.
	// `shell` is used for shell.
	ProvisionerTypes []string `toml:"provisioner_types"`

	// A map of provisioner configurations.
	Provisioners map[string]*provisioner `toml:"provisioners"`
}

// build.DeepCopy makes a deep copy of the build and returns it.
func (b *build) DeepCopy() build {
	copy := &build{
		BuilderTypes:       []string{},
		Builders:           map[string]*builder{},
		PostProcessorTypes: []string{},
		PostProcessors:     map[string]*postProcessor{},
		ProvisionerTypes:   []string{},
		Provisioners:       map[string]*provisioner{},
	}

	if b.BuilderTypes != nil || len(b.BuilderTypes) > 0 {
		copy.BuilderTypes = b.BuilderTypes
	}

	if b.PostProcessorTypes != nil || len(b.PostProcessorTypes) > 0 {
		copy.PostProcessorTypes = b.PostProcessorTypes
	}

	if b.ProvisionerTypes != nil || len(b.ProvisionerTypes) > 0 {
		copy.ProvisionerTypes = b.ProvisionerTypes
	}

	for k, v := range b.Builders {
		copy.Builders[k] = v.DeepCopy()
	}

	for k, v := range b.PostProcessors {
		copy.PostProcessors[k] = v.DeepCopy()
	}

	for k, v := range b.Provisioners {
		copy.Provisioners[k] = v.DeepCopy()
	}

	return *copy
}

// templateSection is used as an embedded type. All Packer build template 
// sections settings can be modeled with these elements.
type templateSection struct {
	// Settings are string settings in "key=value" format.
	Settings []string

	// Arrays are the string array settings.
	Arrays map[string]interface{}
}

// templateSection.DeepCopy updates its information with new via a deep copy.
func (t *templateSection) DeepCopy(new templateSection) {
	//Deep Copy of settings
	t.Settings = make([]string, len(new.Settings))
	copy(t.Settings, new.Settings)

	// make a deep copy of the Arrays(map[string]interface)
	t.Arrays = deepcopy.MapStringInterface(new.Arrays)
}

// builder represents a builder Packer template section.
type builder struct {
	templateSection
}

// builder.DeepCopy copies the builder values instead of the pointers.
func (b *builder) DeepCopy() *builder {
	var c *builder
	c = &builder{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(b.templateSection)
	return c
}

// mergeSettings the settings section of a builder. New values supercede existing ones.
func (b *builder) mergeSettings(new []string) {
	if new == nil {
		return
	}
	b.Settings = mergeSettingsSlices(b.Settings, new)
}

// mergeVMSettings Merge the VMSettings section of a builder. New values supercede existing ones.
//
// question sanity....
//      so this updated the builder, which then affects both builders
//	how should this work?
//      Set default template for each distro
//		b = distro template
//		d = default template
//		b should deep copy d

func (b *builder) mergedVMSettings(new []string) []string {
	if new == nil {
		return nil
	}

	var merged []string
	old := deepcopy.InterfaceToSliceString(b.Arrays[VMSettings])
	merged = mergeSettingsSlices(old, new)

	if b.Arrays == nil {
		b.Arrays = map[string]interface{}{}
	}

	return merged
}

// settingsToMap converts the Settings section, which is a []string of embedded
// key:value pairs in the form of: 'key=value'. First, the key:value pair is
// extracted from each setting. Then variable replacement is performed on each
// value so that the final, runtime value is created.
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

// postProcessor: type for handling the post-processor section of the configs.
type postProcessor struct {
	templateSection
}

// DeepCopy copies the postProcessor values instead of the pointers.
func (p *postProcessor) DeepCopy() *postProcessor {
	if p == nil {
		return nil
	}

	var c *postProcessor
	c = &postProcessor{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(p.templateSection)
	return c
}

// mergeSettings  merges the settings section of a post-processor with the
// passed slice of settings. New values supercede existing ones.
func (p *postProcessor) mergeSettings(new []string) {
	if new == nil {
		return
	}

	if p.Settings == nil {
		p.Settings = new
		return
	}

	// merge the keys
	// go through all the keys and do the appropriate action
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

// mergeArrays merges the settings section of a post-processor with the passed
// slice of settings. New values supercede existing ones.
func (p *postProcessor) mergeArrays(m map[string]interface{}) {
	if m == nil {
		return
	}

	if p.Arrays == nil {
		p.Arrays = m
	}

	// TODO merge the keys
	// go through all the keys and do the appropriate action
	//	p.Settings = mergeSettingsSlices(p.Settings, new)
}

// provisioner: type for common elements for provisioners.
type provisioner struct {
	templateSection
}

// DeepCopy copies the provisioner values instead of the pointers.
func (p *provisioner) DeepCopy() *provisioner {
	var c *provisioner
	c = &provisioner{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(p.templateSection)
	return c
}

// mergeSettings  merges the settings section of a provisioner with the passed
// slice of settings. New values supercede existing ones.
func (p *provisioner) mergeSettings(new []string) {
	if new == nil {
		return
	}
	if p.Settings == nil {
		p.Settings = new
		return
	}
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

/*
// provisioner.mergeArrays merges the passed map with the existing Arrays.
// depending on the array element, the merge may either be an update or a
// replacement of existing, e.g. some Arrays, like only and except replace
// any existing value instead of updating it.
//
// TODO figure out how to represent an unsetting of an existing element, i.e
// if p has Array["except"] and that element is missing from new, the existing
// Array["except"] will not be deleted. Another way to represent an unsetting
// needs to be done.
func (p *provisioner) mergeSettings(new map[string]interface{}) {
	if new == nil {
		return
	}

	// merge the keys

	// go through all the keys and do the appropriate action
//	p.Settings = mergeSettingsSlices(p.Settings, new)
}
*/
/*
// Go through all of the Settings and convert them to a map. Each setting is
// parsed into its constituent parts. The value then goes through variable
// replacement to ensure that the settings are properly resolved.
func (p *provisioner) settingsToMap(Type string, r *rawTemplate) (map[string]interface{}, error) {
//	var k, v string
//	var err error

	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type

	for _, s := range p.Settings {
		jww.TRACE.Printf("%v\n", s)
/*
		k, v = parseVar(s)
		v = r.replaceVariables(v)

		switch k {
		case "execute_command":
			// Get the command from the specified file
			var c []string

			if c, err = commandsFromFile(v); err != nil {
				jww.ERROR.Print(err.Error())
				return nil, err
			}
			v = c[0]
		}
*
//		m[k] = v
	}

	// Add except array.
	if p.Except != nil {
		m["except"] = p.Except
	}

	// Add only array.
	if p.Only != nil {
		m["only"] = p.Only
	}

	jww.TRACE.Printf("Provisioners Map: %v\n",m)

	return m, nil
}
*/

/*
func (p *provisioner) setScripts(new []string) {
	// Scripts are only replaced if it has values, otherwise the existing values are used.
	if len(new) > 0 {
		p.Scripts = new
	}
}
*/

// defaults is used to store Rancher application level defaults for Packer templates.
type defaults struct {
	IODirInf
	PackerInf
	BuildInf
	build
	load   sync.Once
	loaded bool
}

// LoadOnce ensures that the default configs get loaded once. Uses a mutex to 
// prevent race conditions as there can be concurrent processing of Packer 
// templates. When loaded, it sets the loaded boolean so that it only needs to
// be called when it hasn't been loaded.
func (d *defaults) LoadOnce() error {
	var err error

	loadFunc := func() {
		name := os.Getenv(EnvDefaultsFile)

		if name == "" {
			err = errors.New("could not retrieve the default Settings because the " + EnvDefaultsFile + " environment variable was not set. Either set it or check your rancher.cfg setting")
			jww.CRITICAL.Print(err.Error())
			return
		}

		if _, err = toml.DecodeFile(name, d); err != nil {
			jww.CRITICAL.Print(err.Error())
			return
		}

		return
	}

	d.load.Do(loadFunc)

	// Don't need to log this as the loadFunc logged already logged it
	if err != nil {
		return err
	}

	d.loaded = true
	return nil
}

// BuildInf is a container for information about a specific build.
type BuildInf struct {
	// Name is the name for the build. This may be an assigned value from
	// a TOML file setting.
	Name string

	// BuildName is the name of the build. This is either the name, as
	// specified in the build.toml file, or a generated name for -distro
	// flag based builds.
	BuildName string `toml:"build_name"`

	// TODO redo Ubuntu so that it supports custom urls.
	// BaseURL's usage is distro dependent:
	//	Ubuntu: url for Ubuntu's release server-it is not allowed to be
	//		empty.
	//	CentOS: by default, this is empty. If it is populated, its value
	//		is used instead of the randomly generated mirror url. 
	BaseURL   string `toml:"base_url"`
}

// update update's the current values with the passed, if the passed value is
// not an empty string.
func (i *BuildInf) update(new BuildInf) {
	if new.Name != "" {
		i.Name = new.Name
	}

	if new.BuildName != "" {
		i.BuildName = new.BuildName
	}

	return
}

// IODirInf is used to store information about where Rancher can find and put
// things. Source files are always in a SrcDir, e.g. HTTPSrcDir is the source
// directory for the HTTP directory. The destination directory is always a Dir,
// e.g. HTTPDir is the destination directory for the HTTP directory.
type IODirInf struct {
	// The directory in which the command files are located
	CommandsSrcDir string `toml:"commands_src_dir"`

	// The directory that will be used for the HTTP setting.
	HTTPDir string `toml:"http_dir"`

	// The directory that is the source for files to be copied to the HTTP
	// directory, HTTPDir
	HTTPSrcDir string `toml:"http_src_dir"`

	// The directory that the output artifacts will be written to.
	OutDir string `toml:"out_dir"`

	// The directory that scripts for the Packer template will be copied to.
	ScriptsDir string `toml:"scripts_dir"`

	// The directory that contains the scripts that will be copied.
	ScriptsSrcDir string `toml:"scripts_src_dir"`

	// The directory that contains the source files for this build.
	SrcDir string `toml:"src_dir"`
}


// update updates IODirInf with the passed values, when they exist.
func (i *IODirInf) update(new IODirInf) {
	if new.CommandsSrcDir != "" {
		i.CommandsSrcDir = appendSlash(new.CommandsSrcDir)
	}

	if new.HTTPDir != "" {
		i.HTTPDir = appendSlash(new.HTTPDir)
	}

	if new.HTTPSrcDir != "" {
		i.HTTPSrcDir = appendSlash(new.HTTPSrcDir)
	}

	if new.OutDir != "" {
		i.OutDir = appendSlash(new.OutDir)
	}

	if new.ScriptsDir != "" {
		i.ScriptsDir = appendSlash(new.ScriptsDir)
	}

	if new.ScriptsSrcDir != "" {
		i.ScriptsSrcDir = appendSlash(new.ScriptsSrcDir)
	}

	if new.SrcDir != "" {
		i.SrcDir = appendSlash(new.SrcDir)
	}

	return
}

// check to see if the dirinf is set
func (i *IODirInf) check() error {
	if i.HTTPDir == "" {
		err := errors.New("ioDirInf.Check: HTTPDir directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}

	if i.HTTPSrcDir == "" {
		err := errors.New("ioDirInf.Check: HTTPSrcDir directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}

	if i.OutDir == "" {
		err := errors.New("ioDirInf.Check: output directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}

	if i.SrcDir == "" {
		err := errors.New("ioDirInf.Check: SrcDir directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}

	if i.ScriptsDir == "" {
		err := errors.New("ioDirInf.Check: ScriptsDir directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}

	if i.ScriptsSrcDir == "" {
		err := errors.New("ioDirInf.Check: ScriptsSrcDir directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}

	return nil
}

// PackerInf is used to store information about a Packer Template. In Packer,
// these fields are optional, put used here because they are always printed out
// in a template as custom creation of template output hasn't been written--it
// may never be written.
type PackerInf struct {
	MinPackerVersion string `toml:"min_packer_version" json:"min_packer_version"`
	Description      string `toml:"description" json:"description"`
}

func (i *PackerInf) update(new PackerInf) {
	if new.MinPackerVersion != "" {
		i.MinPackerVersion = new.MinPackerVersion
	}

	if new.Description != "" {
		i.Description = new.Description
	}

	return
}

// Struct to hold the details of supported distros. From this information a
// user should be able to build a Packer template by only executing the
// following, at minimum:
//	$ rancher build -distro=ubuntu
// All settings can be overridden. The information here represents the standard
// box configuration for its respective distribution.
type distro struct {
	IODirInf
	PackerInf
	BuildInf

	// The supported Architectures, which can differ per distro. The labels can also
	// differ, e.g. amd64 and x86_64.
	Arch []string `toml:"Arch"`

	// Supported iso Images, e.g. server, minimal, etc.
	Image []string `toml:"Image"`

	// Supported Releases: the supported Releases are the Releases available for
	// download from that distribution's download page. Archived and unsupported
	// Releases are not used.
	Release []string `toml:"Release"`

	// The default Image configuration for this distribution. This usually consists of
	// things like Release, Architecture, Image type, etc.
	DefImage []string `toml:"default_Image"`

	// The configurations needed to generate the default settings for a build for this
	// distribution.
	build
}

// To add support for a distribution, the information about it must be added to
// the supported. file, in addition to adding the code to support it to the
// application.
type supported struct {
	Distro map[string]*distro
	load   sync.Once
	loaded bool
}

// Ensures that the supported distro information only get loaded once. Uses a
// mutex to prevent race conditions as there can be concurrent processing of
// Packer templates. When loaded, it sets the loaded boolean so that it only
// needs to be called when it hasn't been loaded.
func (s *supported) LoadOnce() error {
	var err error

	loadFunc := func() {
		name := os.Getenv(EnvSupportedFile)

		if name == "" {
			err = errors.New("could not retrieve the Supported information because the " + EnvSupportedFile + " environment variable was not set. Either set it or check your rancher.cfg setting")
			jww.CRITICAL.Print(err.Error())
			return
		}

		if _, err = toml.DecodeFile(name, &s); err != nil {
			jww.CRITICAL.Print(err.Error())
			return
		}

		s.loaded = true
		return
	}

	s.load.Do(loadFunc)

	// Don't need to log the error as loadFunc already did it.
	if err != nil {
		return err
	}

	return nil
}

// Struct to hold the builds.
type builds struct {
	Build  map[string]*rawTemplate
	load   sync.Once
	loaded bool
}

// LoadOnce ensures that the build information only get loaded once. Uses a 
// mutex to prevent race conditions as there can be concurrent processing of 
// Packer templates. When loaded, it sets the loaded boolean so that it only
// needs to be called when it hasn't been loaded.
func (b *builds) LoadOnce() error {
	var err error

	loadFunc := func() {
		name := os.Getenv(EnvBuildsFile)

		if name == "" {
			err = errors.New("could not retrieve the Builds configurations because the " + EnvBuildsFile + " environment variable was not set. Either set it or check your rancher.cfg setting")
			jww.CRITICAL.Print(err.Error())
			return
		}

		if _, err = toml.DecodeFile(name, &b); err != nil {
			jww.CRITICAL.Print(err.Error())
			return
		}

		b.loaded = true
		return
	}

	b.load.Do(loadFunc)

	// Don't need to log the error as loadFunc already handled that.
	if err != nil {
		return err
	}

	return nil
}

// buildLists contains lists of builds.
type buildLists struct {
	List map[string]list
}

// list is a slice of 1 or more named builds.
type list struct {
	Builds []string
}

// Load loads the buildLists from its respective TOML file.
func (b *buildLists) Load() error {
	// Load the build lists.
	name := os.Getenv(EnvBuildListsFile)

	if name == "" {
		err := errors.New("could not retrieve the BuildLists file because the " + EnvBuildListsFile + " environment variable was not set. Either set it or check your rancher.cfg setting")
		jww.ERROR.Print(err.Error())
		return err
	}

	if _, err := toml.DecodeFile(name, &b); err != nil {
		jww.ERROR.Print(err.Error())
		return err
	}

	return nil
}
