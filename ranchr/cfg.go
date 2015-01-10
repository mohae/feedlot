// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package ranchr implements the creation of Packer templates from Rancher
// build definitions.
package ranchr

import (
	"fmt"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// Contains most of the information for Packer templates within a Rancher Build.
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

// templateSection is used as an embedded type.
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
	t.Arrays = deepcopy.Iface(new.Arrays).(map[string]interface{})
}

// mergeArrays merges the arrays section of a template builder
func (t *templateSection) mergeArrays(old map[string]interface{}, new map[string]interface{}) map[string]interface{} {
	if old == nil && new == nil {
		return nil
	}
	if old == nil {
		return new
	}
	if new == nil {
		return old
	}
	// both are populated, merge them.
	merged := map[string]interface{}{}
	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(old, new)
	// Process using the keys.
	for _, v := range keys {
		// If the element for this key doesn't exist in new, add old.
		if _, ok := new[v]; !ok {
			merged[v] = old[v]
			continue
		}
		// Otherwise use the new value
		merged[v] = new[v]
	}
	return merged
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
func (b *builder) mergeVMSettings(new []string) []string {
	if new == nil {
		return nil
	}
	var merged []string
	old := deepcopy.InterfaceToSliceStrings(b.Arrays[VMSettings])
	merged = mergeSettingsSlices(old, new)
	if b.Arrays == nil {
		b.Arrays = map[string]interface{}{}
	}
	return merged
}

// Type for handling the post-processor section of the configs.
type postProcessor struct {
	templateSection
}

// postProcessor.DeepCopy copies the postProcessor values instead of the pointers.
func (p *postProcessor) DeepCopy() *postProcessor {
	var c *postProcessor
	c = &postProcessor{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(p.templateSection)
	return c
}

// postProcessor.mergeSettings  merges the settings section of a post-processor
// with the passed slice of settings. New values supercede existing ones.
func (p *postProcessor) mergeSettings(new []string) {
	if new == nil {
		return
	}
	if p.Settings == nil {
		p.Settings = new
		return
	}
	// merge the keys
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

// postProcessor.mergeArrays wraps templateSection.mergeArrays
func (p *postProcessor) mergeArrays(m map[string]interface{}) {
	// merge the arrays:
	p.Arrays = p.templateSection.mergeArrays(p.Arrays, m)
}

// provisioner: type for common elements for provisioners.
type provisioner struct {
	templateSection
}

// postProcessor.DeepCopy copies the postProcessor values instead of the pointers.
func (p *provisioner) DeepCopy() *provisioner {
	var c *provisioner
	c = &provisioner{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(p.templateSection)
	return c
}

// provisioner.mergeSettings  merges the settings section of a post-processor
// with the passed slice of settings. New values supercede existing ones.
func (p *provisioner) mergeSettings(new []string) {
	if new == nil {
		return
	}
	if p.Settings == nil {
		p.Settings = new
		return
	}
	// merge the keys
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

// provisioner.mergeArrays wraps templateSection.mergeArrays
func (p *provisioner) mergeArrays(m map[string]interface{}) {
	// merge the arrays:
	p.Arrays = p.templateSection.mergeArrays(p.Arrays, m)
}

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

// Ensures that the default configs get loaded once. Uses a mutex to prevent
// race conditions as there can be concurrent processing of Packer templates.
// When loaded, it sets the loaded boolean so that it only needs to be called
// when it hasn't been loaded.
func (d *defaults) LoadOnce() error {
	var err error
	loadFunc := func() {
		name := os.Getenv(EnvDefaultsFile)
		if name == "" {
			err = fmt.Errorf("unable to retrieve the default settings: %q was not set; check your 'rancher.cfg'", EnvBuildsFile)
			jww.CRITICAL.Print(err)
			return
		}
		_, err = toml.DecodeFile(name, d)
		if err != nil {
			jww.CRITICAL.Print(err)
			return
		}
		return
	}
	d.load.Do(loadFunc)
	d.loaded = true
	return err
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
	BaseURL   string `toml:"base_url"`
}

func (i *BuildInf) update(new BuildInf) {
	if new.Name != "" {
		i.Name = new.Name
	}
	if new.BuildName != "" {
		i.BuildName = new.BuildName
	}
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
}

// check to see if the dirinf is set
func (i *IODirInf) check() error {
	if i.HTTPDir == "" {
		err := fmt.Errorf("HTTPDir directory not set")
		jww.ERROR.Print(err)
		return err
	}
	if i.HTTPSrcDir == "" {
		err := fmt.Errorf("HTTPSrcDir directory not set")
		jww.ERROR.Print(err)
		return err
	}
	if i.OutDir == "" {
		err := fmt.Errorf("output directory not set")
		jww.ERROR.Print(err)
		return err
	}
	if i.SrcDir == "" {
		err := fmt.Errorf("SrcDir directory not set")
		jww.ERROR.Print(err)
		return err
	}
	if i.ScriptsDir == "" {
		err := fmt.Errorf("ScriptsDir directory not set")
		jww.ERROR.Print(err.Error())
		return err
	}
	if i.ScriptsSrcDir == "" {
		err := fmt.Errorf("ScriptsSrcDir directory not set")
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
	loadFunc := func() {
		name := os.Getenv(EnvSupportedFile)
		if name == "" {
			err := fmt.Errorf("%s not set, unable to retrieve the Supported information", EnvSupportedFile)
			jww.CRITICAL.Print(err)
		}
		_, err := toml.DecodeFile(name, &s)
		if err != nil {
			jww.CRITICAL.Print(err)
			return
		}
		s.loaded = true
		return
	}
	s.load.Do(loadFunc)
	if !s.loaded {
		return fmt.Errorf("an error occurred while loading the Supported information, please check the log")
	}
	return nil
}

// Struct to hold the builds.
type builds struct {
	Build  map[string]*rawTemplate
	load   sync.Once
	loaded bool
}

// Ensures that the build information only get loaded once. Uses a mutex to
// prevent race conditions as there can be concurrent processing of Packer
// templates. When loaded, it sets the loaded boolean so that it only needs to
// be called when it hasn't been loaded.
func (b *builds) LoadOnce() error {
	var err error
	loadFunc := func() {
		name := os.Getenv(EnvBuildsFile)
		if name == "" {
			err = fmt.Errorf("%s not set, unable to retrieve the Build configurations", EnvBuildsFile)
			jww.CRITICAL.Print(err)
			return
		}
		_, err = toml.DecodeFile(name, &b)
		if err != nil {
			jww.CRITICAL.Print(err)
			return
		}
		b.loaded = true
		return
	}
	b.load.Do(loadFunc)
	return nil
}

// Contains lists of builds.
type buildLists struct {
	List map[string]list
}

// A list contains 1 or more builds.
type list struct {
	Builds []string
}

// This is a normal load, no mutex, as this is only called once.
func (b *buildLists) Load() error {
	// Load the build lists.
	name := os.Getenv(EnvBuildListsFile)
	if name == "" {
		err := fmt.Errorf("%s not set, unable to retrieve the BuildLists file", EnvBuildListsFile)
		jww.ERROR.Print(err)
		return err
	}
	if _, err := toml.DecodeFile(name, &b); err != nil {
		jww.ERROR.Print(err)
		return err
	}
	return nil
}
