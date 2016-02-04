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
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	cjsn "github.com/mohae/cjson"
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
	Provisioners map[string]provisioner `toml:"provisioners" json:"provisioners"`
}

// copy makes a deep copy of the build and returns it.
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

// setTypes goes through each component map and check's if the
// templateSection Type is set.  If it isn't, the ID (key) is used to
// set it.
func (b *build) setTypes() {
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

// templateSection is used as an embedded type.
type templateSection struct {
	// Type is the actual Packer component type, this may or may not be the
	// same as the map key (ID).
	Type string `toml:"type" json:"type"`
	// Settings are string settings in "key=value" format.
	Settings []string `toml:"settings" json:"settings"`
	// Arrays are the string array settings.
	Arrays map[string]interface{} `toml:"arrays" json:"arrays"`
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

// mergeArrays merges the received array with the current one.
func (t *templateSection) mergeArrays(n map[string]interface{}) {
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

// postProcessor.mergeSettings  merges the settings section of a
// post-processor with the passed slice of settings. New values supercede
// existing ones.
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
	// If empty, no region filtering will be applied.
	Region string `toml:"region" json:"region"`
	// Country to use when selecting the image mirror; for use with CentOS.
	// This corresponds to the 'Country' column in the
	// https://centos.org/download/full-mirrorlist.csv file.
	//
	// If empty, no country filtering will be applied.
	//
	//  For Region 'US', this is used as the state field.
	Country string `toml:"country" json:"country"`
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
	// If empty, no sponsor filtering will be applied.
	Sponsor string `toml:"sponsor" json:"country"`
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
	if v.Region != "" {
		b.Region = v.Region
	}
	if v.Country != "" {
		b.Country = v.Country
	}
}

// IODirInf is used to store information about where Rancher can find and put
// things. Source files are always in a SourceDir.
type IODirInf struct {
	// Include the packer component name in the path. If true, the
	// component.String() value will be added as the parent of the output
	// resource path: i.e. OutDir/component.String()/resource_name.  This is a
	// pointer so that whether or not this setting was actually set can be
	// determined, otherwise determining whether it was an explicit false or
	// empty would not be possible.
	IncludeComponentString *bool `toml:"include_component_string" json:"include_component_string"`
	// The directory to use for example runs
	OutputDir string `toml:"output_dir" json:"output_dir"`
	// If the output dir path is relative to the conf_dir.  If true, the path is
	// resolved relative to the conf_dir.  Otherwise, the path is used as is.
	// This is a pointer so that whether or not this setting was actually set can
	// be determined, otherwise determining whether it was an explicit false or
	// empty would not be possible.
	OutputDirIsRelative *bool `toml:"output_dir_is_relative" json:"output_dir_is_relative"`
	// The directory that contains the source files for this build.
	SourceDir string `toml:"source_dir" json:"source_dir"`
	// If the source dir path is relative to the conf_dir.  If true, the path is
	// resolved relative to the conf_dir.  Otherwise, the path is used as is.
	// This is a pointer so that whether or not this setting was actually set can
	// be determined, otherwise determining whether it was an explicit false or
	// empty would not be possible.
	SourceDirIsRelative *bool `toml:"source_dir_is_relative" json:"source_dir_is_relative"`
}

// Only update when a value exists; empty strings don't count as being set.
func (i *IODirInf) update(v IODirInf) {
	if v.OutputDir != "" {
		i.OutputDir = appendSlash(v.OutputDir)
	}
	if v.OutputDirIsRelative != nil {
		i.OutputDirIsRelative = v.OutputDirIsRelative
	}
	if v.SourceDir != "" {
		i.SourceDir = appendSlash(v.SourceDir)
	}
	if v.SourceDirIsRelative != nil {
		i.SourceDirIsRelative = v.SourceDirIsRelative
	}
	if v.IncludeComponentString != nil {
		i.IncludeComponentString = v.IncludeComponentString
	}
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
	name, format, err := findConfigFile(getConfigFile(p, fmt.Sprintf("%s.%s", "default", contour.GetString(Format))))
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	switch format {
	case TOML:
		_, err := toml.DecodeFile(name, &d)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	case JSON:
		var buff []byte
		buff, err = ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &d)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	default:
		jww.ERROR.Println(ErrUnsupportedFormat)
		return ErrUnsupportedFormat
	}
	d.build.setTypes()
	d.loaded = true
	return nil
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
	name, format, err := findConfigFile(getConfigFile(p, "supported"))
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	switch format {
	case TOML:
		_, err := toml.DecodeFile(name, &s.Distro)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	case JSON:
		var buff []byte
		buff, err = ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &s.Distro)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	default:
		jww.ERROR.Println(ErrUnsupportedFormat)
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

// Load the build information from the provided name.
func (b *builds) Load(name string) error {
	if name == "" {
		err := errors.New("build: name was empty")
		jww.ERROR.Println(err)
		return err
	}
	switch CfgFormatFromString(contour.GetString(Format)) {
	case TOML:
		_, err := toml.DecodeFile(name, &b.Build)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	case JSON:
		buff, err := ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &b.Build)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	default:
		jww.ERROR.Println(ErrUnsupportedFormat)
		return ErrUnsupportedFormat
	}
	// get the dir of the filepath
	dir := filepath.Dir(name)
	// get the path info from the name
	for _, v := range b.Build {
		v.build.setTypes()
		v.setSourceDir(dir)
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
	name, format, err := findConfigFile(getConfigFile(p, "build_list"))
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	switch format {
	case TOML:
		_, err := toml.DecodeFile(name, &b.List)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	case JSON:
		var buff []byte
		buff, err = ioutil.ReadFile(name)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
		err = cjsn.Unmarshal(buff, &b.List)
		if err != nil {
			err = fmt.Errorf(fmt.Sprintf("%s: %s", name, err))
			jww.ERROR.Println(err)
			return err
		}
	default:
		jww.ERROR.Println(ErrUnsupportedFormat)
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

// SetCfgFile set's the appCFg from the app's cfg file and then applies any
// env vars that have been set. After this, settings can only be updated
// programmatically or via command-line flags.
//
// The default cfg file may not be the one found as the app config file may
// be in a different format. SetCfg first looks for it in the configured
// location. If it is not found, the alternate format is checked.
//
// Since Rancher supports operations without a config file, not finding one
// is not an error state.
//
// Currently supported config file formats:
//    TOML
//    JSON || CJSN
func SetCfgFile() error {
	// find the actual config filename, it may have a different extension as
	// formats can have more than one accepted extension, this is mainly to
	// handle CJSON vs JSON though.  The error is ignored becuase it doesn't
	// matter if the file doesn't exist.
	fname, _, _ := findConfigFile(contour.GetString(CfgFile))
	if fname == "" {
		return nil
	}
	contour.UpdateCfgFile(CfgFile, fname)
	err := contour.SetCfg()
	if err != nil {
		jww.ERROR.Print(err)
	}
	return err
}

// Take a config file name and checks to see if it exists.  If it doesn't
// exist, it checks to see if the file can be found under an alternate
// extension by checking what config format Rancher is set to use and
// iterating through the list of supported exts for that format.  If a file
// exists under a particular file + ext combination, that is returned.  If
// no match is found, the error on the original filename is returned so that
// the message information is consistent with what is expected.
func findConfigFile(fname string) (string, CfgFormat, error) {
	cf := CfgFormatFromString(contour.GetString(Format))
	_, err := os.Stat(fname)
	if err == nil {
		return fname, cf, nil
	}
	// if the file isn't found, look for it according to format extensions
	var exts []string
	switch cf {
	case JSON:
		exts = []string{"json", "jsn", "cjson", "cjsn", "JSON", "JSN", "CJSON", "CJSN"}
	case TOML:
		exts = []string{"toml", "tml", "TOML", "TML"}
	default:
		return "", UnsupportedCfgFormat, fmt.Errorf("unsupported config format: %s", contour.GetString(Format))
	}
	name := strings.TrimSuffix(fname, filepath.Ext(fname))
	for _, ext := range exts {
		n := fmt.Sprintf("%s.%s", name, ext)
		_, err := os.Stat(n)
		if err == nil {
			return n, cf, nil
		}
	}
	// nothing found
	return "", cf, err
}

// GetConfigFile returns the location of the provided conf file. This accounts
// for examples. An empty
//
// If the p field has a value, it is used as the dir path, instead of the
// confDir,
func getConfigFile(p, name string) string {
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
	if fname != "supported" {
		p = filepath.Join(p, contour.GetString(ConfDir))
	}
	if contour.GetBool(Example) {
		// example files always end in '.example'
		return filepath.Join(contour.GetString(ExampleDir), p, name)
	}
	return filepath.Join(p, name)
}
