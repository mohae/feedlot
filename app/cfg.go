// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package app

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	cjsn "github.com/mohae/cjson"
	"github.com/mohae/contour"
	"github.com/mohae/deepcopy"
	"github.com/mohae/feedlot/conf"
	"github.com/mohae/feedlot/log"
)

// Componenter is an interface for Packer components, i.e. builder,
// post-processor, and provisioner.
type Componenter interface {
	getType() string
}

// Contains most of the information for Packer templates within a Feedlot
// Build.  The keys of the maps are the IDs
type Build struct {
	// Targeted builders: either the ID of the builder section or the Packer
	// component type value can be used, e.g. ID: "vbox" or "virtualbox-iso".
	BuilderIDs []string `toml:"builder_ids" json:"builder_ids"`
	// A map of Builder configurations.  When using a VMWare or VirtualBox
	// builder, there is usually a 'common' builder, which has settings common
	// to both VMWare and VirtualBox.
	Builders map[string]BuilderC `toml:"builders" json:"builders"`
	// Targeted post-processors: either the ID of the post-processor or the
	// Packer component type value can be used, e.g. ID: "docker" or
	// "docker-push".
	PostProcessorIDs []string `toml:"post_processor_ids" json:"post_processor_ids"`
	// A map of Post-Processor configurations.
	PostProcessors map[string]PostProcessorC `toml:"post_processors" json:"post_processors"`
	// Targeted provisioners: either the ID of the provisioners or the Packer
	// component type value can be used, e.g. ID: "preprocess" or "shell".
	ProvisionerIDs []string `toml:"provisioner_ids" json:"provisioner_ids"`
	// A map of {rovisioner configurations.
	Provisioners map[string]ProvisionerC `toml:"provisioners" json:"provisioners"`
}

// Copy makes a deep copy of the Build and returns it.
func (b *Build) Copy() *Build {
	return deepcopy.Copy(b).(*Build)
}

// setTypes goes through each component map and check's if the
// templateSection Type is set.  If it isn't, the ID (key) is used to
// set it.
func (b *Build) setTypes() {
	for k, v := range b.Builders {
		if len(v.Type) == 0 {
			v.Type = k
			b.Builders[k] = v
		}
	}
	for k, v := range b.PostProcessors {
		if len(v.Type) == 0 {
			v.Type = k
			b.PostProcessors[k] = v
		}
	}
	for k, v := range b.Provisioners {
		if len(v.Type) == 0 {
			v.Type = k
			b.Provisioners[k] = v
		}
	}
}

// TemplateSection is used as an embedded type.
type TemplateSection struct {
	// Type is the actual Packer component type, this may or may not be the
	// same as the map key (ID).
	Type string `toml:"type" json:"type"`
	// Settings are string settings in "key=value" format.
	Settings []string `toml:"settings" json:"settings"`
	// Arrays are the string array settings.
	Arrays map[string]interface{} `toml:"arrays" json:"arrays"`
}

// mergeArrays merges the received array with the current one.
func (t *TemplateSection) mergeArrays(n map[string]interface{}) {
	if n == nil {
		return
	}
	if t.Arrays == nil {
		t.Arrays = n
		return
	}
	// both are populated, merge them.
	merged := map[string]interface{}{}
	// Get the all keys from both maps
	keys := mergeKeysFromMaps(t.Arrays, n)
	// Process using the keys.
	for _, v := range keys {
		// If the element for this key doesn't exist in new, add old.
		if _, ok := n[v]; !ok {
			merged[v] = t.Arrays[v]
			continue
		}
		// Otherwise use the new value
		merged[v] = n[v]
	}
	t.Arrays = merged
}

// BuilderC represents a builder component of a Packer template.
type BuilderC struct {
	TemplateSection
}

// getType returns the type for this component.  This is on builder instead of
// templateSection because the component needs to fulfill the interface,
// not the templateSection.
func (b BuilderC) getType() string {
	return b.Type
}

// Copy makes a deepcopy of the Builder component.
func (b *BuilderC) Copy() BuilderC {
	new := deepcopy.Copy(b).(*BuilderC)
	return *new
}

// mergeSettings the settings section of a Builder. New values supersede
// existing ones.
func (b *BuilderC) mergeSettings(sl []string) error {
	if sl == nil {
		return nil
	}
	var err error
	b.Settings, err = mergeSettingsSlices(b.Settings, sl)
	if err != nil {
		return Error{slug: "merge settings", err: err}
	}
	return nil
}

// PostProcessorC represents a Packer post-processor component.
type PostProcessorC struct {
	TemplateSection
}

// getType returns the type for this component. This is on postProcessor
// instead of templateSection because the component needs to fulfill the
// interface, not the templateSection.
func (p PostProcessorC) getType() string {
	return p.Type
}

// Copy makes a deep copy of the PostProcessor component.
func (p *PostProcessorC) Copy() PostProcessorC {
	new := deepcopy.Copy(p).(*PostProcessorC)
	return *new
}

// mergeSettings merges the settings section of a post-processor with the
// passed slice of settings. New values supercede existing ones.
func (p *PostProcessorC) mergeSettings(sl []string) error {
	if sl == nil {
		return nil
	}
	if p.Settings == nil {
		p.Settings = sl
		return nil
	}
	// merge the keys
	var err error
	p.Settings, err = mergeSettingsSlices(p.Settings, sl)
	if err != nil {
		return Error{slug: "merge settings", err: err}
	}
	return nil
}

// ProvisionerC representa Packer provisioner component.
type ProvisionerC struct {
	TemplateSection
}

// getType returns the type for this component.  This is on provisioner
// instead of templateSection because the component needs to fulfill the
// interface, not the templateSection.
func (p ProvisionerC) getType() string {
	return p.Type
}

// Copy makes a deep copy of the Provisioner component.
func (p *ProvisionerC) Copy() ProvisionerC {
	new := deepcopy.Copy(p).(*ProvisionerC)
	return *new
}

// mergeSettings merges the settings section of a post-processor with the
// passed slice of settings. New values supercede existing ones.
func (p *ProvisionerC) mergeSettings(sl []string) error {
	if sl == nil {
		return nil
	}
	if p.Settings == nil {
		p.Settings = sl
		return nil
	}
	// merge the keys
	var err error
	p.Settings, err = mergeSettingsSlices(p.Settings, sl)
	if err != nil {
		return Error{slug: "merge settings", err: err}
	}
	return nil
}

// BuildInf is a container for information about a specific build.
type BuildInf struct {
	// Name is the name for the build. This may be an assigned value from
	// a TOML file setting.
	Name string `toml:"name" json:"name"`
	// BuildName is the name of the build. This is either the name, as
	// specified in the build.toml file, or a generated name for -distro
	// flag based builds.
	BuildName string `toml:"build_name" json:"build_name"`
	// BaseURL is the base url for the iso.
	BaseURL string `toml:"base_url" json:"base_url"`
	// Region to use when selecting the image mirror; for use with CentOS.
	// This corresponds to the 'Region' column in the
	// https://centos.org/download/full-mirrorlist.csv file.
	//
	// If empty, no region filtering will be applied.  A pointer is used for
	// detection of set vs not set.
	Region *string `toml:"region" json:"region"`
	// Country to use when selecting the image mirror; for use with CentOS.
	// This corresponds to the 'Country' column in the
	// https://centos.org/download/full-mirrorlist.csv file.
	//
	// If empty, no country filtering will be applied.
	//
	//  For Region 'US', this is used as the state field.  A pointer is used
	// for detection of set vs not set.
	Country *string `toml:"country" json:"country"`
	// Sponsor to use when selecting the image mirror: for use with CentOS.
	// This corresponds to the 'Sponsor' column in the
	// https://centos.org/download/full-mirrorlist.csv file.
	//
	// If a value is specified, the country setting is ignored so that
	// Rackspace and Oregon State University, OSUOSL, are not filtered out.
	//
	// For Oregon State University, aside from it's name, OSUOSL and osuosl
	// are accepted values.
	//
	// If empty, no sponsor filtering will be applied.  A pointer is used for
	// detection of set vs not set.
	Sponsor *string `toml:"sponsor" json:"sponsor"`
}

func (b *BuildInf) update(v BuildInf) {
	if v.Name != "" {
		b.Name = v.Name
	}
	if v.BuildName != "" {
		b.BuildName = v.BuildName
	}
	if v.BaseURL != "" {
		b.BaseURL = v.BaseURL
	}
	if v.Region != nil && *v.Region != "" {
		b.Region = v.Region
	}
	if v.Country != nil && *v.Country != "" {
		b.Country = v.Country
	}
	if v.Sponsor != nil && *v.Sponsor != "" {
		b.Sponsor = v.Sponsor
	}
}

// IODirInf is used to store information about where Feedlot can find and put
// things. Source files are always in a SourceDir.
type IODirInf struct {
	// Include the packer component name in the path. If true, the
	// component.String() value will be added as the parent of the output
	// resource path: i.e. OutDir/component.String()/resource_name.  This is a
	// pointer so that whether or not this setting was actually set can be
	// determined, otherwise determining whether it was an explicit false or
	// empty would not be possible.
	IncludeComponentString *bool `toml:"include_component_string" json:"include_component_string"`
	// PackerOutputDir is the output directory for the Packer artificts, when
	// applicable.  This is usually referenced in a Builder's output directory,
	// e.g. "output_dir=:packer_output_dir"
	PackerOutputDir string `toml:"packer_output_dir" json:"packer_output_dir"`
	// The directory that contains the source files for a build.
	SourceDir string `toml:"source_dir" json:"source_dir"`
	// If the source dir path is relative to the conf_dir.  If true, the path
	// is resolved relative to the conf_dir.  Otherwise, the path is used as
	// is.  This is a pointer so that whether or not this setting was actually
	// set can be determined, otherwise determining whether it was an explicit
	// false or empty would not be possible.
	SourceDirIsRelative *bool `toml:"source_dir_is_relative" json:"source_dir_is_relative"`
	// The output directory for a generated Packer template and its resources.
	TemplateOutputDir string `toml:"template_output_dir" json:"template_output_dir"`
	// If the template output dir path is relative to the current working
	// directory.  If true, the path is resolved relative to the working
	// directory, otherwise, the path is used as is.  This is a pointer so that
	// whether or not this setting was actually set can be determined,
	// otherwise determining whether it was an explicit false or empty would
	// not be possible.
	TemplateOutputDirIsRelative *bool `toml:"template_output_dir_is_relative" json:"template_output_dir_is_relative"`
}

// Only update when a value exists; empty strings don't count as being set.
func (i *IODirInf) update(v IODirInf) {
	if v.TemplateOutputDir != "" {
		i.TemplateOutputDir = v.TemplateOutputDir
	}
	// the path should end with "/"
	i.TemplateOutputDir = appendSlash(i.TemplateOutputDir)
	var b bool
	// treat nils as false
	if v.TemplateOutputDirIsRelative != nil {
		i.TemplateOutputDirIsRelative = v.TemplateOutputDirIsRelative
	}
	if i.TemplateOutputDirIsRelative == nil {
		i.TemplateOutputDirIsRelative = &b
	}
	if v.PackerOutputDir != "" {
		i.PackerOutputDir = v.PackerOutputDir
	}
	// the path should end with "/"
	i.PackerOutputDir = appendSlash(i.PackerOutputDir)

	if v.SourceDir != "" {
		i.SourceDir = appendSlash(v.SourceDir)
	}
	// the path should end with "/"
	i.SourceDir = appendSlash(i.SourceDir)
	// treat nils as false
	if v.SourceDirIsRelative != nil {
		i.SourceDirIsRelative = v.SourceDirIsRelative
	}
	if i.SourceDirIsRelative == nil {
		i.SourceDirIsRelative = &b
	}
	if v.IncludeComponentString != nil {
		i.IncludeComponentString = v.IncludeComponentString
	}
}

// check to see if the dirinf is set, if not, set them to their defaults
func (i *IODirInf) check() {
	if i.TemplateOutputDir == "" {
		i.TemplateOutputDir = fmt.Sprintf("%sbuildname", contour.GetString(conf.ParamDelimStart))
	}
	if i.PackerOutputDir == "" {
		i.PackerOutputDir = fmt.Sprintf("%sbuildname", contour.GetString(conf.ParamDelimStart))
	}
	if i.SourceDir == "" {
		i.SourceDir = "src"
	}
}

// PackerInf is used to store information about a Packer Template. In
// Packer, these fields are optional, put used here because they are
// always printed out in a template as custom creation of template output
// hasn't been written--it may never be written.
type PackerInf struct {
	MinPackerVersion string `toml:"min_packer_version" json:"min_packer_version"`
	Description      string `toml:"description" json:"description"`
}

func (p *PackerInf) update(v PackerInf) {
	if v.MinPackerVersion != "" {
		p.MinPackerVersion = v.MinPackerVersion
	}
	if v.Description != "" {
		p.Description = v.Description
	}
}

// Defaults is used to store Feedlot application level defaults for Packer templates.
type Defaults struct {
	IODirInf
	PackerInf
	BuildInf
	Build
	loaded bool `toml:"-"`
}

// Load loads the default settings. If the defaults have already been loaded
// nothing is done.
func (d *Defaults) Load(p string) error {
	if d.loaded {
		return nil
	}
	name, format, err := conf.ConfFilename(conf.FindConfFile(p, fmt.Sprintf("%s.%s", "default", contour.GetString(conf.Format))))
	if err != nil {
		fmt.Errorf("load defaults: %s", err)
		log.Error(err)
		return err
	}
	switch format {
	case conf.TOML:
		_, err := toml.DecodeFile(name, &d)
		if err != nil {
			err = fmt.Errorf("load defaults: %s: %s", name, err)
			log.Error(err)
			return err
		}
	case conf.JSON:
		var buff []byte
		buff, err = ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("load defaults: %s: %s", name, err)
			log.Error(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &d)
		if err != nil {
			err = fmt.Errorf("load defaults: %s: %s", name, err)
			log.Error(err)
			return err
		}
	default:
		err := fmt.Errorf("load defaults: %s: %s", contour.GetString(conf.Format), conf.ErrUnsupportedFormat)
		log.Error(err)
		return err
	}
	d.Build.setTypes()
	d.loaded = true
	return nil
}

// Struct to hold the details of supported distros. From this information a
// user should be able to build a Packer template by only executing the
// following, at minimum:
//   $ feedlot build -distro=ubuntu
// All settings can be overridden. The information here represents the standard
// box configuration for its respective distribution.
type SupportedDistro struct {
	IODirInf
	PackerInf
	BuildInf
	// The supported Architectures, which can differ per distro. The labels can also
	// differ, e.g. amd64 and x86_64.
	Arch []string `toml:"arch" json:"arch"`
	// Supported iso Images, e.g. server, minimal, etc.
	Image []string `toml:"image" json:"image"`
	// Supported Releases: the supported Releases are the Releases available for
	// download from that distribution's download page. Archived and unsupported
	// Releases are not used.
	Release []string `toml:"release" json:"release"`
	// The default Image configuration for this distribution. This usually consists of
	// things like Release, Architecture, Image type, etc.
	DefImage []string `toml:"default_image" json:"default_image"`
	// The configurations needed to generate the default settings for a build for this
	// distribution.
	Build
}

// To add support for a distribution, the information about it must be added to
// the supported. file, in addition to adding the code to support it to the
// application.
type SupportedDistros struct {
	Distros map[string]*SupportedDistro
	loaded  bool
}

// Load the supported distro info.
func (s *SupportedDistros) Load(p string) error {
	name, format, err := conf.ConfFilename(conf.FindConfFile(p, "supported"))
	if err != nil {
		err = fmt.Errorf("load supported: %s", err)
		log.Error(err)
		return err
	}
	switch format {
	case conf.TOML:
		_, err := toml.DecodeFile(name, &s.Distros)
		if err != nil {
			err = fmt.Errorf("load supported: %s: %s", name, err)
			log.Error(err)
			return err
		}
	case conf.JSON:
		var buff []byte
		buff, err = ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("load supported: %s: %s", name, err)
			log.Error(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &s.Distros)
		if err != nil {
			err = fmt.Errorf("load supported: %s: %s", name, err)
			log.Error(err)
			return err
		}
	default:
		err := fmt.Errorf("load supported: %s: %s", name, conf.ErrUnsupportedFormat)
		log.Error(err)
		return err
	}
	s.loaded = true
	return nil
}

// Struct to hold the builds.
type Builds struct {
	Templates map[string]*RawTemplate
	loaded    bool
}

// Load the build information from the provided name.
func (b *Builds) Load(name string) error {
	if name == "" {
		err := errors.New("load build: no build name specified")
		log.Error(err)
		return err
	}
	switch conf.ParseConfFormat(contour.GetString(conf.Format)) {
	case conf.TOML:
		_, err := toml.DecodeFile(name, &b.Templates)
		if err != nil {
			err = fmt.Errorf("load build %s: %s", name, err)
			log.Error(err)
			return err
		}
	case conf.JSON:
		buff, err := ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("load build %s: %s", name, err)
			log.Error(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &b.Templates)
		if err != nil {
			err = fmt.Errorf("load build %s: %s", name, err)
			log.Error(err)
			return err
		}
	default:
		err := fmt.Errorf("load build %s: %s", name, conf.ErrUnsupportedFormat)
		log.Error(err)
		return err
	}
	for _, v := range b.Templates {
		v.Build.setTypes()
	}

	b.loaded = true
	return nil
}

// getBuildTemplate returns the requested build template, or an error if it
// can't be found. Th
func getBuildTemplate(name string) (*RawTemplate, error) {
	var r *RawTemplate
	var err error
	for _, blds := range BuildDefs {
		for n, bTpl := range blds.Templates {
			if n == name {
				r = bTpl.Copy()
				r.BuildName = name
				goto found
			}
		}
	}
	err = fmt.Errorf("build not found: %s", name)
	return nil, err
found:
	log.Debugf("build %s found\n", name)
	return r, nil
}

// Contains lists of builds.
type BuildLists struct {
	Lists map[string]List
}

// A List contains 1 or more builds.
type List struct {
	Builds []string
}

// Load loads the build lists. It accepts a path prefix; which is mainly used
// for testing ATM.
func (b *BuildLists) Load(p string) error {
	// Load the build lists.
	name, format, err := conf.ConfFilename(conf.FindConfFile(p, "build_list"))
	if err != nil {
		err = fmt.Errorf("load build list: %s: %s", name, err)
		log.Error(err)
		return err
	}
	switch format {
	case conf.TOML:
		_, err := toml.DecodeFile(name, &b.Lists)
		if err != nil {
			err = fmt.Errorf("load build list: %s: %s", name, err)
			log.Error(err)
			return err
		}
	case conf.JSON:
		var buff []byte
		buff, err = ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf("load build list: %s: %s", name, err)
			log.Error(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &b.Lists)
		if err != nil {
			err = fmt.Errorf("load build list: %s: %s", name, err)
			log.Error(err)
			return err
		}
	default:
		err := fmt.Errorf("load build list: %s: %s", name, conf.ErrUnsupportedFormat)
		log.Error(err)
		return err
	}
	return nil
}

// Get returns the requested build list, or an error
func (b *BuildLists) Get(s string) (List, error) {
	l, ok := b.Lists[s]
	if !ok {
		return List{}, fmt.Errorf("%s: unknown build list", s)
	}
	return l, nil
}

// mergeKeysFromComponentMaps takes a variadic array of packer component maps
// and returns a merged, de-duped slice of keys for those maps.
func mergeKeysFromComponentMaps(m ...map[string]Componenter) []string {
	cnt := 0
	keys := make([][]string, len(m))
	// For each passed interface
	for i, tmpM := range m {
		cnt = 0
		tmpK := make([]string, len(tmpM))
		for k := range tmpM {
			tmpK[cnt] = k
			cnt++
		}
		keys[i] = tmpK
	}
	// Merge the slices, de-dupes keys.
	return MergeSlices(keys...)
}
