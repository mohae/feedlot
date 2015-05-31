// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package ranchr implements the creation of Packer templates from Rancher
// build definitions.
package app

import (
	"fmt"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/mohae/contour"
	"github.com/mohae/deepcopy"
)

// Contains most of the information for Packer templates within a Rancher
// Build.
type build struct {
	// Targeted builders: the values are consistent with Packer's, e.g.
	// `virtualbox.iso` is used for VirtualBox.
	BuilderTypes []string `toml:"builder_types"`
	// A map of builder configuration. There should always be a `common`
	// builder, which has settings common to both VMWare and VirtualBox.
	Builders map[string]builder `toml:"builders"`
	// Targeted post-processors: the values are consistent with Packer's, e.g.
	// `vagrant` is used for Vagrant.
	PostProcessorTypes []string `toml:"post_processor_types"`
	// A map of post-processor configurations.
	PostProcessors map[string]postProcessor `toml:"post_processors"`
	// Targeted provisioners: the values are consistent with Packer's, e.g.
	// `shell` is used for shell.
	ProvisionerTypes []string `toml:"provisioner_types"`
	// A map of provisioner configurations.
	Provisioners map[string]provisioner `toml:"provisioners"`
}

// build.DeepCopy makes a deep copy of the build and returns it.
func (b *build) DeepCopy() build {
	newB := build{
		BuilderTypes:       []string{},
		Builders:           map[string]builder{},
		PostProcessorTypes: []string{},
		PostProcessors:     map[string]postProcessor{},
		ProvisionerTypes:   []string{},
		Provisioners:       map[string]provisioner{},
	}
	if b.BuilderTypes != nil || len(b.BuilderTypes) > 0 {
		newB.BuilderTypes = make([]string, len(b.BuilderTypes), len(b.BuilderTypes))
		copy(newB.BuilderTypes, b.BuilderTypes)
	}
	if b.PostProcessorTypes != nil || len(b.PostProcessorTypes) > 0 {
		newB.PostProcessorTypes = make([]string, len(b.PostProcessorTypes), len(b.PostProcessorTypes))
		copy(newB.PostProcessorTypes, b.PostProcessorTypes)
	}
	if b.ProvisionerTypes != nil || len(b.ProvisionerTypes) > 0 {
		newB.ProvisionerTypes = make([]string, len(b.ProvisionerTypes), len(b.ProvisionerTypes))
		copy(newB.ProvisionerTypes, b.ProvisionerTypes)
	}
	for k, v := range b.Builders {
		newB.Builders[k] = v.DeepCopy()
	}
	for k, v := range b.PostProcessors {
		newB.PostProcessors[k] = v.DeepCopy()
	}
	for k, v := range b.Provisioners {
		newB.Provisioners[k] = v.DeepCopy()
	}
	return newB
}

// templateSection is used as an embedded type.
type templateSection struct {
	// Settings are string settings in "key=value" format.
	Settings []string
	// Arrays are the string array settings.
	Arrays map[string]interface{}
}

// templateSection.DeepCopy updates its information with new via a deep copy.
func (t *templateSection) DeepCopy(ts templateSection) {
	//Deep Copy of settings
	t.Settings = make([]string, len(ts.Settings))
	copy(t.Settings, ts.Settings)
	// make a deep copy of the Arrays(map[string]interface)
	t.Arrays = deepcopy.Iface(ts.Arrays).(map[string]interface{})
}

// mergeArrays merges the arrays section of a template builder
func (t *templateSection) mergeArrays(old map[string]interface{}, n map[string]interface{}) map[string]interface{} {
	if old == nil && n == nil {
		return nil
	}
	if old == nil {
		return n
	}
	if n == nil {
		return old
	}
	// both are populated, merge them.
	merged := map[string]interface{}{}
	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(old, n)
	// Process using the keys.
	for _, v := range keys {
		// If the element for this key doesn't exist in new, add old.
		if _, ok := n[v]; !ok {
			merged[v] = old[v]
			continue
		}
		// Otherwise use the new value
		merged[v] = n[v]
	}
	return merged
}

// builder represents a builder Packer template section.
type builder struct {
	templateSection
}

// builder.DeepCopy copies the builder values instead of the pointers.
func (b *builder) DeepCopy() builder {
	c := builder{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(b.templateSection)
	return c
}

// mergeSettings the settings section of a builder. New values supersede
// existing ones.
func (b *builder) mergeSettings(sl []string) {
	if sl == nil {
		return
	}
	b.Settings = mergeSettingsSlices(b.Settings, sl)
}

// mergeArrays merges the arrays section of a template builder
func (b *builder) mergeArrays(m map[string]interface{}) {
	b.Arrays = b.templateSection.mergeArrays(b.Arrays, m)
}

// mergeVMSettings Merge the VMSettings section of a builder. New values
// supersede existing ones.
// TODO update to work with new arrays processing
/*
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
*/

// Type for handling the post-processor section of the configs.
type postProcessor struct {
	templateSection
}

// postProcessor.DeepCopy copies the postProcessor values instead of the
// pointers.
func (p *postProcessor) DeepCopy() postProcessor {
	c := postProcessor{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(p.templateSection)
	return c
}

// postProcessor.mergeSettings  merges the settings section of a post-processor
// with the passed slice of settings. New values supercede existing ones.
func (p *postProcessor) mergeSettings(sl []string) {
	if sl == nil {
		return
	}
	if p.Settings == nil {
		p.Settings = sl
		return
	}
	// merge the keys
	p.Settings = mergeSettingsSlices(p.Settings, sl)
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

// postProcessor.DeepCopy copies the postProcessor values instead of the
// pointers.
func (p *provisioner) DeepCopy() provisioner {
	c := provisioner{templateSection: templateSection{Settings: []string{}, Arrays: map[string]interface{}{}}}
	c.templateSection.DeepCopy(p.templateSection)
	return c
}

// provisioner.mergeSettings  merges the settings section of a post-processor
// with the passed slice of settings. New values supercede existing ones.
func (p *provisioner) mergeSettings(sl []string) {
	if sl == nil {
		return
	}
	if p.Settings == nil {
		p.Settings = sl
		return
	}
	// merge the keys
	p.Settings = mergeSettingsSlices(p.Settings, sl)
}

// provisioner.mergeArrays wraps templateSection.mergeArrays
func (p *provisioner) mergeArrays(m map[string]interface{}) {
	// merge the arrays:
	p.Arrays = p.templateSection.mergeArrays(p.Arrays, m)
}

// defaults is used to store Rancher application level defaults for Packer templates.
type defaults struct {
	IODirInf
	PackerInf
	BuildInf
	build
	sync.Mutex
	loaded bool
}

// Load loads the defualt settings. If the defaults have already been loaded
// nothing is done.
func (d *defaults) Load() error {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()
	if d.loaded {
		return nil
	}
	name := contour.GetString(DefaultFile)
	if name == "" {
		return filenameNotSetErr("default")
	}
	_, err := toml.DecodeFile(name, d)
	if err != nil {
		return fmt.Errorf("load of default information failed: %s", err)
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
	BaseURL   string `toml:"base_url"`
}

func (i *BuildInf) update(b BuildInf) {
	if b.Name != "" {
		i.Name = b.Name
	}
	if b.BuildName != "" {
		i.BuildName = b.BuildName
	}
}

// IODirInf is used to store information about where Rancher can find and put
// things. Source files are always in a SrcDir.
type IODirInf struct {
	// Include the packer component name in the path. If this is true, the component.String()
	// value will be added as the parent of the output resource: i.e. OutDir/component.String()/resource_name
	IncludeComponentString bool
	// The directory that the output artifacts will be written to.
	OutDir string `toml:"out_dir"`
	// The directory that contains the source files for this build.
	SrcDir string `toml:"src_dir"`
}

func (i *IODirInf) update(inf IODirInf) {
	if inf.OutDir != "" {
		i.OutDir = appendSlash(inf.OutDir)
	}
	if inf.SrcDir != "" {
		i.SrcDir = appendSlash(inf.SrcDir)
	}
}

// check to see if the dirinf is set, if not, set them to their defaults
func (i *IODirInf) check() {
	if i.OutDir == "" {
		i.OutDir = contour.GetString(ParamDelimStart) + "build_name"
	}
	if i.SrcDir == "" {
		i.SrcDir = "src"
	}
}

// PackerInf is used to store information about a Packer Template. In Packer,
// these fields are optional, put used here because they are always printed out
// in a template as custom creation of template output hasn't been written--it
// may never be written.
type PackerInf struct {
	MinPackerVersion string `toml:"min_packer_version" json:"min_packer_version"`
	Description      string `toml:"description" json:"description"`
}

func (i *PackerInf) update(inf PackerInf) {
	if inf.MinPackerVersion != "" {
		i.MinPackerVersion = inf.MinPackerVersion
	}
	if inf.Description != "" {
		i.Description = inf.Description
	}
}

// Struct to hold the details of supported distros. From this information a
// user should be able to build a Packer template by only executing the
// following, at minimum:
//   $ rancher build -distro=ubuntu
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
	sync.Mutex
	loaded bool
}

// Load the supported distro info. If it has already been loaded, nothing is
// done.
func (s *supported) Load() error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	name := contour.GetString(SupportedFile)
	if name == "" {
		return filenameNotSetErr("supported")
	}
	_, err := toml.DecodeFile(name, &s)
	if err != nil {
		return decodeErr(err)
	}
	s.loaded = true
	return nil
}

// Struct to hold the builds.
type builds struct {
	Build map[string]*rawTemplate
	sync.Mutex
	loaded bool
}

// Load the build information. If it has already been loaded, nothing will be done
func (b *builds) Load() error {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	name := contour.GetString(BuildFile)
	if name == "" {
		return filenameNotSetErr("build")
	}
	_, err := toml.DecodeFile(name, &b)
	if err != nil {
		return decodeErr(err)
	}
	b.loaded = true
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
	name := contour.GetString(BuildListFile)
	if name == "" {
		return filenameNotSetErr("build_list")
	}
	if _, err := toml.DecodeFile(name, &b); err != nil {
		return decodeErr(err)
	}
	return nil
}

func filenameNotSetErr(target string) error {
	return fmt.Errorf("%q not set, unable to retrieve the %s file", target, target)
}

func decodeErr(err error) error {
	return fmt.Errorf("decode failed: %s", err)
}
