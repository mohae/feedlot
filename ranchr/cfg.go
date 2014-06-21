// Generate Packer templates and associated files for consumption by Packer.
//
// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

// Provides structs for storing the data from the various TOML files that
// Rancher uses, along with methods associated with the structs.
package ranchr

import (
	"errors"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

type build struct {
	// Contains most of the information for Packer templates within a Rancher Build.
	BuilderType    []string                  `toml:"builder_type"`
	Builders       map[string]builder        `toml:"builders"`
	PostProcessors map[string]postProcessors `toml:"post_processors"`
	Provisioners   map[string]provisioners   `toml:"provisioners"`
}

type builder struct {
	// Defines a representation of the builder section of a Packer template.
	Settings   []string `toml:"Settings"`
	VMSettings []string `toml:"vm_Settings"`
}

func (b *builder) mergeSettings(new []string) {
	// Merge the settings section of a builder. New values supercede existing ones.
	b.Settings = mergeSettingsSlices(b.Settings, new)
}

func (b *builder) mergeVMSettings(new []string) {
	// Merge the VMSettings section of a builder. New values supercede existing ones.
	b.VMSettings = mergeSettingsSlices(b.VMSettings, new)
}

func (b *builder) settingsToMap(r *rawTemplate) map[string]interface{} {
	// Go through all of the Settings and convert them to a map. Each setting
	// is parsed into its constituent parts. The value then goes through
	// variable replacement to ensure that the settings are properly resolved.
	var k, v string
	m := make(map[string]interface{}, len(b.Settings)+len(b.VMSettings))
	for _, s := range b.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		m[k] = v
	}
	return m
}

type postProcessors struct {
	// Type for handling the post-processor section of the configs.
	Settings []string
}

func (p *postProcessors) mergeSettings(new []string) {
	// Merge the settings section of a post-processor. New values supercede existing ones.
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

func (p *postProcessors) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	// Go through all of the Settings and convert them to a map. Each setting
	// is parsed into its constituent parts. The value then goes through
	// variable replacement to ensure that the settings are properly resolved.
	var k, v string
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type
	for _, s := range p.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		m[k] = v
	}
	return m
}

type provisioners struct {
	// Type for handling the provisioners sections of the configs.
	Settings []string `toml:"settings"`
	Scripts  []string `toml:"scripts"`
}

func (p *provisioners) mergeSettings(new []string) {
	// Merge the settings section of a post-processor. New values supercede existing ones.
	p.Settings = mergeSettingsSlices(p.Settings, new)
}

func (p *provisioners) settingsToMap(Type string, r *rawTemplate) map[string]interface{} {
	// Go through all of the Settings and convert them to a map. Each setting
	// is parsed into its constituent parts. The value then goes through
	// variable replacement to ensure that the settings are properly resolved.
	var k, v string
	m := make(map[string]interface{}, len(p.Settings))
	m["type"] = Type
	for _, s := range p.Settings {
		k, v = parseVar(s)
		v = r.replaceVariables(v)
		switch k {
		case "execute_command":
			// Not being able to get the command file won't end the template
			// generation. Instead, the returned error will be used as the
			// setting value.
			// This is probably a bad idea and I should revisit TODO
			if c, err := commandsFromFile(v); err != nil {
				v = "Error: " + err.Error()
				err = nil
			} else {
				v = c[0]
			}
		}
		m[k] = v
	}
	return m
}

func (p *provisioners) setScripts(new []string) {
	// Scripts are only replaced if it has values, otherwise the existing values are used.
	if len(new) > 0 {
		p.Scripts = new
	}
}

type defaults struct {
	// Defaults is used to store Rancher application level defaults for Packer templates.
	IODirInf
	PackerInf
	BuildInf
	build
	load   sync.Once
	loaded bool
}

type BuildInf struct {
	// Information about a specific build.
	Name      string `toml:"name"`
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

	return
}

type IODirInf struct {
	// IODirInf is used to store information about where Rancher can find and put things.
	// Source files are always in a SrcDir, e.g. HTTPSrcDir is the source directory for
	// the HTTP directory. The destination directory is always a Dir, e.g. HTTPDir is the
	// destination directory for the HTTP directory.
	CommandsSrcDir string `toml:"commands_src_dir"`
	HTTPDir        string `toml:"http_dir"`
	HTTPSrcDir     string `toml:"http_src_dir"`
	OutDir         string `toml:"out_dir"`
	ScriptsDir     string `toml:"scripts_dir"`
	ScriptsSrcDir  string `toml:"scripts_src_dir"`
	SrcDir         string `toml:"src_dir"`
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

	return
}

type PackerInf struct {
	// PackerInf is used to store information about a Packer Template. In Packer, these
	// fields are optional, put used here because they are always printed out in a
	// template as custom creation of template output hasn't been written--it may never
	// be written.
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

func (d *defaults) LoadOnce() {
	loadFunc := func() {
		name := os.Getenv(EnvDefaultsFile)
		if name == "" {
			logger.Critical("could not retrieve the default Settings file because the " + EnvDefaultsFile + " ENV 	variable was not set. Either set it or check your rancher.cfg setting")
			return
		}
		if _, err := toml.DecodeFile(name, &d); err != nil {
			logger.Critical(err.Error())
			return
		}
		d.loaded = true
		return
	}
	d.load.Do(loadFunc)
	logger.Debugf("defaults loaded: %v", d)
	return

}

// To add support for a distribution, the information about it must be added to
// the supported. file, in addition to adding the code to support it to the
// application.
type supported struct {
	Distro map[string]distro
	load   sync.Once
	loaded bool
}

type distro struct {
	// Struct to hold the details of supported distros. From this information a user
	// should be able to build a Packer template by only executing the following, at
	// minimum:
	//	$ rancher build -distro=ubuntu
	// All settings can be overridden. The information here represents the standard
	// box configuration for its respective distribution.
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

func (s *supported) LoadOnce() {
	loadFunc := func() {
		name := os.Getenv(EnvSupportedFile)
		if name == "" {
			logger.Critical("could not retrieve the Supported information because the " + EnvSupportedFile + " Env variable was not set. Either set it or check your rancher.cfg setting")
			return
		}
		if _, err := toml.DecodeFile(name, &s); err != nil {
			logger.Critical(err.Error())
			return
		}
		s.loaded = true
		return
	}
	s.load.Do(loadFunc)
	logger.Debugf("supported loaded: %v", s)
	return
}

// Struct to hold the builds.
type builds struct {
	Build  map[string]rawTemplate
	load   sync.Once
	loaded bool
}

func (b *builds) LoadOnce() {
	loadFunc := func() {
		name := os.Getenv(EnvBuildsFile)
		if name == "" {
			logger.Critical("could not retrieve the Builds configurations because the " + EnvBuildsFile + "Env variable was not set. Either set it or check your rancher.cfg setting")
			return
		}
		if _, err := toml.DecodeFile(name, &b); err != nil {
			logger.Critical(err.Error())
			return
		}
		b.loaded = true
		return
	}
	b.load.Do(loadFunc)
	logger.Debugf("builds loaded: %v", b)
	return
}

type buildLists struct {
	// Contains lists of builds.
	List map[string]list
}

type list struct {
	// A list of builds. Each list contains one or more builds.
	Builds []string
}

func (b *buildLists) Load() error {
	// Load the build lists.
	name := os.Getenv(EnvBuildListsFile)
	if name == "" {
		err := errors.New("could not retrieve the BuildLists file because the " + EnvBuildListsFile + " Env variable was not set. Either set it or check your rancher.cfg setting")
		logger.Error(err.Error())
		return err
	}
	if _, err := toml.DecodeFile(name, &b); err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Debugf("buildLists loaded: %v", b)
	return nil
}
