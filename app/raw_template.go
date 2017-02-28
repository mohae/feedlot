package app

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mohae/contour"
	"github.com/mohae/deepcopy"
	jww "github.com/spf13/jwalterweatherman"
)

type InvalidComponentErr struct {
	id   string // component id
	cTyp string // component type
	s    string
}

func (e InvalidComponentErr) Error() string {
	if e.id == "" {
		return fmt.Sprintf("%s: %q: invalid type", e.cTyp, e.s)
	}
	return fmt.Sprintf("id: %s: %q: invalid type", e.id, e.cTyp, e.s)
}

// SettingErr occurs when there is a problem with a packer component
// setting.
type SettingErr struct {
	Key   string
	Value string
	err   error
}

func (e SettingErr) Error() string {
	var s string
	if e.Key == "" {
		s = "\"\": "
	} else {
		s = e.Key + ": "
	}
	if e.Value == "" {
		s += "\"\": "
	} else {
		s += e.Value + ": "
	}
	return s + e.err.Error()
}

// RequiredSettingErr occurs when a setting required by the Packer
// component being processed isn't set.
type RequiredSettingErr struct {
	Key string
}

func (e RequiredSettingErr) Error() string {
	return e.Key + ": required setting not found"
}

var (
	// ErrConfigNotFound occurs when the specified configuration resource
	// cannot be found.  This may result in Feedlot using its defaults.
	ErrConfigNotFound = errors.New("configuration not found")
	// ErrNoCommands occurs when a referenced command file doesn't have any
	// contents.
	ErrNoCommands = errors.New("no commands found")
)

func NewErrConfigNotFound(s string) error {
	return Error{slug: s, err: ErrConfigNotFound}
}

// rawTemplate holds all the information for a Feedlot template. This is used
// to generate the Packer Build.
type rawTemplate struct {
	PackerInf
	IODirInf
	BuildInf
	// Example settings
	IsExample  bool
	ExampleDir string
	// holds release information
	releaseISO releaser
	// the builder specific string for the template's OS and Arch
	osType string
	// Current date in ISO 8601
	date string
	// The character(s) used to identify variables for Feedlot. By default
	// this is a colon, :. Currently only a starting delimeter is supported.
	delim string
	// The distro that this template targets. The type must be a supported
	// type, i.e. defined in supported.toml. The values for type are
	// consistent with Packer values.
	Distro string
	// The architecture for the ISO image, this is either 32bit or 64bit,
	// with the actual values being dependent on the operating system and
	// the target builder.
	Arch string
	// The image for the ISO. This is distro dependent.
	Image string
	// The release, or version, for the ISO. Usage and values are distro
	// dependent, however only version currently supported images that are
	// available on the distro's download site are supported.
	Release string
	// varVals is a variable replacement map used in finalizing the value of strings for
	// which variable replacement is supported.
	varVals map[string]string
	// Contains all the build information needed to create the target Packer
	// template and its associated artifacts.
	build
	// files maps destination files to their sources. These are the actual file locations
	// after they have been resolved. The destination file is the key, the source file
	// is the value
	files map[string]string
	// dirs maps destination directories to their source directories. Everything within
	// the directory will be copied. The same resolution rules apply for dirs as for
	// files. The destination directory is the key, the source directory is the value
	dirs map[string]string
}

// mewRawTemplate returns a rawTemplate with current date in ISO 8601 format.
// This should be called when a rawTemplate with the current date is desired.
func newRawTemplate() *rawTemplate {
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	return &rawTemplate{date: splitDate[0], delim: contour.GetString(ParamDelimStart), files: make(map[string]string), dirs: make(map[string]string)}
}

// copy makes a copy of the template and returns the new copy.
func (r *rawTemplate) copy() *rawTemplate {
	Copy := newRawTemplate()
	Copy.PackerInf = deepcopy.Iface(r.PackerInf).(PackerInf)
	Copy.IODirInf = deepcopy.Iface(r.IODirInf).(IODirInf)
	Copy.BuildInf = deepcopy.Iface(r.BuildInf).(BuildInf)
	Copy.releaseISO = r.releaseISO
	Copy.date = r.date
	Copy.delim = r.delim
	Copy.Distro = r.Distro
	Copy.Arch = r.Arch
	Copy.Image = r.Image
	Copy.Release = r.Release
	for k, v := range r.varVals {
		Copy.varVals[k] = v
	}
	Copy.build = r.build.copy()
	return Copy
}

// r.createPackerTemplate creates a Packer template from the rawTemplate.
// TODO:
//		Write to output
//		Copy resources to output
func (r *rawTemplate) createPackerTemplate() (packerTemplate, error) {
	var err error
	// Resolve the Feedlot variables to their final values.
	r.mergeVariables()
	// General Packer Stuff
	p := packerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description
	// Builders
	p.Builders, err = r.createBuilders()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Post-Processors
	p.PostProcessors, err = r.createPostProcessors()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Provisioners
	p.Provisioners, err = r.createProvisioners()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Return the generated Packer Template.
	return p, nil
}

// replaceVariables checks incoming string for variables and replaces them with
// their values.
func (r *rawTemplate) replaceVariables(s string) string {
	//see if the delim is in the string, if not, nothing to replace
	if strings.Index(s, r.delim) < 0 {
		return s
	}
	// Go through each variable and replace as applicable.
	for vName, vVal := range r.varVals {
		s = strings.Replace(s, vName, vVal, -1)
	}
	return s
}

// r.setDefaults takes the incoming distro settings and merges them with its
// existing settings, which are set to feedlot's defaults, to create the
// default template.
func (r *rawTemplate) setDefaults(d *distro) error {
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(d.IODirInf)
	r.PackerInf.update(d.PackerInf)
	r.BuildInf.update(d.BuildInf)
	// check country, region, sponsor stuff: if nil set to empty string
	var s string
	if r.Country == nil {
		r.Country = &s
	}
	if r.Region == nil {
		r.Region = &s
	}
	if r.Sponsor == nil {
		r.Sponsor = &s
	}
	// update def image stuff
	for _, v := range d.DefImage {
		k, vv := parseVar(v)
		switch k {
		case "arch":
			r.Arch = vv
		case "image":
			r.Image = vv
		case "release":
			r.Release = vv
		}
	}
	// If defined, BuilderTypes override any prior BuilderTypes Settings
	if d.BuilderIDs != nil {
		r.BuilderIDs = d.BuilderIDs
	}
	// If defined, PostProcessorTypes override any prior PostProcessorTypes Settings
	if d.PostProcessorIDs != nil {
		r.PostProcessorIDs = d.PostProcessorIDs
	}
	// If defined, ProvisionerTypes override any prior ProvisionerTypes Settings
	if d.ProvisionerIDs != nil {
		r.ProvisionerIDs = d.ProvisionerIDs
	}
	// merge the build portions.
	err := r.updateBuilders(d.Builders)
	if err != nil {
		return err
	}
	err = r.updatePostProcessors(d.PostProcessors)
	if err != nil {
		return err
	}
	err = r.updateProvisioners(d.Provisioners)
	if err != nil {
		return err
	}
	return nil
}

// r.updateBuildSettings merges Settings between an old and new template.
// Note:  Arch, Image, and Release are not updated here as how these fields are
// updated depends on whether this is a build from a distribution's default
// template or from a defined build template.
func (r *rawTemplate) updateBuildSettings(bld *rawTemplate) error {
	r.IODirInf.update(bld.IODirInf)
	r.updateSourceDirSetting()
	err := r.updateTemplateOutputDirSetting()
	if err != nil {
		return err
	}
	r.PackerInf.update(bld.PackerInf)
	r.BuildInf.update(bld.BuildInf)
	if bld.Arch != "" {
		r.Arch = bld.Arch
	}
	if bld.Image != "" {
		r.Image = bld.Image
	}
	if bld.Release != "" {
		r.Release = bld.Release
	}
	// If defined, Builders override any prior builder Settings.
	if bld.BuilderIDs != nil && len(bld.BuilderIDs) > 0 {
		r.BuilderIDs = bld.BuilderIDs
	}
	//   if nil don't do anything (this means prior settings are used, e.g. default)
	// For post_processor_ids and provisioner_ids, the following logic is used:
	//   if len == 0 unset. A len of 0 means that the build template purposely unsets
	//     any build
	//   if len > 0 replace the existing types with the builder's.
	if bld.PostProcessorIDs != nil {
		r.PostProcessorIDs = bld.PostProcessorIDs
	}
	if bld.ProvisionerIDs != nil {
		r.ProvisionerIDs = bld.ProvisionerIDs
	}
	// merge the build portions.
	r.updateBuilders(bld.Builders)
	r.updatePostProcessors(bld.PostProcessors)
	r.updateProvisioners(bld.Provisioners)
	return nil
}

// updateTemplateOutputDirSetting updates the template_output_dir setting
// if the template_output_dir_setting_is_relative flag is true.  Any Feedlot
// variables in the source_dir setting are not resolved.
func (r *rawTemplate) updateTemplateOutputDirSetting() error {
	if *r.IODirInf.TemplateOutputDirIsRelative {
		dir, err := os.Getwd()
		if err != nil {
			return Error{"template output dir error: could not get working directory", err}
		}
		r.IODirInf.TemplateOutputDir = filepath.Join(dir, r.IODirInf.TemplateOutputDir)
	}
	return nil
}

// updateSourceDirSetting updates the source_dir if the source_dir_is_relative
// flag is true.  Any Feedlot variables in the source_dir setting are not
// resolved.
func (r *rawTemplate) updateSourceDirSetting() {
	if *r.IODirInf.SourceDirIsRelative {
		r.IODirInf.SourceDir = filepath.Join(contour.GetString(ConfDir), r.IODirInf.SourceDir)
	}
}

// mergeVariables goes through the template variables and finalizes the
// values of any :vars found within the strings.
//
// Supported:
//  distro               the name of the distro
//  release              the release version being used
//  arch                 the target architecture for the build
//  image                the image used, e.g. server
//  date                 the current datetime, time.Now()
//  build_name           the name of the build template
//  template_output_dir  the directory to write the template build output to
//  packer_output_dir    the directory to write the Packer build artifact to
//  source_dir           the directory of any source files used in the build*
//
// Note: source_dir must be set. Feedlot searches for referenced files and
// uses source_dir/distro as the last search directory. This directory is
// also used as the base directory for any specified src directories.
func (r *rawTemplate) mergeVariables() {
	// Get the delim and set the replacement map, resolve name information
	r.setBaseVarVals()
	// get final value for name first
	r.Name = r.replaceVariables(r.Name)
	r.varVals[r.delim+"name"] = r.Name
	// then merge the sourc and out dirs and set them
	r.SourceDir = r.replaceVariables(r.SourceDir)
	r.TemplateOutputDir = r.replaceVariables(r.TemplateOutputDir)
	r.PackerOutputDir = r.replaceVariables(r.PackerOutputDir)
	r.varVals[r.delim+"template_output_dir"] = r.TemplateOutputDir
	r.varVals[r.delim+"packer_output_dir"] = r.PackerOutputDir
	r.varVals[r.delim+"source_dir"] = r.SourceDir
}

// setBaseVarVals sets the varVals for the base variables
func (r *rawTemplate) setBaseVarVals() {
	r.varVals = map[string]string{
		r.delim + "distro":     r.Distro,
		r.delim + "release":    r.Release,
		r.delim + "arch":       r.Arch,
		r.delim + "image":      r.Image,
		r.delim + "date":       r.date,
		r.delim + "build_name": r.BuildName,
	}
}

// mergeString does a variable replacement on the passed string and returns
// the finalized value. If the passed string is empty, the default value, d, is
// returned
func (r *rawTemplate) mergeString(s, d string) string {
	if s == "" {
		return d
	}
	return r.replaceVariables(s)
}

// ISOInfo sets the ISO info for the template's supported distro type. This
// also sets the builder specific string, when applicable.
// TODO: these should use new functions in release.go. instead of creating the
// structs here
func (r *rawTemplate) ISOInfo(builderType Builder, settings []string) error {
	var k, v, checksumType string
	var err error
	// Only the iso_checksum_type is needed for this.
	for _, s := range settings {
		k, v = parseVar(s)
		switch k {
		case "iso_checksum_type":
			checksumType = v
		}
	}
	switch r.Distro {
	case CentOS.String():
		r.releaseISO = &centos{
			release: release{
				iso: iso{
					BaseURL:      r.BaseURL,
					ChecksumType: checksumType,
				},
				Arch:    r.Arch,
				Distro:  r.Distro,
				Image:   r.Image,
				Release: r.Release,
			},
			region:  *r.Region,
			country: *r.Country,
			sponsor: *r.Sponsor,
		}
		err = r.releaseISO.setVersionInfo()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
		err = r.releaseISO.SetISOInfo()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
		r.osType, err = r.releaseISO.(*centos).getOSType(builderType)
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	case Debian.String():
		r.releaseISO = &debian{
			release: release{
				iso: iso{
					BaseURL:      r.BaseURL,
					ChecksumType: checksumType,
				},
				Arch:    r.Arch,
				Distro:  r.Distro,
				Image:   r.Image,
				Release: r.Release,
			},
		}
		err = r.releaseISO.setVersionInfo()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
		err = r.releaseISO.SetISOInfo()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
		r.osType, err = r.releaseISO.(*debian).getOSType(builderType)
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	case Ubuntu.String():
		r.releaseISO = &ubuntu{
			release: release{
				iso: iso{
					BaseURL:      r.BaseURL,
					ChecksumType: checksumType,
				},
				Arch:    r.Arch,
				Distro:  r.Distro,
				Image:   r.Image,
				Release: r.Release,
			},
		}
		err = r.releaseISO.setVersionInfo()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
		err = r.releaseISO.SetISOInfo()
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
		r.osType, err = r.releaseISO.(*ubuntu).getOSType(builderType)
		if err != nil {
			jww.ERROR.Println(err)
			return err
		}
	default:
		err := fmt.Errorf("unable to set ISO related information for the unsupported distro: %q", r.Distro)
		jww.ERROR.Println(err)
		return err
	}
	return nil
}

// commandsFromFile returns the commands within the requested file, if it can
// be found. No validation of the contents is done.
func (r *rawTemplate) commandsFromFile(name, component string) (commands []string, err error) {
	// find the file
	src, err := r.findCommandFile(name, component)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	// always close what's been opened and check returned error
	defer func() {
		cerr := f.Close()
		if cerr != nil && err == nil {
			err = cerr
		}
	}()
	//New Reader for the string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return nil, err
	}
	return commands, nil
}

// findCommandFile locates the requested command file. If a match cannot be
// found, an os.ErrNotExist is returned. Any other errors will result in a
// termination of the search.
//
// The request string is build with the following order:
//    commands/{name}
//    {name}
//
// findSource is called to handle the actual location of the file. If
// no match is found an os.ErrNotExist will be returned.
func (r *rawTemplate) findCommandFile(name, component string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("the passed command filename was empty")
	}
	findPath := filepath.Join("commands", name)
	src, err := r.findSource(findPath, component, false)
	// return the error for any error other than ErrNotExist
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	// if err is nil, the source was found
	if err == nil {
		return src, nil
	}
	return r.findSource(name, component, false)
}

// findSource searches for the specified sub-path using Feedlot's algorithm
// for finding the correct location.  Passed names may include relative path
// information and may be either a filename or a directory.  Releases may
// have "."'s in them.  In addition to searching for the requested source
// within the point release, the "." are stripped out and the resulting value
// is searched: e.g. 14.04 becomes 1404,or numericRelease.  The base release
// number is also checked: e.g. 14, the releaseBase, is searched for 14.04.
//
// Search order:
// distro+release+arch+build_name+component
//  source_dir/distro/release/arch/build_name/component/
//  source_dir/distro/release/arch/build_name/
//  source_dir/distro/numericRelease/arch/build_name/component/
//  source_dir/distro/numericRelease/arch/build_name/
//  source_dir/distro/releaseBase/arch/build_name/component/
//  source_dir/distro/releaseBase/arch/build_name/
// distro+release
//  source_dir/distro/release/build_name/component/
//  source_dir/distro/release/build_name/
//  source_dir/distro/numericRelease/build_name/component/
//  source_dir/distro/numericRelease/build_name/
//  source_dir/distro/releaseBase/build_name/component/
//  source_dir/distro/releaseBase/build_name/
// distro+arch
//  source_dir/distro/arch/build_name/component/
//  source_dir/distro/arch/build_name/
// distro
//  source_dir/distro/build_name/component/
//  source_dir/distro/build_name/
// root
//  source_dir/build_name/component/
//  source_dir/build_name/
//
// without build_name
// distro+release+arch
//  source_dir/distro/release/arch/component/
//  source_dir/distro/release/arch/
//  source_dir/distro/numericRelease/arch/component/
//  source_dir/distro/numericRelease/arch/
//  source_dir/distro/releaseBase/arch/component/
//  source_dir/distro/releaseBase/arch/
// distro+release
//  source_dir/distro/release/component/
//  source_dir/distro/release/
//  source_dir/distro/numericRelease/component/
//  source_dir/distro/numericRelease/
//  source_dir/distro/releaseBase/component/
//  source_dir/distro/releaseBase/
// distro+arch
//  source_dir/distro/arch/component/
//  source_dir/distro/arch/
// distro
//  source_dir/distro/component/
//  source_dir/distro/
// root
//  source_dir/component/
//  source_dir/
//
// If the component has a - in it, e.g. salt-masterless, component checks
// will be done on both the value and its base, i.e. salt.
//
// If the passed path is not found, an os.ErrNotExist will be returned
//
// p is the path slug to find, this is the value from a setting.
// component is the name of the component from which p was a setting.
//
// TODO: is isDir necessary?  For now, it is a legacy setting from the
// original code.
func (r *rawTemplate) findSource(p, component string, isDir bool) (string, error) {
	if p == "" {
		return "", errors.New("cannot find source: no path received")
	}
	// build a slice of release values to append to search paths.  An empty
	// string is the first element because the first path to search is
	// source_dir/distro/build_name.
	empty := []string{""}
	rInf := []string{r.Release}
	rParts := strings.Split(r.Release, ".")
	var numRelease string
	if len(rParts) > 1 {
		for _, v := range rParts {
			numRelease += v
		}
		// numeric release
		rInf = append(rInf, numRelease)
		// base
		rInf = append(rInf, rParts[0])
	}
	// check source_dir/distro/release/arch/build_name
	paths := r.buildSearchPaths(r.Distro, filepath.Join(r.Arch, r.BuildName), rInf)
	path, err := r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/distro/release/build_name
	paths = r.buildSearchPaths(r.Distro, r.BuildName, rInf)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/distro/arch/build_name
	paths = r.buildSearchPaths(r.Distro, filepath.Join(r.Arch, r.BuildName), empty)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/distro/build_name
	paths = r.buildSearchPaths(r.Distro, r.BuildName, empty)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/build_name
	paths = r.buildSearchPaths("", r.BuildName, empty)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}

	// try searches w/o build names
	// check source_dir/distro/release/arch
	paths = r.buildSearchPaths(r.Distro, r.Arch, rInf)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/distro/release/
	paths = r.buildSearchPaths(r.Distro, "", rInf)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/distro/arch
	paths = r.buildSearchPaths(r.Distro, r.Arch, empty)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir/distro
	paths = r.buildSearchPaths(r.Distro, "", empty)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}
	// check source_dir
	paths = r.buildSearchPaths("", "", empty)
	path, err = r.checkSourcePaths(p, component, paths)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	if path != "" {
		return path, nil
	}

	jww.TRACE.Printf("findSource:  %s not found", p)
	// not found, return an error
	return "", &os.PathError{"find", filepath.ToSlash(p), os.ErrNotExist}
}

// buildSearchPaths builds a slice of paths to search based on what it
// receives.
// for each release element:  path = source_dir + root + release + base
func (r *rawTemplate) buildSearchPaths(root, base string, release []string) []string {
	var paths []string
	for _, v := range release {
		paths = append(paths, filepath.Join(r.SourceDir, root, v, base))
	}
	return paths
}

// checkSourcePaths checks to see if the requested source exists in the
// received paths.
//
// First the path is checked with the component appended.  If a match isn't
// found there the path is checked as is.  If the component is "", only the
// path is checked.
//
// If a match is found, the path will be returned.  If a non os.ErrNotExist
// error occurs, that error will be returned; otherwise os.ErrNotExist will
// be returned
func (r *rawTemplate) checkSourcePaths(p, component string, paths []string) (string, error) {
	searchC := []string{component}
	// if the component has a - in the name, check the base of the component
	// e.g. chef-solo's base is chef
	cParts := strings.Split(component, "-")
	if len(cParts) > 1 {
		searchC = append(searchC, cParts[0])
	}

	for _, path := range paths {
		for _, c := range searchC {
			tmp := filepath.Join(path, c, p)
			inf, err := os.Stat(tmp)
			if err == nil {
				// if it's a dir, append the slash
				if inf.IsDir() {
					tmp += string(os.PathSeparator)
				}
				jww.TRACE.Printf("findSource:  %s found", tmp)
				return filepath.ToSlash(tmp), nil
			}
			if !os.IsNotExist(err) {
				return "", err
			}
		}
		// if the component was empty, no need to search w/o it as the
		// filepath.Join with the empty string results in the same path as below.
		if component == "" {
			continue
		}
		tmp := filepath.Join(path, p)
		inf, err := os.Stat(tmp)
		if err == nil {
			if inf.IsDir() {
				tmp += string(os.PathSeparator)
			}
			jww.TRACE.Printf("findSource:  %s found", tmp)
			return filepath.ToSlash(tmp), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}
	return "", os.ErrNotExist
}

// checkPath checks to see
// buildOutPath builds the full output path of the passed path, p, and returns
// that value.  If the template is set to include the component string as the
// parent directory, it is added to the path.
func (r *rawTemplate) buildOutPath(component, p string) string {
	if r.IncludeComponentString != nil && *r.IncludeComponentString && component != "" {
		component = strings.ToLower(component)
		return path.Join(r.TemplateOutputDir, component, p)
	}
	return path.Join(r.TemplateOutputDir, p)
}

// buildTemplateResourcePath builds the path that will be added to the Packer
// template for the passed path, p, and returns that value.  If the template is
// set to include the component string as the parent directory, it is added to
// the path.
//
// All paths in the template output use '/'.
func (r *rawTemplate) buildTemplateResourcePath(component, p string, slashSuffix bool) string {
	if r.IncludeComponentString != nil && *r.IncludeComponentString && component != "" {
		component = strings.ToLower(component)
		p = path.Join(strings.ToLower(component), p)
	}
	// If this is a dir, append with a slash.
	if slashSuffix {
		p = appendSlash(p)
	}
	return p
}

// setExampleDir sets the SourceDir and TemplateOutputDir for example template
// builds. If either the Dir starts with 1 or more parent dirs, '../', they
// will elided from the Dir before prepending the SourceDir path with the
// Example directory.
//
// src = example/src
// ../src = example/src
// ../../src = example/src
// src/foo = example/src/foo
func (r *rawTemplate) setExampleDirs() {
	var i int
	var part string
	var parts []string
	if r.SourceDir == "" {
		r.SourceDir = r.ExampleDir
		goto outDir
	}
	parts = strings.Split(r.SourceDir, string(filepath.Separator))
	for i, part = range parts {
		if part != ".." {
			break
		}
	}
	if i > 0 {
		if r.ExampleDir != "" {
			r.SourceDir = filepath.Join(parts[i:]...)
		}
	}
	r.SourceDir = path.Join(r.ExampleDir, r.SourceDir)
outDir:
	if r.TemplateOutputDir == "" {
		r.TemplateOutputDir = r.ExampleDir
		return
	}
	parts = strings.Split(r.TemplateOutputDir, string(filepath.Separator))
	for i, part = range parts {
		if part != ".." {
			break
		}
	}
	if i > 0 {
		if r.ExampleDir != "" {
			r.TemplateOutputDir = path.Join(parts[i:]...)
		}
	}
	r.TemplateOutputDir = path.Join(r.ExampleDir, r.TemplateOutputDir)
}
