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
	"github.com/mohae/feedlot/conf"
	"github.com/mohae/feedlot/log"
)

type InvalidComponentErr struct {
	id   string // component id
	cTyp string // component type
	s    string
}

func (e InvalidComponentErr) Error() string {
	if e.id == "" {
		return fmt.Sprintf("%q: invalid %s", e.s, e.cTyp)
	}
	return fmt.Sprintf("id: %s: %q: invalid %s", e.s, e.id, e.cTyp)
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

type EmptyPathErr struct {
	s string
}

func (e EmptyPathErr) Error() string {
	return e.s + ": empty path"
}

// ErrNoCommands occurs when a referenced command file doesn't have any
// contents.
var ErrNoCommands = errors.New("no commands found")

// RawTemplate holds all the information for a Feedlot template. This is used
// to generate the Packer Build.
type RawTemplate struct {
	PackerInf
	IODirInf
	BuildInf
	// Example settings
	IsExample  bool
	ExampleDir string
	// holds release information
	ReleaseISO Releaser
	// the builder specific string for the template's OS and Arch
	OSType string
	// Current date in ISO 8601
	Date string
	// The character(s) used to identify variables for Feedlot. By default
	// this is a colon, :. Currently only a starting delimeter is supported.
	Delim string
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
	// VarVals is a variable replacement map used in finalizing the value of strings for
	// which variable replacement is supported.
	VarVals map[string]string
	// Contains all the build information needed to create the target Packer
	// template and its associated artifacts.
	Build
	// Files maps destination files to their sources. These are the actual file locations
	// after they have been resolved. The destination file is the key, the source file
	// is the value
	Files map[string]string
	// Dirs maps destination directories to their source directories. Everything within
	// the directory will be copied. The same resolution rules apply for dirs as for
	// files. The destination directory is the key, the source directory is the value
	Dirs map[string]string
}

// mewRawTemplate returns a rawTemplate with current date in ISO 8601 format.
// This should be called when a rawTemplate with the current date is desired.
func newRawTemplate() *RawTemplate {
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	return &RawTemplate{Date: splitDate[0], Delim: contour.GetString(conf.ParamDelimStart), Files: make(map[string]string), Dirs: make(map[string]string)}
}

// Copy makes a deep copy of the template and returns the new copy.
func (r *RawTemplate) Copy() *RawTemplate {
	return deepcopy.Copy(r).(*RawTemplate)
}

// r.createPackerTemplate creates a Packer template from the rawTemplate.
// TODO:
//		Write to output
//		Copy resources to output
func (r *RawTemplate) createPackerTemplate() (PackerTemplate, error) {
	log.Infof("%s: create packer template", r.Name)
	var err error
	// Resolve the Feedlot variables to their final values.
	r.mergeVariables()
	// General Packer Stuff
	p := PackerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description
	// Builders
	p.Builders, err = r.createBuilders()
	if err != nil {
		err = Error{slug: r.BuildInf.Name, err: err}
		log.Error(err)
		return p, err
	}
	// Post-Processors
	p.PostProcessors, err = r.createPostProcessors()
	if err != nil {
		err = Error{slug: r.BuildInf.Name, err: err}
		log.Error(err)
		return p, err
	}
	// Provisioners
	p.Provisioners, err = r.createProvisioners()
	if err != nil {
		err = Error{slug: r.BuildInf.Name, err: err}
		log.Error(err)
		return p, err
	}
	// Return the generated Packer Template.
	log.Infof("%s: packer template created", r.Name)
	return p, nil
}

// replaceVariables checks incoming string for variables and replaces them with
// their values.
func (r *RawTemplate) replaceVariables(s string) string {
	log.Debugf("replace vars: %s", s)
	//see if the delim is in the string, if not, nothing to replace
	if strings.Index(s, r.Delim) < 0 {
		return s
	}
	// Go through each variable and replace as applicable.
	for vName, vVal := range r.VarVals {
		s = strings.Replace(s, vName, vVal, -1)
	}
	log.Debugf("vars replaced: %s", s)
	return s
}

// r.setDefaults takes the incoming distro settings and merges them with its
// existing settings, which are set to feedlot's defaults, to create the
// default template.
func (r *RawTemplate) setDefaults(d *SupportedDistro) error {
	log.Infof("%s: set defaults from distro settings", r.Name)
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
		log.Debugf("%s: override builder IDs with %v", r.Name, d.BuilderIDs)
		r.BuilderIDs = d.BuilderIDs
	}
	// If defined, PostProcessorTypes override any prior PostProcessorTypes Settings
	if d.PostProcessorIDs != nil {
		log.Debugf("%s: override post-processor IDs with %v", r.Name, d.PostProcessorIDs)
		r.PostProcessorIDs = d.PostProcessorIDs
	}
	// If defined, ProvisionerTypes override any prior ProvisionerTypes Settings
	if d.ProvisionerIDs != nil {
		log.Debugf("%s: override provisioner IDs with %v", r.Name, d.ProvisionerIDs)
		r.ProvisionerIDs = d.ProvisionerIDs
	}
	// merge the build portions.
	err := r.updateBuilders(d.Builders)
	if err != nil {
		return Error{slug: "set builder defaults", err: err}
	}
	err = r.updatePostProcessors(d.PostProcessors)
	if err != nil {
		return Error{slug: "set builder defaults", err: err}
	}
	err = r.updateProvisioners(d.Provisioners)
	if err != nil {
		return Error{slug: "set builder defaults", err: err}
	}
	log.Infof("%s: defaults set from distro settings", r.Name)
	return nil
}

// r.updateBuildSettings merges Settings between an old and new template.
// Note:  Arch, Image, and Release are not updated here as how these fields are
// updated depends on whether this is a build from a distribution's default
// template or from a defined build template.
func (r *RawTemplate) updateBuildSettings(bld *RawTemplate) error {
	r.IODirInf.update(bld.IODirInf)
	r.updateSourceDirSetting()
	err := r.updateTemplateOutputDirSetting()
	if err != nil {
		return err
	}
	r.PackerInf.update(bld.PackerInf)
	r.BuildInf.update(bld.BuildInf)
	if bld.Arch != "" {
		log.Debugf("%s: set arch from %s: %s", r.Name, bld.Name, bld.Arch)
		r.Arch = bld.Arch
	}
	if bld.Image != "" {
		log.Debugf("%s: set image from %s: %s", r.Name, bld.Name, bld.Image)
		r.Image = bld.Image
	}
	if bld.Release != "" {
		log.Debugf("%s: set release from %s: %s", r.Name, bld.Name, bld.Release)
		r.Release = bld.Release
	}
	// If defined, Builders override any prior builder Settings.
	if len(bld.BuilderIDs) > 0 {
		log.Debugf("%s: set builder ids from %s: %s", r.Name, bld.Name, bld.BuilderIDs)
		r.BuilderIDs = bld.BuilderIDs
	}
	//   if nil don't do anything (this means prior settings are used, e.g. default)
	// For post_processor_ids and provisioner_ids, the following logic is used:
	//   if len == 0 unset. A len of 0 means that the build template purposely unsets
	//     any build
	//   if len > 0 replace the existing types with the builder's.
	if bld.PostProcessorIDs != nil {
		log.Debugf("%s: set post-processor ids from %s: %s", r.Name, bld.Name, bld.PostProcessorIDs)
		r.PostProcessorIDs = bld.PostProcessorIDs
	}
	if bld.ProvisionerIDs != nil {
		log.Debugf("%s: set provisioner ids from %s: %s", r.Name, bld.Name, bld.ProvisionerIDs)
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
func (r *RawTemplate) updateTemplateOutputDirSetting() error {
	if *r.IODirInf.TemplateOutputDirIsRelative {
		log.Debugf("%s: template output dir is relative", r.Name)
		dir, err := os.Getwd()
		if err != nil {
			return Error{"template output dir: get working directory", err}
		}
		r.IODirInf.TemplateOutputDir = filepath.Join(dir, r.IODirInf.TemplateOutputDir)
		log.Debugf("%s: template output dir is now %s", r.Name, r.IODirInf.TemplateOutputDir)
	}
	return nil
}

// updateSourceDirSetting updates the source_dir if the source_dir_is_relative
// flag is true.  Any Feedlot variables in the source_dir setting are not
// resolved.
func (r *RawTemplate) updateSourceDirSetting() {
	if *r.IODirInf.SourceDirIsRelative {
		log.Debugf("%s: template source dir is relative", r.Name)
		r.IODirInf.SourceDir = filepath.Join(contour.GetString(conf.Dir), r.IODirInf.SourceDir)
		log.Debugf("%s: template source dir is now %s", r.Name, r.IODirInf.TemplateOutputDir)
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
func (r *RawTemplate) mergeVariables() {
	// Get the delim and set the replacement map, resolve name information
	r.setBaseVarVals()
	// get final value for name first
	r.Name = r.replaceVariables(r.Name)
	r.VarVals[r.Delim+"name"] = r.Name
	// then merge the sourc and out dirs and set them
	r.SourceDir = r.replaceVariables(r.SourceDir)
	r.TemplateOutputDir = r.replaceVariables(r.TemplateOutputDir)
	r.PackerOutputDir = r.replaceVariables(r.PackerOutputDir)
	r.VarVals[r.Delim+"template_output_dir"] = r.TemplateOutputDir
	r.VarVals[r.Delim+"packer_output_dir"] = r.PackerOutputDir
	r.VarVals[r.Delim+"source_dir"] = r.SourceDir
	log.Debugf("%s: merged SourceDir: %s", r.Name, r.SourceDir)
	log.Debugf("%s: merged TemplateOutputDir: %s", r.Name, r.TemplateOutputDir)
	log.Debugf("%s: merged PackerOutputDir: %s", r.Name, r.PackerOutputDir)
}

// setBaseVarVals sets the varVals for the base variables
func (r *RawTemplate) setBaseVarVals() {
	r.VarVals = map[string]string{
		r.Delim + "distro":     r.Distro,
		r.Delim + "release":    r.Release,
		r.Delim + "arch":       r.Arch,
		r.Delim + "image":      r.Image,
		r.Delim + "date":       r.Date,
		r.Delim + "build_name": r.BuildName,
	}
}

// mergeString does a variable replacement on the passed string and returns
// the finalized value. If the passed string is empty, the default value, d, is
// returned
func (r *RawTemplate) mergeString(s, d string) string {
	if s == "" {
		return d
	}
	return r.replaceVariables(s)
}

// ISOInfo sets the ISO info for the template's supported distro type. This
// also sets the builder specific string, when applicable.
// TODO: these should use new functions in release.go. instead of creating the
// structs here
func (r *RawTemplate) ISOInfo(builderType Builder, settings []string) error {
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
		r.ReleaseISO = &centos{
			release: release{
				ISO: ISO{
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
		err = r.ReleaseISO.setVersionInfo()
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
		err = r.ReleaseISO.SetISOInfo()
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
		r.OSType, err = r.ReleaseISO.(*centos).getOSType(builderType)
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
	case Debian.String():
		r.ReleaseISO = &debian{
			release: release{
				ISO: ISO{
					BaseURL:      r.BaseURL,
					ChecksumType: checksumType,
				},
				Arch:    r.Arch,
				Distro:  r.Distro,
				Image:   r.Image,
				Release: r.Release,
			},
		}
		err = r.ReleaseISO.setVersionInfo()
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
		err = r.ReleaseISO.SetISOInfo()
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
		r.OSType, err = r.ReleaseISO.(*debian).getOSType(builderType)
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
	case Ubuntu.String():
		r.ReleaseISO = &ubuntu{
			release: release{
				ISO: ISO{
					BaseURL:      r.BaseURL,
					ChecksumType: checksumType,
				},
				Arch:    r.Arch,
				Distro:  r.Distro,
				Image:   r.Image,
				Release: r.Release,
			},
		}
		err = r.ReleaseISO.setVersionInfo()
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
		err = r.ReleaseISO.SetISOInfo()
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
		r.OSType, err = r.ReleaseISO.(*ubuntu).getOSType(builderType)
		if err != nil {
			err = Error{slug: "iso info", err: err}
			log.Error(err)
			return err
		}
	default:
		err := fmt.Errorf("iso info: %s: unsupported distro", r.Distro)
		log.Error(err)
		return err
	}
	return nil
}

// commandsFromFile returns the commands within the requested file, if it can
// be found. No validation of the contents is done.
func (r *RawTemplate) commandsFromFile(name, component string) (commands []string, err error) {
	// find the file
	src, err := r.findCommandFile(name, component)
	if err != nil {
		return nil, err
	}
	log.Debugf("%s: command file: %s", r.Name, src)
	f, err := os.Open(src)
	if err != nil {
		return nil, Error{slug: "get commands", err: err}
	}
	// always close what's been opened and check returned error
	defer func() {
		cerr := f.Close()
		if cerr != nil && err == nil {
			err = Error{slug: "get commands", err: cerr}
		}
	}()
	//New Reader for the string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		commands = append(commands, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		err = Error{slug: "get commands", err: err}
		return nil, err
	}
	log.Debugf("%s: %s: %d commands found", r.Name, src, len(commands))
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
func (r *RawTemplate) findCommandFile(name, component string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("find command file: empty filename")
	}
	findPath := filepath.Join("commands", name)
	src, err := r.findSource(findPath, component, false)
	// return the error for any error other than ErrNotExist
	if err != nil && err != os.ErrNotExist {
		return "", Error{slug: "find command file", err: err}
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
func (r *RawTemplate) findSource(p, component string, isDir bool) (src string, err error) {
	if p == "" {
		return "", EmptyPathErr{"find source"}
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

	log.Debugf("findSource: %s not found", p)
	// not found, return an error
	return "", Error{slug: filepath.ToSlash(p), err: os.ErrNotExist}
}

// buildSearchPaths builds a slice of paths to search based on what it
// receives.
// for each release element:  path = source_dir + root + release + base
func (r *RawTemplate) buildSearchPaths(root, base string, release []string) []string {
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
func (r *RawTemplate) checkSourcePaths(p, component string, paths []string) (string, error) {
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
			log.Debugf("%s: check for source at %s", r.Name, tmp)
			inf, err := os.Stat(tmp)
			if err == nil {
				// if it's a dir, append the slash
				if inf.IsDir() {
					tmp += string(os.PathSeparator)
				}
				log.Debugf("%s: source found at %s", r.Name, tmp)
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
		log.Debugf("%s: check for source at %s", r.Name, tmp)
		inf, err := os.Stat(tmp)
		if err == nil {
			if inf.IsDir() {
				tmp += string(os.PathSeparator)
			}
			log.Debugf("%s: source found at %s", r.Name, tmp)
			return filepath.ToSlash(tmp), nil
		}
		if !os.IsNotExist(err) {
			return "", err
		}
	}
	log.Debugf("%s: source for component %s: %s not found", r.Name, component, p)
	return "", os.ErrNotExist
}

// checkPath checks to see
// buildOutPath builds the full output path of the passed path, p, and returns
// that value.  If the template is set to include the component string as the
// parent directory, it is added to the path.
func (r *RawTemplate) buildOutPath(component, p string) string {
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
func (r *RawTemplate) buildTemplateResourcePath(component, p string, slashSuffix bool) string {
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
func (r *RawTemplate) setExampleDirs() {
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
	log.Debugf("%s: example output dir: %s", r.Name, r.TemplateOutputDir)

	if r.TemplateOutputDir == "" {
		r.TemplateOutputDir = r.ExampleDir
		log.Debugf("%s: example output dir: %s", r.Name, r.TemplateOutputDir)
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
	log.Debugf("%s: example output dir: %s", r.Name, r.TemplateOutputDir)
}

// IsEmptyPathERror returns if the error is an empty path
func IsEmptyPathErr(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(EmptyPathErr)
	return ok
}
