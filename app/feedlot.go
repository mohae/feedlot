// Generate Packer templates and associated files for consumption by Packer.
//
// Copyright 2014 Joel Scoble. All Rights Reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//

package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/mohae/contour"
	"github.com/mohae/feedlot/conf"
	"github.com/mohae/feedlot/log"
)

// ErrEmptyParam occurs when the received parameter was empy.
var ErrEmptyParam = errors.New("received an empty parameter, expected a value")

// Error is the generic error wrapper used by this package.
type Error struct {
	slug string
	err  error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.slug, e.err)
}

type UnsupportedDistroErr struct {
	s string
}

func (e UnsupportedDistroErr) Error() string {
	return e.s + ": unsupported distro"
}

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

var casedDistros = [...]string{
	"Unsupported Distro",
	"CentOS",
	"Debian",
	"Ubuntu",
}

// CasedString returns the distro's name with proper casing.
func (d Distro) CasedString() string { return casedDistros[d] }

// ParseDistro returns the Distro constant for the passed string or
// unsupported.
//
// All incoming strings are normalized to lowercase.
func ParseDistro(s string) Distro {
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
var Builds map[string]builds

// Defaults for each supported distribution
var DistroDefaults distroDefaults

func init() {
	Builds = map[string]builds{}
}

// distroDefaults contains the defaults for all supported distros and a flag
// whether its been set or not.
type distroDefaults struct {
	Templates map[Distro]rawTemplate
	IsSet     bool
}

// GetTemplate returns a deep copy of the default template for the passed
// distro name. If the distro does not exist, an error is returned.
func (d *distroDefaults) GetTemplate(n string) (*rawTemplate, error) {
	var t rawTemplate
	var ok bool
	t, ok = d.Templates[ParseDistro(n)]
	if !ok {
		err := fmt.Errorf("unsupported distro: %s", n)
		log.Error(err)
		return nil, err
	}
	return t.Copy(), nil
}

// Set sets the default templates for each distro.
func (d *distroDefaults) Set() error {
	dflts := &defaults{}
	err := dflts.Load("")
	if err != nil {
		err = Error{slug: "set distro defaults", err: err}
		log.Error(err)
		return err
	}
	// get the source settings from the defaults; If the source_dir setting isn't set
	// return an error
	if len(dflts.IODirInf.SourceDir) == 0 {
		err = Error{slug: "set distro defaults", err: RequiredSettingErr{"source_dir"}}
		log.Error(err)
		return err
	}

	s := &supported{}
	err = s.Load("")
	if err != nil {
		err = Error{slug: "set distro defaults", err: err}
		log.Error(err)
		return err
	}
	d.Templates = map[Distro]rawTemplate{}
	// Generate the default settings for each distro.
	for k, v := range s.Distro {
		// See if the base url exists for non centos distros
		// It isn't required for debian because automatic resolution of iso
		// information is not supported.
		if v.BaseURL == "" && k != CentOS.String() {
			err = Error{slug: fmt.Sprintf("set distro defaults: %s", k), err: RequiredSettingErr{"base_url"}}
			log.Error(err)
			return err

		}
		// Create the struct for the default settings
		tmp := newRawTemplate()
		// First assign it all the default settings.

		/*
			tmp.BuildInf = deepcopy.Iface(dflts.BuildInf).(BuildInf)
			tmp.IODirInf = deepcopy.Iface(dflts.IODirInf).(IODirInf)
			tmp.PackerInf = deepcopy.Iface(dflts.PackerInf).(PackerInf)
			tmp.build = dflts.build.copy()
			tmp.Distro = strings.ToLower(k)
		*/
		// Now update it with the distro settings.
		tmp.BaseURL = appendSlash(v.BaseURL)
		tmp.Arch, tmp.Image, tmp.Release = getDefaultISOInfo(v.DefImage)
		err = tmp.setDefaults(v)
		if err != nil {
			err = Error{slug: fmt.Sprintf("set distro defaults: %s", k), err: err}
			log.Error(err)
			return err
		}
		d.Templates[ParseDistro(k)] = *tmp
	}
	DistroDefaults.IsSet = true
	return nil
}

// loadBuilds accepts a list of builds and loads the build information for
// them. Since we don't know everything that is going to be used, we load all
// build configuration files. A Feedlot configuration directory can have 0 or
// more build configuration files and any number of subdirectories.
//
// A build configuration file is any file that ends in ".fmt" and isn't name
// build_list.fmt", "defualt.fmt", "feedlot.fmt", or "supported.fmt".
//
// Subdirectories are called environments, envs, and are a way to namespace
// builds. An envs' name is the same as the subdirectories name. Env names can
// be concatonated together, using the env_separator_char as the separator; '-'
// is the default value.
//
// The sourceDir and sourceDirIsRelative settings from the defaults file is
// passed so that each build template's Packer source directory can be set
// if the template doesn't define its own.
// TODO: add env support
func loadBuilds() error {
	log.Debug("loading builds")
	// index all the files in the configuration directory, including subdir
	// this should be sorted
	cDir := contour.GetString(conf.Dir)
	if contour.GetBool(conf.Example) {
		cDir = filepath.Join(contour.GetString(conf.ExampleDir), cDir)
	}
	// names come from os.FileInfo.Name() results
	// TODO: add handling of dir names and recursive for envs support
	_, fnames, err := indexDir(cDir)
	if err != nil {
		return Error{slug: "load builds", err: err}
	}
	// for each file
	for _, fname := range fnames {
		// get the file name, without the extension
		ext := filepath.Ext(fname)
		file := strings.TrimSuffix(fname, ext)
		// skip non-build files.
		switch file {
		case "build_list":
			continue
		case "default":
			continue
		case "feedlot":
			continue
		case "supported":
			continue
		}
		fname = filepath.Join(cDir, fname)
		log.Debugf("loading build file %s", fname)
		b := builds{}
		err := b.Load(fname)
		if err != nil {
			return Error{slug: "load builds", err: err}
		}
		Builds[fname] = b
	}
	log.Debug("builds loaded")
	return nil
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
	return 0, fmt.Errorf("cannot determine len: expected a slice, got %q", reflect.TypeOf(v).Kind().String())
}

// MergeSlices takes a variadic input of []string and returns a string slice
// with all of the values within the slices merged.
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
// elements are returned in order of first encounter: duplicate keys are
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
	tempSl := make([]string, len(s1))
	copy(tempSl, s1)
	// Go through every element in the second slice.
	for _, v := range s2 {
		var found bool
		// See if the key already exists
		for _, tmp := range tempSl {
			if v == tmp {
				// it already exists
				found = true
				break
			}
		}
		if !found {
			tempSl = append(tempSl, v)
		}
	}
	return tempSl
}

// mergeSettingsSlices merges two slices of settings. In cases of a key
// collision, the second slice, s2, takes precedence. There are no duplicates
// at the end of this operation.
//
// Since settings use  embedded key=value pairs, the key is extracted from each
// value and matches are performed on the key only as the value will be
// different if the key appears in both slices.
func mergeSettingsSlices(s1 []string, s2 []string) ([]string, error) {
	if len(s1) == 0 && len(s2) == 0 {
		return nil, nil
	}
	// Make a slice with a length equal to the sum of the two input slices.
	merged := make([]string, len(s1)+len(s2))
	// Copy the first slice.
	i := copy(merged, s1)
	// if nothing was copied, i == 0 , just copy the 2nd slice.
	if i == 0 {
		copy(merged, s2)
		return merged, nil
	}
	ms1 := map[string]string{}
	// Create a map of variables from the first slice for comparison reasons.
	ms1 = varMapFromSlice(s1)
	if ms1 == nil {
		return nil, errors.New("unable to create merge comparison maps")
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
			if indx >= 0 {
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
	return ret, nil
}

// varMapFromSlice creates a map from the passed slice. A Rancher var string
// contains a key=value string. Whitespace before and after the key and value
// are ignored, but whitespace within the key and value are preserved. The key
// is everything up to the first '='. As such, a value may contain any number of
// '=' tokends but the key may not contain any.
func varMapFromSlice(vars []string) map[string]string {
	if vars == nil {
		log.Info("unable to create a Packer Settings map because no variables were received")
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
			log.Errorf("unknown default key: %s", k)
		}
	}
	return arch, image, release
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

// copyFile copies a file the source file to the destination
func copyFile(src string, dst string) (written int64, err error) {
	if src == "" {
		return 0, errors.New("copy file: source was empty")
	}
	if dst == "" {
		return 0, errors.New("copy file: destination was empty")
	}
	// get the destination directory
	dstDir := filepath.Dir(dst)
	if dstDir == "." {
		return 0, fmt.Errorf("copy file: destination, %q, did not include a directory", dst)
	}
	// Create the scripts dir and copy each script from sript_src to out_dir/scripts/
	// while keeping track of success/failures.
	err = os.MkdirAll(dstDir, os.FileMode(0766))
	if err != nil {
		return 0, Error{slug: "copy file", err: err}
	}
	var fs, fd *os.File
	// Open the source file
	fs, err = os.Open(src)
	if err != nil {
		return 0, Error{slug: "copy file", err: err}
	}
	defer func() {
		cerr := fs.Close()
		if cerr != nil && err == nil {
			err = Error{slug: "copy file", err: cerr}
		}
	}()
	// Open the destination, create or truncate as needed.
	fd, err = os.Create(dst)
	if err != nil {
		return 0, Error{slug: "copy file", err: err}
	}
	defer func() {
		cerr := fd.Close()
		if cerr != nil && err == nil {
			err = Error{slug: "copy file", err: cerr}
		}
	}()
	return io.Copy(fd, fs)
}

// copyDir takes 2 directory paths and copies the contents from src to dest.
func copyDir(srcDir string, dstDir string) error {
	exists, err := pathExists(srcDir)
	if err != nil {
		return Error{slug: "copy dir", err: err}
	}
	if !exists {
		return fmt.Errorf("copy dir: %s does not exist", srcDir)
	}
	dir := Archive{}
	err = dir.DirWalk(srcDir)
	if err != nil {
		return Error{slug: "copy dir", err: err}
	}
	for _, file := range dir.Files {
		if file.info == nil {
			return fmt.Errorf("copy dir: %s does not exist", file.p)
		}
		// skip non-regular files
		if !file.info.Mode().IsRegular() {
			continue
		}
		_, err = copyFile(filepath.Join(srcDir, file.p), filepath.Join(dstDir, file.p))
		if err != nil {
			return Error{slug: "copy dir", err: err}

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

// mergeKeysFromComponentMaps takes a variadic array of packer component maps
// and returns a merged, de-duped slice of keys for those maps.
func mergeKeysFromMaps(m ...map[string]interface{}) []string {
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
		return path.Join(d, p)
	}
	return p
}

// getUniqueFilename takes the path of the file to be created along with a date
// layout and checks to see if it exists. If it doesn't exist, it is returned
// as the filename to use. Otherwise, it goes through the steps below until an
// os.IsNotExist() error is returned. This is used for situations
// where there might be a filename collision and the existing file is to be
// preserved in some manner, e.g. archives or log files.
//
// If the filepath and name already exists, the current formatted date is
// appended to it using the received layout. The name is then appended with a
// sequence number, starting at 1, and checked for existence until no file is
// found. The first filename that results in an os.IsNotExist(err) is
// returned as the filename to use.
//
// Any non os.IsNotExist() error is returned as an error.
//
// There is a special check made for tar.gz, as this is the default extension
// for the compressed archives of templates; otherwise, it is assumed that the
// extension is the text after the last "." in the path.
func getUniqueFilename(p, layout string) (string, error) {
	// see if file exists; if it doesn't we're done.
	_, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return "", Error{slug: "get unique filename", err: err}
	}
	var base, ext string
	dir := filepath.Dir(p)
	if strings.HasSuffix(p, ".tar.gz") {
		ext = ".tar.gz"
	} else {
		ext = path.Ext(p)
	}
	base = path.Base(strings.TrimSuffix(p, ext))
	// cache the path fragment in case we need to use a sequence
	if layout != "" {
		now := time.Now().Format(layout)
		base = fmt.Sprintf("%s.%s", base, now)
	}
	// check for a unique name while appending a sequence.
	i := 1
	for {
		newPath := path.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		_, err = os.Stat(newPath)
		if err != nil {
			if os.IsNotExist(err) {
				return filepath.ToSlash(newPath), nil
			}
			return "", Error{slug: "get unique filename", err: err}
		}
		i++
	}
}

// indexDir indexes the passed directory and returns its contents as two lists:
// directory names and file names. Any error encountered results in termination
// of indexing and returns.
func indexDir(s string) (dirs, files []string, err error) {
	// nothing to index
	if s == "" {
		return nil, nil, ErrEmptyParam
	}
	fi, err := os.Stat(s)
	if err != nil {
		return nil, nil, err
	}
	if !fi.IsDir() {
		return nil, nil, fmt.Errorf("index dir: %s is not a directory", s)
	}
	f, err := os.Open(s)
	if err != nil {
		return nil, nil, Error{slug: "index dir", err: err}
	}
	defer f.Close()
	fis, err := f.Readdir(-1)
	if err != nil {
		return nil, nil, Error{slug: "index dir", err: err}
	}
	for _, fi := range fis {
		if fi.IsDir() {
			dirs = append(dirs, fi.Name())
			continue
		}
		files = append(files, fi.Name())
	}
	return dirs, files, nil
}
