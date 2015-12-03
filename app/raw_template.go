package app

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mohae/contour"
	jww "github.com/spf13/jwalterweatherman"
)

// rawTemplate holds all the information for a Rancher template. This is used
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
	return &rawTemplate{date: splitDate[0], delim: contour.GetString(ParamDelimStart), files: make(map[string]string), dirs: make(map[string]string)}
}

// copy makes a copy of the template and returns the new copy.
func (r *rawTemplate) copy() *rawTemplate {
	Copy := newRawTemplate()
	Copy.PackerInf = r.PackerInf
	Copy.IODirInf = r.IODirInf
	Copy.BuildInf = r.BuildInf
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
	// Return the generated Packer Template
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
// existing settings, which are set to rancher's defaults, to create the
// default template.
func (r *rawTemplate) setDefaults(d *distro) error {
	// merges Settings between an old and new template.
	// Note: Arch, Image, and Release are not updated here as how these fields
	// are updated depends on whether this is a build from a distribution's
	// default template or from a defined build template.
	r.IODirInf.update(d.IODirInf)
	r.PackerInf.update(d.PackerInf)
	r.BuildInf.update(d.BuildInf)
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
func (r *rawTemplate) updateBuildSettings(bld *rawTemplate) {
	r.IODirInf.update(bld.IODirInf)
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
	// For post_processor_types and provisioner_types the following logic is used
	// if nil don't do anything (this means prior settings are used, e.g. default)
	// if len == 0 unset. A len of 0 means that the build template purposely unsets
	//   any build
	// if len > 0 replace the existing types with the builder's.

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
}

// mergeVariables goes through the template variables and finalizes the values
// of any :vars found within the strings.
//
// Supported:
//  distro                   the name of the distro
//  release                  the release version being used
//  arch                     the target architecture for the build
//  image                    the image used, e.g. server
//  date                     the current datetime, time.Now()
//  build_name               the name of the build template
//  output_dir                  the directory to write the build output to
//  source_dir                  the directory of any source files used in the build*
//
// Note: source_dir must be set. Rancher searches for referenced files and uses
// source_dir/distro as the last search directory. This directory is also used as
// the base directory for any specified src directories.
//
// TODO should there be a flag to not prefix src paths with source_dir to allow for
// specification of files that are not in src? If the flag is set to not prepend
// source_dir, source_dir could still be used by adding it to the specific variable.
func (r *rawTemplate) mergeVariables() {
	// Get the delim and set the replacement map, resolve name information
	r.setBaseVarVals()
	// get final value for name first
	r.Name = r.replaceVariables(r.Name)
	r.varVals[r.delim+"name"] = r.Name
	// then merge the sourc and out dirs and set them
	r.mergeSourceDir()
	r.mergeOutDir()
	r.varVals[r.delim+"output_dir"] = r.OutputDir
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

// mergeVariable does a variable replacement on the passed string and returns
// the finalized value. If the passed string is empty, the default value, d, is
// returned
func (r *rawTemplate) mergeString(s, d string) string {
	if s == "" {
		return d
	}
	return strings.TrimSuffix(r.replaceVariables(s), "/")
}

// mergeSourceDir sets whether or not a custom source directory was used, does any
// necessary variable replacement, and normalizes the string to not end in /
func (r *rawTemplate) mergeSourceDir() {
	// variable replacement is only necessary if the SourceDir has the variable delims
	if !strings.Contains(r.SourceDir, r.delim) {
		// normalize to no ending /
		r.SourceDir = strings.TrimSuffix(r.replaceVariables(r.SourceDir), "/")
		return
	}
	// normalize to no ending /
	r.SourceDir = strings.TrimSuffix(r.replaceVariables(r.SourceDir), "/")
}

// mergeOutDir resolves the output_dir for this template.
func (r *rawTemplate) mergeOutDir() {
	// variable replacement is only necessary if the SourceDir has the variable delims
	if !strings.Contains(r.OutputDir, r.delim) {
		// normalize to no ending /
		r.OutputDir = strings.TrimSuffix(r.replaceVariables(r.OutputDir), "/")
		return
	}
	// normalize to no ending /
	r.OutputDir = strings.TrimSuffix(r.replaceVariables(r.OutputDir), "/")
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
		r.osType, err = r.releaseISO.(*centos).getOSType(builderType.String())
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

// commandsFromFile returns the commands within the requested file, if it can
// be found. No validation of the contents is done.
func (r *rawTemplate) commandsFromFile(component, name string) (commands []string, err error) {
	// find the file
	src, err := r.findCommandFile(component, name)
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
// findComponentSource is called to handle the actual location of the file. If
// no match is found an os.ErrNotExist will be returned.
func (r *rawTemplate) findCommandFile(component, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("the passed command filename was empty")
	}
	findPath := path.Join("commands", name)
	src, err := r.findComponentSource(component, findPath, false)
	// return the error for any error other than ErrNotExist
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	// if err is nil, the source was found
	if err == nil {
		return src, nil
	}
	return r.findComponentSource(component, name, false)
}

// findComponentSource attempts to locate the source file or directory referred
// to in p for the requested component and return it's actual location within
// the source_dir.  If the component is not empty, it is added to the path to see
// if there are any component specific files that match.  If none are found,
// just the path is used.  Any match is returned, otherwise an os.ErrNotFound
// error is returned.  Any other error encountered will also be returned.
//
// The search path is built, in order of precedence:
//    component/path
//    component-base/path
//    path
//
// Component is the name of the packer component that this path belongs to,
// e.g. vagrant, chef-client, shell, etc.  The component-base is the base name
// of the packer component that this path belongs to, if applicable, e.g.
// chef-client's base would be chef as would chef-solo's.
func (r *rawTemplate) findComponentSource(component, p string, isDir bool) (string, error) {
	if p == "" {
		return "", fmt.Errorf("cannot find source, no path received")
	}
	var tmpPath string
	var err error
	// if len(cParts) > 1, there was a - and component-base processing should be done
	if component != "" {
		component = strings.ToLower(component)
		tmpPath, err = r.findSource(path.Join(component, p), isDir)
		if err != nil && err != os.ErrNotExist {
			return "", fmt.Errorf("%s: %s", p, err)
		}
		if err == nil {
			return tmpPath, nil
		}
		cParts := strings.Split(component, "-")
		if len(cParts) > 1 {
			// first element is the base
			tmpPath, err = r.findSource(path.Join(cParts[0], p), isDir)
			if err != nil && err != os.ErrNotExist {
				return "", fmt.Errorf("%s: %s", p, err)
			}
			if err == nil {
				return tmpPath, nil
			}
		}
	}
	// look for the source as using just the passed path
	tmpPath, err = r.findSource(p, isDir)
	if err != nil {
		// If the file didn't exist and this is an example, it's not an error
		if err == os.ErrNotExist && r.IsExample {
			return "", nil
		}
		// Otherwise return the error
		return "", fmt.Errorf("%s file %q: %s", component, p, err)
	}
	return tmpPath, nil

}

// findSource searches for the specified sub-path using Rancher's algorithm for
// finding the correct location.  Passed names may include relative path
// information and may be either a filename or a directory.  Releases may have
// "."'s in them.  In addition to searching for the requested source within the
// point release, the "." are stripped out and the resulting value is searched:
// e.g. 14.04 becomes 1404,or numericRelease.  The base release number is also
// checked: e.g. 14, the releaseBase, is searched for 14.04.
// Search order:
//   source_dir/build_name/
//   source_dir/distro/build_name/
//   source_dir/distro/release/build_name/
//   source_dir/distro/numericRelease/build_name/
//   source_dir/distro/releaseBase/build_name/
//   source_dir/distro/release/arch/
//   source_dir/distro/releaseBase/arch/
//   source_dir/distro/release/
//   source_dir/distro/releaseBase/
//   source_dir/distro/arch
//   source_dir/distro/
//   source_dir/
//
// If the passed path is not found, an os.ErrNotExist will be returned
func (r *rawTemplate) findSource(p string, isDir bool) (string, error) {
	if p == "" {
		return "", fmt.Errorf("cannot find source, no path received")
	}
	releaseParts := strings.Split(r.Release, ".")
	var numericRelease string
	if len(releaseParts) > 1 {
		for _, v := range releaseParts {
			numericRelease += v
		}
	}
	// source_dir/:build_name/p
	tmpPath := r.getSourcePath(path.Join(r.BuildName, p), isDir)
	_, err := os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/:build_name/p
	tmpPath = r.getSourcePath(path.Join(r.Distro, r.BuildName, p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/:release/:build_name/p
	tmpPath = r.getSourcePath(path.Join(r.Distro, r.Release, r.BuildName, p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/numericRelease/:build_name/p
	// only if the numericRelease is different than the release
	if numericRelease != r.Release {
		tmpPath = r.getSourcePath(path.Join(r.Distro, numericRelease, r.BuildName, p), isDir)
		_, err = os.Stat(tmpPath)
		if err == nil {
			jww.TRACE.Printf("findSource:  %s found", tmpPath)
			return filepath.ToSlash(tmpPath), nil
		}
		jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	}
	// source_dir/:distro/releaseBase/:build_name/p
	// only if releaseBase is different than the release
	if releaseParts[0] != r.Release {
		tmpPath = r.getSourcePath(path.Join(r.Distro, releaseParts[0], r.BuildName, p), isDir)
		_, err = os.Stat(tmpPath)
		if err == nil {
			jww.TRACE.Printf("findSource:  %s found", tmpPath)
			return filepath.ToSlash(tmpPath), nil
		}
		jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	}
	// source_dir/:distro/:release/:arch/p
	tmpPath = r.getSourcePath(path.Join(r.Distro, r.Release, r.Arch, p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/release/:arch/p
	// only if the numericRelease is different than the release
	if numericRelease != r.Release {
		tmpPath = r.getSourcePath(path.Join(r.Distro, numericRelease, r.Arch, p), isDir)
		_, err = os.Stat(tmpPath)
		if err == nil {
			jww.TRACE.Printf("findSource:  %s found", tmpPath)
			return filepath.ToSlash(tmpPath), nil
		}
		jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	}
	// source_dir/:distro/releaseBase/:arch/p
	// only if releaseBase is different than the release
	if releaseParts[0] != r.Release {
		tmpPath = r.getSourcePath(path.Join(r.Distro, releaseParts[0], r.Arch, p), isDir)
		_, err = os.Stat(tmpPath)
		if err == nil {
			jww.TRACE.Printf("findSource:  %s found", tmpPath)
			return filepath.ToSlash(tmpPath), nil
		}
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/:release/p
	tmpPath = r.getSourcePath(path.Join(r.Distro, r.Release, p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/release/p
	// only if the numericRelease is different than the release
	if numericRelease != r.Release {
		tmpPath = r.getSourcePath(path.Join(r.Distro, numericRelease, p), isDir)
		_, err = os.Stat(tmpPath)
		if err == nil {
			jww.TRACE.Printf("findSource:  %s found", tmpPath)
			return filepath.ToSlash(tmpPath), nil
		}
		jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	}
	// source_dir/:distro/releaseBase/p
	// only if releaseBase is different than the release
	if releaseParts[0] != r.Release {
		tmpPath = r.getSourcePath(path.Join(r.Distro, releaseParts[0], p), isDir)
		_, err = os.Stat(tmpPath)
		if err == nil {
			jww.TRACE.Printf("findSource:  %s found", tmpPath)
			return filepath.ToSlash(tmpPath), nil
		}
		jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	}
	// source_dir/:distro/:arch/p
	tmpPath = r.getSourcePath(path.Join(r.Distro, r.Arch, p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/:distro/p
	tmpPath = r.getSourcePath(path.Join(r.Distro, p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// source_dir/p
	tmpPath = r.getSourcePath(path.Join(p), isDir)
	_, err = os.Stat(tmpPath)
	if err == nil {
		jww.TRACE.Printf("findSource:  %s found", tmpPath)
		return filepath.ToSlash(tmpPath), nil
	}
	jww.TRACE.Printf("findSource:  %s not found", tmpPath)
	// not found, return an error
	return "", os.ErrNotExist
}

// buildOutPath builds the full output path of the passed path, p, and returns
// that value.  If the template is set to include the component string as the
// parent directory, it is added to the path.
func (r *rawTemplate) buildOutPath(component, p string) string {
	if r.includeComponentString() && component != "" {
		component = strings.ToLower(component)
		return path.Join(r.OutputDir, component, p)
	}
	return path.Join(r.OutputDir, p)
}

// buildTemplateResourcePath builds the path that will be added to the Packer
// template for the passed path, p, and returns that value.  If the template is
// set to include the component string as the parent directory, it is added to
// the path.
func (r *rawTemplate) buildTemplateResourcePath(component, p string) string {
	if r.includeComponentString() && component != "" {
		component = strings.ToLower(component)
		return path.Join(strings.ToLower(component), p)
	}
	return p
}

// getSourcePath returns the requested path as a child of the SourceDir.
func (r *rawTemplate) getSourcePath(p string, isDir bool) string {
	if p == "" {
		return ""
	}
	return path.Join(r.SourceDir, p)
}

// setExampleDisr sets the SourceDir and OutDir for example template builds. If
// either the Dir starts with 1 or more parent dirs, '../', they will elided
// from the Dir before prepending the SourceDir path with the Example directory.n
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
	if r.OutputDir == "" {
		r.OutputDir = r.ExampleDir
		return
	}
	parts = strings.Split(r.OutputDir, string(filepath.Separator))
	for i, part = range parts {
		if part != ".." {
			break
		}
	}
	if i > 0 {
		if r.ExampleDir != "" {
			r.OutputDir = path.Join(parts[i:]...)
		}
	}
	r.OutputDir = path.Join(r.ExampleDir, r.OutputDir)
}
