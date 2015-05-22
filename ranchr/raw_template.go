package ranchr

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	jww "github.com/spf13/jwalterweatherman"
)

// rawTemplate holds all the information for a Rancher template. This is used
// to generate the Packer Build.
type rawTemplate struct {
	PackerInf
	IODirInf
	BuildInf
	// holds release information
	releaseISO releaser
	// the builder specific string for the template's OS and Arch
	osType string
	// Current date in ISO 8601
	date string
	// The character(s) used to identify variables for Rancher. By default
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
	// Variable name mapping...currently not supported
	vars map[string]string
	// Contains all the build information needed to create the target Packer
	// template and its associated artifacts.
	build
	// files maps destination files to their sources. These are the actual file locations
	// after they have been resolved. The destination file is the key, the source file
	// is the value
	files map[string]string
	// dirs maps destination directories to their source directories. Everything within
	// the directory will be copied. The same resolution rules apply for dirs as for
	// filies. The destination directory is the key, the source directory is the value
	dirs map[string]string
}

// mewRawTemplate returns a rawTemplate with current date in ISO 8601 format.
// This should be called when a rawTemplate with the current date is desired.
func newRawTemplate() *rawTemplate {
	// Set the date, formatted to ISO 8601
	date := time.Now()
	splitDate := strings.Split(date.String(), " ")
	return &rawTemplate{date: splitDate[0], delim: os.Getenv(EnvParamDelimStart), files: make(map[string]string), dirs: make(map[string]string)}
}

// r.createPackerTemplate creates a Packer template from the rawTemplate that
// can be marshalled to JSON.
func (r *rawTemplate) createPackerTemplate() (packerTemplate, error) {
	var err error
	// Resolve the Rancher variables to their final values.
	r.mergeVariables()
	// General Packer Stuff
	p := packerTemplate{}
	p.MinPackerVersion = r.MinPackerVersion
	p.Description = r.Description
	// Builders
	p.Builders, _, err = r.createBuilders()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Post-Processors
	p.PostProcessors, _, err = r.createPostProcessors()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Provisioners
	p.Provisioners, _, err = r.createProvisioners()
	if err != nil {
		jww.ERROR.Println(err)
		return p, err
	}
	// Now we can create the Variable Section
	// TODO: currently not implemented/supported
	// Return the generated Packer Template
	return p, nil
}

// replaceVariables checks incoming string for variables and replaces them
// with their values.
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
// existing settings, which are set to rancher's defaults, to create the
// default template.
func (r *rawTemplate) setDefaults(d *distro) {
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(d.IODirInf)
	r.PackerInf.update(d.PackerInf)
	r.BuildInf.update(d.BuildInf)
	// If defined, BuilderTypes override any prior BuilderTypes Settings
	if d.BuilderTypes != nil && len(d.BuilderTypes) > 0 {
		r.BuilderTypes = d.BuilderTypes
	}
	// If defined, PostProcessorTypes override any prior PostProcessorTypes Settings
	if d.PostProcessorTypes != nil && len(d.PostProcessorTypes) > 0 {
		r.PostProcessorTypes = d.PostProcessorTypes
	}
	// If defined, ProvisionerTypes override any prior ProvisionerTypes Settings
	if d.ProvisionerTypes != nil && len(d.ProvisionerTypes) > 0 {
		r.ProvisionerTypes = d.ProvisionerTypes
	}
	// merge the build portions.
	r.updateBuilders(d.Builders)
	r.updatePostProcessors(d.PostProcessors)
	r.updateProvisioners(d.Provisioners)
	return
}

// r.updateBuildSettings merges Settings between an old and new template. Note:
// Arch, Image, and Release are not updated here as how these fields are
// updated depends on whether this is a build from a distribution's default
// template or from a defined build template.
func (r *rawTemplate) updateBuildSettings(bld *rawTemplate) {
	r.IODirInf.update(bld.IODirInf)
	r.PackerInf.update(bld.PackerInf)
	r.BuildInf.update(bld.BuildInf)
	// If defined, Builders override any prior builder Settings.
	if bld.BuilderTypes != nil && len(bld.BuilderTypes) > 0 {
		r.BuilderTypes = bld.BuilderTypes
	}
	// If defined, PostProcessorTypes override any prior PostProcessorTypes Settings
	if bld.PostProcessorTypes != nil && len(bld.PostProcessorTypes) > 0 {
		r.PostProcessorTypes = bld.PostProcessorTypes
	}
	// If defined, ProvisionerTypes override any prior ProvisionerTypes Settings
	if bld.ProvisionerTypes != nil && len(bld.ProvisionerTypes) > 0 {
		r.ProvisionerTypes = bld.ProvisionerTypes
	}
	// merge the build portions.
	r.updateBuilders(bld.Builders)
	r.updatePostProcessors(bld.PostProcessors)
	r.updateProvisioners(bld.Provisioners)
}

// mergeVariables goes through the template variables and finalizes the values of any
// :vars found within the strings.
//
// Supported:
//  distro                   the name of the distro
//  release                  the release version being used
//  arch                     the target architecture for the build
//  image                    the image used, e.g. server
//  date                     the current datetime, time.Now()
//  build_name               the name of the build template
//  out_dir                  the directory to write the build output to
//  src_dir                  the directory of any source files used in the build*
//  commands_dir             ???
//  commands_src_dir         the directory of any command files that the build template
//                           uses**
//
//  * src_dir must be set. Rancher searches for referenced files and uses src_dir/distro
//    as the last search directory. This directory is also used as the base directory
//    for any specified src directories.
// TODO should there be a flag to not prefix src paths with src_dir to allow for
// specification of files that are not in src? If the flag is set to not prepend
// src_dir, src_dir could still be used by adding it to the specific variable.
//
//  ** commands_src_dir: if a value is not specified, Rancher will use "commands" as
//  the commands_src_dir, which is expected to be a directory within src_dir/distro/
//  or one of the subdirectories within that path that is part of rancher's search
//  path.
func (r *rawTemplate) mergeVariables() {
	// Get the delim and set the replacement map, resolve name information
	r.setBaseVarVals()
	// get final value for name first
	r.Name = r.replaceVariables(r.Name)
	r.varVals[r.delim+"name"] = r.Name

	// then merge the sourc and out dirs and set them
	r.mergeSrcDir()
	r.mergeOutDir()
	r.varVals[r.delim+"out_dir"] = r.OutDir
	r.varVals[r.delim+"src_dir"] = r.SrcDir

	// set with default, if empty. The default must not have a trailing /
	r.CommandsSrcDir = r.mergeString(r.CommandsSrcDir, "commands")

	// Create a full variable replacement map, know that the SrcDir and OutDir stuff are resolved.
	// Rest of the replacements are done by the packerers.
	r.varVals[r.delim+"commands_src_dir"] = r.CommandsSrcDir
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

// mergeVariable does a variable replacement on the passed string and returns the
// finalized value. If the passed string is empty, the default value, d, is returned
func (r *rawTemplate) mergeString(s, d string) string {
	if s == "" {
		return d
	}
	return strings.TrimSuffix(r.replaceVariables(s), "/")
}

// mergeSrcDir sets whether or not a custom source directory was used, does any
// necessary variable replacement, and normalizes the string to not end in /
func (r *rawTemplate) mergeSrcDir() {
	// variable replacement is only necessary if the SrcDir has the variable delims
	if !strings.Contains(r.SrcDir, r.delim) {
		// normalize to no ending /
		r.SrcDir = strings.TrimSuffix(r.replaceVariables(r.SrcDir), "/")
		return
	}
	// this means that this is a custom src dir. It may also be set to true in the
	// build template w or w/o variables
	r.CustomSrcDir = true
	// normalize to no ending /
	r.SrcDir = strings.TrimSuffix(r.replaceVariables(r.SrcDir), "/")
}

// mergeOutDir resolves the out_dir for this template.  If the build's custom_out_dir
// == true or there are variables are specified in the out_dir, the resolved name is
// used, otherwise the default of out_dir/:distro/:release/:build_name is used as the
// output directory. If the custom_out_dir is false, but variables were specified in
// the out_dir the custom_out_dir flag is set to true.
func (r *rawTemplate) mergeOutDir() {
	// variable replacement is only necessary if the SrcDir has the variable delims
	if !strings.Contains(r.OutDir, r.delim) {
		// normalize to no ending /
		r.OutDir = strings.TrimSuffix(r.replaceVariables(r.OutDir), "/")
		return
	}
	// this means that this is a custom out dir. It may also be set to true in the
	// build template w or w/o variables
	r.CustomOutDir = true
	// normalize to no ending /
	r.OutDir = strings.TrimSuffix(r.replaceVariables(r.OutDir), "/")
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
		r.releaseISO = &centOS{
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
		r.releaseISO.SetISOInfo()
		r.osType, err = r.releaseISO.(*centOS).getOSType(builderType.String())
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
		r.releaseISO.SetISOInfo()
		r.osType, err = r.releaseISO.(*debian).getOSType(builderType.String())
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
		r.releaseISO.SetISOInfo()
		r.osType, err = r.releaseISO.(*ubuntu).getOSType(builderType.String())
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

// Takes the name of the command file, including path relative to the src_dir, and
// returns a slice of shell commands. Each command within the file is separated by a
// newline. Returns error if an error occurs with the file.
//
// Multiple passes may be made in an attempt to find the requested command file. If a
// matching file is found, the contents are read and parsed as the command(s) to be
// returned. At minimum, a command file should have 1 line, though it may have more.
// Each line in the file is an element in the returned slice. The caller is expected
// to properly handle the returned results.
//
// Because the passed name may or may not be the actual location of the requested
// file, this method may check multiple sub-paths of the src_dir in an attempt to find
// the correct one. The first encountered match is used. The source file search
// algorithm is applied.
//
// This method will use the following as the path to search:
//   If the name contains at least 1 directory in the path, that is searched first.
//   commands/{typ}/{name}
//   commands/{typ-base}/{name} <- if applicable
//   commands/{name}
//
// If the typ includes a "-", the typ-base is the portion prior to the first -, e.g.
//   the typ-base of chef-client is chef.
//
// If no match is found, an error is returned.
func (r *rawTemplate) commandsFromFile(component, name string) (commands []string, err error) {
	if name == "" {
		err = fmt.Errorf("the passed Command filename was empty")
		jww.ERROR.Println(err)
		return nil, err
	}
	// findPath is what is actually being looked for. If the name doesn't include
	// directory information, multiple paths patterns may be used, which is what
	// findPath is used for. Initially set to name
	findPath := name
	// see if the passed name includes a directory, if not, prepend with command
	dir := path.Dir(name)
	if dir == "." {
		// name did not include directory, prepend "commands"
		findPath = filepath.Join("commands", name)

		// also prepend the type, if it's not empty
		if component != "" {
			findPath = filepath.Join(component, findPath)
		}
	}

	// find the correct file location
	// TODO finish rewrite
	path, err := r.findSource(findPath)
	// if not found and component information was passed, look for it without the component
	if err != nil && component != "" {
		// see if typ can be reduced to its base
		componentParts := strings.Split(component, "-")
		if len(componentParts) > 1 {
			// the first element is the base, e.g. chef is the base of chef-client
			findPath = filepath.Join(componentParts[0], "commands", name)
		}
		findPath = filepath.Join("commands", name)
		path, err = r.findSource(findPath)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		err = fmt.Errorf("%q not found in %q or any of its searched subdirectories", findPath, r.SrcDir)
		jww.ERROR.Println(err)
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		jww.ERROR.Println(err)
		return nil, err
	}
	// always close what's been opened and check returned error
	defer func() {
		cerr := f.Close()
		if cerr != nil && err == nil {
			jww.WARN.Println(cerr)
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
		jww.WARN.Println(err)
		return nil, err
	}
	return commands, nil
}

// findComponentSource attempts to locate the source file or directory referred to in
// p for the requested component and return it's actual location within the src_dir.
// If the component is not empty, it is added to the path to see if there are any
// component specific files that match. If none are found, just the path is used.
// Any match is returned, otherwise an os.ErrNotFound error is returned. An other
// error encountered will also be returned.
//
// The search path is built, in order of precedence:
//    component/path
//    component-base/path
//    path
//
//  component is the name of the packer component that this path belongs to, e.g
//    vagrant, chef-client, shell, etc.
//  component-base is the base name of the packer component that this path belongs
//    to, if applicable, e.g. chef-client's base would be chef as would chef-solo's.
func (r *rawTemplate) findComponentSource(component, p string) (string, error) {
	var tmpPath string
	var err error
	// if len(cParts) > 1, there was a - and component-base processing should be done
	if component != "" {
		tmpPath = filepath.Join(component, p)
		tmpPath, err = r.findSource(tmpPath)
		if err != nil && err != os.ErrNotExist {
			return "", err
		}
		if err == nil {
			return tmpPath, nil
		}
		cParts := strings.Split(component, "-")
		if len(cParts) > 1 {
			// first element is the base
			tmpPath = filepath.Join(cParts[0], p)
			tmpPath, err = r.findSource(tmpPath)
			if err != nil && err != os.ErrNotExist {
				return "", err
			}
			if err == nil {
				return tmpPath, nil
			}
		}
	}
	// look for the source as using just the passed path
	tmpPath, err = r.findSource(p)
	if err == nil {
		return tmpPath, nil
	}
	return "", err
}

// findSource searches for the specified sub-path using Rancher's algorithm for finding
// the correct location. Passed names may include relative path information and may be
// either a filename or a directory. Releases may have "."'s in them. In addition to
// searching for the requested source within the point release, the "." are stripped
// out and the resulting value is searched; e.g. 14.04 becomes 1404. The base release
// number is also checked; e.g. 14 is searched for 14.04.
// Search order:
//   src_dir/distro/release/build_name/
//   src_dir/distro/releaseBase/build_name/
//   src_dir/distro/build_name/
//   src_dir/build_name/
//   src_dir/distro/release/arch/
//   src_dir/distro/releaseBase/arch/
//   src_dir/distro/release/
//   src_dir/distro/releaseBase/
//   src_dir/distro/arch
//   src_dir/distro/
//   src_dir/
//
// If the passed poth is not found, an os.ErrNotExist will be returned
func (r *rawTemplate) findSource(p string) (string, error) {
	if p == "" {
		return "", fmt.Errorf("cannot find source, no path received")
	}
	releaseParts := strings.Split(r.Release, ".")
	var release string
	if len(releaseParts) > 1 {
		for _, v := range releaseParts {
			release += v
		}
	}
	// src_dir/:distro/:release/:build_name/p
	tmpPath := filepath.Join(r.SrcDir, r.Distro, r.Release, r.BuildName, p)
	_, err := os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/release/:build_name/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, release, r.BuildName, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/releaseBase/:build_name/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, releaseParts[0], r.BuildName, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/:build_name/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, r.BuildName, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:build_name/p
	tmpPath = filepath.Join(r.SrcDir, r.BuildName, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/:release/:arch/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, r.Release, r.Arch, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/release/:arch/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, release, r.Arch, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/releaseBase/:arch/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, releaseParts[0], r.Arch, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/:release/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, r.Release, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/release/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, release, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/releaseBase/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, releaseParts[0], p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/:arch/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, r.Arch, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/:distro/p
	tmpPath = filepath.Join(r.SrcDir, r.Distro, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	// src_dir/p
	tmpPath = filepath.Join(r.SrcDir, p)
	_, err = os.Stat(tmpPath)
	if err == nil {
		return tmpPath, nil
	}
	return "", os.ErrNotExist
}
