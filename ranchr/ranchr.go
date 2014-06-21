// Generate Packer templates and associated files for consumption by Packer.
//
// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

// A Ranch is where Ranchers get their work done, same here. This package is
// just a way of organizing code for Rancher. It also contains the package
// level variables and sets up logging.
package ranchr

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	_ "time"

	"github.com/BurntSushi/toml"
	seelog "github.com/cihub/seelog"
)

var logger seelog.LoggerInterface

var (
	appName = "RANCHER"

	EnvConfig          = appName + "_CONFIG"
	EnvLogging         = appName + "_LOGGING"
	EnvLogFile         = appName + "_LOG_FILE"
	EnvLogLevel        = appName + "_LOG_LEVEL"
	EnvBuildsFile      = appName + "_BUILDS_FILE"
	EnvBuildListsFile  = appName + "_BUILD_LISTS_FILE"
	EnvDefaultsFile    = appName + "_DEFAULTS_FILE"
	EnvSupportedFile   = appName + "_SUPPORTED_FILE"
	EnvParamDelimStart = appName + "_PARAM_DELIM_START"
)

var (
	BuilderCommon = "common"
	BuilderVBox   = "virtualbox-iso"
	BuilderVMWare = "vmware-iso"
)

var (
	MessageEmptyValue = "Value expected but %v was empty"
)

var (
	supportedDistros  *supported
	supportedDefaults map[string]rawTemplate
	supportedBuilds   builds
	supportedLoaded   bool
)

type appConfig struct {
	Logging         string
	DefaultsFile    string `toml:"defaults_file"`
	SupportedFile   string `toml:"Supported_file"`
	BuildsFile      string `toml:"builds_file"`
	BuildListsFile  string `toml:"build_lists_file"`
	ParamDelimStart string `toml:"param_delim_start"`
}

type ArgsFilter struct {
	Arch    string
	Distro  string
	Image   string
	Release string
}

// Set the environment variables, if they do not already exist.
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
	var config appConfig
	var tmp string
	tmp = os.Getenv(EnvConfig)
	if tmp == "" {
		tmp = "rancher.cfg"
	}
	if _, err = toml.DecodeFile(tmp, &config); err != nil {
		logger.Error(err.Error())
		return err
	}
	tmp = os.Getenv(EnvDefaultsFile)
	if tmp == "" {
		if err = os.Setenv(EnvDefaultsFile, config.DefaultsFile); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	tmp = os.Getenv(EnvSupportedFile)
	if tmp == "" {
		if err = os.Setenv(EnvSupportedFile, config.SupportedFile); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	tmp = os.Getenv(EnvBuildsFile)
	if tmp == "" {
		if err = os.Setenv(EnvBuildsFile, config.BuildsFile); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	tmp = os.Getenv(EnvBuildListsFile)
	if tmp == "" {
		if err = os.Setenv(EnvBuildListsFile, config.BuildListsFile); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	tmp = os.Getenv(EnvParamDelimStart)
	if tmp == "" {
		if err = os.Setenv(EnvParamDelimStart, config.ParamDelimStart); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	tmp = os.Getenv(EnvLogging)
	if tmp == "" {
		if err = os.Setenv(EnvLogging, config.Logging); err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	return nil
}

func loadSupported() error {
	var err error
	if supportedDistros, supportedDefaults, err = distrosInf(); err != nil {
		logger.Error(err.Error())
		err = errors.New("Load of Default and Supported information failed. Please check prior log entries for more information")
		return err
	}
	supportedBuilds.LoadOnce()
	if supportedBuilds.loaded == false {
		err := errors.New("Load of Build information failed. Please check prior log entries for more information")
		logger.Error(err.Error())
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
	d.LoadOnce()
	if d.loaded == false {
		err := errors.New("Loading of the defaults file failed. Please check the log for more info.")
		return s, nil, err
	}
	s.LoadOnce()
	if s.loaded == false {
		err := errors.New("Loading of the supported file failed. Please check the log for more info.")
		return s, nil, err
	}
	dd := map[string]rawTemplate{}
	if dd, err = setDistrosDefaults(d, s); err != nil {
		logger.Error(err.Error())
		return s, nil, err
	}
	return s, dd, nil
}

func BuildDistro(a ArgsFilter) error {
	if !supportedLoaded {
		if err := loadSupported(); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	if err := buildPackerTemplateFromDistro(a); err != nil {
		logger.Error(err.Error())
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

	logger.Infof("Packer template built for %v using: %s", a.Distro, argString)
	return nil

}

// Create Packer templates from specified build templates.
func buildPackerTemplateFromDistro(a ArgsFilter) error {
	var d rawTemplate
	var found bool
	var err error
	if a.Distro == "" {
		err = errors.New("Cannot build a packer template because no target distro information was passed.")
		logger.Error(err.Error())
		return err
	}
	// Get the default for this distro, if one isn't found then it isn't Supported.
	if d, found = supportedDefaults[a.Distro]; !found {
		err = errors.New("Cannot build a packer template from passed distro: " + a.Distro + " is not Supported. Please pass a Supported distribution.")
		logger.Error(err.Error())
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
	rTpl.createDistroTemplate(d)
	rTpl.BuildName = d.Type + "-" + d.Release + "-" + d.Arch + "-" + d.Image
	// Now that the raw template has been made, create a Packer template out of it
	if pTpl, err = rTpl.createPackerTemplate(); err != nil {
		logger.Error(err.Error())
		return err
	}
	// Get the scripts for this build, if any.
	var scripts []string
	scripts = rTpl.ScriptNames()
	// Create the JSON version of the Packer template. This also handles creation of
	// the build directory and copying all files that the Packer template needs to the
	// build directory.
	// TODO break this call up or rename the function
	logger.Trace("Distro based template built; build the template for JSON")
	if err = pTpl.TemplateToFileJSON(rTpl.IODirInf, rTpl.BuildInf, scripts); err != nil {
		logger.Error(err.Error())
		return err
	}
	logger.Info("Created Packer template and associated build directory for " + d.BuildName)
	return nil
}

// Processes the passed build requests. Returns either a message providing
// information about the processing of the requested builds or an error.
func BuildBuilds(buildNames ...string) (string, error) {
	if !supportedLoaded {
		if err := loadSupported(); err != nil {
			logger.Error(err.Error())
			return "", err
		}
	}
	var errorCount, builtCount int
	nBuilds := len(buildNames)
	doneCh := make(chan error, nBuilds)
	for i := 0; i < nBuilds; i++ {
		go buildPackerTemplateFromNamedBuild(buildNames[i], doneCh)
	}
	for i := 0; i < nBuilds; i++ {
		err := <-doneCh
		if err != nil {
			errorCount++
		} else {
			builtCount++
		}
	}
	/*
		for _, buildName := range buildNames {
			if err := buildPackerTemplateFromNamedBuild(buildName); err != nil {
				logger.Error(err.Error())
				errorCount++
			} else {
				builtCount++
			}
		}
	*/
	return fmt.Sprintf("%v Build requests were processed with %v successfully processed and %v resulting in an error.", errorCount+builtCount, builtCount, errorCount), nil
}

func buildPackerTemplateFromNamedBuild(buildName string, doneCh chan error) {
	if buildName == "" {
		err := errors.New("unable to build a Packer Template from a named build: no build name was passed")
		logger.Error(err.Error())
		doneCh <- err
		return
	}
	if !supportedLoaded {
		if err := loadSupported(); err != nil {
			logger.Error(err.Error())
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
		logger.Error(err.Error())
		doneCh <- err
		return
	}
	// See if the distro default exists.
	if tpl, ok = supportedDefaults[bld.Type]; !ok {
		err := errors.New("Requested distribution, " + bld.Type + ", is not Supported. The Packer template for the requested build could not be created.")
		logger.Error(err.Error())
		doneCh <- err
		return
	}
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
		logger.Error(err.Error())
		doneCh <- err
		return
	}
	var scripts []string
	scripts = tpl.ScriptNames()
	if err = pTpl.TemplateToFileJSON(tpl.IODirInf, tpl.BuildInf, scripts); err != nil {
		logger.Error(err.Error())
		doneCh <- err
		return
	}
	logger.Info("Created Packer template and associated build directory for build:" + buildName + ".")
	doneCh <- nil
	return
}

func commandsFromFile(name string) (commands []string, err error) {
	// Takes the name of the command file, including path, and returns a slice of
	// shell commands. Each command within the file is separated by a newline.
	// Returns error if an error occurs with the file.
	if name == "" {
		err = errors.New("the passed Command filename was empty")
		logger.Error(err.Error())
		return
	}
	f, errf := os.Open(name)
	if errf != nil {
		err = errf
		logger.Error(errf.Error())
		return
	}
	// always close what's been opened and check returned error
	defer func() {
		if err = f.Close(); err != nil {
			logger.Warn(err.Error())
		}
	}()
	//New Reader for the string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		logger.Warn(err.Error())
		return
	}
	return commands, nil
}

func setDistrosDefaults(d *defaults, s *supported) (map[string]rawTemplate, error) {
	logger.Tracef("defaults: %v\nsupported %v", d, s)
	// Create the default and Supported info struct for the Supported distros.
	dd := map[string]rawTemplate{}
	for k, v := range s.Distro {
		if v.BaseURL == "" && k != "centos" {
			err := errors.New(k + " does not have its BaseURL configured.")
			logger.Critical(err.Error())
			return nil, err

		}
		tmp := newRawTemplate()
		tmp.BuildInf = d.BuildInf
		tmp.IODirInf = d.IODirInf
		tmp.PackerInf = d.PackerInf
		tmp.build = d.build	
		tmp.Type = k
		tmp.BaseURL = appendSlash(v.BaseURL)
		tmp.Arch, tmp.Image, tmp.Release = getDefaultISOInfo(v.DefImage)
		tmp.BuildInf.update(v.BuildInf)
		tmp.IODirInf.update(v.IODirInf)
		tmp.PackerInf.update(v.PackerInf)
		tmp.mergeDistroSettings(v)
		dd[k] = tmp
	}
	return dd, nil
}

func mergeSlices(s1 []string, s2 []string) []string {
	// Takes two slices and returns the merged list without duplicates.
	// The elements are returned in order of first encounter-duplicate keys
	// are discarded.
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

func mergeSettingsSlices(s1 []string, s2 []string) []string {
	// MergeSlices merges two slices. In cases of key collisions, the second slice,
	// s2, takes precedence. There are no duplicates at the end of this operation.
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
	ms1 = varMapFromSlice(s1)
	if ms1 == nil {
		logger.Critical("Unable to create a variable map from the passed slice.")
	}
	// Make a slice with a length equal to the sum of the two input slices.
	tempSl := make([]string, len(s1)+len(s2))
	copy(tempSl, s1)
	i := len(s1) - 1
	indx := 0
	var k string
	for _, v := range s2 {
		k, _ = parseVar(v)
		if _, ok := ms1[k]; ok {
			// This key already exists. Find it and update it.
			indx = keyIndexInVarSlice(k, tempSl)
			if indx < 0 {
				logger.Warn("The key, " + k + ", was not updated to '" + v + "' because it was not found in the target slice.")
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

func varMapFromSlice(vars []string) map[string]interface{} {
	// Converts a slice to a map[string]interface{} from a Rancher var string.
	// A Rancher var string contains a key=value string.
	if vars == nil {
		logger.Warn("Unable to create a Packer Settings map because no variables were received")
		return nil
	}
	vmap := make(map[string]interface{})
	for _, variable := range vars {
		k, v := parseVar(variable)
		vmap[k] = v
	}
	return vmap
}

func parseVar(s string) (k string, v string) {
	// parseVar: takes a string in the form of `key=value` and returns the key-value pair.
	if s == "" {
		return
	}
	arr := strings.SplitN(s, "=", 2)
	k = strings.Trim(arr[0], " ")
	if len(arr) == 2 {
		v = strings.Trim(arr[1], " ")
	}
	return
}

func keyIndexInVarSlice(key string, sl []string) int {
	// Searches for the passed key in the slice and returns its index if found, or
	// -1 if not found; 0 is a valid index on a slice. The string to seArch is in
	// the form of 'key=value'.
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

func getVariableName(s string) (string, error) {
	if s == "" {
		err := errors.New("no variable name was passed")
		logger.Error(err.Error())
		return "", err
	}
	return "{{user `" + s + "` }}", nil
}

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
			logger.Warn("Unknown default key: " + k)
		}
	}
	return
}

func getMergedBuilders(old map[string]builder, new map[string]builder) map[string]builder {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return old
	}
	// Get all the keys in both maps
	keys1 := make([]string, len(old))
	cnt := 0
	for k, _ := range old {
		keys1[cnt] = k
		cnt++
	}
	keys2 := make([]string, len(new))
	cnt = 0
	for k, _ := range new {
		keys2[cnt] = k
		cnt++
	}
	// Merge this slice down to get a final list of keys.
	var keys3 []string
	keys3 = mergeSlices(keys1, keys2)
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

func getMergedPostProcessors(old map[string]postProcessors, new map[string]postProcessors) map[string]postProcessors {
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
	if len(new) <= 0 {
		return old
	}
	// Go through each PostProcessors and merge new Settings with the old
	// Settings.
	tmp := map[string]postProcessors{}
	for k, v := range new {
		p := postProcessors{}
		p = old[k]
		p.mergeSettings(v.Settings)
		tmp[k] = p
	}
	return tmp
}

func getMergedProvisioners(old map[string]provisioners, new map[string]provisioners) map[string]provisioners {
	// merges the new config with the old. The updates occur as follows:
	//
	//	* The existing configuration is used when no `new` provisioners are
	//	  specified.
	//	* When 1 or more `new` provisioners are specified, they will replace
	//        all existing provisioners. In this situation, if a provisioners
	//	  exists in the `old` map but it does not exist in the `new` map, that
	//        provisioners will be orphaned.
	if len(new) <= 0 {
		return old
	}
	tmp := map[string]provisioners{}
	for k, v := range new {
		p := provisioners{}
		p = old[k]
		p.mergeSettings(v.Settings)
		if len(v.Scripts) > 0 {
			p.setScripts(v.Scripts)
		}
		tmp[k] = p
	}
	return tmp
}

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

func trimSuffix(s string, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func copyFile(srcDir string, destDir string, file string) (written int64, err error) {
	if srcDir == "" {
		err := errors.New("copyFile: no source directory passed")
		logger.Error(err.Error())
		return 0, err
	}
	if destDir == "" {
		err := errors.New("copyFile: no destination directory passed")
		logger.Error(err.Error())
		return 0, err
	}
	if file == "" {
		err := errors.New("copyFile: no filename passed")
		logger.Error(err.Error())
		return 0, err
	}

	srcDir = appendSlash(srcDir)
	destDir = appendSlash(destDir)
	src := srcDir + file
	dest := destDir + file
	// Create the scripts dir and copy each script from sript_src to out_dir/scripts/
	// while keeping track of success/failures.
	if err = os.MkdirAll(destDir, os.FileMode(0766)); err != nil {
		logger.Error(err.Error())
		return 0, err
	}
	var fs, fd *os.File
	// Open the source script
	if fs, err = os.Open(src); err != nil {
		logger.Error(err.Error())
		return 0, err
	}
	defer fs.Close()
	// Open the destination, create or truncate as needed.
	fd, err = os.Create(dest)
	if err != nil {
		logger.Error(err.Error())
		return 0, err
	}
	defer fd.Close()
	return io.Copy(fd, fs)
}

func copyDirContent(srcDir string, destDir string) error {
	// takes 2 directory paths and copies the contents from src to dest
	//get the contents of srcDir
	// The archive struct should be renamed to something more appropriate
	exists, err := pathExists(srcDir)
	if err != nil {
		return err
	}
	if !exists {
		err = errors.New("Source, " + srcDir + ", does not exist. Nothing copied.")
		logger.Info(err.Error())
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
			logger.Error(err.Error())
			return err
		}
		if file.info.IsDir() {
			if err := os.MkdirAll(file.p, os.FileMode(0766)); err != nil {
				logger.Error(err.Error())
				return err
			}
		} else {
			if _, err := copyFile(srcDir, destDir, file.info.Name()); err != nil {
				logger.Error(err.Error())
				return err
			}
		}
	}
	return nil
}

func deleteDirContent(dir string) error {
	// deletes the contents of a directory
	// see if the directory exists first, actually any error results in the
	// same handling so just return on any error instead of doing an
	// os.IsNotExist(err)
	logger.Debugf("dir: %s", dir)
	var dirs []string
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			logger.Error(err.Error())
			return err
		}
	}
	dirInf := directory{}
	dirInf.DirWalk(dir)
	logger.Tracef("dirIng: %+v", dirInf)
	dir = appendSlash(dir)
	for _, file := range dirInf.Files {
		logger.Tracef("process: %v", dir+file.p)
		if file.info.IsDir() {
			dirs = append(dirs, dir+file.p)
			logger.Tracef("added directory: %v", dir+file.p)
		} else {
			if err := os.Remove(dir + file.p); err != nil {
				logger.Error(err.Error())
				return err
			}
			logger.Tracef("deleted: %v", dir+file.p)
		}
	}
	// all the files should now be deleted so its safe to delete the directories
	// do this in reverse order
	for i := len(dirs) - 1; i >= 0; i-- {
		logger.Tracef("process directory: %v", dirs[i])
		if err := os.Remove(dirs[i]); err != nil {
			logger.Error(err.Error())
			return err
		}
	}
	return nil
}

// Given a string, a position, and a length, return the substring containted
// within. If the requested index + requested length is greater than the length
// of the string, the string contents from the index to the end of the string
// will be returned instead.
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
