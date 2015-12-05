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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mohae/cjsn"
	"github.com/mohae/contour"
	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

// Componenter is an interface for Packer components, i.e. builder,
// post-processor, and provisioner.
type Componenter interface {
	getType() string
}

// Contains most of the information for Packer templates within a Rancher
// Build.  The keys of the maps are the IDs
type build struct {
	// Targeted builders: either the ID of the builder section or the Packer
	// component type value can be used, e.g. ID: "vbox" or "virtualbox-iso".
	BuilderIDs []string `toml:"builder_ids" json:"builder_ids"`
	// A map of builder configurations.  When using a VMWare or VirtualBox
	// builder, there is usually a 'common' builder, which has settings common
	// to both VMWare and VirtualBox.
	Builders map[string]builder `toml:"builders" json:"builders"`
	// Targeted post-processors: either the ID of the post-processor or the
	// Packer component type value can be used, e.g. ID: "docker" or
	// "docker-push".
	PostProcessorIDs []string `toml:"post_processor_ids" json:"post_processor_ids"`
	// A map of post-processor configurations.
	PostProcessors map[string]postProcessor `toml:"post_processors" json:"post_processors"`
	// Targeted provisioners: either the ID of the provisioners or the Packer
	// component type value can be used, e.g. ID: "preprocess" or "shell".
	ProvisionerIDs []string `toml:"provisioner_ids" json:"provisioner_ids"`
	// A map of provisioner configurations.
	Provisioners map[string]provisioner `toml:"provisioners"`
}

// build.DeepCopy makes a deep copy of the build and returns it.
func (b *build) copy() build {
	newB := build{
		Builders:       map[string]builder{},
		PostProcessors: map[string]postProcessor{},
		Provisioners:   map[string]provisioner{},
	}
	if b.BuilderIDs != nil {
		newB.BuilderIDs = make([]string, len(b.BuilderIDs), len(b.BuilderIDs))
		copy(newB.BuilderIDs, b.BuilderIDs)
	}
	if b.PostProcessorIDs != nil {
		newB.PostProcessorIDs = make([]string, len(b.PostProcessorIDs), len(b.PostProcessorIDs))
		copy(newB.PostProcessorIDs, b.PostProcessorIDs)
	}
	if b.ProvisionerIDs != nil {
		newB.ProvisionerIDs = make([]string, len(b.ProvisionerIDs), len(b.ProvisionerIDs))
		copy(newB.ProvisionerIDs, b.ProvisionerIDs)
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
	// Type is the actual Packer component type, this may or may not be the
	// same as the map key (ID).
	Type string
	// Settings are string settings in "key=value" format.
	Settings []string
	// Arrays are the string array settings.
	Arrays map[string]interface{}
}

// templateSection.DeepCopy updates its information with new via a deep copy.
func (t *templateSection) DeepCopy(ts templateSection) {
	// Copy Type
	t.Type = ts.Type
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
	keys = mergeKeysFromMaps(old, n)
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

// getType returns the type for this component.  This is on builder instead of
// templateSection because the component needs to fulfill the interface,
// not the templateSection.
func (b builder) getType() string {
	return b.Type
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

// getType returns the type for this component.  This is on postProcessor
// instead of templateSection because the component needs to fulfill the
// interface, not the templateSection.
func (p postProcessor) getType() string {
	return p.Type
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

// getType returns the type for this component.  This is on provisioner
// instead of templateSection because the component needs to fulfill the
// interface, not the templateSection.
func (p provisioner) getType() string {
	return p.Type
}

// provisioner.DeepCopy copies the postProcessor values instead of the
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
	case "cjsn", "json":
		var b []byte
		var err, alterr error
		b, err = ioutil.ReadFile(name)
		if err != nil {
			// try alternate
			name, _ = getAltJSONName(name)
			b, alterr = ioutil.ReadFile(name)
			if alterr != nil {
				return decodeErr(name, err)
			}
		}
		err = cjsn.Unmarshal(b, &d)
		if err != nil {
			return decodeErr(name, err)
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
	case "cjsn", "json":
		var b []byte
		var err, alterr error
		b, err = ioutil.ReadFile(name)
		if err != nil {
			// try alternate
			name, _ = getAltJSONName(name)
			b, alterr = ioutil.ReadFile(name)
			if alterr != nil {
				// use the original error because the filename will
				// be what the user expects
				return decodeErr(name, err)
			}
		}
		err = cjsn.Unmarshal(b, &s.Distro)
		if err != nil {
			return decodeErr(name, err)
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
	case "cjsn", "json", "jsn":
		by, err := ioutil.ReadFile(name)
		if err != nil {
			return decodeErr(name, err)
		}
		err = cjsn.Unmarshal(by, &b.Build)
		if err != nil {
			return decodeErr(name, err)
		}
	default:
		return ErrUnsupportedFormat
	}
	// populate Type where applicable
	for name, bld := range b.Build {
		for k, tmp := range bld.build.Builders {
			if len(tmp.Type) == 0 {
				tmp.Type = k
				bld.build.Builders[k] = tmp
			}
		}
		for k, tmp := range bld.build.PostProcessors {
			if len(tmp.Type) == 0 {
				tmp.Type = k
				bld.build.PostProcessors[k] = tmp
			}
		}
		for k, tmp := range bld.build.Provisioners {
			if len(tmp.Type) == 0 {
				tmp.Type = k
				bld.build.Provisioners[k] = tmp
			}
		}
		b.Build[name] = bld
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
func (bl *buildLists) Load(p string) error {

	// Load the build lists.
	name := GetConfFile(p, BuildList)
	if name == "" {
		return filenameNotSetErr(BuildList)
	}
	switch contour.GetString(Format) {
	case "toml", "tml":
		_, err := toml.DecodeFile(name, &bl.List)
		if err != nil {
			return decodeErr(name, err)
		}
	case "cjsn", "json":
		var b []byte
		var err, alterr error
		b, err = ioutil.ReadFile(name)
		if err != nil {
			// try alternate
			name, _ = getAltJSONName(name)
			b, alterr = ioutil.ReadFile(name)
			if alterr != nil {
				return decodeErr(name, err)
			}
		}
		err = cjsn.Unmarshal(b, &bl.List)
		if err != nil {
			return decodeErr(name, err)
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
//    JSON || CJSN
func SetCfgFile() error {
	// determine the correct file, This is done here because it's less ugly than the alternatives.
	fname := contour.GetString(CfgFile)
	_, err := os.Stat(fname)
	if err != nil && os.IsNotExist(err) {
		// if this is json, try cjsn
		var ok bool
		fname, ok = getAltJSONName(fname)
		if !ok {
			return nil
		}
		_, err := os.Stat(fname)
		if err != nil && os.IsNotExist(err) {
			return nil
		}
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

// checks to see if the file ext is either .json or .cjsn.  If it is, it
// returns the name with the alternate ext, otherwise it returns false.
// This allows for transparent support of either JSON or CJSN.
func getAltJSONName(fname string) (string, bool) {
	ext := path.Ext(fname)
	switch ext {
	case ".json":
		return fmt.Sprintf("%s.cjsn", strings.TrimSuffix(fname, ext)), true
	case ".cjsn":
		return fmt.Sprintf("%s.json", strings.TrimSuffix(fname, ext)), true
	default:
		return "", false
	}
}
