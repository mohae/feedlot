// Generate Packer templates and associated files for consumption by Packer.
//
// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

// Package ranchr is a package for organizing Rancher code. It also contains the package
// level variables and sets up logging.
package ranchr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	appName            = "RANCHER"

	// EnvConfig is the name of the environment variable name for the config file.
	EnvConfig          = appName + "_CONFIG"

	// EnvBuildsFile is the name of the environment variable name for the builds file.
	EnvBuildsFile      = appName + "_BUILDS_FILE"

	// EnvBuildListsFile is the name of the environment variable name for the build lists file.
	EnvBuildListsFile  = appName + "_BUILD_LISTS_FILE"

	// EnvDefaultsFile is the name of the environment variable name for the defaults file.
	EnvDefaultsFile    = appName + "_DEFAULTS_FILE"

	// EnvSupportedFile is the name of the environment variable name for the supported file.
	EnvSupportedFile   = appName + "_SUPPORTED_FILE"

	// EnvParamDelimStart is the name of the environment variable name for the delimter that starts Rancher variables.
	EnvParamDelimStart = appName + "_PARAM_DELIM_START"

	// EnvLogToFile is the name of the environment variable name for whether or not Rancher logs to a file.
	EnvLogToFile       = appName + "_LOG_TO_FILE"

	// EnvLogFilename is the name of the environment variable name for the log filename, if logging to file is enabled..
	EnvLogFilename     = appName + "_LOG_FILENAME"

	// EnvLogLevelFile is the name of the environment variable name for the file output's log level.
	EnvLogLevelFile    = appName + "_LOG_LEVEL_FILE"

	// EnvLogLevelStdout is the name of the environment variable name for stdout's log level.
	EnvLogLevelStdout  = appName + "_LOG_LEVEL_STDOUT"
)

var (
	// BuilderCommon is the name of the common builder section in the toml files.
	BuilderCommon = "common"

	// BuilderVBox is the name of the VirtualBox builder section in the toml files.
	BuilderVBox   = "virtualbox-iso"

	// BuilderVMWare is the name of the VMWare builder section in the toml files.
	BuilderVMWare = "vmware-iso"

	// PostProcessorVagrant is the name of the Vagrant PostProcessor
	PostProcessorVagrant = "vagrant"

	// ProvisionerAnsible is the name of the Ansible Provisioner
	ProvisionerAnsible = "ansible-local"

	// ProvisionerFileUploads is the name of the Salt Provisioner
	ProvisionerFileUploads = "file"

	// ProvisionerSalt is the name of the Salt Provisioner
	ProvisionerSalt = "salt-masterless"

	// ProvisionerShellScripts is the name of the Shell Scripts Provisioner
	ProvisionerShellScripts = "shell"
)

var (
	supportedDistros  *supported
	supportedDefaults map[string]rawTemplate
	supportedBuilds   builds
	supportedLoaded   bool
)

// AppConfig contains the current Rancher configuration...loaded at start-up.
var AppConfig appConfig

type appConfig struct {
	BuildsFile      string `toml:"builds_file"`
	BuildListsFile  string `toml:"build_lists_file"`
	DefaultsFile    string `toml:"defaults_file"`
	LogToFile       bool   `toml:"log_to_file"`
	LogFilename     string `toml:"log_filename"`
	LogLevelFile    string `toml:"log_level_file"`
	LogLevelStdout  string `toml:"log_level_stdout"`
	ParamDelimStart string `toml:"param_delim_start"`
	SupportedFile   string `toml:"Supported_file"`
}

// ArgsFilter has all the valid commandline flags for the build-subcommand.
type ArgsFilter struct {
	// Arch is a distribution specific string for the OS's target 
	// architecture.
	Arch    string

	// Distro is the name of the distribution, this value is consistent
	// with Packer.
	Distro  string

	// Image is the type of ISO image that is to be used. This is a 
	// distribution specific value.
	Image   string

	// Release is the release number or string of the ISO that is to be 
	// used. The valid values are distribution specific.
	Release string
}

// SetEnv sets the environment variables, if they do not already exist.
//
// The location of the rancher.cfg file can be overridden by setting its ENV
// variable prior to running Rancher. In addition, any of the other Rancher
// TOML file locations can be overridden by setting their corresponding ENV
// variable prior to running Rancher. The settings in the rancher.cfg file are
// only used if their corresponding ENV variables aren't set.
//
// ENV variables are used by rancher for the location of its TOML files and
// Rancher's logging settings.
func SetEnv() error {
	var err error
	var tmp string
	tmp = os.Getenv(EnvConfig)

	if tmp == "" {
		tmp = "rancher.cfg"
	}

	if _, err = toml.DecodeFile(tmp, &AppConfig); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	tmp = os.Getenv(EnvBuildsFile)

	if tmp == "" {

		if err = os.Setenv(EnvBuildsFile, AppConfig.BuildsFile); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvBuildListsFile)

	if tmp == "" {

		if err = os.Setenv(EnvBuildListsFile, AppConfig.BuildListsFile); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvDefaultsFile)

	if tmp == "" {

		if err = os.Setenv(EnvDefaultsFile, AppConfig.DefaultsFile); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvLogToFile)

	if tmp == "" {

		if err = os.Setenv(EnvLogToFile, strconv.FormatBool(AppConfig.LogToFile)); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvLogFilename)

	if tmp == "" {

		if err = os.Setenv(EnvLogFilename, AppConfig.LogFilename); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvLogLevelFile)

	if tmp == "" {

		if err = os.Setenv(EnvLogLevelFile, AppConfig.LogLevelFile); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvLogLevelStdout)

	if tmp == "" {

		if err = os.Setenv(EnvLogLevelStdout, AppConfig.LogLevelStdout); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	if tmp == "" {

		if err = os.Setenv(EnvParamDelimStart, AppConfig.ParamDelimStart); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	tmp = os.Getenv(EnvSupportedFile)

	if tmp == "" {

		if err = os.Setenv(EnvSupportedFile, AppConfig.SupportedFile); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	return nil
}

// Load the default and supported configuration and create the distro defaults
// out of them.
func loadSupported() error {
	var err error

	if supportedDistros, supportedDefaults, err = distrosInf(); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	err = supportedBuilds.LoadOnce()

	if err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	supportedLoaded = true
	return nil
}

// Load the application defaults and the default settings for each supported
// distribution. These get merged to create the default settings for each
// distribution as the configured default settings for a supported distro,
// as defined in the supported.toml, only define distro specific settings
// and the settings that the supported configuration overrides.
func distrosInf() (*supported, map[string]rawTemplate, error) {
	d := &defaults{}
	s := &supported{}
	var err error

	err = d.LoadOnce()

	if err != nil {
		jww.ERROR.Println(err.Error())
		return s, nil, err
	}

	err = s.LoadOnce()

	if err != nil {
		jww.ERROR.Println(err.Error())
		return s, nil, err
	}

	dd := map[string]rawTemplate{}

	if dd, err = setDistrosDefaults(d, s); err != nil {
		jww.ERROR.Println(err.Error())
		return s, nil, err
	}

	return s, dd, nil
}

// BuildDistro creates a build based on the target distro's defaults. The 
// ArgsFilter contains information on the target distro and any overrides
// that are to be applied to the build.
// Returns an error or nil if successful.
func BuildDistro(a ArgsFilter) error {
	if !supportedLoaded {

		if err := loadSupported(); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	if err := buildPackerTemplateFromDistro(a); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	argString := ""

	if a.Arch != "" {
		argString += "Arch=" + a.Arch
	}

	if a.Image != "" {

		if argString != "" {
			argString += ", "
		}

		argString += "Image=" + a.Image
	}

	if a.Release != "" {

		if argString != "" {
			argString += ", "
		}

		argString += "Release=" + a.Release

	}

	jww.INFO.Printf("Packer template built for %v using: %s", a.Distro, argString)
	return nil

}

// Create Packer templates from specified build templates.
func buildPackerTemplateFromDistro(a ArgsFilter) error {
	var d rawTemplate
	var found bool

	if a.Distro == "" {
		err := errors.New("Cannot build a packer template because no target distro information was passed.")
		jww.ERROR.Println(err.Error())
		return err
	}

	// Get the default for this distro, if one isn't found then it isn't Supported.
	if d, found = supportedDefaults[a.Distro]; !found {
		err := errors.New("Cannot build a packer template from passed distro: " + a.Distro + " is not Supported. Please pass a Supported distribution.")
		jww.ERROR.Println(err.Error())
		return err
	}

	// If any overrides were passed, set them.
	if a.Arch != "" {
		d.Arch = a.Arch
	}

	if a.Image != "" {
		d.Image = a.Image
	}

	if a.Release != "" {
		d.Release = a.Release
	}

	//	d.BuildName = ":type-:release-:arch-:image-rancher"
	// Now everything can get put in a template
	rTpl := newRawTemplate()
	pTpl := packerTemplate{}
	var err error
	rTpl.createDistroTemplate(d)

	// Since distro builds don't actually have a build name, we create one
	// out of the args used to create it.
	rTpl.BuildName = d.Type + "-" + d.Release + "-" + d.Arch + "-" + d.Image

	// Now that the raw template has been made, create a Packer template out of it
	if pTpl, err = rTpl.createPackerTemplate(); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	// Get the scripts for this build, if any.
	var scripts []string
	scripts = rTpl.ScriptNames()

	// Create the JSON version of the Packer template. This also handles creation of
	// the build directory and copying all files that the Packer template needs to the
	// build directory.
	// TODO break this call up or rename the function
	jww.TRACE.Println("Distro based template built; build the template for JSON")

	if err := pTpl.TemplateToFileJSON(rTpl.IODirInf, rTpl.BuildInf, scripts); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	jww.INFO.Println("Created Packer template and associated build directory for " + d.BuildName)
	return nil
}

// BuildBuilds manages the process of creating Packer Build templates out of 
// the passed build names. All builds are done concurrently. 
// Returns either a message providing information about the processing of the
// requested builds or an error.
func BuildBuilds(buildNames ...string) (string, error) {
	if buildNames[0] == "" {
		err := errors.New("Nothing to build. No build name was passed")
		jww.ERROR.Println(err.Error())
		return "", err
	}

	// Only load supported if it hasn't been loaded. Even though LoadSupported
	// uses a mutex to control access to prevent race conditions, no need to
	// call it if its already loaded.
	if !supportedLoaded {

		if err := loadSupported(); err != nil {
			jww.ERROR.Println(err.Error())
			return "", err
		}

	}

	// Make as many channels as there are build requests.
	var errorCount, builtCount int
	nBuilds := len(buildNames)
	doneCh := make(chan error, nBuilds)

	// Process each build request
	for i := 0; i < nBuilds; i++ {
		go buildPackerTemplateFromNamedBuild(buildNames[i], doneCh)
	}

	// Wait for channel done responses.
	for i := 0; i < nBuilds; i++ {
		err := <-doneCh

		if err != nil {
			return "", err
			errorCount++
		} else {
			builtCount++
		}

	}

	return fmt.Sprintf("Create Packer templates from named builds: %v Builds were successfully processed and %v Builds resulted in an error.", builtCount, errorCount), nil
}

// buildPackerTemplateFromNamedBuild creates a Packer tmeplate and associated
// artifacts for the passed build.
func buildPackerTemplateFromNamedBuild(buildName string, doneCh chan error) {
	if buildName == "" {
		err := errors.New("unable to build a Packer Template from a named build: no build name was passed")
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	if !supportedLoaded {

		if err := loadSupported(); err != nil {
			jww.ERROR.Println(err.Error())
			doneCh <- err
			return
		}

	}

	var tpl, bld rawTemplate
	var ok bool
	// Check the type and create the defaults for that type, if it doesn't already exist.
	tpl = rawTemplate{}
	bld = rawTemplate{}
	bld, ok = supportedBuilds.Build[buildName]

	if !ok {
		err := errors.New("Unable to create template for the requested build, " + buildName + ". Requested Build definition was not found.")
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	// See if the distro default exists.
	if tpl, ok = supportedDefaults[bld.Type]; !ok {
		err := errors.New("Requested distribution, " + bld.Type + ", is not Supported. The Packer template for the requested build could not be created.")
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	// Set build iso information overrides, if any.
	if bld.Arch != "" {
		tpl.Arch = bld.Arch
	}

	if bld.Image != "" {
		tpl.Image = bld.Image
	}

	if bld.Release != "" {
		tpl.Release = bld.Release
	}

	bld.BuildName = buildName

	// create build template() then call create packertemplate
	tpl.build = supportedDefaults[bld.Type].build
	tpl.mergeBuildSettings(bld)
	pTpl := packerTemplate{}
	var err error

	if pTpl, err = tpl.createPackerTemplate(); err != nil {
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	// Process the scripts for the Packer template.
	var scripts []string
	scripts = tpl.ScriptNames()

	if err = pTpl.TemplateToFileJSON(tpl.IODirInf, tpl.BuildInf, scripts); err != nil {
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	jww.INFO.Println("Created Packer template and associated build directory for build:" + buildName + ".")
	doneCh <- nil
	return
}

// Takes the name of the command file, including path, and returns a slice of
// shell commands. Each command within the file is separated by a newline.
// Returns error if an error occurs with the file.
func commandsFromFile(name string) (commands []string, err error) {
	if name == "" {
		err = errors.New("the passed Command filename was empty")
		jww.ERROR.Println(err.Error())
		return commands, err
	}

	f, errf := os.Open(name)
	if errf != nil {
		err = errf
		jww.ERROR.Println(errf.Error())
		return commands, err
	}

	// always close what's been opened and check returned error
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			jww.WARN.Println(cerr.Error())
			err = cerr
		}
	}()

	//New Reader for the string
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}

	if err = scanner.Err(); err != nil {
		jww.WARN.Println(err.Error())
		return
	}

	return commands, nil
}

// setDistrosDefaults takes the defaults and suppported settings and creates
// default build settings for each supported distro.
func setDistrosDefaults(d *defaults, s *supported) (map[string]rawTemplate, error) {
	jww.TRACE.Printf("defaults: %v\nsupported %v", d, s)
	// Create the distro default map
	dd := map[string]rawTemplate{}

	// Generate the default settings for each distro.
	for k, v := range s.Distro {

		// See if the base url exists for non centos distros
		if v.BaseURL == "" && k != "centos" {
			err := errors.New(k + " does not have its BaseURL configured.")
			jww.CRITICAL.Println(err.Error())
			return nil, err

		}

		// Create the struct for the default settings
		tmp := newRawTemplate()

		// First assign it all the default settings.
		tmp.BuildInf = d.BuildInf
		tmp.IODirInf = d.IODirInf
		tmp.PackerInf = d.PackerInf
		tmp.build = d.build
		tmp.Type = k

		// Now update it with the distro settings.
		tmp.BaseURL = appendSlash(v.BaseURL)
		tmp.Arch, tmp.Image, tmp.Release = getDefaultISOInfo(v.DefImage)
		tmp.BuildInf.update(v.BuildInf)
		tmp.IODirInf.update(v.IODirInf)
		tmp.PackerInf.update(v.PackerInf)
		tmp.mergeDistroSettings(v)

		// assign it to the return slice
		dd[k] = tmp
	}

	return dd, nil
}

// Takes two slices and returns the de-duped, merged list. The elements are
// returned in order of first encounter-duplicate keys are discarded.
func mergeSlices(s1 []string, s2 []string) []string {
	// If nothing is received return nothing
	if (s1 == nil || len(s1) <= 0) && (s2 == nil || len(s2) <= 0) {
		return nil
	}

	if s1 == nil || len(s1) <= 0 {
		return s2
	}

	if s2 == nil || len(s2) == 0 {
		return s1
	}

	// Make a slice with a length equal to the sum of the two input slices.
	tempSl := make([]string, len(s1)+len(s2))
	copy(tempSl, s1)
	i := len(s1) - 1
	var found bool

	// Go through every element in the second slice.
	for _, v := range s2 {

		// See if the key already exists
		for k, tmp := range s1 {

			if v == tmp {
				// it already exists
				found = true
				tempSl[k] = v
				break
			}

		}

		if !found {
			i++
			tempSl[i] = v
		}

		found = false

	}

	// Shrink the slice back down.
	retSl := make([]string, i+1)
	copy(retSl, tempSl)

	return retSl
}

// mergeSettingsSlices merges two slices of settings. In cases of a key
// collision, the second slice, s2, takes precedence. There are no duplicates
// at the end of this operation.
//
// Since settings use  embedded key=value pairs, the key is extracted from each
// value and matches are performed on the key only as the value will be
// different if the key appears in both slices.
func mergeSettingsSlices(s1 []string, s2 []string) []string {
	if (s1 == nil || len(s1) <= 0) && (s2 == nil || len(s2) <= 0) {
		return nil
	}

	if s1 == nil || len(s1) <= 0 {
		return s2
	}

	if s2 == nil || len(s2) <= 0 {
		return s1
	}

	ms1 := map[string]interface{}{}
	// Create a map of variables from the first slice for comparison reasons.
	ms1 = varMapFromSlice(s1)

	if ms1 == nil {
		jww.CRITICAL.Println("Unable to create a variable map from the passed slice.")
	}

	// Make a slice with a length equal to the sum of the two input slices.
	tempSl := make([]string, len(s1)+len(s2))
	copy(tempSl, s1)
	i := len(s1) - 1
	indx := 0
	var k string

	// For each element in the second slice, get the key. If it already
	// exists, update the existing value, otherwise add it to the merged
	// slice
	for _, v := range s2 {
		k, _ = parseVar(v)

		if _, ok := ms1[k]; ok {
			// This key already exists. Find it and update it.
			indx = keyIndexInVarSlice(k, tempSl)

			if indx < 0 {
				jww.WARN.Println("The key, " + k + ", was not updated to '" + v + "' because it was not found in the target slice.")
			} else {
				tempSl[indx] = v
			}

		} else {
			i++
			tempSl[i] = v
		}

	}

	// Shrink the slice back down.
	retSl := make([]string, i+1)
	copy(retSl, tempSl)

	return retSl
}

// varMapFromSlice creates a map from the passed slice. A Rancher var string
// contains a key=value string.
func varMapFromSlice(vars []string) map[string]interface{} {
	if vars == nil {
		jww.WARN.Println("Unable to create a Packer Settings map because no variables were received")
		return nil
	}

	vmap := make(map[string]interface{})
	// Go through each element and create a map entry from it.
	for _, variable := range vars {
		k, v := parseVar(variable)
		vmap[k] = v
	}

	return vmap
}

// parseVar: takes a string in the form of `key=value` and returns the key-value pair.
func parseVar(s string) (k string, v string) {
	if s == "" {
		return
	}

	// The key is assumed to be everything before the first equal sign.
	// The value is assumed to be everything after the first equal sign and
	// may also contain equal signs.
	// Both the key and value can have leading and trailing spaces. These
	// will be trimmed.
	arr := strings.SplitN(s, "=", 2)
	k = strings.Trim(arr[0], " ")

	// If the split resulted in 2 elements (key & value), get the trimmed
	// value.
	if len(arr) == 2 {
		v = strings.Trim(arr[1], " ")
	}

	return
}

// Searches for the passed key in the slice and returns its index if found, or
// -1 if not found; 0 is a valid index on a slice. The string to seArch is in
// the form of 'key=value'.
func keyIndexInVarSlice(key string, sl []string) int {
	//Go through the slice and find the matching key
	for i, s := range sl {
		k, _ := parseVar(s)

		// if the keys match, return its index.
		if k == key {
			return i
		}

	}

	// If we've gotten here, it wasn't found. Return -1 (not found)
	return -1
}

// getPackerVariableFromString takes the passed string and creates a Packer
// variable from it and returns that string.
func getPackerVariableFromString(s string) string {
	if s == "" {
		return s
	}
	return "{{user `" + s + "` }}"
}

// getDefaultISOInfo accepts a slice of strings and returns Arch, Image, and
// Release info extracted from that slice.
func getDefaultISOInfo(d []string) (Arch string, Image string, Release string) {

	for _, val := range d {
		k, v := parseVar(val)

		switch k {
		case "arch":
			Arch = v
		case "image":
			Image = v
		case "release":
			Release = v
		default:
			jww.WARN.Println("Unknown default key: " + k)
		}

	}

	return
}

// getMergedBuilders merges old and new builder settings nad returns the
// resulting builder.
func getMergedBuilders(old map[string]builder, new map[string]builder) map[string]builder {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return old
	}

	// Get all the keys in both maps
	// Get the all keys from both maps
	var keys[]string
	keys = keysFromMaps(old, new)

	bM := map[string]builder{}
	for _, v := range keys3 {
		b := builder{}
		b = old[v]
		b.mergeSettings(new[v].Settings)
		b.mergeVMSettings(new[v].VMSettings)
		bM[v] = b
	}

	return bM
}

// Merges the new config with the old. The updates occur as follows:
//
//	* The existing configuration is used when no `new` postProcessors are
//	  specified.
//	* When 1 or more `new` postProcessors are specified, they will replace
//        all existing postProcessors. In this situation, if a postProcessor
//	  exists in the `old` map but it does not exist in the `new` map, that
//        postProcessor will be orphaned.
// If there isn't a new config, return the existing as there are no
// overrides
func getMergedPostProcessors(old map[string]postProcessors, new map[string]postProcessors) map[string]postProcessors {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return old
	}

	// Get the all keys from both maps
	var keys[]string
	keys = keysFromMaps(old, new)

	pM := map[string]postProcessors{}

	for _, v := range keys {
		p := postProcessors{}
		p = old[v]
		p.mergeVMSettings(new[v].VMSettings)
		pM[v] = p
	}

	return pM
}

// merges the new config with the old. The updates occur as follows:
//
//	* The existing configuration is used when no `new` provisioners are
//	  specified.
//	* When 1 or more `new` provisioners are specified, they will replace
//        all existing provisioners. In this situation, if a provisioners
//	  exists in the `old` map but it does not exist in the `new` map, that
//        provisioners will be orphaned.
func getMergedProvisioners(old map[string]provisioners, new map[string]provisioners) map[string]provisioners {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return old
	}

	// Get the all keys from both maps
	var keys[]string
	keys = keysFromMaps(old, new)

	pM := map[string]provisioners{}

	for _, v := range keys {
		p := provisioners{}
		p = old[v]
		p.mergeVMSettings(new[v].VMSettings)
		pM[v] = p
	}

	return pM
}

// appendSlash appends a slash to the passed string. If the string already ends
// in a slash, nothing is done.
func appendSlash(s string) string {
	// Don't append empty strings
	if s == "" {
		return s
	}

	if !strings.HasSuffix(s, "/") {
		s += "/"
	}

	return s
}

// trimSuffix trims the passed suffix from the passed string, if it exists.
func trimSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}

	return s
}

// copyFile copies a file from source directory to destination directory. It
// returns either the number of bytes written or an error.
func copyFile(file string, srcDir string, destDir string) (written int64, err error) {
	if srcDir == "" {
		err := errors.New("copyFile: no source directory passed")
		jww.ERROR.Println(err.Error())
		return 0, err
	}

	if destDir == "" {
		err := errors.New("copyFile: no destination directory passed")
		jww.ERROR.Println(err.Error())
		return 0, err
	}

	if file == "" {
		err := errors.New("copyFile: no filename passed")
		jww.ERROR.Println(err.Error())
		return 0, err
	}

	srcDir = appendSlash(srcDir)
	destDir = appendSlash(destDir)
	src := srcDir + file
	dest := destDir + file

	// Create the scripts dir and copy each script from sript_src to out_dir/scripts/
	// while keeping track of success/failures.
	if err = os.MkdirAll(destDir, os.FileMode(0766)); err != nil {
		jww.ERROR.Println(err.Error())
		return 0, err
	}

	var fs, fd *os.File

	// Open the source script
	if fs, err = os.Open(src); err != nil {
		jww.ERROR.Println(err.Error())
		return 0, err
	}
	defer fs.Close()

	// Open the destination, create or truncate as needed.
	fd, err = os.Create(dest)
	if err != nil {
		jww.ERROR.Println(err.Error())
		return 0, err
	}
	defer func() {
		if cerr := fd.Close(); cerr != nil && err == nil {
			jww.WARN.Println(cerr.Error())
			err = cerr
		}
	}()

	return io.Copy(fd, fs)
}

// copyDirContent takes 2 directory paths and copies the contents from src to
// dest get the contents of srcDir.
func copyDirContent(srcDir string, destDir string) error {
	exists, err := pathExists(srcDir)
	if err != nil {
		return err
	}

	if !exists {
		err = errors.New("Source, " + srcDir + ", does not exist. Nothing copied.")
		jww.INFO.Println(err.Error())
		return err
	}

	dir := Archive{}
	err = dir.DirWalk(srcDir)

	if err != nil {
		return err
	}

	for _, file := range dir.Files {

		if file.info == nil {
			// if the info is empty, whatever this entry represents
			// doesn't actually exist.
			err := errors.New(file.p + " does not exist")
			jww.ERROR.Println(err.Error())
			return err
		}

		if file.info.IsDir() {

			if err := os.MkdirAll(file.p, os.FileMode(0766)); err != nil {
				jww.ERROR.Println(err.Error())
				return err
			}

		} else {

			if _, err := copyFile(file.info.Name(), srcDir, destDir); err != nil {
				jww.ERROR.Println(err.Error())
				return err
			}

		}

	}
	return nil
}

// deleteDirContent deletes the contents of a directory.
func deleteDirContent(dir string) error {
	jww.DEBUG.Printf("dir: %s", dir)
	var dirs []string
	// see if the directory exists first, actually any error results in the
	// same handling so just return on any error instead of doing an
	// os.IsNotExist(err)
	if _, err := os.Stat(dir); err != nil {

		if os.IsNotExist(err) {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	dirInf := directory{}
	dirInf.DirWalk(dir)
	jww.TRACE.Printf("dirIng: %+v", dirInf)
	dir = appendSlash(dir)

	for _, file := range dirInf.Files {
		jww.TRACE.Printf("process: %v", dir+file.p)

		if file.info.IsDir() {
			dirs = append(dirs, dir+file.p)
			jww.TRACE.Printf("added directory: %v", dir+file.p)
		} else {

			if err := os.Remove(dir + file.p); err != nil {
				jww.ERROR.Println(err.Error())
				return err
			}

			jww.TRACE.Printf("deleted: %v", dir+file.p)
		}
	}

	// all the files should now be deleted so its safe to delete the directories
	// do this in reverse order
	for i := len(dirs) - 1; i >= 0; i-- {

		jww.TRACE.Printf("process directory: %v", dirs[i])

		if err := os.Remove(dirs[i]); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	return nil
}

// Substring returns a substring of s starting at i for a length of l chars.
// If the requested index + requested length is greater than the length of the
// string, the string contents from the index to the end of the string will be
// returned instead. Note this assumes UTF-8, i.e. uses rune.
func Substring(s string, i, l int) string {
	if i <= 0 {
		return ""
	}

	if l <= 0 {
		return ""
	}

	r := []rune(s)
	length := i + l

	if length > len(r) {
		length = len(r)
	}

	return string(r[i:length])
}

func pathExists(p string) (bool, error) {
	_, err := os.Stat(p)

	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// Takes a variadic array of maps and returns a merged key
func getKeysFromMaps(m ...map[string]interface{}) []string {
	mapCnt := 0
	keyCnt :=0
	var tmpK []string
	var keys [][]string
        for _, tmpM := range m {
		// Get all the keys
		for k := range tmpM {
			keys[cnt] = k
			cnt++
		}
	}

	keys3 = mergeSlices(keys1, keys2)
	}

	return keys
}
		
