// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

// Package ranchr implements the creation of Packer templates from Rancher
// build definitions.
package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mohae/contour"
	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// Contains most of the information for Packer templates within a Rancher
// Build.
type build struct {
	// Targeted builders: the values are consistent with Packer's, e.g.
	// `virtualbox.iso` is used for VirtualBox.
	BuilderTypes []string `toml:"builder_types" json:"builder_types"`
	// A map of builder configuration. There should always be a `common`
	// builder, which has settings common to both VMWare and VirtualBox.
	Builders map[string]builder `toml:"builders"`
	// Targeted post-processors: the values are consistent with Packer's, e.g.
	// `vagrant` is used for Vagrant.
	PostProcessorTypes []string `toml:"post_processor_types" json:"post_processor_types"`
	// A map of post-processor configurations.
	PostProcessors map[string]postProcessor `toml:"post_processors" json:"post_processors"`
	// Targeted provisioners: the values are consistent with Packer's, e.g.
	// `shell` is used for shell.
	ProvisionerTypes []string `toml:"provisioner_types" json:"provisioner_types"`
	// A map of provisioner configurations.
	Provisioners map[string]provisioner `toml:"provisioners"`
}

// build.DeepCopy makes a deep copy of the build and returns it.
func (b *build) copy() build {
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
func (b *builder) mergeSettings(sl []string) error {
	if sl == nil {
		return nil
	}
	var err error
	b.Settings, err = mergeSettingsSlices(b.Settings, sl)
	if err != nil {
		return fmt.Errorf("merge of builder settings failed: %s", err)
	}
	return nil
}

// mergeArrays merges the arrays section of a template builder
func (b *builder) mergeArrays(m map[string]interface{}) {
	b.Arrays = b.templateSection.mergeArrays(b.Arrays, m)
}

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
func (p *postProcessor) mergeSettings(sl []string) error {
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
		return fmt.Errorf("merge of post-processor settings failed: %s", err)
	}
	return nil
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
func (p *provisioner) mergeSettings(sl []string) error {
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
		return fmt.Errorf("merge of provisioner settings failed: %s", err)
	}
	return nil
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
	loaded bool
}

// Load loads the defualt settings. If the defaults have already been loaded
// nothing is done.
func (d *defaults) Load(p string) error {
	if d.loaded {
		return nil
	}
	name := GetConfFile(p, Default)
	switch contour.GetString(Format) {
	case "toml", "tml":
		_, err := toml.DecodeFile(name, &d)
		if err != nil {
			return decodeErr(name, err)
		}
	case "json", "jsn":
		f, err := os.Open(name)
		if err != nil {
			return decodeErr(name, err)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err != nil {
			return decodeErr(name, err)
		}
		for {
			err := dec.Decode(&d)
			if err == io.EOF {
				break
			}
			if err != nil {
				return decodeErr(name, err)
			}
		}
	default:
		return ErrUnsupportedFormat
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
	BuildName string `toml:"build_name" json:"build_name"`
	BaseURL   string `toml:"base_url" json:"base_url"`
}

func (i *BuildInf) update(b BuildInf) {
	if b.Name != "" {
		i.Name = b.Name
	}
	if b.BuildName != "" {
		i.BuildName = b.BuildName
	}
	if b.BaseURL != "" {
		i.BaseURL = b.BaseURL
	}
}

// IODirInf is used to store information about where Rancher can find and put
// things. Source files are always in a SourceDir.
type IODirInf struct {
	// Include the packer component name in the path. Even though it is used as a bool,
	// it is defined as a string so that it makes absense from a template detectable.
	// Any value that strconv.ParseBool can parse to true is accepted as true. If the
	// value is empty, parsed to false, or cannot be properly parsed, false is assumed.
	// If this is true, the component.String() value will be added as the parent of the
	// output resource: i.e. OutDir/component.String()/resource_name
	IncludeComponentString string `toml:"include_component_string" json:"include_component_string"`
	// The directory to use for example runs
	OutputDir string `toml:"output_dir" json:"output_dir"`
	// The directory that contains the source files for this build.
	SourceDir string `toml:"source_dir" json:"source_dir"`
}

func (i *IODirInf) update(inf IODirInf) {
	if inf.OutputDir != "" {
		i.OutputDir = appendSlash(inf.OutputDir)
	}
	if inf.SourceDir != "" {
		i.SourceDir = appendSlash(inf.SourceDir)
	}
	if inf.IncludeComponentString != "" {
		i.IncludeComponentString = inf.IncludeComponentString
	}
}

// includeComponentString returns whether, or not, the name of the component
// should be included. Any string that results in a parse error will be
// evaluated as false, otherwise any string that strconv.ParseBool() can parse
// is valid.
//
// Any value that errors is evaluated to false
func (i *IODirInf) includeComponentString() bool {
	b, _ := strconv.ParseBool(i.IncludeComponentString)
	return b
}

// check to see if the dirinf is set, if not, set them to their defaults
func (i *IODirInf) check() {
	if i.OutputDir == "" {
		i.OutputDir = fmt.Sprintf("%sbuildname", contour.GetString(ParamDelimStart))
	}
	if i.SourceDir == "" {
		i.SourceDir = "src"
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
	DefImage []string `toml:"default_image" json:"default_image"`
	// The configurations needed to generate the default settings for a build for this
	// distribution.
	build
}

// To add support for a distribution, the information about it must be added to
// the supported. file, in addition to adding the code to support it to the
// application.
type supported struct {
	Distro map[string]*distro
	loaded bool
}

// Load the supported distro info.
func (s *supported) Load(p string) error {
	name := GetConfFile(p, Supported)
	switch contour.GetString(Format) {
	case "toml", "tml":
		_, err := toml.DecodeFile(name, &s.Distro)
		if err != nil {
			return decodeErr(name, err)
		}
	case "json", "jsn":
		f, err := os.Open(name)
		if err != nil {
			return decodeErr(name, err)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		for {
			err := dec.Decode(&s.Distro)
			if err == io.EOF {
				break
			}
			if err != nil {
				return decodeErr(name, err)
			}
		}
	default:
		return ErrUnsupportedFormat
	}
	s.loaded = true
	return nil
}

// Struct to hold the builds.
type builds struct {
	Build  map[string]*rawTemplate
	loaded bool
}

// Load the build information.
func (b *builds) Load(name string) error {
	if name == "" {
		return filenameNotSetErr("build")
	}
	switch contour.GetString(Format) {
	case "toml", "tml":
		_, err := toml.DecodeFile(name, &b.Build)
		if err != nil {
			return decodeErr(name, err)
		}
	case "json", "jsn":
		f, err := os.Open(name)
		if err != nil {
			return decodeErr(name, err)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		for {
			err := dec.Decode(&b.Build)
			if err == io.EOF {
				break
			}
			if err != nil {
				return decodeErr(name, err)
			}
		}
	default:
		return ErrUnsupportedFormat
	}
	b.loaded = true
	return nil
}

// getBuildTemplate returns the requested build template, or an error if it
// can't be found. Th
func getBuildTemplate(name string) (*rawTemplate, error) {
	var r *rawTemplate
	for _, blds := range Builds {
		for n, bTpl := range blds.Build {
			if n == name {
				r = bTpl.copy()
				r.BuildName = name
				goto found
			}
		}
	}
	return nil, fmt.Errorf("build not found: %s", name)
found:
	return r, nil
}

// Contains lists of builds.
type buildLists struct {
	List map[string]list
}

// A list contains 1 or more builds.
type list struct {
	Builds []string
}

// Load loads the build lists. It accepts a path prefix; which is mainly used
// for testing ATM.
func (b *buildLists) Load(p string) error {

	// Load the build lists.
	name := GetConfFile(p, BuildList)
	if name == "" {
		return filenameNotSetErr(BuildList)
	}
	switch contour.GetString(Format) {
	case "toml", "tml":
		_, err := toml.DecodeFile(name, &b.List)
		if err != nil {
			return decodeErr(name, err)
		}
	case "json", "jsn":
		f, err := os.Open(name)
		if err != nil {
			return decodeErr(name, err)
		}
		defer f.Close()
		dec := json.NewDecoder(f)
		if err != nil {
			return decodeErr(name, err)
		}
		for {
			err := dec.Decode(&b.List)
			if err == io.EOF {
				break
			}
			if err != nil {
				return decodeErr(name, err)
			}
		}
	default:
		return ErrUnsupportedFormat
	}
	return nil
}

// Get returns the requested build list, or an error
func (b *buildLists) Get(s string) (list, error) {
	l, ok := b.List[s]
	if !ok {
		return list{}, fmt.Errorf("%s is not a valid build_list name", s)
	}
	return l, nil
}

// SetCfgFile set's the appCFg from the app's cfg file and then applies any env
// vars that have been set. After this, settings can only be updated
// programmatically or via command-line flags.
//
// The default cfg file may not be the one found as the app config file may be
// in a different format. SetCfg first looks for it in the configured location.
// If it is not found, the alternate format is checked.
//
// Since Rancher supports operations without a config file, not finding one is
// not an error state.
//
// Currently supported config file formats:
//    TOML
//    JSON
func SetCfgFile() error {
	// determine the correct file, This is done here because it's less ugly than the alternatives.
	_, err := os.Stat(contour.GetString(CfgFile))
	if err != nil && os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	err = contour.SetCfg()
	if err != nil {
		jww.ERROR.Print(err)
	}
	return err
}

// GetConfFile returns the location of the provided conf file. This accounts
// for examples. An empty
//
// If the p field has a value, it is used as the dir path, instead of the
// confDir,
func GetConfFile(p, name string) string {
	if name == "" {
		return name
	}
	var fname string
	// save the filename and add an extension to it if it doesn't exist
	if filepath.Ext(name) == "" {
		fname = name
		name = fmt.Sprintf("%s.%s", name, contour.GetString(Format))
	} else {
		fname = strings.TrimSuffix(name, filepath.Ext(name))
	}
	// if the path wasn't passed, use the confdir, unless this file is the supported
	// file. A path is prefixed to supported file only if this func receives one;
	// the ConfDir is not used for supported.
	if fname != Supported {
		p = filepath.Join(p, contour.GetString(ConfDir))
	}
	if contour.GetBool(Example) {
		// example files always end in '.example'
		return filepath.Join(contour.GetString(ExampleDir), p, name)
	}
	return filepath.Join(p, name)
}
