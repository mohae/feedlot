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
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	json "github.com/mohae/customjson"
	jww "github.com/spf13/jwalterweatherman"
)

var (
	appName = "RANCHER"

	// EnvRancerhFiile is the name of the environment variable name for Rancher's config file.
	EnvRancherFile = appName + "_CONFIG"

	// EnvBuildsFile is the name of the environment variable name for the builds file.
	EnvBuildsFile = appName + "_BUILDS_FILE"

	// EnvBuildListsFile is the name of the environment variable name for the build lists file.
	EnvBuildListsFile = appName + "_BUILD_LISTS_FILE"

	// EnvDefaultsFile is the name of the environment variable name for the defaults file.
	EnvDefaultsFile = appName + "_DEFAULTS_FILE"

	// EnvSupportedFile is the name of the environment variable name for the supported file.
	EnvSupportedFile = appName + "_SUPPORTED_FILE"

	// EnvParamDelimStart is the name of the environment variable name for the delimter that starts Rancher variables.
	EnvParamDelimStart = appName + "_PARAM_DELIM_START"

	// EnvLogToFile is the name of the environment variable name for whether or not Rancher logs to a file.
	EnvLogToFile = appName + "_LOG_TO_FILE"

	// EnvLogFilename is the name of the environment variable name for the log filename, if logging to file is enabled..
	EnvLogFilename = appName + "_LOG_FILENAME"

	// EnvLogLevelFile is the name of the environment variable name for the file output's log level.
	EnvLogLevelFile = appName + "_LOG_LEVEL_FILE"

	// EnvLogLevelStdout is the name of the environment variable name for stdout's log level.
	EnvLogLevelStdout = appName + "_LOG_LEVEL_STDOUT"
)

var (
	// BuilderCommon is the name of the common builder section in the toml files.
	BuilderCommon = "common"

	// BuilderVirtualBoxISO is the name of the VirtualBox builder section in the toml files.
	BuilderVirtualBoxISO = "virtualbox-iso"

	// BuilderVirtualBoxOVF is the name of the VirtualBox builder section in the toml files.
	BuilderVirtualBoxOVF = "virtualbox-ovf"

	// BuilderVMWareISO is the name of the VirtualBox builder section in the toml files.
	BuilderVMWareISO = "vmware-iso"

	// BuilderVMWareOVF is the name of the VirtualBox builder section in the toml files.
	BuilderVMWareOVF = "vmware-ovf"

	// PostProcessorVagrant is the name of the Vagrant PostProcessor
	PostProcessorVagrant = "vagrant"

	// PostProcessorVagrant is the name of the Vagrant PostProcessor
	PostProcessorVagrantCloud = "vagrant-cloud"

	// ProvisionerAnsible is the name of the Ansible Provisioner
	ProvisionerAnsible = "ansible-local"

	// ProvisionerFile is the name of the File Provisioner
	ProvisionerFile = "file"

	// ProvisionerSalt is the name of the Salt Provisioner
	ProvisionerSalt = "salt-masterless"

	// ProvisionerShell is the name of the Shell Provisioner
	ProvisionerShell = "shell"

	// ProvisionerShellScripts is an alias for the Shell provisioner as the
	// Shell provisioner is technically the Shell Script provisioner, in
	// the Packer documentation
	ProvisionerShellScripts = ProvisionerShell
)

var (
	// indent: default indent to use for marshal stuff
	indent = "    "

	//VMSetting is the constant for builders with vm-settings
	VMSettings = "vm_settings"
)

var Builds *builds
var Distros distroDefaults

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
	Arch string

	// Distro is the name of the distribution, this value is consistent
	// with Packer.
	Distro string

	// Image is the type of ISO image that is to be used. This is a
	// distribution specific value.
	Image string

	// Release is the release number or string of the ISO that is to be
	// used. The valid values are distribution specific.
	Release string
}

// distroDefaults contains the defaults for all supported distros and a flag
// whether its been set or not.
type distroDefaults struct {
	Templates map[string]*rawTemplate
	IsSet     bool
}

// GetTemplate returns a deep copy of the default template for the passed
// distro name. If the distro does not exist, an error is returned.
func (d *distroDefaults) GetTemplate(n string) (*rawTemplate, error) {
	var t *rawTemplate
	var ok bool

	if t, ok = d.Templates[n]; !ok {
		return t, fmt.Errorf("distroDefaults.GetTemplate: The requested Distro, " + n + " is not supported. No template to return")
	}

	copy := newRawTemplate()
	copy.PackerInf = t.PackerInf
	copy.IODirInf = t.IODirInf
	copy.BuildInf = t.BuildInf
	copy.releaseISO = t.releaseISO
	copy.date = t.date
	copy.delim = t.delim
	copy.Type = t.Type
	copy.Arch = t.Arch
	copy.Image = t.Image
	copy.Release = t.Release
	copy.varVals = t.varVals
	copy.vars = t.vars
	copy.build = t.build.DeepCopy()

	return copy, nil
}

// Set sets the default templates for each distro.
func (d *distroDefaults) Set() error {
	dflts := &defaults{}
	if err := dflts.LoadOnce(); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	s := &supported{}
	if err := s.LoadOnce(); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	d.Templates = map[string]*rawTemplate{}

	// Generate the default settings for each distro.
	for k, v := range s.Distro {
		// See if the base url exists for non centos distros
		if v.BaseURL == "" && k != "centos" {
			err := errors.New("setDistroDefaults: " + k + " does not have its BaseURL configured.")
			jww.CRITICAL.Println(err.Error())
			return err

		}

		// Create the struct for the default settings
		tmp := newRawTemplate()
		// First assign it all the default settings.
		tmp.BuildInf = dflts.BuildInf
		tmp.IODirInf = dflts.IODirInf
		tmp.PackerInf = dflts.PackerInf
		tmp.build = dflts.build.DeepCopy()
		tmp.Type = k

		// Now update it with the distro settings.
		tmp.BaseURL = appendSlash(v.BaseURL)
		tmp.Arch, tmp.Image, tmp.Release = getDefaultISOInfo(v.DefImage)
		tmp.setDefaults(v)
		d.Templates[k] = tmp
	}

	Distros.IsSet = true
	return nil
}

/*
// rancherBuilds holds the build definitions for rancher.
type rancherBuilds struct {
	Build *build
}

//
func (r *rancherBuilds) Get(n string) build

*/

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
	tmp = os.Getenv(EnvRancherFile)

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

// Set DistroDefaults
func loadBuilds() error {
	var err error
	Builds = &builds{}
	if err = Builds.LoadOnce(); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}
	return nil
}

// BuildDistro creates a build based on the target distro's defaults. The
// ArgsFilter contains information on the target distro and any overrides
// that are to be applied to the build.
// Returns an error or nil if successful.
func BuildDistro(a ArgsFilter) error {
	if !Distros.IsSet {

		if err := Distros.Set(); err != nil {
			jww.ERROR.Println(err.Error())
			return err
		}

	}

	fmt.Println("BuildDistro:\n" + json.MarshalIndentToString(Distros, "", indent))
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

	jww.INFO.Printf("BuildDistro: Packer template built for %v using: %s", a.Distro, argString)
	return nil

}

// Create Packer templates from specified build templates.
func buildPackerTemplateFromDistro(a ArgsFilter) error {
	jww.DEBUG.Println("buildPackerTemplateFromDistro: Enter")
	var t *rawTemplate
	var err error

	if a.Distro == "" {
		err = errors.New("buildPackerTemplateFromDistro: Cannot build a packer template because no target distro information was passed.")
		jww.ERROR.Println(err.Error())
		return err
	}

	// Get the default for this distro, if one isn't found then it isn't Supported.
	if t, err = Distros.GetTemplate(a.Distro); err != nil {
		err = errors.New("buildPackerTemplateFromDistro: Cannot build a packer template from passed distro: " + a.Distro + " is not Supported. Please pass a Supported distribution.")
		jww.ERROR.Println(err.Error())
		return err
	}

	// If any overrides were passed, set them.
	if a.Arch != "" {
		t.Arch = a.Arch
	}

	if a.Image != "" {
		t.Image = a.Image
	}

	if a.Release != "" {
		t.Release = a.Release
	}

	t.BuildName = ":type-:release-:arch-:image-rancher"

	//	// make a copy of the .
	//	rTpl := newRawTemplate()
	//	rTpl.updateBuilders(d.Builders)

	// Since distro builds don't actually have a build name, we create one
	// out of the args used to create it.
	t.BuildName = t.Type + "-" + t.Release + "-" + t.Arch + "-" + t.Image

	jww.TRACE.Printf("\trawtemplate: %v\n", json.MarshalIndentToString(t, "", indent))
	pTpl := packerTemplate{}

	// Now that the raw template has been made, create a Packer template out of it
	if pTpl, err = t.createPackerTemplate(); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	jww.TRACE.Printf("\tpackerTemplate:  %v\n", json.MarshalIndentToString(pTpl, "", indent))

	// Get the scripts for this build, if any.
	var scripts []string
	scripts = t.ScriptNames()

	// Create the JSON version of the Packer template. This also handles creation of
	// the build directory and copying all files that the Packer template needs to the
	// build directory.
	if err := pTpl.create(t.IODirInf, t.BuildInf, scripts); err != nil {
		jww.ERROR.Println(err.Error())
		return err
	}

	jww.INFO.Println("Created Packer template and associated build directory for " + t.BuildName)
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
	if !Distros.IsSet {

		if err := Distros.Set(); err != nil {
			jww.ERROR.Println(err.Error())
			return "", err
		}

	}

	// First load the build information
	loadBuilds()

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

	var tpl, bld *rawTemplate
	var ok bool
	// Check the type and create the defaults for that type, if it doesn't already exist.
	tpl = &rawTemplate{}
	bld = &rawTemplate{}

	if bld, ok = Builds.Build[buildName]; !ok {
		err := errors.New("Unable to create template for the requested build, " + buildName + ". Requested Build definition was not found.")
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	// See if the distro default exists.
	if tpl, ok = Distros.Templates[bld.Type]; !ok {
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
	tpl.build = Distros.Templates[bld.Type].build
	tpl.updateBuildSettings(bld)

	pTpl := packerTemplate{}
	var err error

	if pTpl, err = tpl.createPackerTemplate(); err != nil {
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}

	_ = pTpl
	// Process the scripts for the Packer template.
	var scripts []string
	scripts = tpl.ScriptNames()

	if err = pTpl.create(tpl.IODirInf, tpl.BuildInf, scripts); err != nil {
		jww.ERROR.Println(err.Error())
		doneCh <- err
		return
	}
	jww.INFO.Println("Created Packer template and associated build directory for build:" + buildName + ".")
	doneCh <- nil

	return
}

// getSliceLenFromIface takes an interface that's assumed to be a slice and
// returns its length. If it is not a slice, an error is returned.
func getSliceLenFromIface(v interface{}) (int, error) {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		sl := reflect.ValueOf(v)
		return sl.Len(), nil
	}

	return 0, fmt.Errorf("err: getSliceLenFromIface expected a slice, go" + reflect.TypeOf(v).Kind().String())
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

// MergeSlices takes a variadic input of []string and returns a string slice
// with all of the values within the slices merged, later occurrences of the
// same key override previous.
func MergeSlices(s ...[]string) []string {
	// If nothing is received return nothing
	if s == nil {
		return nil
	}

	// If there is only 1, there is nothing to merge
	if len(s) == 1 {
		return s[0]
	}

	// Otherwise merge slices, starting with s1 & s2
	var merged []string

	for _, tmpS := range s {
		merged = mergeSlices(merged, tmpS)
	}

	return merged
}

// mergeSlices Takes two slices and returns the de-duped, merged list. The elements are
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
// Slice structure: ptr | len | cap
// Copying a slice means the slice structure is copied but the underlying array
// is not copied so now you have two slices that both point to the underlying array
func mergeSettingsSlices(s1 []string, s2 []string) []string {
	l1 := len(s1)
	l2 := len(s2)

	if l1 == 0 && l2 == 0 {
		return nil
	}

	// Make a slice with a length equal to the sum of the two input slices.
	merged := make([]string, l1+l2)

	// Copy the first slice.
	i := copy(merged, s1)

	// if nothing was copied, i == 0 , just copy the 2nd slice.
	if i == 0 {
		copy(merged, s2)
		return merged
	}

	ms1 := map[string]string{}
	// Create a map of variables from the first slice for comparison reasons.
	ms1 = varMapFromSlice(s1)

	if ms1 == nil {
		jww.CRITICAL.Println("Unable to create a variable map from the passed slice.")
	}

	// For each element in the second slice, get the key. If it already
	// exists, update the existing value, otherwise add it to the merged
	// slice
	var indx int
	var v, key string
	for _, v = range s2 {
		key, _ = parseVar(v)
		if _, ok := ms1[key]; ok {
			// This key already exists. Find it and update it.
			indx = keyIndexInVarSlice(key, merged)

			if indx < 0 {
				jww.WARN.Println("The key, " + key + ", was not updated to '" + v + "' because it was not found in the target slice.")
			} else {
				merged[indx] = v
			}

		} else {
			// i is the index of the next element to add, a result of
			// i being set to the count of the items copied, which is
			// 1 greater than the index, or the index of the next item
			// should it exist. Instead, it is updated after adding the
			// new value as, after add, i points to the current element.
			merged[i] = v
			fmt.Printf("\tadded indx=%v:\t%v\n", i, v)
			i++
		}

	}

	// Shrink the slice back down to == its length
	ret := make([]string, i)
	copy(ret, merged)
	return ret
}

// varMapFromSlice creates a map from the passed slice. A Rancher var string
// contains a key=value string.
func varMapFromSlice(vars []string) map[string]string {
	if vars == nil {
		jww.WARN.Println("Unable to create a Packer Settings map because no variables were received")
		return nil
	}

	vmap := make(map[string]string)
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

// merges the new config with the old. The updates occur as follows:
//
//	* The existing configuration is used when no `new` provisioners are
//	  specified.
//	* When 1 or more `new` provisioners are specified, they will replace
//        all existing provisioners. In this situation, if a provisioners
//	  exists in the `old` map but it does not exist in the `new` map, that
//        provisioners will be orphaned.
func getMergedProvisioners(old map[string]provisioner, new map[string]provisioner) map[string]provisioner {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return old
	}
	// Convert to an interface.
	var ifaceOld map[string]interface{} = make(map[string]interface{}, len(old))
	for i, o := range old {
		ifaceOld[i] = o
	}

	// Convert to an interface.
	var ifaceNew map[string]interface{} = make(map[string]interface{}, len(new))
	for i, n := range new {
		ifaceNew[i] = n
	}

	// Get the all keys from both maps
	var keys []string
	keys = mergedKeysFromMaps(ifaceOld, ifaceNew)

	pM := map[string]provisioner{}

	for _, v := range keys {
		p := provisioner{}
		p = old[v]
		//		p.mergeSettings(new[v].Settings)
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

// mergedKeysFromMaps takes a variadic array of maps and returns a merged
// slice of keys for those maps.
func mergedKeysFromMaps(m ...map[string]interface{}) []string {
	cnt := 0
	types := make([][]string, len(m))

	// For each passed interface
	for i, tmpM := range m {
		cnt = 0
		tmpK := make([]string, len(tmpM))
		for k := range tmpM {
			tmpK[cnt] = k
			cnt++
		}
		types[i] = tmpK
	}

	// Merge the slices, de-dupes keys.
	mergedKeys := MergeSlices(types...)

	return mergedKeys
}
