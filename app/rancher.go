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
package app

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	jww "github.com/spf13/jwalterweatherman"
)

// supported distros
const (
	UnsupportedDistro Distro = iota
	CentOS
	Debian
	Ubuntu
)

// Distro is the distribution type
type Distro int

var distros = [...]string{
	"unsupported distro",
	"centos",
	"debian",
	"ubuntu",
}

func (d Distro) String() string { return distros[d] }

// DistroFromString returns the Distro constant for the passed string or
// unsupported.
//
// All incoming strings are normalized to lowercase.
func DistroFromString(s string) Distro {
	s = strings.ToLower(s)
	switch s {
	case "centos":
		return CentOS
	case "debian":
		return Debian
	case "ubuntu":
		return Ubuntu
	}
	return UnsupportedDistro
}

// indent: default indent to use for marshal stuff
var indent = "    "

// Defined builds
var Builds *builds

// Defaults for each supported distribution
var DistroDefaults distroDefaults

// distroDefaults contains the defaults for all supported distros and a flag
// whether its been set or not.
type distroDefaults struct {
	Templates map[Distro]*rawTemplate
	IsSet     bool
}

// GetTemplate returns a deep copy of the default template for the passed
// distro name. If the distro does not exist, an error is returned.
func (d *distroDefaults) GetTemplate(n string) (*rawTemplate, error) {
	var t *rawTemplate
	var ok bool
	t, ok = d.Templates[DistroFromString(n)]
	if !ok {
		err := fmt.Errorf("unsupported distro: %s", n)
		jww.ERROR.Println(err)
		return t, err
	}
	Copy := newRawTemplate()
	Copy.PackerInf = t.PackerInf
	Copy.IODirInf = t.IODirInf
	Copy.BuildInf = t.BuildInf
	Copy.releaseISO = t.releaseISO
	Copy.date = t.date
	Copy.delim = t.delim
	Copy.Distro = t.Distro
	Copy.Arch = t.Arch
	Copy.Image = t.Image
	Copy.Release = t.Release
	for k, v := range t.varVals {
		Copy.varVals[k] = v
	}
	Copy.build = t.build.DeepCopy()
	return Copy, nil
}

// Set sets the default templates for each distro.
func (d *distroDefaults) Set() error {
	dflts := &defaults{}
	err := dflts.Load()
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	s := &supported{}
	err = s.Load()
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	d.Templates = map[Distro]*rawTemplate{}
	// Generate the default settings for each distro.
	for k, v := range s.Distro {
		// See if the base url exists for non cento distros
		// It isn't required for debian because automatic resolution of iso
		// information is not supported.
		if v.BaseURL == "" && k != CentOS.String() {
			err = fmt.Errorf("%s requires a BaseURL, none provided", k)
			jww.CRITICAL.Println(err)
			return err

		}
		// Create the struct for the default settings
		tmp := newRawTemplate()
		// First assign it all the default settings.
		tmp.BuildInf = dflts.BuildInf
		tmp.IODirInf = dflts.IODirInf
		tmp.PackerInf = dflts.PackerInf
		tmp.build = dflts.build.DeepCopy()
		tmp.Distro = k
		// Now update it with the distro settings.
		tmp.BaseURL = appendSlash(v.BaseURL)
		tmp.Arch, tmp.Image, tmp.Release = getDefaultISOInfo(v.DefImage)
		tmp.setDefaults(v)
		d.Templates[DistroFromString(k)] = tmp
	}
	DistroDefaults.IsSet = true
	return nil
}

// Set DistroDefaults
func loadBuilds() error {
	Builds = &builds{}
	err := Builds.Load()
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	return nil
}

// BuildDistro creates a build based on the target distro's defaults. The
// ArgsFilter contains information on the target distro and any overrides that
// are to be applied to the build.
//
// Returns an error or nil if successful.
func BuildDistro(a ArgsFilter) error {
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	}
	err := buildPackerTemplateFromDistro(a)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// TODO: what does this argString processing do, or supposed to do? and document it this time!
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
	return nil

}

// Create Packer templates from specified build templates.
func buildPackerTemplateFromDistro(a ArgsFilter) error {
	var rTpl *rawTemplate
	if a.Distro == "" {
		err := fmt.Errorf("cannot build a packer template because no target distro information was passed")
		jww.ERROR.Println(err)
		return err
	}
	// Get the default for this distro, if one isn't found then it isn't Supported.
	rTpl, err := DistroDefaults.GetTemplate(a.Distro)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// If any overrides were passed, set them.
	if a.Arch != "" {
		rTpl.Arch = a.Arch
	}
	if a.Image != "" {
		rTpl.Image = a.Image
	}
	if a.Release != "" {
		rTpl.Release = a.Release
	}
	rTpl.BuildName = ":type-:release-:arch-:image-rancher"

	//	// make a copy of the .
	//	rTpl := newRawTemplate()
	//	rTpl.updateBuilders(d.Builders)

	// Since distro builds don't actually have a build name, we create one
	// out of the args used to create it.
	// TODO: given the above, should this be done? Or should the buildname for distro
	//       builds be merged later?
	rTpl.BuildName = fmt.Sprintf("%s-%s-%s-%s", rTpl.Distro, rTpl.Release, rTpl.Arch, rTpl.Image)
	pTpl := packerTemplate{}
	// Now that the raw template has been made, create a Packer template out of it
	pTpl, err = rTpl.createPackerTemplate()
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	// Create the JSON version of the Packer template. This also handles creation of
	// the build directory and copying all files that the Packer template needs to the
	// build directory.
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.dirs, rTpl.files)
	if err != nil {
		jww.ERROR.Println(err)
		return err
	}
	return nil
}

// BuildBuilds manages the process of creating Packer Build templates out of
// the passed build names. All builds are done concurrently.  Returns either a
// message providing information about the processing of the requested builds
// or an error.
func BuildBuilds(buildNames ...string) (string, error) {
	if buildNames[0] == "" {
		err := fmt.Errorf("Nothing to build. No build name was passed")
		jww.ERROR.Println(err)
		return "", err
	}
	// Only load supported if it hasn't been loaded. Even though LoadSupported
	// uses a mutex to control access to prevent race conditions, no need to
	// call it if its already loaded.
	if !DistroDefaults.IsSet {
		err := DistroDefaults.Set()
		if err != nil {
			jww.ERROR.Println(err)
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
		err := fmt.Errorf("unable to build Packer Template: no build name was passed")
		jww.ERROR.Println(err)
		doneCh <- err
		return
	}
	var rTpl, bTpl *rawTemplate
	var ok bool
	// Check the type and create the defaults for that type, if it doesn't already exist.
	rTpl = &rawTemplate{}
	bTpl = &rawTemplate{}
	bTpl, ok = Builds.Build[buildName]
	if !ok {
		err := fmt.Errorf("unable to build Packer Template: %q is not a valid build name", buildName)
		jww.ERROR.Println(err)
		doneCh <- err
		return
	}
	// See if the distro default exists.
	rTpl, ok = DistroDefaults.Templates[DistroFromString(bTpl.Distro)]
	if !ok {
		err := fmt.Errorf("unsupported distro: %s", bTpl.Distro)
		jww.ERROR.Println(err)
		doneCh <- err
		return
	}
	// Set build iso information overrides, if any.
	if bTpl.Arch != "" {
		rTpl.Arch = bTpl.Arch
	}
	if bTpl.Image != "" {
		rTpl.Image = bTpl.Image
	}
	if bTpl.Release != "" {
		rTpl.Release = bTpl.Release
	}
	bTpl.BuildName = buildName
	// create build template() then call create packertemplate
	rTpl.build = DistroDefaults.Templates[DistroFromString(bTpl.Distro)].build
	rTpl.updateBuildSettings(bTpl)
	pTpl := packerTemplate{}
	var err error
	pTpl, err = rTpl.createPackerTemplate()
	if err != nil {
		jww.ERROR.Println(err)
		doneCh <- err
		return
	}
	err = pTpl.create(rTpl.IODirInf, rTpl.BuildInf, rTpl.dirs, rTpl.files)
	if err != nil {
		jww.ERROR.Println(err)
		doneCh <- err
		return
	}
	doneCh <- nil
	return
}

// getSliceLenFromIface takes an interface that's assumed to be a slice and
// returns its length. If it is not a slice, an error is returned.
func getSliceLenFromIface(v interface{}) (int, error) {
	if v == nil {
		return 0, nil
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Slice:
		sl := reflect.ValueOf(v)
		return sl.Len(), nil
	}
	return 0, fmt.Errorf("err: getSliceLenFromIface expected a slice, go" + reflect.TypeOf(v).Kind().String())
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

// mergeSlices Takes two slices and returns the de-duped, merged list. The
// elements are returned in order of first encounter-duplicate keys are
// discarded.
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
			indx = indexOfKeyInVarSlice(key, merged)
			if indx < 0 {
				jww.WARN.Printf("%q, was not updated to %q because it was not found in the target", key, v)
			} else {
				merged[indx] = v
			}
			continue
		}
		// i is the index of the next element to add, a result of i being set to the count
		// of the items copied, which is 1 greater than the index, or the index of the next
		// item, should it exist. Instead, it is updated after adding the new value as, after
		// add, i points to the current element.
		merged[i] = v
		i++
	}
	// Shrink the slice back down to == its length
	ret := make([]string, i)
	copy(ret, merged)
	return ret
}

// varMapFromSlice creates a map from the passed slice. A Rancher var string
// contains a key=value string. Whitespace before and after the key and value
// are ignored, but whitespace within the key and value are preserved. The key
// is everything up to the first '='. As such, a value may contain any number of
// '=' tokends but the key may not contain any.
func varMapFromSlice(vars []string) map[string]string {
	if vars == nil {
		jww.WARN.Println("unable to create a Packer Settings map because no variables were received")
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

// parseVar: takes a string in the form of `key=value` and returns the
// key-value pair.
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
	return k, v
}

// indexOfKeyInVarSlice searches for the passed key in the slice and returns
// its index if found, or -1 if not found; 0 is a valid index on a slice. The
// string to search is in the form of 'key=value'.
func indexOfKeyInVarSlice(key string, sl []string) int {
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
	return fmt.Sprintf("{{user `%s` }}", s)
}

// getDefaultISOInfo accepts a slice of strings and returns Arch, Image, and
// Release info extracted from that slice.
func getDefaultISOInfo(d []string) (arch string, image string, release string) {
	for _, val := range d {
		k, v := parseVar(val)
		switch k {
		case "arch":
			arch = v
		case "image":
			image = v
		case "release":
			release = v
		default:
			jww.WARN.Printf("unknown default key: %s", k)
		}
	}
	return arch, image, release
}

// getMergedProvisioners merges the new config with the old. The updates follow
// these rules:
//   * The existing configuration is used when no `new` provisioners are
//     specified.
//   * When 1 or more `new` provisioners are specified, they will replace all
//     existing provisioners. In this situation, if a provisioners exists in
//     the `old` map but it does not exist in the `new` map, that provisioners
//     will be orphaned.
func getMergedProvisioners(old map[string]provisioner, new map[string]provisioner) map[string]provisioner {
	// If there is nothing new, old equals merged.
	if len(new) <= 0 || new == nil {
		return old
	}
	// Convert to an interface.
	var ifaceOld = make(map[string]interface{}, len(old))
	for i, o := range old {
		ifaceOld[i] = o
	}
	// Convert to an interface.
	var ifaceNew = make(map[string]interface{}, len(new))
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

// copyFile copies a file the source file to the destination
func copyFile(src string, dst string) (written int64, err error) {
	if src == "" {
		return 0, fmt.Errorf("copyfile error: source name was empty")
	}
	if dst == "" {
		return 0, fmt.Errorf("copyfile error: destination name was empty")
	}
	// get the destination directory
	dstDir := path.Dir(dst)
	if dstDir == "." {
		return 0, fmt.Errorf("copyfile error: destination name, %q, did not include a directory", dst)
	}
	// Create the scripts dir and copy each script from sript_src to out_dir/scripts/
	// while keeping track of success/failures.
	err = os.MkdirAll(dstDir, os.FileMode(0766))
	if err != nil {
		jww.ERROR.Println(err)
		return 0, err
	}
	var fs, fd *os.File
	// Open the source file
	fs, err = os.Open(src)
	if err != nil {
		jww.ERROR.Println(err)
		return 0, err
	}
	defer func() {
		cerr := fs.Close()
		if cerr != nil && err == nil {
			jww.WARN.Println(cerr)
			err = cerr
		}
	}()
	// Open the destination, create or truncate as needed.
	fd, err = os.Create(dst)
	if err != nil {
		jww.ERROR.Println(err)
		return 0, err
	}
	defer func() {
		cerr := fd.Close()
		if cerr != nil && err == nil {
			jww.WARN.Println(cerr)
			err = cerr
		}
	}()
	return io.Copy(fd, fs)
}

// copyDir takes 2 directory paths and copies the contents from src to dest get
// the contents of srcDir.
func copyDir(srcDir string, dstDir string) error {
	exists, err := pathExists(srcDir)
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	if !exists {
		err = fmt.Errorf("nothing copied: the source, %s, does not exist", srcDir)
		jww.ERROR.Println(err)
		return err
	}
	dir := Archive{}
	err = dir.DirWalk(srcDir)
	if err != nil {
		jww.ERROR.Print(err)
		return err
	}
	for _, file := range dir.Files {
		if file.info == nil {
			// if the info is empty, whatever this entry represents
			// doesn't actually exist.
			err := fmt.Errorf("%s does not exist", file.p)
			jww.ERROR.Println(err)
			return err
		}
		if file.info.IsDir() {
			err = os.MkdirAll(file.p, os.FileMode(0766))
			if err != nil {
				jww.ERROR.Println(err)
				return err
			}
			continue
		}
		_, err = copyFile(filepath.Join(srcDir, file.p), filepath.Join(dstDir, file.p))
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	}
	return nil
}

// deleteDir deletes the contents of a directory.
func deleteDir(dir string) error {
	var dirs []string
	// see if the directory exists first, actually any error results in the
	// same handling so just return on any error instead of doing an
	// os.IsNotExist(err)
	_, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			jww.ERROR.Println(err)
			return err
		}
	}
	dirInf := directory{}
	dirInf.DirWalk(dir)
	dir = appendSlash(dir)
	for _, file := range dirInf.Files {
		if file.info.IsDir() {
			dirs = append(dirs, dir+file.p)
			continue
		}
		err := os.Remove(dir + file.p)
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	}
	// all the files should now be deleted so its safe to delete the directories
	// do this in reverse order
	for i := len(dirs) - 1; i >= 0; i-- {
		err = os.Remove(dirs[i])
		if err != nil {
			jww.ERROR.Println(err)
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

// mergedKeysFromMaps takes a variadic array of maps and returns a merged slice
// of keys for those maps.
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

// setParentDir takes a directory name and and a path.
//   * If the path does not contain a parent directory, the passed directory
//     is prepended to the path and the new value is returned, otherwise the
//     path is returned; e.g. if
//   * "shell" and "script.sh" are passed as the dir and path, the returned
//     value will be "shell/script.sh", with the "/" normalized to the os
//     specific use. If "shell" and "scripts/script.sh" are passed as the dir
//     and path, the returned value will be "scripts/script.sh".
//   * An empty path will result in an empty string being returned.
//   * An empty directory will result in the path being returned.
func setParentDir(d, p string) string {
	if p == "" {
		return ""
	}
	if d == "" {
		return p
	}
	dir := path.Dir(p)
	if dir == "." {
		return filepath.Join(d, p)
	}
	return p
}

// getUniqueFilename takes the path of the file to be created along with a date
// layout and checks to see if it exists. If it doesn't exist, it is returned
// as the filename to use. Otherwise, it goes through the steps below until an
// "no such file or directory" error is returned. This is used for situations
// where there might be a filename collision and the existing file is to be
// preserved in some manner, e.g. archives or log files.
//
// If the filepath and name already exists, the current formatted date is
// appended to it using the received layout.  If the file doesn't exist, this
// filepath is returned.  If the layout is an empty string, this step is
// skipped.
//
// Otherwise, a sequence number is appended to the filename with the date and
// is checked for collision until no file is found. The first filename that
// results in a "no such file or directory" error is returned as the filename
// to use.
//
// Any non "no such file or directory" error is returned as an error.
func getUniqueFilename(p, layout string) (string, error) {
	_, err := os.Stat(p)
	if err != nil {
		if err.(*os.PathError).Err.Error() == "no such file or directory" {
			return p, nil
		}
		return "", err
	}

	dir, file := path.Split(p)
	parts := strings.Split(file, ".")
	var newPath, basePath string
	// If the path had multiple .'s append everything except last two elements
	for i := 0; i < len(parts)-1; i++ {
		newPath += parts[i]
		if i < len(parts)-2 {
			newPath += "."
		}
	}
	newPath = filepath.Join(dir, newPath)
	// cache the path fragment in case we need to use a sequence
	basePath = newPath
	if layout != "" {
		now := time.Now().Format(layout)
		newPath += "-" + now
		// update basePath
		basePath = newPath
		newPath += "." + parts[len(parts)-1]
		fmt.Println(newPath)
		_, err = os.Stat(newPath)
		if err != nil {
			if err.(*os.PathError).Err.Error() == "no such file or directory" {
				return newPath, nil
			}
			return "", err
		}
	}
	// check for a unique name while appending a sequence.
	i := 1
	for {
		newPath = basePath + "-" + strconv.Itoa(i) + "." + parts[len(parts)-1]
		fmt.Println(newPath)
		_, err = os.Stat(newPath)
		if err != nil {
			if err.(*os.PathError).Err.Error() == "no such file or directory" {
				return newPath, nil
			}
			return "", err
		}
		i++
	}
}
